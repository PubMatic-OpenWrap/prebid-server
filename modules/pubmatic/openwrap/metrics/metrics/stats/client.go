package stats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/golang/glog"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is a StatClient. All stats related operation will be done using this.
type Client struct {
	config     *Config
	httpClient HttpClient
	endpoint   string
	// logger    logger // TODO : ???
	pubChan   chan stat
	pubTicker *time.Ticker
	statMap   map[string]int
	// mu        sync.Mutex // TODO : not needed anymore
}

// NewClient will validate the Config provided and return a new Client
func NewClient(cfg *Config) (*Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid stats client configurations:%s", err.Error())
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   time.Duration(cfg.DialTimeout) * time.Second,
				KeepAlive: time.Duration(cfg.KeepAliveDuration) * time.Minute,
			}).DialContext,
			MaxIdleConns:          cfg.MaxIdleConns,
			MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
			ResponseHeaderTimeout: 30 * time.Second,
		},
		//  Timeout: , //TODO: do we need this timeout ??
	}

	u := url.URL{
		Scheme:     "http",
		Host:       net.JoinHostPort(cfg.Host, cfg.Port),
		Path:       "stat",
		ForceQuery: true,
	}

	c := &Client{
		config:     cfg,
		httpClient: client,
		endpoint:   u.String(),
		// logger:    lgr,
		pubChan:   make(chan stat, statsChanLen), // TODO : statsChanLen should it be configurable ???
		pubTicker: time.NewTicker(time.Duration(cfg.PublishingInterval) * time.Minute),
		statMap:   make(map[string]int),
	}

	go c.process()
	// go c.startStatsCollector()
	// go c.startStatsPublisher()

	return c, nil
}

// PublishStat will push a stat to pubChan channel.
func (sc *Client) PublishStat(key string, value int) {
	sc.pubChan <- stat{Key: key, Value: value}
}

// process function will keep listening on the pubChan
// It will publish the stats to server if
// (1) number of stats reaches the PublishingThreshold or,
// (2) PublishingInterval timeout occurs

func (sc *Client) process() {

	for {
		select {
		case stat := <-sc.pubChan:
			val := sc.statMap[stat.Key] // if key is absent then val will be 0
			sc.statMap[stat.Key] = stat.Value + val

			if len(sc.statMap) >= sc.config.PublishingThreshold {
				sc.prepareStatsForPublishing()
				sc.pubTicker.Reset(time.Duration(sc.config.PublishingInterval) * time.Minute)
			}

		case <-sc.pubTicker.C:
			sc.prepareStatsForPublishing()
		}
	}
}

// func (sc *Client) startStatsCollector() {
// 	for {
// 		select {
// 		case stat := <-sc.pubChan:
// 			// key := stat.validateStatKey()
// 			key := stat.Key
// 			// val, ok := sc.statMap[key]
// 			sc.mu.Lock()
// 			val, ok := sc.statMap[key] //calling this after mu.lock
// 			if ok {
// 				sc.statMap[key] = stat.Value + val
// 			} else {
// 				sc.statMap[key] = stat.Value
// 			}
// 			sc.mu.Unlock()
// 			if len(sc.statMap) >= sc.config.PublishingThreshold {
// 				sc.prepareStatsForPublishing() //it will wait till data published to stats server
// 				sc.pubTicker.Reset(time.Duration(sc.config.PublishingInterval) * time.Minute)
// 			}
// 		}
// 	}
// }

// func (sc *Client) startStatsPublisher() {
// 	for {
// 		select {
// 		case <-sc.pubTicker.C:
// 			sc.prepareStatsForPublishing()
// 		}
// 	}
// }

func (sc *Client) prepareStatsForPublishing() {
	if len(sc.statMap) != 0 {
		collectedStats := sc.statMap
		sc.statMap = map[string]int{}
		go sc.publishStatsToServer(collectedStats)
	}
}

func (sc *Client) publishStatsToServer(statMap map[string]int) int {

	sb, err := json.Marshal(statMap)
	if err != nil {
		glog.Errorf("[stats_fail] Json unmarshal fail: %v", err)
		return statusSetupFail
	}
	// glog.Info("[stats] Stats to be sent to server: %s", string(sb))

	req, err := http.NewRequest(http.MethodPost, sc.endpoint, io.NopCloser(bytes.NewBuffer(sb)))
	if err != nil {
		glog.Errorf("[stats_fail] Failed to form request to sent stats to server: %v", err)
		return statusSetupFail
	}

	req.Header.Add(contentType, applicationJSON)

	for retry := 0; retry < sc.config.Retries; retry++ {

		startTime := time.Now()
		resp, err := sc.httpClient.Do(req)
		elapsedTime := time.Since(startTime)

		if err == nil && resp != nil && resp.StatusCode == http.StatusOK {
			// glog.Info("[stats_success] Stats sent to server successfully")
			return statusPublishSuccess
		}

		code := 0
		if resp != nil {
			code = resp.StatusCode
			if resp.Body != nil {
				resp.Body.Close()
			}
		}

		if retry >= (sc.config.Retries - 1) {
			glog.Errorf("[stats_fail] retry:[%d] status:[%d] nstats:[%d] time:[%v] error:[%v]", retry, code, len(statMap), elapsedTime, err)
			break
		}

		// glog.Info("[stats_retry] retry:[%d] status:[%d] nstats:[%v] time:[%v] error:[%v]", retry, code, len(statMap), elapsedTime, err)
		if sc.config.retryInterval > 0 {
			time.Sleep(time.Duration(sc.config.retryInterval) * time.Second)
		}
	}
	return statusPublishFail
}
