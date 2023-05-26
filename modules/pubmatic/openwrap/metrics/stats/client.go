package stats

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// Client is a StatClient. All stats related operation will be done using this.
type Client struct {
	config    *Config
	client    *http.Client
	endpoint  string
	logger    logger
	pubChan   chan stat
	pubTicker *time.Ticker
	statMap   map[string]int
	mu        sync.Mutex
}

// NewClient will validate the Config provided and return a new Client
func NewClient(cfg Config, l logger) (*Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, errors.Wrap(err, "invalid stats client configurations")
	}

	hc := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   time.Duration(cfg.DialTimeout) * time.Second,
				KeepAlive: time.Duration(cfg.KeepAliveDuration) * time.Minute,
			}).DialContext,
			MaxIdleConns:          cfg.MaxIdleConns,
			MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
			ResponseHeaderTimeout: 30 * time.Second,
		},
	}

	u := url.URL{
		Scheme:     "http",
		Host:       net.JoinHostPort(cfg.Host, cfg.Port),
		Path:       "stat",
		ForceQuery: true,
	}

	c := &Client{
		config:    &cfg,
		client:    hc,
		endpoint:  u.String(),
		logger:    l,
		pubChan:   make(chan stat, statsChanLen),
		pubTicker: time.NewTicker(time.Duration(cfg.PublishingInterval) * time.Minute),
		statMap:   make(map[string]int),
	}

	go c.startStatsCollector()
	go c.startStatsPublisher()

	return c, nil
}

// PublishStat will push a stat to pubChan channel.
func (sc *Client) PublishStat(key string, value int) {
	var buf bytes.Buffer
	buf.Reset()
	buf.WriteString(key)
	buf.WriteString(sc.config.keyPostFix)
	str := buf.String()
	sc.pubChan <- stat{Key: str, Value: value}
}

func (sc *Client) startStatsCollector() {
	for {
		select {
		case stat := <-sc.pubChan:
			key := stat.validateStatKey()
			val, ok := sc.statMap[key]
			sc.mu.Lock()

			if ok {
				sc.statMap[key] = stat.Value + val
			} else {
				sc.statMap[key] = stat.Value
			}

			sc.mu.Unlock()

			if len(sc.statMap) >= sc.config.PublishingThreshold {
				sc.prepareStatsForPublishing() //it will wait till data published to stats server
				sc.pubTicker.Reset(time.Duration(sc.config.PublishingInterval) * time.Minute)
			}
		}
	}
}

func (sc *Client) startStatsPublisher() {
	for {
		select {
		case <-sc.pubTicker.C:
			sc.prepareStatsForPublishing()
		}
	}
}

func (sc *Client) prepareStatsForPublishing() {
	if len(sc.statMap) != 0 {
		sc.mu.Lock()

		collectedStats := sc.statMap
		sc.statMap = map[string]int{}

		sc.mu.Unlock()

		go sc.publishStatsToServer(collectedStats)
	}
}

func (sc *Client) publishStatsToServer(statMap map[string]int) {
	sb, err := json.Marshal(statMap)
	if nil != err {
		sc.logger.Error("Failed while marshaling stats: %v", err)
		return
	}
	sc.logger.Info("Stats to be sent to server: %s", string(sb))

	req, err := http.NewRequest(http.MethodPost, sc.endpoint, nil)
	if err != nil {
		sc.logger.Error("Failed to form request to sent stats to server: %v", err)
	}

	req.Header.Add(contentType, applicationJSON)

	for i := 0; i <= sc.config.Retries; i++ {
		req.Body = ioutil.NopCloser(bytes.NewBuffer(sb))
		rsp, err := sc.client.Do(req)
		if err != nil || rsp.StatusCode != http.StatusOK {
			if err != nil {
				sc.logger.Error("Failed attempt %d to publish collected stats to server: %v", i+1, err)
			} else if rsp.StatusCode != http.StatusOK {
				sc.logger.Error("Failed attempt %d to publish collected stats to server. Server responded with %d", i+1, rsp.StatusCode)
			}
			if i == sc.config.Retries {
				sc.logger.Error("Stats lost, failed last attempt to publish stats")
			} else {
				sc.logger.Info("Will try publishing again in %d seconds", sc.config.retryInterval)
				time.Sleep(time.Duration(sc.config.retryInterval) * time.Second)
			}
			continue
		}
		sc.logger.Info("Stats sent to server successfully")
		break
	}
}
