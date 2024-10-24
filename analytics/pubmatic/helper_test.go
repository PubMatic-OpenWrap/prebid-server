package pubmatic

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/analytics"
	"github.com/prebid/prebid-server/v2/analytics/pubmatic/mhttp"
	mock_mhttp "github.com/prebid/prebid-server/v2/analytics/pubmatic/mhttp/mock"
	mock_metrics "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/wakanda"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func TestPrepareLoggerURL(t *testing.T) {
	type args struct {
		wlog        *WloggerRecord
		loggerURL   string
		gdprEnabled int
	}
	tests := []struct {
		name     string
		args     args
		owlogger string
	}{
		{
			name: "nil_wlog",
			args: args{
				wlog:        nil,
				loggerURL:   "http://t.pubmatic.com/wl",
				gdprEnabled: 1,
			},
			owlogger: "",
		},
		{
			name: "gdprEnabled=1",
			args: args{
				wlog: &WloggerRecord{
					record: record{
						PubID:     10,
						ProfileID: "1",
						VersionID: "0",
					},
				},
				loggerURL:   "http://t.pubmatic.com/wl",
				gdprEnabled: 1,
			},
			owlogger: `http://t.pubmatic.com/wl?gdEn=1&json={"pubid":10,"pid":"1","pdvid":"0","dvc":{},"ft":0}&pubid=10`,
		},
		{
			name: "gdprEnabled=0",
			args: args{
				wlog: &WloggerRecord{
					record: record{
						PubID:            10,
						ProfileID:        "1",
						VersionID:        "0",
						CustomDimensions: "age=23;traffic=media",
					},
				},
				loggerURL:   "http://t.pubmatic.com/wl",
				gdprEnabled: 0,
			},
			owlogger: `http://t.pubmatic.com/wl?json={"pubid":10,"pid":"1","pdvid":"0","dvc":{},"ft":0,"cds":"age=23;traffic=media"}&pubid=10`,
		},
		{
			name: "private endpoint",
			args: args{
				wlog: &WloggerRecord{
					record: record{
						PubID:            5,
						ProfileID:        "5",
						VersionID:        "1",
						CustomDimensions: "age=23;traffic=media",
					},
				},
				loggerURL:   "http://10.172.141.11/wl",
				gdprEnabled: 0,
			},
			owlogger: `http://10.172.141.11/wl?json={"pubid":5,"pid":"5","pdvid":"1","dvc":{},"ft":0,"cds":"age=23;traffic=media"}&pubid=5`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owlogger := PrepareLoggerURL(tt.args.wlog, tt.args.loggerURL, tt.args.gdprEnabled)
			decodedOwlogger, _ := url.QueryUnescape(owlogger)
			assert.Equal(t, tt.owlogger, decodedOwlogger, tt.name)
		})
	}
}
func TestGetGdprEnabledFlag(t *testing.T) {
	tests := []struct {
		name          string
		partnerConfig map[int]map[string]string
		gdprFlag      int
	}{
		{
			name:          "Empty partnerConfig",
			partnerConfig: make(map[int]map[string]string),
			gdprFlag:      0,
		},
		{
			name: "partnerConfig without versionlevel cfg",
			partnerConfig: map[int]map[string]string{
				2: {models.GDPR_ENABLED: "1"},
			},
			gdprFlag: 0,
		},
		{
			name: "partnerConfig without GDPR_ENABLED",
			partnerConfig: map[int]map[string]string{
				models.VersionLevelConfigID: {"any": "1"},
			},
			gdprFlag: 0,
		},
		{
			name: "partnerConfig with invalid GDPR_ENABLED",
			partnerConfig: map[int]map[string]string{
				models.VersionLevelConfigID: {models.GDPR_ENABLED: "non-int"},
			},
			gdprFlag: 0,
		},
		{
			name: "partnerConfig with GDPR_ENABLED=1",
			partnerConfig: map[int]map[string]string{
				models.VersionLevelConfigID: {models.GDPR_ENABLED: "1"},
			},
			gdprFlag: 1,
		},
		{
			name: "partnerConfig with GDPR_ENABLED=2",
			partnerConfig: map[int]map[string]string{
				models.VersionLevelConfigID: {models.GDPR_ENABLED: "2"},
			},
			gdprFlag: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gdprFlag := getGdprEnabledFlag(tt.partnerConfig)
			assert.Equal(t, tt.gdprFlag, gdprFlag, tt.name)
		})
	}
}
func TestSendMethod(t *testing.T) {
	// initialise global variables
	mhttp.Init(1, 1, 1, 2000)
	// init mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		rctx    *models.RequestCtx
		url     string
		headers http.Header
	}
	tests := []struct {
		name                    string
		args                    args
		getMetricsEngine        func() *mock_metrics.MockMetricsEngine
		getMockMultiHttpContext func() *mock_mhttp.MockMultiHttpContextInterface
	}{
		{
			name: "send success",
			args: args{
				rctx: &models.RequestCtx{
					PubIDStr:     "5890",
					ProfileIDStr: "1",
					Endpoint:     models.EndpointV25,
				},
				url: "http://10.172.11.11/wl",
				headers: http.Header{
					"key": []string{"val"},
				},
			},
			getMetricsEngine: func() *mock_metrics.MockMetricsEngine {
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				mockEngine.EXPECT().RecordSendLoggerDataTime(gomock.Any())
				return mockEngine
			},
			getMockMultiHttpContext: func() *mock_mhttp.MockMultiHttpContextInterface {
				mockHttpCtx := mock_mhttp.NewMockMultiHttpContextInterface(ctrl)
				mockHttpCtx.EXPECT().AddHttpCall(gomock.Any())
				mockHttpCtx.EXPECT().Execute().Return(0, 0)
				return mockHttpCtx
			},
		},
		{
			name: "send fail",
			args: args{
				rctx: &models.RequestCtx{
					PubIDStr:      "5890",
					ProfileIDStr:  "1",
					Endpoint:      models.EndpointV25,
					KADUSERCookie: &http.Cookie{},
				},
				url: "http://10.172.11.11/wl",
				headers: http.Header{
					"key": []string{"val"},
				},
			},
			getMetricsEngine: func() *mock_metrics.MockMetricsEngine {
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				mockEngine.EXPECT().RecordPublisherWrapperLoggerFailure("5890", "")
				return mockEngine
			},
			getMockMultiHttpContext: func() *mock_mhttp.MockMultiHttpContextInterface {
				mockHttpCtx := mock_mhttp.NewMockMultiHttpContextInterface(ctrl)
				mockHttpCtx.EXPECT().AddHttpCall(gomock.Any())
				mockHttpCtx.EXPECT().Execute().Return(0, 1)
				return mockHttpCtx
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.rctx.MetricsEngine = tt.getMetricsEngine()
			send(tt.args.rctx, tt.args.url, tt.args.headers, tt.getMockMultiHttpContext())
		})
	}
}

func TestRestoreBidResponse(t *testing.T) {
	type args struct {
		ao   analytics.AuctionObject
		rctx *models.RequestCtx
	}
	tests := []struct {
		name    string
		args    args
		want    *openrtb2.BidResponse
		wantErr string
	}{
		{
			name: "Endpoint is not AppLovinMax",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						ID: "test-case-1",
					},
				},
				rctx: &models.RequestCtx{
					Endpoint: models.EndpointV25,
				},
			},
			want: &openrtb2.BidResponse{
				ID: "test-case-1",
			},
		},
		{
			name: "NBR is not nil",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						ID:  "test-case-1",
						NBR: ptrutil.ToPtr(nbr.InvalidProfileConfiguration),
					},
				},
				rctx: &models.RequestCtx{
					Endpoint: models.EndpointAppLovinMax,
				},
			},
			want: &openrtb2.BidResponse{
				ID:  "test-case-1",
				NBR: ptrutil.ToPtr(nbr.InvalidProfileConfiguration),
			},
		},
		{
			name: "failed to unmarshal BidResponse.SeatBid[0].Bid[0].Ext",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						ID: "test-case-1",
						SeatBid: []openrtb2.SeatBid{
							{
								Bid: []openrtb2.Bid{
									{
										ID:  "123",
										Ext: json.RawMessage(`{`),
									},
								},
							},
						},
					},
				},
				rctx: &models.RequestCtx{
					Endpoint: models.EndpointAppLovinMax,
				},
			},
			want: &openrtb2.BidResponse{
				ID: "test-case-1",
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:  "123",
								Ext: json.RawMessage(`{`),
							},
						},
					},
				},
			},
			wantErr: "unexpected end of JSON input",
		},
		{
			name: "signaldata not present in ext",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						ID: "test-case-1",
						SeatBid: []openrtb2.SeatBid{
							{
								Bid: []openrtb2.Bid{
									{
										ID:  "123",
										Ext: json.RawMessage(`{"signalData1": "{\"matchedimpression\":{\"appnexus\":50,\"pubmatic\":50}}\r\n"}`),
									},
								},
							},
						},
					},
				},
				rctx: &models.RequestCtx{
					Endpoint: models.EndpointAppLovinMax,
				},
			},
			want: &openrtb2.BidResponse{
				ID: "test-case-1",
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:  "123",
								Ext: json.RawMessage(`{"signalData1": "{\"matchedimpression\":{\"appnexus\":50,\"pubmatic\":50}}\r\n"}`),
							},
						},
					},
				},
			},
			wantErr: "signal data not found in the response",
		},
		{
			name: "failed to unmarshal signaldata",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						ID:    "123",
						BidID: "bid-id-1",
						Cur:   "USD",
						SeatBid: []openrtb2.SeatBid{
							{
								Seat: "pubmatic",
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp_1",
										Ext:   json.RawMessage(`{"signaldata": "{"id":123}"}"`),
									},
								},
							},
						},
					},
				},
				rctx: &models.RequestCtx{
					Endpoint: models.EndpointAppLovinMax,
				},
			},
			want: &openrtb2.BidResponse{
				ID:    "123",
				BidID: "bid-id-1",
				Cur:   "USD",
				SeatBid: []openrtb2.SeatBid{
					{
						Seat: "pubmatic",
						Bid: []openrtb2.Bid{
							{
								ID:    "bid-id-1",
								ImpID: "imp_1",
								Ext:   json.RawMessage(`{"signaldata": "{"id":123}"}"`),
							},
						},
					},
				},
			},
			wantErr: `invalid character 'i' after object key:value pair`,
		},
		{
			name: "valid AppLovinMax Response",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						ID:    "123",
						BidID: "bid-id-1",
						Cur:   "USD",
						SeatBid: []openrtb2.SeatBid{
							{
								Seat: "pubmatic",
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp_1",
										Ext:   json.RawMessage(`{"signaldata":"{\"id\":\"123\",\"seatbid\":[{\"bid\":[{\"id\":\"bid-id-1\",\"impid\":\"imp_1\",\"price\":0}],\"seat\":\"pubmatic\"}],\"bidid\":\"bid-id-1\",\"cur\":\"USD\",\"ext\":{\"matchedimpression\":{\"appnexus\":50,\"pubmatic\":50}}}\r\n"}`),
									},
								},
							},
						},
					},
				},
				rctx: &models.RequestCtx{
					Endpoint: models.EndpointAppLovinMax,
				},
			},
			want: &openrtb2.BidResponse{
				ID:    "123",
				BidID: "bid-id-1",
				Cur:   "USD",
				SeatBid: []openrtb2.SeatBid{
					{
						Seat: "pubmatic",
						Bid: []openrtb2.Bid{
							{
								ID:    "bid-id-1",
								ImpID: "imp_1",
							},
						},
					},
				},
				Ext: json.RawMessage(`{"matchedimpression":{"appnexus":50,"pubmatic":50}}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RestoreBidResponse(tt.args.rctx, tt.args.ao)
			if err != nil {
				assert.Equal(t, tt.wantErr, err.Error(), tt.name)
			}
			assert.Equal(t, tt.want, tt.args.ao.Response, tt.name)
		})
	}
}

func TestWloggerRecord_logProfileMetaData(t *testing.T) {
	type fields struct {
		record record
	}
	type args struct {
		rctx *models.RequestCtx
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantRecord record
	}{
		{
			name: "all feilds are empty",
			args: args{
				rctx: &models.RequestCtx{},
			},
			fields: fields{
				record: record{},
			},
			wantRecord: record{},
		},
		{
			name: "all feilds are set",
			args: args{
				rctx: &models.RequestCtx{
					ProfileType:           1,
					ProfileTypePlatform:   2,
					AppPlatform:           3,
					AppIntegrationPath:    ptrutil.ToPtr(4),
					AppSubIntegrationPath: ptrutil.ToPtr(5),
				},
			},
			fields: fields{
				record: record{},
			},
			wantRecord: record{
				ProfileType:           1,
				ProfileTypePlatform:   2,
				AppPlatform:           3,
				AppIntegrationPath:    ptrutil.ToPtr(4),
				AppSubIntegrationPath: ptrutil.ToPtr(5),
			},
		},
		{
			name: "appIntegrationPath and appSubIntegrationPath are nil",
			args: args{
				rctx: &models.RequestCtx{
					ProfileType:           1,
					ProfileTypePlatform:   2,
					AppPlatform:           3,
					AppIntegrationPath:    nil,
					AppSubIntegrationPath: nil,
				},
			},
			fields: fields{
				record: record{},
			},
			wantRecord: record{
				ProfileType:           1,
				ProfileTypePlatform:   2,
				AppPlatform:           3,
				AppIntegrationPath:    nil,
				AppSubIntegrationPath: nil,
			},
		},
		{
			name: "appIntegrationPath and appSubIntegrationPath are not nil but less than 0",
			args: args{
				rctx: &models.RequestCtx{
					ProfileType:           1,
					ProfileTypePlatform:   2,
					AppPlatform:           3,
					AppIntegrationPath:    ptrutil.ToPtr(-1),
					AppSubIntegrationPath: ptrutil.ToPtr(-1),
				},
			},
			fields: fields{
				record: record{},
			},
			wantRecord: record{
				ProfileType:           1,
				ProfileTypePlatform:   2,
				AppPlatform:           3,
				AppIntegrationPath:    nil,
				AppSubIntegrationPath: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wlog := &WloggerRecord{
				record: tt.fields.record,
			}
			wlog.logProfileMetaData(tt.args.rctx)
			assert.Equal(t, tt.wantRecord, wlog.record, tt.name)
		})
	}
}

func TestSetWakandaWinningBidFlag(t *testing.T) {
	type args struct {
		wakandaDebug wakanda.WakandaDebug
		response     *openrtb2.BidResponse
	}
	tests := []struct {
		name string
		args args
		want wakanda.WakandaDebug
	}{
		{
			name: "all_empty_parameters",
			args: args{},
			want: nil,
		},
		{
			name: "only_wakanda_empty",
			args: args{
				wakandaDebug: nil,
				response: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									Price: 5,
								},
							},
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "only_response_empty",
			args: args{
				wakandaDebug: &wakanda.Debug{},
				response:     nil,
			},
			want: &wakanda.Debug{},
		},
		{
			name: "no_seatbid",
			args: args{
				wakandaDebug: &wakanda.Debug{},
				response:     &openrtb2.BidResponse{},
			},
			want: &wakanda.Debug{},
		},
		{
			name: "no_bid",
			args: args{
				wakandaDebug: &wakanda.Debug{},
				response: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{},
					},
				},
			},
			want: &wakanda.Debug{},
		},
		{
			name: "no_price",
			args: args{
				wakandaDebug: &wakanda.Debug{},
				response: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{},
							},
						},
					},
				},
			},
			want: &wakanda.Debug{},
		},
		{
			name: "non_zero_price",
			args: args{
				wakandaDebug: &wakanda.Debug{},
				response: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									Price: 5,
								},
							},
						},
					},
				},
			},
			want: &wakanda.Debug{DebugData: wakanda.DebugData{WinningBid: true}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setWakandaWinningBidFlag(tt.args.wakandaDebug, tt.args.response)
			assert.Equal(t, tt.want, tt.args.wakandaDebug)
		})
	}
}

func TestSetWakandaObject(t *testing.T) {
	type args struct {
		rCtx      *models.RequestCtx
		ao        *analytics.AuctionObject
		loggerURL string
	}
	testCases := []struct {
		name string
		args args
		want *models.RequestCtx
	}{
		{
			name: "rctx is empty",
			args: args{
				rCtx:      &models.RequestCtx{},
				ao:        &analytics.AuctionObject{},
				loggerURL: "",
			},
			want: &models.RequestCtx{},
		},
		{
			name: "wakanda is disabled",
			args: args{
				rCtx:      &models.RequestCtx{WakandaDebug: &wakanda.Debug{Enabled: false}},
				ao:        &analytics.AuctionObject{},
				loggerURL: "",
			},
			want: &models.RequestCtx{WakandaDebug: &wakanda.Debug{Enabled: false}},
		},
		{
			name: "wakanda is enabled",
			args: args{
				rCtx: &models.RequestCtx{WakandaDebug: &wakanda.Debug{Enabled: true}},
				ao: &analytics.AuctionObject{
					RequestWrapper: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{},
					},
					Response: &openrtb2.BidResponse{},
				},
				loggerURL: "",
			},
			want: &models.RequestCtx{WakandaDebug: &wakanda.Debug{Enabled: true, DebugData: wakanda.DebugData{WinningBid: false, HTTPResponseBody: "{\"id\":\"\"}", OpenRTB: &openrtb2.BidRequest{}, Logger: json.RawMessage{}}}},
		},
		{
			name: "wakanda enabled with valid flow",
			args: args{
				rCtx: &models.RequestCtx{WakandaDebug: &wakanda.Debug{Enabled: true}},
				ao: &analytics.AuctionObject{
					RequestWrapper: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{
							Imp: []openrtb2.Imp{
								{
									ID: "imp_1",
								},
							},
						},
					},
					Response: &openrtb2.BidResponse{
						ID:    "123",
						BidID: "bid-id-1",
						Cur:   "USD",
						SeatBid: []openrtb2.SeatBid{
							{
								Seat: "pubmatic",
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp_1",
										Price: 5,
										Ext:   json.RawMessage(`{"signaldata":"{\"id\":\"123\",\"seatbid\":[{\"bid\":[{\"id\":\"bid-id-1\",\"impid\":\"imp_1\",\"price\":5}],\"seat\":\"pubmatic\"}],\"bidid\":\"bid-id-1\",\"cur\":\"USD\",\"ext\":{\"matchedimpression\":{\"appnexus\":50,\"pubmatic\":50}}}\r\n"}`),
									},
								},
							},
						},
					},
				},
				loggerURL: "",
			},
			want: &models.RequestCtx{WakandaDebug: &wakanda.Debug{Enabled: true, DebugData: wakanda.DebugData{WinningBid: true, HTTPResponseBody: "{\"id\":\"123\",\"seatbid\":[{\"bid\":[{\"id\":\"bid-id-1\",\"impid\":\"imp_1\",\"price\":5,\"ext\":{\"signaldata\":\"{\\\"id\\\":\\\"123\\\",\\\"seatbid\\\":[{\\\"bid\\\":[{\\\"id\\\":\\\"bid-id-1\\\",\\\"impid\\\":\\\"imp_1\\\",\\\"price\\\":5}],\\\"seat\\\":\\\"pubmatic\\\"}],\\\"bidid\\\":\\\"bid-id-1\\\",\\\"cur\\\":\\\"USD\\\",\\\"ext\\\":{\\\"matchedimpression\\\":{\\\"appnexus\\\":50,\\\"pubmatic\\\":50}}}\\r\\n\"}}],\"seat\":\"pubmatic\"}],\"bidid\":\"bid-id-1\",\"cur\":\"USD\"}",
				Logger: json.RawMessage{},
				OpenRTB: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{
							ID: "imp_1",
						},
					},
				}}}},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			setWakandaObject(tt.args.rCtx, tt.args.ao, tt.args.loggerURL)
			assert.Equal(t, tt.want, tt.args.rCtx)
		})
	}
}
