package stats

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {

	type args struct {
		cfg *config
	}

	type want struct {
		err error
		cfg *config
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "empty_endpoint",
			args: args{
				cfg: &config{
					Endpoint: "",
				},
			},
			want: want{
				err: fmt.Errorf("stat server endpoint cannot be empty"),
				cfg: &config{
					Endpoint: "",
				},
			},
		},
		{
			name: "lower_values_than_min_limit",
			args: args{
				cfg: &config{
					Endpoint:            "10.10.10.10/stat",
					PublishingInterval:  0,
					DialTimeout:         0,
					KeepAliveDuration:   0,
					MaxIdleConns:        -1,
					MaxIdleConnsPerHost: -1,
					PublishingThreshold: 0,
				},
			},
			want: want{
				err: nil,
				cfg: &config{
					Endpoint:              "10.10.10.10/stat",
					PublishingInterval:    minPublishingInterval,
					DialTimeout:           minDialTimeout,
					KeepAliveDuration:     minKeepAliveDuration,
					MaxIdleConns:          0,
					MaxIdleConnsPerHost:   0,
					PublishingThreshold:   minPublishingThreshold,
					MaxChannelLength:      minChannelLength,
					ResponseHeaderTimeout: minResponseHeaderTimeout,
					PoolMaxWorkers:        minPoolWorker,
					PoolMaxCapacity:       minPoolCapacity,
				},
			},
		},
		{
			name: "high_PublishingInterval_than_max_limit",
			args: args{
				cfg: &config{
					Endpoint:           "10.10.10.10/stat",
					PublishingInterval: 10,
				},
			},
			want: want{
				err: nil,
				cfg: &config{
					Endpoint:              "10.10.10.10/stat",
					PublishingInterval:    maxPublishingInterval,
					DialTimeout:           minDialTimeout,
					KeepAliveDuration:     minKeepAliveDuration,
					MaxIdleConns:          0,
					MaxIdleConnsPerHost:   0,
					PublishingThreshold:   minPublishingThreshold,
					MaxChannelLength:      minChannelLength,
					ResponseHeaderTimeout: minResponseHeaderTimeout,
					PoolMaxWorkers:        minPoolWorker,
					PoolMaxCapacity:       minPoolCapacity,
				},
			},
		},
		{
			name: "high_Retries_than_maxRetriesAllowed",
			args: args{
				cfg: &config{
					Endpoint:           "10.10.10.10/stat",
					PublishingInterval: 3,
					Retries:            100,
				},
			},
			want: want{
				err: nil,
				cfg: &config{
					Endpoint:              "10.10.10.10/stat",
					PublishingInterval:    3,
					DialTimeout:           minDialTimeout,
					KeepAliveDuration:     minKeepAliveDuration,
					MaxIdleConns:          0,
					MaxIdleConnsPerHost:   0,
					PublishingThreshold:   minPublishingThreshold,
					Retries:               5,
					retryInterval:         minRetryDuration,
					MaxChannelLength:      minChannelLength,
					ResponseHeaderTimeout: minResponseHeaderTimeout,
					PoolMaxWorkers:        minPoolWorker,
					PoolMaxCapacity:       minPoolCapacity,
				},
			},
		},
		{
			name: "valid_Retries_value",
			args: args{
				cfg: &config{
					Endpoint:           "10.10.10.10/stat",
					PublishingInterval: 3,
					Retries:            5,
				},
			},
			want: want{
				err: nil,
				cfg: &config{
					Endpoint:              "10.10.10.10/stat",
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.cfg.validate()
			assert.Equal(t, err, tt.want.err, "Mismatched error")
			assert.Equal(t, tt.args.cfg, tt.want.cfg, "Mismatched config")
		})
	}
}
