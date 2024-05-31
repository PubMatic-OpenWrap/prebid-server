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
				mockEngine.EXPECT().RecordPublisherWrapperLoggerFailure("5890", "1", "")
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

func TestSetWakandaWinningBidFlag(t *testing.T) {
	type args struct {
		wakandaDebug *wakanda.Debug
		response     *openrtb2.BidResponse
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "all_empty_parameters",
			args: args{},
			want: false,
		},
		{
			name: "only_wakanda_empty",
			args: args{
				wakandaDebug: nil,
				response:     &openrtb2.BidResponse{},
			},
			want: false,
		},
		{
			name: "only_response_empty",
			args: args{
				wakandaDebug: &wakanda.Debug{},
				response:     nil,
			},
			want: false,
		},
		{
			name: "no_seatbid",
			args: args{
				wakandaDebug: &wakanda.Debug{},
				response:     &openrtb2.BidResponse{},
			},
			want: false,
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
			want: false,
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
			want: false,
		},
		{
			name: "zero_price",
			args: args{
				wakandaDebug: &wakanda.Debug{},
				response: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									Price: 0,
								},
							},
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setWakandaWinningBidFlag(tt.args.wakandaDebug, tt.args.response)
			actual := false
			if tt.args.wakandaDebug != nil {
				actual = tt.args.wakandaDebug.DebugData.WinningBid
			}
			assert.Equal(t, tt.want, actual)
		})
	}
}

func TestSetWakandaObject(t *testing.T) {
	testCases := []struct {
		description string
		rCtx        *models.RequestCtx
		ao          *analytics.AuctionObject
		loggerURL   string
	}{
		// {
		// 	description: "rctx is empty",
		// 	rCtx:        &models.RequestCtx{},
		// 	ao:          &analytics.AuctionObject{},
		// 	loggerURL:   "",
		// },
		{
			description: "wakanda is disabled",
			rCtx:        &models.RequestCtx{WakandaDebug: &wakanda.Debug{Enabled: false}},
			ao:          &analytics.AuctionObject{},
			loggerURL:   "",
		},
		{
			description: "wakanda is enabled",
			rCtx:        &models.RequestCtx{WakandaDebug: &wakanda.Debug{Enabled: true}},
			ao: &analytics.AuctionObject{
				RequestWrapper: &openrtb_ext.RequestWrapper{
					BidRequest: &openrtb2.BidRequest{},
				},
				Response: &openrtb2.BidResponse{},
			},
			loggerURL: "",
		},
		{
			description: "wakanda enabled with valid flow",
			rCtx:        &models.RequestCtx{WakandaDebug: &wakanda.Debug{Enabled: true}},
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
									Ext:   json.RawMessage(`{"signaldata":"{\"id\":\"123\",\"seatbid\":[{\"bid\":[{\"id\":\"bid-id-1\",\"impid\":\"imp_1\",\"price\":0}],\"seat\":\"pubmatic\"}],\"bidid\":\"bid-id-1\",\"cur\":\"USD\",\"ext\":{\"matchedimpression\":{\"appnexus\":50,\"pubmatic\":50}}}\r\n"}`),
								},
							},
						},
					},
				},
			},
			loggerURL: ow.cfg.Endpoint + `?json={"pubid":5890,"pid":"0","pdvid":"0","sl":1,"s":[{"sid":"uuid","sn":"sn","au":"au","ps":[{"pn":"pubmatic","bc":"pubmatic","kgpv":"","kgpsv":"","psz":"0x0","af":"","eg":0,"en":0,"l1":0,"l2":0,"t":0,"wb":0,"bidid":"bid-id-1","origbidid":"bid-id-1","di":"-1","dc":"","db":0,"ss":1,"mi":0,"ocpm":0,"ocry":"USD","fv":10.1,"frv":10.1}]}],"dvc":{},"fmv":"model-version","fsrc":2,"ft":1,"ffs":2,"fp":"provider1"}&pubid=5890`,
		},
	}
	for _, test := range testCases {
		setWakandaObject(test.rCtx, test.ao, test.loggerURL)
	}
}
