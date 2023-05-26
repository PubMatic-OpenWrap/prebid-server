package stats

/*
Package contains a TCP basted client library for the internal PubMatic stats server
Following is how you use the this package

import (
	"fmt"
	"git.pubmatic.com/PubMatic/go-common/logger"
	"git.pubmatic.com/PubMatic/go-common/tcpstats"
	"time"
)

type statLogger struct{}

func (l statLogger) Error(format string, args ...interface{}) {
	logger.Error(format, args...)
}

func (l statLogger) Info(format string, args ...interface{}) {
	logger.Info(format, args...)
}

func main() {
	l := statLogger{}

	cfg := tcpstats.Config{
		Host:                "192.168.0.1",
		Port:                "80",
		Server:              "s",
		DC:                  "d",
		PublishingInterval:  2,
		Retries:             1,
		DialTimeout:         3,
		KeepAliveDuration:   30,
		MaxIdleConns:        1,
		MaxIdleConnsPerHost: 2,
	}

	c, err := tcpstats.NewClient(cfg, l)
	if err != nil {
		logger.Fatal("error creating stats client: %v", err)
	}

	go func() {
		for {
			c.PublishStat(fmt.Sprintf("%s:%s:%d", []interface{}{"a", "b", -1}...), 4)
			time.Sleep(1 * time.Second)
		}
	}()

	time.Sleep(10 * time.Minute)
}

*/
