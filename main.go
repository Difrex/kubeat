package main

import (
	"encoding/json"
	"fmt"
	"time"

	"runtime"

	"os"

	"github.com/Difrex/kubeat/beater"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type LogMessage struct {
	PodName    string                 `json:"pod_name"`
	Namespace  string                 `json:"namespace"`
	Container  string                 `json:"container"`
	Message    string                 `json:"message"`
	SenderTime time.Time              `json:"sender_time"`
	Meta       map[string]interface{} `json:"meta"`
}

const (
	version = "0.1"
)

func main() {

	var client *kubernetes.Clientset
	var config *rest.Config
	if isInK8S() {
		log.Info("K8S launch detected")
		c, err := rest.InClusterConfig()
		handleError(err)
		config = c
	} else {
		log.Info("Outside K8S launch detected")
		c, err := clientcmd.BuildConfigFromFlags("", configPath)
		handleError(err)
		config = c
	}

	client, err := kubernetes.NewForConfig(config)
	handleError(err)

	podLogs := beater.NewPodLogs("difrex", client, config)

	pods, err := client.CoreV1().Pods("difrex").List(metav1.ListOptions{})
	handleError(err)

	ignored := ignoredPods()
	for _, pod := range pods.Items {
		if !ignored.isIgnored(pod.Name) && pod.Status.Phase == "Running" {
			ch := make(chan bool)
			go podLogs.Run(pod.Name, ch, "")
			podLogs.Add(pod.Name, ch)
		}
	}

	go podLogs.Watch()

	ticker := time.NewTicker(time.Duration(tickTime) * time.Second)

	for t := range ticker.C {
		log.Info(t.Unix(), " Num of logwatchers: ", podLogs.Len())
		log.Info(t.Unix(), " Num of CGOCalls: ", runtime.NumCgoCall())
	}
}

func logSender(ns, pod, message, con string) error {
	log := LogMessage{
		Namespace:  ns,
		PodName:    pod,
		Message:    message,
		Container:  con,
		SenderTime: time.Now(),
	}
	data, err := json.Marshal(log)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func isInK8S() bool {
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return true
	}
	return false
}
