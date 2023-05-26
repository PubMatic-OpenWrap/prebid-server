package stats

import (
	"fmt"
	"strings"
	// "github.com/pm-nilesh-chate/prebid-server/metrics"
)

type StatsTCP struct {
	statsClient *Client
}

type statLogger struct{}

func (l statLogger) Error(format string, args ...interface{}) {
	// logger.Error(format, args...)
}

func (l statLogger) Info(format string, args ...interface{}) {
	// logger.Debug(format, args...)
}

func initTCPStatsClient(statIP, statPort, server, dc string,
	pubInterval, pubThreshold, retries, dialTimeout, keepAliveDur, maxIdleConn, maxIdleConnPerHost int) (*StatsTCP, error) {

	cgf := Config{
		Host:                statIP,
		Port:                statPort,
		Server:              server,
		DC:                  dc,
		PublishingInterval:  pubInterval,
		PublishingThreshold: pubThreshold,
		Retries:             retries,
		DialTimeout:         dialTimeout,
		KeepAliveDuration:   keepAliveDur,
		MaxIdleConns:        maxIdleConn,
		MaxIdleConnsPerHost: maxIdleConnPerHost,
	}

	sc, err := NewClient(cgf, statLogger{})
	if err != nil {
		// logger.Error("Failed to connect to stats server via TCP")
		return nil, err
	}

	return &StatsTCP{statsClient: sc}, nil
}

func formStatKeyWithTrimmedDcPlaceHolder(statIndex int, params ...interface{}) string {
	statKeyFmt := statKeys[statIndex].Fmt
	indexToTrim := strings.LastIndex(statKeyFmt, ":")
	statKeyFmt = statKeyFmt[:indexToTrim]
	return fmt.Sprintf(statKeyFmt, params...)
}

func (st *StatsTCP) RecordOpenWrapServerPanicStats() {
	st.statsClient.PublishStat(formStatKeyWithTrimmedDcPlaceHolder(statsKeyOpenWrapServerPanic), 1)
}
