package stats

import (
	"errors"
	"fmt"
)

// Config will have the information required to initialise a stats client
type Config struct {
	Host                string
	Port                string
	Server              string
	DC                  string
	PublishingInterval  int // In minutes
	PublishingThreshold int
	Retries             int
	DialTimeout         int // In seconds
	KeepAliveDuration   int // In seconds
	MaxIdleConns        int
	MaxIdleConnsPerHost int

	retryInterval int // In seconds
	keyPostFix    string
}

func (c *Config) validate() (err error) {
	if c.Host == "" || c.Port == "" {
		return errors.New("stat server host and port cannot be empty")
	}

	if c.Server == "" {
		c.Server = "svr0"
	}

	if c.DC == "" {
		c.DC = "dc0"
	}

	c.keyPostFix = fmt.Sprintf(":%s:%s", c.DC, c.Server)

	if c.PublishingInterval < minPublishingInterval {
		c.PublishingInterval = minPublishingInterval
	} else if c.PublishingInterval > maxPublishingInterval {
		c.PublishingInterval = maxPublishingInterval
	}

	if c.Retries > 0 {
		maxRetriesAllowed := (c.PublishingInterval * 60) / minRetryDuration

		if c.Retries > maxRetriesAllowed {
			c.Retries = maxRetriesAllowed
			c.retryInterval = minRetryDuration
		} else {
			c.retryInterval = (c.PublishingInterval * 60) / c.Retries
		}

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

	return nil
}
