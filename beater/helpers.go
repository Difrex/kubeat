package beater

import (
	"os"
	"regexp"
	"strings"
)

type ignored []*regexp.Regexp

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
