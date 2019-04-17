package beater

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"regexp"

	"strings"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	containersErrorRe string = `.*choose one of: \[(.*)\] .*`
)

type PodLogs struct {
	Channels  map[string]chan bool
	Client    *kubernetes.Clientset
	Config    *rest.Config
	Ignored   string
	Namespace string

	sc     *SenderConfig
	sender *Sender
	mux    sync.Mutex
}

// Add adds control channel
func (p *PodLogs) Add(pod string, ch chan bool) {
	p.mux.Lock()
	p.Channels[pod] = ch
	p.mux.Unlock()
}

// Del deletes pod control channel
func (p *PodLogs) Del(pod string) {
	p.mux.Lock()
	new := make(map[string]chan bool)
	if p.Channels != nil {
		for k, v := range p.Channels {
			if k != pod {
				new[k] = v
			}
		}
	}
	p.Channels = new
	p.mux.Unlock()
}

func (p *PodLogs) Len() int {
	return len(p.Channels)
}

func (p *PodLogs) PodTicker() {
	ignored := ignoredPods(p.Ignored)
	ticker := time.NewTicker(time.Second * time.Duration(30))
	for c := range ticker.C {
		log.Warn("New tick in pod watcher")

		timeout := int64(10)
		pods, err := p.Client.CoreV1().Pods("difrex").List(metav1.ListOptions{TimeoutSeconds: &timeout})
		if err != nil {
			log.Error("Error on tick: ", c, " ", err.Error())
			continue
		}

		log.Info("Get pods: ", len(pods.Items))
		for _, pod := range pods.Items {
			if _, ok := p.Channels[pod.Name]; !ok && pod.Status.Phase == "Running" && !ignored.isIgnored(pod.Name) {
				ch := make(chan bool)
				p.Add(pod.Name, ch)
				go p.Run(pod.Name, ch, "")
			}
		}
	}
}

// Watch watches for k8s events
func (p *PodLogs) Watch() {

	ignored := ignoredPods(p.Ignored)
	go p.PodTicker()
	go p.sender.Ticker()

	w, err := p.Client.CoreV1().Pods(p.Namespace).Watch(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	ch := w.ResultChan()

	for {
		event := <-ch
		e := marshalEvent(event)
		name := e.Name()

		log.Warn("New event received: ", event.Type)
		switch event.Type {
		case "MODIFIED":
			ch := make(chan bool)
			if podCh, ok := p.Channels[name]; !ok && e.State() == "Running" && !ignored.isIgnored(name) {
				p.Add(name, ch)
				go p.Run(name, ch, "")
			} else if ok && e.State() != "Running" {
				p.Stop(podCh)
			}
		case "ADDED":
			ch := make(chan bool)
			if _, ok := p.Channels[name]; !ok && e.State() == "Running" && !ignored.isIgnored(name) {
				p.Add(name, ch)
				go p.Run(name, ch, "")
			}
		case "DELETED":
			if ch, ok := p.Channels[name]; ok {
				p.Del(name)
				ch <- true
			}
		}
	}
}

func (p *PodLogs) Stop(ch chan bool) {
	log.Warn("Trying stopping a watcher")
	ch <- true
}

// Run runs the logwatcher
func (p *PodLogs) Run(pod string, ch chan bool, con string) {
	log.Warn("Trying to start watcher for pod ", pod, con)
	c := http.DefaultClient

	podApi := p.Config.Host + "/api/v1/namespaces/" + p.Namespace + "/pods/" + pod
	var req *http.Request
	if con == "" {
		r, err := http.NewRequest("GET", podApi+"/log?follow=true&tailLines=10", nil)
		if err != nil {
			p.Del(pod)
			return
		}
		req = r
	} else {
		r, err := http.NewRequest(
			"GET",
			podApi+"/log?follow=true&tailLines=10&container="+con,
			nil)
		if err != nil {
			p.Del(pod)
			return
		}
		req = r
	}

	req.Header.Add("Authorization", "Bearer "+p.Config.BearerToken)

	resp, err := c.Do(req)
	if err != nil {
		log.Error(err)
		return
	}
	defer resp.Body.Close()

	var stop bool
	go func(s *bool, ch chan bool) {
		stop = <-ch
	}(&stop, ch)

	if resp.StatusCode == 400 {
		e := &LogRequestError{}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
			p.Del(pod)
			return
		}

		log.Error(string(data))

		err = json.Unmarshal(data, &e)
		if err != nil {
			log.Error(err)
			p.Del(pod)
			return
		}

		cons := e.Containers()
		if len(cons) > 0 {
			for _, container := range cons {
				c := make(chan bool)
				go p.Run(pod, c, container)
			}
		}
		return
	}

	reader := bufio.NewReader(resp.Body)
	for {
		if stop {
			log.Warn("Stopping logwatcher fro Pod: ", pod)
			p.Del(pod)
			return
		}
		line, err := reader.ReadBytes('\n')
		if err != nil && err == io.EOF {
			log.Errorf("Status code is %d for pod %s", resp.StatusCode, pod)
			p.Del(pod)
			return
		} else if err != nil {
			log.Errorf("%s %s %s", err.Error(), p.Namespace, pod)
			p.Del(pod)
			continue
		}

		p.mux.Lock()
		p.sender.Send(p.Namespace, pod, string(line), con)
		p.mux.Unlock()
		// err = logSender(p.Namespace, pod, string(line), con)
		// if err != nil {
		// 	log.Error(err)
		// 	continue
		// }
	}
}

func (e WatchEvent) Name() string {
	return e.Object.Metadata.Name
}

func (e WatchEvent) State() string {
	return e.Object.Status.Phase
}

func marshalEvent(event watch.Event) WatchEvent {
	var e WatchEvent

	data, err := json.Marshal(event)
	if err != nil {
		log.Error(err)
		return e
	}

	err = json.Unmarshal(data, &e)
	if err != nil {
		log.Error(err)
	}
	return e
}

func NewPodLogs(namespace string, client *kubernetes.Clientset, config *rest.Config) *PodLogs {
	podLogs := &PodLogs{
		Namespace: namespace,
		Channels:  make(map[string]chan bool),
		Client:    client,
		Config:    config,
		sc:        GetSenderConfigFromFlags(),
	}

	err := podLogs.NewSender()
	if err != nil {
		panic(err)
	}

	return podLogs
}

type LogRequestError struct {
	Kind       string      `json:"kind"`
	APIVersion string      `json:"apiVersion"`
	Metadata   interface{} `json:"metadata"`
	Status     string      `json:"status"`
	Message    string      `json:"message"`
	Reason     string      `json:"reason"`
	Code       int         `json:"code"`
}

func (l *LogRequestError) Containers() (containers []string) {
	re, err := regexp.Compile(containersErrorRe)
	if err != nil {
		log.Error(err)
		return
	}
	if re.MatchString(l.Message) {
		m := re.FindStringSubmatch(l.Message)
		if len(m) > 1 {
			for _, con := range strings.Split(m[1], " ") {
				containers = append(containers, con)
			}
		}
	}

	return
}
