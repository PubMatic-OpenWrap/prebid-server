package stats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/alitto/pond"
	"github.com/golang/glog"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// TrySubmit attempts to send a task to this worker pool for execution. If the queue is full,
// it will not wait for a worker to become idle. It returns true if it was able to dispatch
// the task and false otherwise.
type WorkerPool interface {
	TrySubmit(task func()) bool
}

// Client is a StatClient. All stats related operation will be done using this.
type Client struct {
	config       *config
	httpClient   HttpClient
	endpoint     string
	pubChan      chan stat
	pubTicker    *time.Ticker
	statMap      map[string]int
	shutDownChan chan struct{}
	pool         WorkerPool
}

// NewClient will validate the Config provided and return a new Client
func NewClient(cfg *config) (*Client, error) {
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
			ResponseHeaderTimeout: time.Duration(cfg.ResponseHeaderTimeout) * time.Second,
		},
	}

	c := &Client{
		config:       cfg,
		httpClient:   client,
		endpoint:     cfg.Endpoint,
		pubChan:      make(chan stat, cfg.MaxChannelLength),
		pubTicker:    time.NewTicker(time.Duration(cfg.PublishingInterval) * time.Minute),
		statMap:      make(map[string]int),
		shutDownChan: make(chan struct{}),
		pool:         pond.New(cfg.PoolMaxWorkers, cfg.PoolMaxCapacity),
	}

	go c.process()

	return c, nil
}

// ShutdownProcess will perform the graceful shutdown operation
func (sc *Client) ShutdownProcess() {
	sc.shutDownChan <- struct{}{}
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
			sc.statMap[stat.Key] = sc.statMap[stat.Key] + stat.Value
			if len(sc.statMap) >= sc.config.PublishingThreshold {
				sc.prepareStatsForPublishing()
				sc.pubTicker.Reset(time.Duration(sc.config.PublishingInterval) * time.Minute)
			}

		case <-sc.pubTicker.C:
			sc.prepareStatsForPublishing()

		case <-sc.shutDownChan:
			sc.prepareStatsForPublishing()
			return
		}
	}
}

// prepareStatsForPublishing creates copy of map containing stat-key and value
// and calls publishStatsToServer to publishes it to the stat-server
func (sc *Client) prepareStatsForPublishing() {
	if len(sc.statMap) != 0 {
		collectedStats := sc.statMap
		sc.statMap = map[string]int{}
		status := sc.pool.TrySubmit(func() {
			sc.publishStatsToServer(collectedStats)
		})
		if !status {
			glog.Errorf("[stats_fail] Failed to submit the publishStatsToServer task containing %d record to pool", len(collectedStats))
		}
	}
}

// publishStatsToServer sends the stats to the stat-server
// in case of failure, it retries to send for Client.config.Retries number of times.
func (sc *Client) publishStatsToServer(statMap map[string]int) int {

	statJson, err := json.Marshal(statMap)
	if err != nil {
		glog.Errorf("[stats_fail] Json unmarshal fail: %v", err)
		return statusSetupFail
	}

	glog.V(3).Infof("[stats] nstats:[%d] data:[%s]", len(statMap), statJson)
	req, err := http.NewRequest(http.MethodPost, sc.endpoint, bytes.NewBuffer(statJson))
	if err != nil {
		glog.Errorf("[stats_fail] Failed to form request to sent stats to server: %v", err)
		return statusSetupFail
	}

	req.Header.Add(contentType, applicationJSON)
	for retry := 0; retry < sc.config.Retries; retry++ {

		startTime := time.Now()
		resp, err := sc.httpClient.Do(req)
		elapsedTime := time.Since(startTime)

		code := 0
		if resp != nil {
			code = resp.StatusCode
			defer resp.Body.Close()
		}

		if err == nil && code == http.StatusOK {
			glog.Infof("[stats_success] retry:[%d] nstats:[%d] time:[%v]", retry, len(statMap), elapsedTime)
			return statusPublishSuccess
		}

		if retry == (sc.config.Retries - 1) {
			glog.Errorf("[stats_fail] retry:[%d] status:[%d] nstats:[%d] time:[%v] error:[%v]", retry, code, len(statMap), elapsedTime, err)
			break
		}

		if sc.config.retryInterval > 0 {
			time.Sleep(time.Duration(sc.config.retryInterval) * time.Second)
		}
	}

	return statusPublishFail
}
