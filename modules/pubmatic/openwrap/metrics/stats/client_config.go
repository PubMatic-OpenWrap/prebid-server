package stats

import (
	"errors"
)

// config will have the information required to initialise a stats client
type config struct {
	Endpoint              string // stat-server's endpoint
	PublishingInterval    int    // interval (in minutes) to publish stats to server
	PublishingThreshold   int    // publish stats if number of stat-records present in map is higher than this threshold
	Retries               int    // max retries to publish stats to server
	DialTimeout           int    // http connection dial-timeout (in seconds)
	KeepAliveDuration     int    // http connection keep-alive-duration (in minutes)
	MaxIdleConns          int    // maximum idle connections across all hosts
	MaxIdleConnsPerHost   int    // maximum idle connections per host
	retryInterval         int    // if failed to publish stat then wait for retryInterval seconds for next attempt
	ResponseHeaderTimeout int    // amount of time (in seconds) to wait for server's response header
	MaxChannelLength      int    // max number of stat keys
	PoolMaxWorkers        int    // max number of workers that will actually send the data to stats-server
	PoolMaxCapacity       int    // number of tasks that can be submitted to the pool without blocking
}

func (c *config) validate() (err error) {
	if c.Endpoint == "" {
		return errors.New("stat server endpoint cannot be empty")
	}

	if c.PublishingInterval < minPublishingInterval {
		c.PublishingInterval = minPublishingInterval
	} else if c.PublishingInterval > maxPublishingInterval {
		c.PublishingInterval = maxPublishingInterval
	}

	if c.Retries > 0 {
		if c.Retries > (c.PublishingInterval*60)/minRetryDuration {
			c.Retries = (c.PublishingInterval * 60) / minRetryDuration
			c.retryInterval = minRetryDuration
		} else {
			c.retryInterval = (c.PublishingInterval * 60) / c.Retries
		}
	}

	if c.DialTimeout < minDialTimeout {
		c.DialTimeout = minDialTimeout
	}

	if c.KeepAliveDuration < minKeepAliveDuration {
		c.KeepAliveDuration = minKeepAliveDuration
	}

	if c.MaxIdleConns < 0 {
		c.MaxIdleConns = 0
	}

	if c.MaxIdleConnsPerHost < 0 {
		c.MaxIdleConnsPerHost = 0
	}

	if c.PublishingThreshold < minPublishingThreshold {
		c.PublishingThreshold = minPublishingThreshold
	}

	if c.ResponseHeaderTimeout < minResponseHeaderTimeout {
		c.ResponseHeaderTimeout = minResponseHeaderTimeout
	}

	if c.MaxChannelLength < minChannelLength {
		c.MaxChannelLength = minChannelLength
	}

	if c.PoolMaxWorkers < minPoolWorker {
		c.PoolMaxWorkers = minPoolWorker
	}

	if c.PoolMaxCapacity < minPoolCapacity {
		c.PoolMaxCapacity = minPoolCapacity
	}

	return nil
}
