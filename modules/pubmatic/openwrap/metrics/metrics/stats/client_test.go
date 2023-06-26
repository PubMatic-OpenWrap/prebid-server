package stats

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/metrics/stats/mock"

	// "github.com/pm-nilesh-chate/prebid-server/modules/pubmatic/openwrap/metrics/stats/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {

	type args struct {
		cfg *Config
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
				cfg: &Config{
					Host: "",
				},
			},
			want: want{
				err:        fmt.Errorf("invalid stats client configurations:stat server host and port cannot be empty"),
				statClient: nil,
			},
		},
		{
			name: "valid_config",
			args: args{
				cfg: &Config{
					Host:                "10.10.10.10",
					Port:                "8000",
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
					config: &Config{
						Host:                "10.10.10.10",
						Port:                "8000",
						PublishingInterval:  3,
						DialTimeout:         minDialTimeout,
						KeepAliveDuration:   minKeepAliveDuration,
						MaxIdleConns:        0,
						MaxIdleConnsPerHost: 0,
						PublishingThreshold: minPublishingThreshold,
						Retries:             5,
						retryInterval:       36,
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
					endpoint:  "http://10.10.10.10:8000/stat?",
					pubChan:   make(chan stat, statsChanLen),
					pubTicker: time.NewTicker(time.Duration(3) * time.Minute),
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
		assert.Equal(t, expectedClient.endpoint, actualClient.endpoint, "Mismatched endpoint")
		assert.Equal(t, cap(expectedClient.pubChan), cap(actualClient.pubChan), "Mismatched pubChan capacity")
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

	tests := []struct {
		name               string
		args               args
		expectedChanLength int
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
			expectedChanLength: 2,
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
			for stat := range client.pubChan {
				assert.Equal(t, stat.Value, tt.args.keyVal[stat.Key])
			}
		})
	}
}

func TestPrepareStatsForPublishing(t *testing.T) {

	type args struct {
		client *Client
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "statMap_should_be_empty",
			args: args{
				client: &Client{
					statMap: map[string]int{
						"key1": 10,
						"key2": 20,
					},
					config: &Config{
						Retries: 1,
					},
					httpClient: http.DefaultClient,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.client.prepareStatsForPublishing()
			assert.Equal(t, len(tt.args.client.statMap), 0)
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
					config: &Config{
						Retries: 1,
					},
					httpClient: mockClient,
				},
				statsMap: map[string]int{
					"key": 10,
				},
			},
			setup: func() {
				mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: 500, Body: http.NoBody}, nil)
			},
			expStatusCode: statusPublishFail,
		},
		{
			name: "server_responds_with_error_multi_retries",
			args: args{
				statClient: &Client{
					endpoint: "http://any-random-server.com",
					config: &Config{
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
				mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: 500}, nil).Times(3)
			},
			expStatusCode: statusPublishFail,
		},
		{
			name: "first_attempt_fail_second_attempt_success",
			args: args{
				statClient: &Client{
					endpoint: "http://any-random-server.com",
					config: &Config{
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
					mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: 500}, nil),
					mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: 200}, nil),
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
		client    *Client
		sleepTime time.Duration
	}

	tests := []struct {
		name  string
		args  args
		setup func(*Client)
	}{
		{
			name: "PublishingThreshold_limit_reached",
			args: args{
				client: &Client{
					statMap: map[string]int{},
					config: &Config{
						Retries:            1,
						PublishingInterval: 1,
					},
					pubChan:    make(chan stat, 2),
					httpClient: http.DefaultClient,
					pubTicker:  time.NewTicker(1 * time.Minute),
				},
				sleepTime: time.Second * 3,
			},
			setup: func(client *Client) {
				client.pubChan <- stat{Key: "key1", Value: 1}
				client.pubChan <- stat{Key: "key2", Value: 2}
			},
		},
		{
			name: "PublishingInterval_timer_timeouts",
			args: args{
				client: &Client{
					statMap: map[string]int{},
					config: &Config{
						Retries:            1,
						PublishingInterval: 1,
					},
					pubChan:    make(chan stat, 10),
					httpClient: http.DefaultClient,
					pubTicker:  time.NewTicker(2 * time.Second),
				},
				sleepTime: time.Second * 5,
			},
			setup: func(client *Client) {
				client.pubChan <- stat{Key: "key1", Value: 1}
				client.pubChan <- stat{Key: "key2", Value: 2}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.args.client

			go tt.setup(client) // push stats into the channel
			go client.process()
			time.Sleep(tt.args.sleepTime) // wait time till stats-client publish stats to server

			assert.Equal(t, len(client.statMap), 0)
		})
	}
}
