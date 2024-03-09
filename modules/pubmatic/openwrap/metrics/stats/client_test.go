package stats

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/alitto/pond"
	"github.com/golang/mock/gomock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics/stats/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {

	type args struct {
		cfg *config
	}

	type want struct {
		err        error
		statClient *Client
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "invalid_config",
			args: args{
				cfg: &config{
					Endpoint: "",
				},
			},
			want: want{
				err:        fmt.Errorf("invalid stats client configurations:stat server endpoint cannot be empty"),
				statClient: nil,
			},
		},
		{
			name: "valid_config",
			args: args{
				cfg: &config{
					Endpoint:            "10.10.10.10:8080/stat",
					PublishingInterval:  3,
					DialTimeout:         minDialTimeout,
					KeepAliveDuration:   minKeepAliveDuration,
					MaxIdleConns:        0,
					MaxIdleConnsPerHost: 0,
					PublishingThreshold: minPublishingThreshold,
					Retries:             5,
					retryInterval:       36,
				},
			},
			want: want{
				err: nil,
				statClient: &Client{
					config: &config{
						Endpoint:              "10.10.10.10:8080/stat",
						PublishingInterval:    3,
						DialTimeout:           minDialTimeout,
						KeepAliveDuration:     minKeepAliveDuration,
						MaxIdleConns:          0,
						MaxIdleConnsPerHost:   0,
						PublishingThreshold:   minPublishingThreshold,
						Retries:               5,
						retryInterval:         36,
						MaxChannelLength:      minChannelLength,
						ResponseHeaderTimeout: minResponseHeaderTimeout,
						PoolMaxWorkers:        minPoolWorker,
						PoolMaxCapacity:       minPoolCapacity,
					},
					httpClient: &http.Client{
						Transport: &http.Transport{
							DialContext: (&net.Dialer{
								Timeout:   time.Duration(minDialTimeout) * time.Second,
								KeepAlive: time.Duration(minKeepAliveDuration) * time.Minute,
							}).DialContext,
							MaxIdleConns:          0,
							MaxIdleConnsPerHost:   0,
							ResponseHeaderTimeout: 30 * time.Second,
						},
					},
					endpoint:  "10.10.10.10:8080/stat",
					pubChan:   make(chan stat, minChannelLength),
					pubTicker: time.NewTicker(time.Duration(3) * time.Minute),
					statMap:   map[string]int{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.args.cfg)
			assert.Equal(t, tt.want.err, err, "Mismatched error")
			compareClient(tt.want.statClient, client, t)
		})
	}
}

func compareClient(expectedClient, actualClient *Client, t *testing.T) {

	if expectedClient != nil && actualClient != nil {
		assert.Equal(t, expectedClient.endpoint, actualClient.endpoint, "Mismatched endpoint")
		assert.Equal(t, expectedClient.config, actualClient.config, "Mismatched config")
		assert.Equal(t, cap(expectedClient.pubChan), cap(actualClient.pubChan), "Mismatched pubChan capacity")
		assert.Equal(t, expectedClient.statMap, actualClient.statMap, "Mismatched statMap")
	}

	if expectedClient != nil && actualClient == nil {
		t.Errorf("actualClient is expected to be non-nil")
	}

	if actualClient != nil && expectedClient == nil {
		t.Errorf("actualClient is expected to be nil")
	}
}

func TestPublishStat(t *testing.T) {

	type args struct {
		keyVal      map[string]int
		maxChanSize int
	}

	type want struct {
		keyVal      map[string]int
		channelSize int
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "push_multiple_stat",
			args: args{
				keyVal: map[string]int{
					"key1": 10,
					"key2": 20,
				},
				maxChanSize: 2,
			},
			want: want{
				keyVal: map[string]int{
					"key1": 10,
					"key2": 20,
				},
				channelSize: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{
				pubChan: make(chan stat, tt.args.maxChanSize),
			}
			for k, v := range tt.args.keyVal {
				client.PublishStat(k, v)
			}

			close(client.pubChan)
			assert.Equal(t, tt.want.channelSize, len(client.pubChan))
			for stat := range client.pubChan {
				assert.Equal(t, stat.Value, tt.want.keyVal[stat.Key])
			}
		})
	}
}

func TestPrepareStatsForPublishing(t *testing.T) {

	type args struct {
		client *Client
	}

	tests := []struct {
		name           string
		args           args
		expectedLength int
	}{
		{
			name: "statMap_should_be_empty",
			args: args{
				client: &Client{
					statMap: map[string]int{
						"key1": 10,
						"key2": 20,
					},
					config: &config{
						Retries: 1,
					},
					httpClient: http.DefaultClient,
					pool:       pond.New(2, 2),
				},
			},
			expectedLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.client.prepareStatsForPublishing()
			assert.Equal(t, len(tt.args.client.statMap), tt.expectedLength)
		})
	}
}

func TestPublishStatsToServer(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockHttpClient(ctrl)

	type args struct {
		statClient *Client
		statsMap   map[string]int
	}

	tests := []struct {
		name          string
		args          args
		expStatusCode int
		setup         func()
	}{
		{
			name: "invalid_url",
			args: args{
				statClient: &Client{
					endpoint: "%%invalid%%url",
				},
				statsMap: map[string]int{
					"key": 10,
				},
			},
			setup:         func() {},
			expStatusCode: statusSetupFail,
		},
		{
			name: "server_responds_with_error",
			args: args{
				statClient: &Client{
					endpoint: "http://any-random-server.com",
					config: &config{
						Retries: 1,
					},
					httpClient: mockClient,
				},
				statsMap: map[string]int{
					"key": 10,
				},
			},
			setup: func() {
				mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: 500, Body: io.NopCloser(bytes.NewReader(nil))}, nil)
			},
			expStatusCode: statusPublishFail,
		},
		{
			name: "server_responds_with_error_multi_retries",
			args: args{
				statClient: &Client{
					endpoint: "http://any-random-server.com",
					config: &config{
						Retries:       3,
						retryInterval: 1,
					},
					httpClient: mockClient,
				},
				statsMap: map[string]int{
					"key": 10,
				},
			},
			setup: func() {
				mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: 500, Body: io.NopCloser(bytes.NewReader(nil))}, nil).Times(3)
			},
			expStatusCode: statusPublishFail,
		},
		{
			name: "first_attempt_fail_second_attempt_success",
			args: args{
				statClient: &Client{
					endpoint: "http://any-random-server.com",
					config: &config{
						Retries:       3,
						retryInterval: 1,
					},
					httpClient: mockClient,
				},
				statsMap: map[string]int{
					"key": 10,
				},
			},
			setup: func() {
				gomock.InOrder(
					mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: 500, Body: io.NopCloser(bytes.NewReader(nil))}, nil),
					mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(nil))}, nil),
				)
			},
			expStatusCode: statusPublishSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			statusCode := tt.args.statClient.publishStatsToServer(tt.args.statsMap)
			assert.Equal(t, tt.expStatusCode, statusCode)
		})
	}
}

func TestProcess(t *testing.T) {
	type args struct {
		client *Client
	}

	tests := []struct {
		name        string
		args        args
		setup       func(*Client)
		getMockPool func(wg *sync.WaitGroup) (*gomock.Controller, WorkerPool)
	}{
		{
			name: "PublishingThreshold_limit_reached",
			args: args{
				client: &Client{
					statMap: map[string]int{},
					config: &config{
						Retries:             1,
						PublishingInterval:  1,
						PublishingThreshold: 2,
					},
					pubChan:      make(chan stat, 2),
					pubTicker:    time.NewTicker(1 * time.Minute),
					shutDownChan: make(chan struct{}),
				},
			},
			setup: func(client *Client) {
				client.PublishStat("key1", 1)
				client.PublishStat("key2", 2)
			},
			getMockPool: func(wg *sync.WaitGroup) (*gomock.Controller, WorkerPool) {
				ctrl := gomock.NewController(t)
				mockWorkerPool := mock.NewMockWorkerPool(ctrl)
				mockWorkerPool.EXPECT().TrySubmit(gomock.Any()).DoAndReturn(func(task func()) bool {
					wg.Done()
					return true
				})
				return ctrl, mockWorkerPool
			},
		},
		{
			name: "PublishingInterval_timer_timeouts",
			args: args{
				client: &Client{
					statMap: map[string]int{},
					config: &config{
						Retries:             1,
						PublishingInterval:  1,
						PublishingThreshold: 10,
					},
					pubChan:      make(chan stat, 10),
					pubTicker:    time.NewTicker(1 * time.Second),
					shutDownChan: make(chan struct{}),
				},
			},
			setup: func(client *Client) {
				client.PublishStat("key1", 1)
				client.PublishStat("key2", 2)
			},
			getMockPool: func(wg *sync.WaitGroup) (*gomock.Controller, WorkerPool) {
				ctrl := gomock.NewController(t)
				mockWorkerPool := mock.NewMockWorkerPool(ctrl)
				mockWorkerPool.EXPECT().TrySubmit(gomock.Any()).DoAndReturn(func(task func()) bool {
					wg.Done()
					return true
				})
				return ctrl, mockWorkerPool
			},
		},
		{
			name: "graceful_shutdown_process",
			args: args{
				client: &Client{
					statMap: map[string]int{},
					config: &config{
						Retries:             1,
						PublishingThreshold: 5,
					},
					pubChan:      make(chan stat, 10),
					pubTicker:    time.NewTicker(1 * time.Minute),
					shutDownChan: make(chan struct{}),
				},
			},
			setup: func(client *Client) {
				client.PublishStat("key1", 1)
				time.Sleep(1 * time.Second)
				client.ShutdownProcess()
			},
			getMockPool: func(wg *sync.WaitGroup) (*gomock.Controller, WorkerPool) {
				ctrl := gomock.NewController(t)
				mockWorkerPool := mock.NewMockWorkerPool(ctrl)
				mockWorkerPool.EXPECT().TrySubmit(gomock.Any()).DoAndReturn(func(task func()) bool {
					wg.Done()
					return true
				})
				return ctrl, mockWorkerPool
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			wg.Add(1)

			client := tt.args.client
			ctrl, mockPool := tt.getMockPool(&wg)
			defer ctrl.Finish()
			client.pool = mockPool

			go client.process()

			tt.setup(client)

			wg.Wait()
		})
	}
}
