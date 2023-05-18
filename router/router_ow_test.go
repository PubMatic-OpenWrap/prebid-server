package router

import (
	"errors"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/analytics"
	analyticsConf "github.com/prebid/prebid-server/analytics/config"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/currency"
	"github.com/prebid/prebid-server/floors"
	"github.com/prebid/prebid-server/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		floorFetcher  *floors.PriceFloorFetcher
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
				floorFetcher:  &floors.PriceFloorFetcher{},
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
		pubid, bidder string
		code          openrtb3.NonBidStatusCode
	}

	type want struct {
		expectToGetRecord bool
	}

	tests := []struct {
		description string
		args        args
		want        want
	}{
		{
			description: "nil g_metric",
			args:        args{},
			want: want{
				expectToGetRecord: false,
			},
		},
		{
			description: "non-nil g_metric",
			args: args{
				pubid:  "11",
				bidder: "Pubmatic",
				code:   102,
			},
			want: want{
				expectToGetRecord: true,
			},
		},
	}

	for _, test := range tests {

		metricsMock := &metrics.MetricsEngineMock{}
		if test.want.expectToGetRecord {
			metricsMock.Mock.On("RecordRejectedBids", mock.Anything, mock.Anything, mock.Anything).Return()
		}
		if test.description != "nil g_metric" {
			g_metrics = metricsMock
		}
		// CallRecordRejectedBids will panic if g_metrics is non-nil and if there is no call to RecordRejectedBids
		CallRecordNonBids(test.args.pubid, test.args.bidder, test.args.code)
	}
}
