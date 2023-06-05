package stats

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {

	type args struct {
		cfg *Config
	}

	type want struct {
		err error
		cfg *Config
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "empty_host",
			args: args{
				cfg: &Config{
					Host: "",
				},
			},
			want: want{
				err: fmt.Errorf("stat server host and port cannot be empty"),
				cfg: &Config{
					Host: "",
				},
			},
		},
		{
			name: "empty_port",
			args: args{
				cfg: &Config{
					Host: "10.10.10.10",
					Port: "",
				},
			},
			want: want{
				err: fmt.Errorf("stat server host and port cannot be empty"),
				cfg: &Config{
					Host: "10.10.10.10",
					Port: "",
				},
			},
		},
		{
			name: "lower_values_than_min_limit",
			args: args{
				cfg: &Config{
					Host:                "10.10.10.10",
					Port:                "8000",
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
				cfg: &Config{
					Host:                "10.10.10.10",
					Port:                "8000",
					PublishingInterval:  minPublishingInterval,
					DialTimeout:         minDialTimeout,
					KeepAliveDuration:   minKeepAliveDuration,
					MaxIdleConns:        0,
					MaxIdleConnsPerHost: 0,
					PublishingThreshold: minPublishingThreshold,
				},
			},
		},
		{
			name: "high_PublishingInterval_than_max_limit",
			args: args{
				cfg: &Config{
					Host:               "10.10.10.10",
					Port:               "8000",
					PublishingInterval: 10,
				},
			},
			want: want{
				err: nil,
				cfg: &Config{
					Host:                "10.10.10.10",
					Port:                "8000",
					PublishingInterval:  maxPublishingInterval,
					DialTimeout:         minDialTimeout,
					KeepAliveDuration:   minKeepAliveDuration,
					MaxIdleConns:        0,
					MaxIdleConnsPerHost: 0,
					PublishingThreshold: minPublishingThreshold,
				},
			},
		},
		{
			name: "high_Retries_than_maxRetriesAllowed",
			args: args{
				cfg: &Config{
					Host:               "10.10.10.10",
					Port:               "8000",
					PublishingInterval: 3,
					Retries:            100,
				},
			},
			want: want{
				err: nil,
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
					retryInterval:       minRetryDuration,
				},
			},
		},
		{
			name: "valid_Retries_value",
			args: args{
				cfg: &Config{
					Host:               "10.10.10.10",
					Port:               "8000",
					PublishingInterval: 3,
					Retries:            5,
				},
			},
			want: want{
				err: nil,
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
