package main

import (
	"flag"
	"regexp"
	"strings"
)

var (
	// Do not accept self logs
	ignorePod         string
	configPath        string
	senderConfigPath  string
	kubeSkipTLSVerify bool
	tickTime          int
)

type ignored []*regexp.Regexp

func init() {
	flag.StringVar(&ignorePod, "ignore-pod", "", "regexp for ignoring self logs")
	flag.StringVar(&configPath, "kube-config", "", "absolute path to the kubectl config")
	flag.StringVar(&senderConfigPath, "sender-config", "", "absolute path to the sender.json")
	flag.BoolVar(&kubeSkipTLSVerify, "kube-skip-tls-verify", false, "skip k8s TLS verification")
	flag.IntVar(&tickTime, "tick-time", 60, "Metrics tick")

	flag.Parse()
}

func (i ignored) isIgnored(name string) bool {
	if i == nil {
		return false
	}
	for _, r := range i {
		if r.MatchString(name) {
			return true
		}
	}
	return false
}

func ignoredPods() ignored {
	if ignorePod == "" {
		return nil
	}
	regexps := strings.Split(ignorePod, ",")
	var ignoredRegexps ignored

	for _, i := range regexps {
		r, err := regexp.Compile(i)
		if err != nil {
			panic("Can't compile regexp: " + err.Error())
		}
		if r != nil {
			ignoredRegexps = append(ignoredRegexps, r)
		}
	}

	return ignoredRegexps
}
