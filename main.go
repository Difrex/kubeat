package main

import (
	"time"

	"runtime"

	"os"

	"io/ioutil"

	"github.com/Difrex/kubeat/beater"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	version       = "0.1"
	namespacePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
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

	podLogs := beater.NewPodLogs(getNamespace(), client, config)
	podLogs.SkipVerify = kubeSkipTLSVerify
	podLogs.Ignored = ignorePod

	go podLogs.PodTicker()

	ticker := time.NewTicker(time.Duration(tickTime) * time.Second)

	for t := range ticker.C {
		log.Info(t.Unix(), " Num of logwatchers: ", podLogs.Len())
		log.Info(t.Unix(), " Num of CGOCalls: ", runtime.NumCgoCall())
		log.Info(t.Unix(), " Num of goroutines ", runtime.NumGoroutine())
	}
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

func getNamespace() string {
	f, err := os.Open(namespacePath)
	if os.IsExist(err) {
		data, err := ioutil.ReadAll(f)
		handleError(err)
		return string(data)
	}
	f.Close()
	return namespace
}
