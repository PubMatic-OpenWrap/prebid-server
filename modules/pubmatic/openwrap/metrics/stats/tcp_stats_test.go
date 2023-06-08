package stats

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitTCPStatsClient(t *testing.T) {

	type args struct {
		endpoint string
		pubInterval, pubThreshold, retries, dialTimeout,
		keepAliveDur, maxIdleConn, maxIdleConnPerHost, respHeaderTimeout,
		maxChannelLength, poolMaxWorkers, poolMaxCapacity int
	}

	type want struct {
		client *StatsTCP
		err    error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "returns_error",
			args: args{
				endpoint:           "",
				pubInterval:        10,
				pubThreshold:       10,
				retries:            3,
				dialTimeout:        10,
				keepAliveDur:       10,
				maxIdleConn:        10,
				maxIdleConnPerHost: 10,
			},
			want: want{
				client: nil,
				err:    fmt.Errorf("invalid stats client configurations:stat server endpoint cannot be empty"),
			},
		},
		{
			name: "returns_valid_client",
			args: args{
				endpoint:           "http://10.10.10.10:8000/stat",
				pubInterval:        10,
				pubThreshold:       10,
				retries:            3,
				dialTimeout:        10,
				keepAliveDur:       10,
				maxIdleConn:        10,
				maxIdleConnPerHost: 10,
			},
			want: want{
				client: &StatsTCP{
					statsClient: &Client{
						endpoint: "http://10.10.10.10:8000/stat",
						httpClient: &http.Client{
							Transport: &http.Transport{
								DialContext: (&net.Dialer{
									Timeout:   10 * time.Second,
									KeepAlive: 10 * time.Minute,
								}).DialContext,
								MaxIdleConns:          10,
								MaxIdleConnsPerHost:   10,
								ResponseHeaderTimeout: 30 * time.Second,
							},
						},
						config: &config{
							Endpoint:              "http://10.10.10.10:8000/stat",
							PublishingInterval:    5,
							PublishingThreshold:   1000,
							Retries:               3,
							DialTimeout:           10,
							KeepAliveDuration:     15,
							MaxIdleConns:          10,
							MaxIdleConnsPerHost:   10,
							retryInterval:         100,
							MaxChannelLength:      1000,
							ResponseHeaderTimeout: 30,
							PoolMaxWorkers:        minPoolWorker,
							PoolMaxCapacity:       minPoolCapacity,
						},
						pubChan: make(chan stat, 1000),
						statMap: map[string]int{},
					},
				},
				err: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := initTCPStatsClient(tt.args.endpoint,
				tt.args.pubInterval, tt.args.pubThreshold, tt.args.retries, tt.args.dialTimeout, tt.args.keepAliveDur,
				tt.args.maxIdleConn, tt.args.maxIdleConnPerHost, tt.args.respHeaderTimeout, tt.args.maxChannelLength,
				tt.args.poolMaxWorkers, tt.args.poolMaxCapacity)

			assert.Equal(t, tt.want.err, err)
			if err == nil {
				compareClient(tt.want.client.statsClient, client.statsClient, t)
			}
		})
	}
}
