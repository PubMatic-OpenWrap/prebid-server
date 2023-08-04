package config

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	cfg "github.com/prebid/prebid-server/config"
	metrics_cfg "github.com/prebid/prebid-server/metrics/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	mock "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/stats"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/nbr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricsEngine(t *testing.T) {

	type args struct {
		owConfig        *config.Config
		metricsRegistry metrics_cfg.MetricsRegistry
		metricsCfg      *cfg.Metrics
	}
	type want struct {
		expectNilEngine  bool
		err              error
		metricsEngineCnt int
	}
	testCases := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid_configurations",
			args: args{
				owConfig: &config.Config{
					Stats: stats.Stats{
						Endpoint:    "http://example.com",
						UseHostName: true,
					},
				},
				metricsRegistry: metrics_cfg.MetricsRegistry{
					metrics_cfg.PrometheusRegistry: prometheus.NewRegistry(),
				},
				metricsCfg: &cfg.Metrics{
					Prometheus: cfg.PrometheusMetrics{
						Port:             14404,
						Namespace:        "ow",
						Subsystem:        "pbs",
						TimeoutMillisRaw: 10,
					},
				},
			},
			want: want{
				expectNilEngine:  false,
				err:              nil,
				metricsEngineCnt: 2,
			},
		},
		{
			name: "empty_stat_config_and_nil_metrics_config",
			args: args{
				owConfig: &config.Config{
					Stats: stats.Stats{
						Endpoint: "",
					},
				},
				metricsRegistry: metrics_cfg.MetricsRegistry{
					metrics_cfg.PrometheusRegistry: prometheus.NewRegistry(),
				},
				metricsCfg: nil,
			},
			want: want{
				expectNilEngine: true,
				err:             fmt.Errorf("metric-engine is not configured"),
			},
		},
		{
			name: "empty_stat_config_and_nil_metrics_registry",
			args: args{
				owConfig: &config.Config{
					Stats: stats.Stats{
						Endpoint: "",
					},
				},
				metricsRegistry: metrics_cfg.MetricsRegistry{
					metrics_cfg.PrometheusRegistry: nil,
				},
				metricsCfg: &cfg.Metrics{
					Prometheus: cfg.PrometheusMetrics{},
				},
			},
			want: want{
				expectNilEngine: true,
				err:             fmt.Errorf("metric-engine is not configured"),
			},
		},
		{
			name: "empty_stat_and_valid_metrics_cfg_and_registry",
			args: args{
				owConfig: &config.Config{
					Stats: stats.Stats{
						Endpoint: "",
					},
				},
				metricsRegistry: metrics_cfg.MetricsRegistry{
					metrics_cfg.PrometheusRegistry: prometheus.NewRegistry(),
				},
				metricsCfg: &cfg.Metrics{
					Prometheus: cfg.PrometheusMetrics{},
				},
			},
			want: want{
				expectNilEngine:  false,
				err:              nil,
				metricsEngineCnt: 1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualOutput, actualError := NewMetricsEngine(tc.args.owConfig, tc.args.metricsCfg, tc.args.metricsRegistry)
			assert.Equal(t, tc.want.expectNilEngine, actualOutput == nil)
			assert.Equal(t, tc.want.err, actualError)
			assert.Equal(t, tc.want.metricsEngineCnt, len(actualOutput))
		})
	}
}

func TestRecordFunctionForMultiMetricsEngine(t *testing.T) {

	ctrl := gomock.NewController(t)
	mockEngine := mock.NewMockMetricsEngine(ctrl)
	defer ctrl.Finish()

	// set the variables
	publisher := "5890"
	profile := "123"
	partner := "pubmatic"
	impCount := 1
	platform := "video"
	responseTime := 1
	endpoint := "in-app"
	versionID := "1"
	errorCode := 10
	processingTime := 10
	method := "GET"
	maxDuration := 20
	minDuration := 10
	aliasBidder := "pubmatic-2"
	adFormat := "banner"
	dealId := "pubdeal"
	host := "sv3:xyz1234"

	// set the expectations
	mockEngine.EXPECT().RecordOpenWrapServerPanicStats(host, method)
	mockEngine.EXPECT().RecordPublisherPartnerNoCookieStats(publisher, partner)
	mockEngine.EXPECT().RecordPartnerResponseErrors(publisher, partner, models.PartnerErrTimeout)
	mockEngine.EXPECT().RecordPartnerConfigErrors(publisher, profile, partner, models.PartnerErrSlotNotMapped)

	mockEngine.EXPECT().RecordPublisherProfileRequests(publisher, profile)
	mockEngine.EXPECT().RecordPublisherInvalidProfileImpressions(publisher, profile, impCount)
	mockEngine.EXPECT().RecordNobidErrPrebidServerRequests(publisher, nbr.AllPartnerThrottled)
	mockEngine.EXPECT().RecordNobidErrPrebidServerResponse(publisher)
	mockEngine.EXPECT().RecordInvalidCreativeStats(publisher, partner)
	mockEngine.EXPECT().RecordPlatformPublisherPartnerReqStats(platform, publisher, partner)
	mockEngine.EXPECT().RecordPlatformPublisherPartnerResponseStats(platform, publisher, partner)
	mockEngine.EXPECT().RecordPublisherResponseEncodingErrorStats(publisher)
	mockEngine.EXPECT().RecordPartnerResponseTimeStats(publisher, partner, responseTime)
	mockEngine.EXPECT().RecordPublisherResponseTimeStats(publisher, responseTime)
	mockEngine.EXPECT().RecordPublisherWrapperLoggerFailure(publisher, profile, versionID)
	mockEngine.EXPECT().RecordCacheErrorRequests(endpoint, publisher, profile)
	mockEngine.EXPECT().RecordPublisherInvalidProfileRequests(endpoint, publisher, profile)
	mockEngine.EXPECT().RecordBadRequests(endpoint, errorCode)
	mockEngine.EXPECT().RecordPrebidTimeoutRequests(publisher, profile)
	mockEngine.EXPECT().RecordSSTimeoutRequests(publisher, profile)
	mockEngine.EXPECT().RecordUidsCookieNotPresentErrorStats(publisher, profile)
	mockEngine.EXPECT().RecordVideoInstlImpsStats(publisher, profile)
	mockEngine.EXPECT().RecordImpDisabledViaConfigStats(adFormat, publisher, profile)
	mockEngine.EXPECT().RecordPreProcessingTimeStats(publisher, processingTime)
	mockEngine.EXPECT().RecordStatsKeyCTVPrebidFailedImpression(errorCode, publisher, profile)
	mockEngine.EXPECT().RecordCTVRequests(endpoint, platform)
	mockEngine.EXPECT().RecordPublisherRequests(endpoint, publisher, platform)
	mockEngine.EXPECT().RecordCTVHTTPMethodRequests(endpoint, publisher, method)
	mockEngine.EXPECT().RecordCTVInvalidReasonCount(errorCode, publisher)
	mockEngine.EXPECT().RecordCTVReqImpsWithDbConfigCount(publisher)
	mockEngine.EXPECT().RecordCTVReqImpsWithReqConfigCount(publisher)
	mockEngine.EXPECT().RecordAdPodGeneratedImpressionsCount(impCount, publisher)
	mockEngine.EXPECT().RecordRequestAdPodGeneratedImpressionsCount(impCount, publisher)
	mockEngine.EXPECT().RecordReqImpsWithContentCount(publisher, models.ContentTypeSite)
	mockEngine.EXPECT().RecordAdPodImpressionYield(maxDuration, minDuration, publisher)
	mockEngine.EXPECT().RecordCTVReqCountWithAdPod(publisher, profile)
	mockEngine.EXPECT().RecordPBSAuctionRequestsStats()
	mockEngine.EXPECT().RecordInjectTrackerErrorCount(adFormat, publisher, partner)
	mockEngine.EXPECT().RecordBidResponseByDealCountInPBS(publisher, profile, aliasBidder, dealId)
	mockEngine.EXPECT().RecordBidResponseByDealCountInHB(publisher, profile, aliasBidder, dealId)
	mockEngine.EXPECT().RecordPartnerTimeoutInPBS(publisher, profile, aliasBidder)
	mockEngine.EXPECT().RecordVideoImpDisabledViaConnTypeStats(publisher, profile)
	mockEngine.EXPECT().Shutdown()

	// create the multi-metric engine
	multiMetricEngine := MultiMetricsEngine{}
	multiMetricEngine = append(multiMetricEngine, mockEngine)

	// call the functions
	multiMetricEngine.RecordOpenWrapServerPanicStats(host, method)
	multiMetricEngine.RecordPublisherPartnerNoCookieStats(publisher, partner)
	multiMetricEngine.RecordPartnerResponseErrors(publisher, partner, models.PartnerErrTimeout)
	multiMetricEngine.RecordPartnerConfigErrors(publisher, profile, partner, models.PartnerErrSlotNotMapped)
	multiMetricEngine.RecordPublisherProfileRequests(publisher, profile)
	multiMetricEngine.RecordPublisherInvalidProfileImpressions(publisher, profile, impCount)
	multiMetricEngine.RecordNobidErrPrebidServerRequests(publisher, nbr.AllPartnerThrottled)
	multiMetricEngine.RecordNobidErrPrebidServerResponse(publisher)
	multiMetricEngine.RecordInvalidCreativeStats(publisher, partner)
	multiMetricEngine.RecordPlatformPublisherPartnerReqStats(platform, publisher, partner)
	multiMetricEngine.RecordPlatformPublisherPartnerResponseStats(platform, publisher, partner)
	multiMetricEngine.RecordPublisherResponseEncodingErrorStats(publisher)
	multiMetricEngine.RecordPartnerResponseTimeStats(publisher, partner, responseTime)
	multiMetricEngine.RecordPublisherResponseTimeStats(publisher, responseTime)
	multiMetricEngine.RecordPublisherWrapperLoggerFailure(publisher, profile, versionID)
	multiMetricEngine.RecordCacheErrorRequests(endpoint, publisher, profile)
	multiMetricEngine.RecordPublisherInvalidProfileRequests(endpoint, publisher, profile)
	multiMetricEngine.RecordBadRequests(endpoint, errorCode)
	multiMetricEngine.RecordPrebidTimeoutRequests(publisher, profile)
	multiMetricEngine.RecordSSTimeoutRequests(publisher, profile)
	multiMetricEngine.RecordUidsCookieNotPresentErrorStats(publisher, profile)
	multiMetricEngine.RecordVideoInstlImpsStats(publisher, profile)
	multiMetricEngine.RecordImpDisabledViaConfigStats(adFormat, publisher, profile)
	multiMetricEngine.RecordPreProcessingTimeStats(publisher, processingTime)
	multiMetricEngine.RecordStatsKeyCTVPrebidFailedImpression(errorCode, publisher, profile)
	multiMetricEngine.RecordCTVRequests(endpoint, platform)
	multiMetricEngine.RecordPublisherRequests(endpoint, publisher, platform)
	multiMetricEngine.RecordCTVHTTPMethodRequests(endpoint, publisher, method)
	multiMetricEngine.RecordCTVInvalidReasonCount(errorCode, publisher)
	multiMetricEngine.RecordCTVReqImpsWithDbConfigCount(publisher)
	multiMetricEngine.RecordCTVReqImpsWithReqConfigCount(publisher)
	multiMetricEngine.RecordAdPodGeneratedImpressionsCount(impCount, publisher)
	multiMetricEngine.RecordRequestAdPodGeneratedImpressionsCount(impCount, publisher)
	multiMetricEngine.RecordReqImpsWithContentCount(publisher, models.ContentTypeSite)
	multiMetricEngine.RecordAdPodImpressionYield(maxDuration, minDuration, publisher)
	multiMetricEngine.RecordCTVReqCountWithAdPod(publisher, profile)
	multiMetricEngine.RecordPBSAuctionRequestsStats()
	multiMetricEngine.RecordInjectTrackerErrorCount(adFormat, publisher, partner)
	multiMetricEngine.RecordBidResponseByDealCountInPBS(publisher, profile, aliasBidder, dealId)
	multiMetricEngine.RecordBidResponseByDealCountInHB(publisher, profile, aliasBidder, dealId)
	multiMetricEngine.RecordPartnerTimeoutInPBS(publisher, profile, aliasBidder)
	multiMetricEngine.RecordVideoImpDisabledViaConnTypeStats(publisher, profile)
	multiMetricEngine.Shutdown()
}
