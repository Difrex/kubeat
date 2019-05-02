package beater

import (
	"os"
	"regexp"
	"strings"

	"k8s.io/api/core/v1"
)

type ignored []*regexp.Regexp

const (
	annotation_key   = "kubeat_disable"
	annotation_value = "yes"
)

func checkAnnotation(annotations map[string]string) bool {
	if annotations == nil {
		return false
	}

	if v, ok := annotations[annotation_key]; ok {
		if v == annotation_value {
			return true
		}
	}
	return false
}

func (i ignored) isIgnored(pod v1.Pod) bool {
	if checkAnnotation(pod.Annotations) {
		return true
	}
	if i == nil {
		return false
	}
	for _, r := range i {
		if r.MatchString(pod.Name) {
			return true
		}
	}
	return false
}

func ignoredPods(ignorePod string) ignored {
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

func isInK8S() bool {
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return true
	}
	return false
}

func getESCredsFromEnv() (string, string) {
	user := os.Getenv(ELASTIC_ENV_USERNAME)
	pass := os.Getenv(ELASTIC_ENV_PASSWORD)
	return user, pass
}
