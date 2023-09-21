package adunitconfig

import (
	"regexp"
	"sync"
)

// Check https://pkg.go.dev/github.com/umisama/go-regexpcache#section-readme for Regex-Caching
var (
	regexMapContainer regexMap
)

// Compile parses a regular expression.
// This compatible with regexp.Compile but this uses a cache.
func Compile(str string) (*regexp.Regexp, error) {
	return regexMapContainer.Get(str)
}

// Match checks whether a textual regular expression matches a string.
// This compatible with regexp.MatchString but this uses a cache.
func MatchString(pattern string, s string) (matched bool, err error) {
	re, err := Compile(pattern)
	if err != nil {
		return false, err
	}
	return re.MatchString(s), nil
}

type regexMap struct {
	regexps map[string]*regexp.Regexp
	mu      *sync.RWMutex
}

func newContainer() regexMap {
	return regexMap{
		regexps: make(map[string]*regexp.Regexp),
		mu:      &sync.RWMutex{},
	}
}

func (s *regexMap) Get(str string) (*regexp.Regexp, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	re, ok := s.regexps[str]
	if ok {
		return re, nil
	}

	var err error

	re, err = regexp.Compile(str)

	if err != nil {
		return nil, err
	}
	s.regexps[str] = re

	return re, nil
}

func init() {
	regexMapContainer = newContainer()
}
