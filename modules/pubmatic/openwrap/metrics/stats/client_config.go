package stats

import (
	"errors"
)

// Config will have the information required to initialise a stats client
type Config struct {
	Host                string
	Port                string
	PublishingInterval  int // interval (in minutes) to publish stats to server
	PublishingThreshold int // publish stats if number of stat-records present in map is higher than this threshold
	Retries             int // max retries to publish stats to server
	DialTimeout         int // http connection dial-timeout (in seconds)
	KeepAliveDuration   int // http connection keep-alive-duration (in minutes)
	MaxIdleConns        int // maximum idle connections across all hosts
	MaxIdleConnsPerHost int // maximum idle connections per host
	retryInterval       int // if failed to publish stat then wait for retryInterval seconds for next attempt

	// Timeout int TODO : http connection timeout ???
	// maxChannelLength int TODO:  should we add ???

	// TODO : remove
	// Server              string
	// DC                  string
	// keyPostFix          string
}

func (c *Config) validate() (err error) {
	if c.Host == "" || c.Port == "" {
		return errors.New("stat server host and port cannot be empty")
	}

	// if c.Server == "" {
	// 	c.Server = "svr0"
	// }

	// if c.DC == "" {
	// 	c.DC = "dc0"
	// }

	// c.keyPostFix = fmt.Sprintf(":%s:%s", c.DC, c.Server)

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
			c.retryInterval = (c.PublishingInterval * 60) / c.Retries //180/5 36
		}

		// TODO : Why ???
		// if c.Retries > (c.PublishingInterval*60)/minRetryDuration {
		// 	c.Retries = (c.PublishingInterval * 60) / minRetryDuration
		// 	c.retryInterval = minRetryDuration
		// } else {
		// 	c.retryInterval = (c.PublishingInterval * 60) / c.Retries
		// }
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
