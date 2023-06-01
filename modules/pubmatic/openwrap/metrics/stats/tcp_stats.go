package stats

import (
	"fmt"

	"github.com/golang/glog"
)

type StatsTCP struct {
	statsClient *Client
}

func initTCPStatsClient(statIP, statPort, server, dc string,
	pubInterval, pubThreshold, retries, dialTimeout, keepAliveDur, maxIdleConn, maxIdleConnPerHost int) (*StatsTCP, error) {

	cfg := Config{
		Host: statIP,
		Port: statPort,
		// Server: server,
		// DC:                  dc,
		PublishingInterval:  pubInterval,
		PublishingThreshold: pubThreshold,
		Retries:             retries,
		DialTimeout:         dialTimeout,
		KeepAliveDuration:   keepAliveDur,
		MaxIdleConns:        maxIdleConn,
		MaxIdleConnsPerHost: maxIdleConnPerHost,
	}

	sc, err := NewClient(&cfg)
	if err != nil {
		glog.Error("Failed to connect to stats server via TCP")
		return nil, err
	}

	return &StatsTCP{statsClient: sc}, nil
}

func (st *StatsTCP) RecordOpenWrapServerPanicStats() {
	st.statsClient.PublishStat(fmt.Sprintf(statKeys[statsKeyOpenWrapServerPanic]), 1)
}
