package beater

import (
	"bufio"
	"crypto/tls"
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
	Channels      map[string]chan bool
	Client        *kubernetes.Clientset
	Config        *rest.Config
	Ignored       string
	Namespace     string
	SkipVerify    bool
	EnableWatcher bool

	tick   int
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

	if !p.EnableWatcher {
		go p.sender.Ticker()
	}

	ignored := ignoredPods(p.Ignored)
	ticker := time.NewTicker(time.Second * time.Duration(p.tick))
	for c := range ticker.C {
		log.Warn("New tick in pod watcher")

		timeout := int64(10)
		pods, err := p.Client.CoreV1().Pods(p.Namespace).List(metav1.ListOptions{TimeoutSeconds: &timeout})
		if err != nil {
			log.Error("Error on tick: ", c, " ", err.Error())
			continue
		}

		log.Info("Get pods: ", len(pods.Items))
		for _, pod := range pods.Items {
			if podCh, ok := p.Channels[pod.Name]; !ok && pod.Status.Phase == "Running" && !ignored.isIgnored(pod.Name) {
				ch := make(chan bool)
				p.Add(pod.Name, ch)
				go p.Run(pod.Name, ch, "")
			} else if ok && pod.Status.Phase != "Running" || ignored.isIgnored(pod.Name) {
				p.Stop(podCh)
			}
		}
	}
}

// Watch watches for k8s events
func (p *PodLogs) Watch() {

	go p.PodTicker()
	go p.sender.Ticker()

	ignored := ignoredPods(p.Ignored)

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
			ch := make(chan bool, 1)
			if podCh, ok := p.Channels[name]; !ok && e.State() == "Running" && !ignored.isIgnored(name) {
				p.Add(name, ch)
				go p.Run(name, ch, "")
			} else if ok && e.State() != "Running" {
				p.Stop(podCh)
			}
		case "ADDED":
			ch := make(chan bool, 1)
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
	log.Warnf("Trying to start watcher for pod %s-%s", pod, con)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: p.SkipVerify}
	c := &http.Client{}

	podApi := p.Config.Host + "/api/v1/namespaces/" + p.Namespace + "/pods/" + pod
	var req *http.Request
	if con == "" {
		r, err := http.NewRequest("GET", podApi+"/log?follow=true&tailLines=10", nil)
		if err != nil {
			p.Del(pod)
			if c, ok := p.Channels[pod+"-"+con]; ok {
				p.Stop(c)
				p.Del(pod + "-" + con)
			}
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
			if c, ok := p.Channels[pod+"-"+con]; ok {
				p.Stop(c)
				p.Del(pod + "-" + con)
			}
			return
		}
		req = r
	}

	req.Header.Add("Authorization", "Bearer "+p.Config.BearerToken)

	resp, err := c.Do(req)
	if err != nil {
		log.Error(err)
		p.Del(pod)
		if c, ok := p.Channels[pod+"-"+con]; ok {
			p.Stop(c)
			p.Del(pod + "-" + con)
		}
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
			if c, ok := p.Channels[pod+"-"+con]; ok {
				p.Stop(c)
				p.Del(pod + "-" + con)
			}
			return
		}

		log.Error(string(data))

		err = json.Unmarshal(data, &e)
		if err != nil {
			log.Error(err)
			p.Del(pod)
			if c, ok := p.Channels[pod+"-"+con]; ok {
				p.Stop(c)
				p.Del(pod + "-" + con)
			}
			return
		}

		cons := e.Containers()
		if len(cons) > 0 {
			for _, container := range cons {
				c := make(chan bool, 1)
				p.Add(pod+"-"+con, c)
				go p.Run(pod, c, container)
			}
		}
		return
	}

	log.Warnf("Watcher for pod %s-%s started", pod, con)
	reader := bufio.NewReader(resp.Body)
	for {
		if stop {
			log.Warn("Stopping logwatcher for Pod: ", pod)
			p.Del(pod)
			if _, ok := p.Channels[pod+"-"+con]; ok {
				p.Del(pod + "-" + con)
			}
			return
		}
		line, err := reader.ReadBytes('\n')
		if err != nil && err == io.EOF {
			log.Errorf("Received EOF for pod %s. Shutdown logwatcher.", pod)
			p.Stop(ch)
			continue
		} else if err != nil {
			log.Errorf("Error received %s for pod %s-%s. Shutdown logwatcher.", err.Error(), pod, con)
			p.Stop(ch)
			continue
		}

		// p.mux.Lock()
		p.sender.Send(p.Namespace, pod, string(line), con)
		// p.mux.Unlock()
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
		Namespace:     namespace,
		Channels:      make(map[string]chan bool),
		Client:        client,
		Config:        config,
		EnableWatcher: isWatcherEnabled(),
		tick:          GetTickFromFlags(),
		sc:            GetSenderConfigFromFlags(),
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
