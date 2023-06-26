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
		statIP, statPort string
		pubInterval, pubThreshold, retries, dialTimeout,
		keepAliveDur, maxIdleConn, maxIdleConnPerHost int
	}

	type want struct {
		client *statsTCP
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
				statIP:             "10.10.10.10",
				statPort:           "",
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
				err:    fmt.Errorf("invalid stats client configurations:stat server host and port cannot be empty"),
			},
		},
		{
			name: "returns_valid_client",
			args: args{
				statIP:             "10.10.10.10",
				statPort:           "8000",
				pubInterval:        10,
				pubThreshold:       10,
				retries:            3,
				dialTimeout:        10,
				keepAliveDur:       10,
				maxIdleConn:        10,
				maxIdleConnPerHost: 10,
			},
			want: want{
				client: &statsTCP{
					statsClient: &Client{
						endpoint: "http://10.10.10.10:8000/stat?",
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
						config: &Config{
							Host:                "10.10.10.10",
							Port:                "8000",
							PublishingInterval:  5,
							PublishingThreshold: 1000,
							Retries:             3,
							DialTimeout:         10,
							KeepAliveDuration:   15,
							MaxIdleConns:        10,
							MaxIdleConnsPerHost: 10,
							retryInterval:       100,
						},
						pubChan: make(chan stat, 1000),
					},
				},
				err: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := initTCPStatsClient(tt.args.statIP, tt.args.statPort,
				tt.args.pubInterval, tt.args.pubThreshold, tt.args.retries, tt.args.dialTimeout, tt.args.keepAliveDur,
				tt.args.maxIdleConn, tt.args.maxIdleConnPerHost)

			assert.Equal(t, tt.want.err, err)
			if err == nil {
				compareClient(tt.want.client.statsClient, client.statsClient, t)
			}
		})
	}
}
