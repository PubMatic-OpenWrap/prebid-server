package openwrap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prebid/prebid-server/v2/config"
	mock_metrics "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockMetrics := mock_metrics.NewMockMetricsEngine(ctrl)
	ow = &OpenWrap{}
	ow.metricEngine = mockMetrics
	type args struct {
		gatherer prometheus.Gatherer
		duration time.Duration
		setup    func()
		endpoint string
	}
	type testCase struct {
		name       string
		args       args
		expectCode int
	}

	tests := []testCase{
		{
			name: "valid test",
			args: args{
				gatherer: prometheus.Gatherers{},
				duration: 10 * time.Second,
				endpoint: "/abc",
				setup: func() {
					mockMetrics.EXPECT().RecordRequest(gomock.Any()).AnyTimes()
				},
			},
			expectCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.setup()
			handler := PrometheusHandler(tt.args.gatherer, tt.args.duration, tt.args.endpoint)
			server := &http.Server{
				Addr:    ":" + "8991",
				Handler: handler,
			}
			server.Shutdown(context.Background())
			req := httptest.NewRequest("GET", tt.args.endpoint, nil)
			rec := httptest.NewRecorder()
			server.Handler.ServeHTTP(rec, req)
			assert.Equal(t, rec.Code, tt.expectCode)

		})
	}
}

func TestInitPrometheusStatsEndpoint(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockMetrics := mock_metrics.NewMockMetricsEngine(ctrl)
	ow = &OpenWrap{}
	ow.metricEngine = mockMetrics
	type args struct {
		endpoint string
		cfg      *config.Configuration
		gatherer *prometheus.Registry
		setup    func()
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
				setup: func() {
					mockMetrics.EXPECT().RecordRequest(gomock.Any()).AnyTimes()
				},
				promMux: http.NewServeMux(),
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
				setup: func() {
					mockMetrics.EXPECT().RecordRequest(gomock.Any()).AnyTimes()
				},
				promMux: http.NewServeMux(),
			},

			want: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.setup()
			initPrometheusStatsEndpoint(tt.args.cfg, tt.args.gatherer, tt.args.promMux)
			req := httptest.NewRequest("GET", tt.args.endpoint, nil)
			rec := httptest.NewRecorder()
			tt.args.promMux.ServeHTTP(rec, req)
			assert.Equal(t, rec.Code, tt.want)
		})
	}
}

func TestInitPrometheusMetricsEndpoint(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockMetrics := mock_metrics.NewMockMetricsEngine(ctrl)
	ow = &OpenWrap{}
	ow.metricEngine = mockMetrics
	type args struct {
		endpoint string
		cfg      *config.Configuration
		setup    func()
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
				setup: func() {
					mockMetrics.EXPECT().RecordRequest(gomock.Any()).AnyTimes()
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
				setup: func() {
					mockMetrics.EXPECT().RecordRequest(gomock.Any()).AnyTimes()
				},
				promMux: http.NewServeMux(),
			},
			want: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.setup()
			initPrometheusMetricsEndpoint(tt.args.cfg, tt.args.promMux)
			req := httptest.NewRequest("GET", tt.args.endpoint, nil)
			rec := httptest.NewRecorder()
			tt.args.promMux.ServeHTTP(rec, req)
			assert.Equal(t, rec.Code, tt.want)
		})
	}
}

func TestGetOpenWrapPrometheusServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockMetrics := mock_metrics.NewMockMetricsEngine(ctrl)
	ow = &OpenWrap{}
	ow.metricEngine = mockMetrics
	type args struct {
		endpoint string
		cfg      *config.Configuration
		setup    func()
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
				setup: func() {
					mockMetrics.EXPECT().RecordRequest(gomock.Any()).AnyTimes()
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
				setup: func() {
					mockMetrics.EXPECT().RecordRequest(gomock.Any()).AnyTimes()
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
				setup: func() {
					mockMetrics.EXPECT().RecordRequest(gomock.Any()).AnyTimes()
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
				setup: func() {
					mockMetrics.EXPECT().RecordRequest(gomock.Any()).AnyTimes()
				},
				gatherer: prometheus.NewRegistry(),
			},
			want: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.setup()
			promServer := GetOpenWrapPrometheusServer(tt.args.cfg, tt.args.gatherer, server)
			defer promServer.Shutdown(context.Background())
			req := httptest.NewRequest("GET", tt.args.endpoint, nil)
			rec := httptest.NewRecorder()
			promServer.Handler.ServeHTTP(rec, req)
			assert.Equal(t, rec.Code, tt.want)
		})
	}
}
