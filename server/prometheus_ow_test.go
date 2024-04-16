package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/prebid/prebid-server/config"
	"github.com/prometheus/client_golang/prometheus"
)

func TestInitPrometheusStatsEndpoint(t *testing.T) {
	type args struct {
		endpoint string
		cfg      *config.Configuration
		gatherer *prometheus.Registry
		promMux  *http.ServeMux
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "valid request for /stats endpoint",
			args: args{
				endpoint: "/stats",
				cfg: &config.Configuration{
					Metrics: config.Metrics{
						Prometheus: config.PrometheusMetrics{
							Port:             9090,
							TimeoutMillisRaw: 12,
						},
					},
				},
				gatherer: &prometheus.Registry{},
				promMux:  http.NewServeMux(),
			},

			want: http.StatusOK,
		},
		{
			name: "invalid request for /stats endpoint",
			args: args{
				endpoint: "/stat",
				cfg: &config.Configuration{
					Metrics: config.Metrics{
						Prometheus: config.PrometheusMetrics{
							Port:             9090,
							TimeoutMillisRaw: 12,
						},
					},
				},
				gatherer: &prometheus.Registry{},
				promMux:  http.NewServeMux(),
			},

			want: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initPrometheusStatsEndpoint(tt.args.cfg, tt.args.gatherer, tt.args.promMux)
			req := httptest.NewRequest("GET", tt.args.endpoint, nil)
			rec := httptest.NewRecorder()
			tt.args.promMux.ServeHTTP(rec, req)
			assert.Equal(t, rec.Code, tt.want)
		})
	}
}

func TestInitPrometheusMetricsEndpoint(t *testing.T) {
	type args struct {
		endpoint string
		cfg      *config.Configuration
		promMux  *http.ServeMux
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "valid request for /metrics endpoint",
			args: args{
				endpoint: "/metrics",
				cfg: &config.Configuration{
					Metrics: config.Metrics{
						Prometheus: config.PrometheusMetrics{
							Port:             9091,
							TimeoutMillisRaw: 12,
						},
					},
				},
				promMux: http.NewServeMux(),
			},
			want: http.StatusOK,
		},
		{
			name: "invalid request for /metrics endpoint",
			args: args{
				endpoint: "/endpoint",
				cfg: &config.Configuration{
					Metrics: config.Metrics{
						Prometheus: config.PrometheusMetrics{
							Port:             9091,
							TimeoutMillisRaw: 12,
						},
					},
				},
				promMux: http.NewServeMux(),
			},
			want: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initPrometheusMetricsEndpoint(tt.args.cfg, tt.args.promMux)
			req := httptest.NewRequest("GET", tt.args.endpoint, nil)
			rec := httptest.NewRecorder()
			tt.args.promMux.ServeHTTP(rec, req)
			assert.Equal(t, rec.Code, tt.want)
		})
	}
}

func TestGetOpenWrapPrometheusServer(t *testing.T) {
	type args struct {
		endpoint string
		cfg      *config.Configuration
		gatherer *prometheus.Registry
	}

	server := &http.Server{
		Addr: ":" + "9092",
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "valid request for /stats endpoint",
			args: args{
				endpoint: "/stats",
				cfg: &config.Configuration{
					Metrics: config.Metrics{
						Prometheus: config.PrometheusMetrics{
							Port:             9092,
							TimeoutMillisRaw: 12,
						},
					},
				},
				gatherer: prometheus.NewRegistry(),
			},
			want: http.StatusOK,
		},
		{
			name: "valid request for /metrics endpoint",
			args: args{
				endpoint: "/stats",
				cfg: &config.Configuration{
					Metrics: config.Metrics{
						Prometheus: config.PrometheusMetrics{
							Port:             9092,
							TimeoutMillisRaw: 12,
						},
					},
				},
				gatherer: prometheus.NewRegistry(),
			},
			want: http.StatusOK,
		},
		{
			name: "invalid request for /stats endpoint",
			args: args{
				endpoint: "/abc",
				cfg: &config.Configuration{
					Metrics: config.Metrics{
						Prometheus: config.PrometheusMetrics{
							Port:             9092,
							TimeoutMillisRaw: 12,
						},
					},
				},
				gatherer: prometheus.NewRegistry(),
			},
			want: http.StatusNotFound,
		},
		{
			name: "valid request for /metrics endpoint",
			args: args{
				endpoint: "/xyz",
				cfg: &config.Configuration{
					Metrics: config.Metrics{
						Prometheus: config.PrometheusMetrics{
							Port:             9092,
							TimeoutMillisRaw: 12,
						},
					},
				},
				gatherer: prometheus.NewRegistry(),
			},
			want: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promServer := getOpenWrapPrometheusServer(tt.args.cfg, tt.args.gatherer, server)
			defer promServer.Shutdown(context.Background())
			req := httptest.NewRequest("GET", tt.args.endpoint, nil)
			rec := httptest.NewRecorder()
			promServer.Handler.ServeHTTP(rec, req)
			assert.Equal(t, rec.Code, tt.want)
		})
	}
}
