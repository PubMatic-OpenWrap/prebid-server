package router

import (
	"errors"
	"strconv"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/prebid/openrtb/v17/openrtb3"
	"github.com/prebid/prebid-server/analytics"
	analyticsConf "github.com/prebid/prebid-server/analytics/config"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/currency"
	"github.com/prebid/prebid-server/metrics"
	metricsConf "github.com/prebid/prebid-server/metrics/config"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	originalSchemaDirectory := schemaDirectory
	originalinfoDirectory := infoDirectory
	defer func() {
		schemaDirectory = originalSchemaDirectory
		infoDirectory = originalinfoDirectory
	}()
	schemaDirectory = "../static/bidder-params"
	infoDirectory = "../static/bidder-info"

	type args struct {
		cfg           *config.Configuration
		rateConvertor *currency.RateConverter
	}
	tests := []struct {
		name    string
		args    args
		wantR   *Router
		wantErr bool
		setup   func()
	}{
		{
			name: "Happy path",
			args: args{
				cfg:           &config.Configuration{},
				rateConvertor: &currency.RateConverter{},
			},
			wantR:   &Router{Router: &httprouter.Router{}},
			wantErr: false,
			setup: func() {
				g_syncers = nil
				g_cfg = nil
				g_ex = nil
				g_accounts = nil
				g_paramsValidator = nil
				g_storedReqFetcher = nil
				g_storedRespFetcher = nil
				g_metrics = nil
				g_analytics = nil
				g_disabledBidders = nil
				g_videoFetcher = nil
				g_activeBidders = nil
				g_defReqJSON = nil
				g_cacheClient = nil
				g_transport = nil
				g_gdprPermsBuilder = nil
				g_tcf2CfgBuilder = nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := New(tt.args.cfg, tt.args.rateConvertor)
			assert.Equal(t, tt.wantErr, err != nil, err)

			assert.NotNil(t, g_syncers)
			assert.NotNil(t, g_cfg)
			assert.NotNil(t, g_ex)
			assert.NotNil(t, g_accounts)
			assert.NotNil(t, g_paramsValidator)
			assert.NotNil(t, g_storedReqFetcher)
			assert.NotNil(t, g_storedRespFetcher)
			assert.NotNil(t, g_metrics)
			assert.NotNil(t, g_analytics)
			assert.NotNil(t, g_disabledBidders)
			assert.NotNil(t, g_videoFetcher)
			assert.NotNil(t, g_activeBidders)
			assert.NotNil(t, g_defReqJSON)
			assert.NotNil(t, g_cacheClient)
			assert.NotNil(t, g_transport)
			assert.NotNil(t, g_gdprPermsBuilder)
			assert.NotNil(t, g_tcf2CfgBuilder)
		})
	}
}

type mockAnalytics []analytics.PBSAnalyticsModule

func (m mockAnalytics) LogAuctionObject(a *analytics.AuctionObject)               {}
func (m mockAnalytics) LogVideoObject(a *analytics.VideoObject)                   {}
func (m mockAnalytics) LogCookieSyncObject(a *analytics.CookieSyncObject)         {}
func (m mockAnalytics) LogSetUIDObject(a *analytics.SetUIDObject)                 {}
func (m mockAnalytics) LogAmpObject(a *analytics.AmpObject)                       {}
func (m mockAnalytics) LogNotificationEventObject(a *analytics.NotificationEvent) {}

func TestRegisterAnalyticsModule(t *testing.T) {

	type args struct {
		modules     []analytics.PBSAnalyticsModule
		g_analytics *analytics.PBSAnalyticsModule
	}

	type want struct {
		err               error
		registeredModules int
	}

	tests := []struct {
		description string
		arg         args
		want        want
	}{
		{
			description: "error if nil module",
			arg: args{
				modules:     []analytics.PBSAnalyticsModule{nil},
				g_analytics: new(analytics.PBSAnalyticsModule),
			},
			want: want{
				registeredModules: 0,
				err:               errors.New("module to be added is nil"),
			},
		},
		{
			description: "register valid module",
			arg: args{
				modules:     []analytics.PBSAnalyticsModule{&mockAnalytics{}, &mockAnalytics{}},
				g_analytics: new(analytics.PBSAnalyticsModule),
			},
			want: want{
				err:               nil,
				registeredModules: 2,
			},
		},
		{
			description: "error if g_analytics is nil",
			arg: args{
				modules:     []analytics.PBSAnalyticsModule{&mockAnalytics{}, &mockAnalytics{}},
				g_analytics: nil,
			},
			want: want{
				err:               errors.New("g_analytics is nil"),
				registeredModules: 0,
			},
		},
	}

	for _, tt := range tests {
		g_analytics = tt.arg.g_analytics
		analyticsConf.EnableAnalyticsModule = func(module, moduleList analytics.PBSAnalyticsModule) (analytics.PBSAnalyticsModule, error) {
			if tt.want.err == nil {
				modules, _ := moduleList.(mockAnalytics)
				modules = append(modules, module)
				return modules, nil
			}
			return nil, tt.want.err
		}
		for _, m := range tt.arg.modules {
			err := RegisterAnalyticsModule(m)
			assert.Equal(t, err, tt.want.err)
		}
		if g_analytics != nil {
			// cast g_analytics to mock analytics
			tmp, _ := (*g_analytics).(mockAnalytics)
			assert.Equal(t, tt.want.registeredModules, len(tmp))
		}
	}
}

func TestCallRecordRejectedBids(t *testing.T) {
	metricEngine := g_metrics
	defer func() {
		g_metrics = metricEngine
	}()

	type args struct {
		pubid, bidder, code string
		metricEngine        metrics.MetricsEngine
	}

	type want struct {
		expectToGetRecord bool
		bidderLossCount   map[string]map[openrtb3.LossReason]float64
	}

	tests := []struct {
		description string
		arg         args
		want        want
	}{
		{
			description: "nil g_metric",
			arg: args{
				metricEngine: nil,
			},
			want: want{
				expectToGetRecord: false,
			},
		},
		{
			description: "non-nil g_metric",
			arg: args{
				metricEngine: metricsConf.NewMetricsEngine(&config.Configuration{
					Metrics: config.Metrics{Prometheus: config.PrometheusMetrics{Port: 1}},
				}, nil, nil),
				pubid:  "11",
				bidder: "Pubmatic",
				code:   "102",
			},
			want: want{
				expectToGetRecord: true,
				bidderLossCount: map[string]map[openrtb3.LossReason]float64{
					"Pubmatic": map[openrtb3.LossReason]float64{
						openrtb3.LossLostToHigherBid: 1,
					},
				},
			},
		},
	}

	for _, test := range tests {
		g_metrics = test.arg.metricEngine
		CallRecordRejectedBids(test.arg.pubid, test.arg.bidder, test.arg.code)

		detailedEngine, ok := g_metrics.(*metricsConf.DetailedMetricsEngine)
		if !ok {
			if test.want.expectToGetRecord {
				t.Errorf("Failed to get metric-engine for test case - [%s]", test.description)
			}
			continue
		}
		metricFamilies, _ := detailedEngine.PrometheusMetrics.Gatherer.Gather()
		isRecorded := false

		for _, metricFamily := range metricFamilies {
			if metricFamily.GetName() == "rejected_bids" {
				for _, metric := range metricFamily.GetMetric() {
					counter := metric.GetCounter().Value
					current_bidder := ""
					current_code := 0

					// verify labels
					for _, label := range metric.GetLabel() {
						switch *label.Name {
						case "pubid":
							if *label.Value != test.arg.pubid {
								t.Errorf("Expected pubid=[%s], got- [%s]", test.arg.pubid, *label.Value)
							}
						case "bidder":
							current_bidder = *label.Value
						case "code":
							current_code, _ = strconv.Atoi(*label.Value)
						default:
							t.Errorf("Unexpected label %s found in metric", *label.Name)
						}
					}
					lossCount := test.want.bidderLossCount[current_bidder]

					// verify counter value
					if *counter != lossCount[openrtb3.LossReason(current_code)] {
						t.Errorf("Counter value mismatch for bidder- [%s], code - [%d], expected - [%f], got - [%f]", current_bidder, current_code, lossCount[openrtb3.LossReason(current_code)], *counter)
					}
					isRecorded = true
				}
			}
		}
		// verify if metric got recorded by metric-engine.
		if test.want.expectToGetRecord != isRecorded {
			t.Errorf("Failed to record rejected_bids for test case - [%s]", test.description)
		}
	}
}
