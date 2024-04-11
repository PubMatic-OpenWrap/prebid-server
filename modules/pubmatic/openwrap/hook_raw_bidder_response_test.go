package openwrap

import (
	"fmt"
	"net/http"

	"testing"

	unWrapCfg "git.pubmatic.com/vastunwrap/config"
	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	mock_metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestHandleRawBidderResponseHook(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMetricsEngine := mock_metrics.NewMockMetricsEngine(ctrl)

	type args struct {
		module              OpenWrap
		payload             hookstage.RawBidderResponsePayload
		moduleInvocationCtx hookstage.ModuleInvocationContext
		url                 string
		wantAdM             bool
		randomNumber        int
	}
	tests := []struct {
		name          string
		args          args
		wantResult    hookstage.HookResult[hookstage.RawBidderResponsePayload]
		expectedBids  []*adapters.TypedBid
		setup         func()
		wantErr       bool
		unwrapRequest func(w http.ResponseWriter, req *http.Request)
	}{
		{
			name: "Empty Request Context",
			args: args{
				module: OpenWrap{},
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{DebugMessages: []string{"error: request-ctx not found in handleRawBidderResponseHook()"}},
			wantErr:    false,
		},
		{
			name: "Set Vast Unwrapper to false in request context with type video",
			args: args{
				module: OpenWrap{
					cfg: config.Config{VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
						MaxWrapperSupport: 5,
						StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
						APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
					}},
					metricEngine: mockMetricsEngine,
				},
				payload: hookstage.RawBidderResponsePayload{
					Bids: []*adapters.TypedBid{
						{
							Bid: &openrtb2.Bid{
								ID:    "Bid-123",
								ImpID: fmt.Sprintf("div-adunit-%d", 123),
								Price: 2.1,
								AdM:   "<div>This is an Ad</div>",
								CrID:  "Cr-234",
								W:     100,
								H:     50,
							},
							BidType: "video",
						}}},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: false}}},
				randomNumber:        1,
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			wantErr:    false,
		},
		{
			name: "Set Vast Unwrapper to false in request context with type video, stats enabled true",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{
							VASTUnwrapStatsPecent: 2,
						},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
					unwrapRequest: func(w http.ResponseWriter, req *http.Request) {
						w.Header().Add("unwrap-status", "0")
						w.Header().Add("unwrap-count", "1")
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write([]byte(inlineXMLAdM))
					},
				},
				payload: hookstage.RawBidderResponsePayload{
					Bids: []*adapters.TypedBid{
						{
							Bid: &openrtb2.Bid{
								ID:    "Bid-123",
								ImpID: fmt.Sprintf("div-adunit-%d", 123),
								Price: 2.1,
								AdM:   vastXMLAdM,
								CrID:  "Cr-234",
								W:     100,
								H:     50,
							},
							BidType: "video",
						}},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: false}}},
				randomNumber:        1,
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordUnwrapRequestStatus("5890", "pubmatic", "0")
				mockMetricsEngine.EXPECT().RecordUnwrapWrapperCount("5890", "pubmatic", "1")
				mockMetricsEngine.EXPECT().RecordUnwrapRequestTime("5890", "pubmatic", gomock.Any())
				mockMetricsEngine.EXPECT().RecordUnwrapRespTime("5890", "1", gomock.Any())
			},
			wantErr: false,
		},
		{
			name: "Set Vast Unwrapper to true in request context with invalid vast xml",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{
							VASTUnwrapStatsPecent: 2,
							VASTUnwrapPecent:      50,
						},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
					unwrapRequest: func(w http.ResponseWriter, req *http.Request) {
						w.Header().Add("unwrap-status", "1")
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write([]byte(invalidVastXMLAdM))
					},
				},
				payload: hookstage.RawBidderResponsePayload{
					Bids: []*adapters.TypedBid{
						{
							Bid: &openrtb2.Bid{
								ID:    "Bid-123",
								ImpID: fmt.Sprintf("div-adunit-%d", 123),
								Price: 2.1,
								AdM:   invalidVastXMLAdM,
								CrID:  "Cr-234",
								W:     100,
								H:     50,
							},
							BidType: "video",
						},
					},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: true}}},
				url:                 UnwrapURL,
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordUnwrapRequestStatus("5890", "pubmatic", "1")
				mockMetricsEngine.EXPECT().RecordUnwrapRequestTime("5890", "pubmatic", gomock.Any())
			},
			wantErr: true,
		},
		{
			name: "Set Vast Unwrapper to true in request context with type video",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{
							VASTUnwrapStatsPecent: 2,
							VASTUnwrapPecent:      50,
						},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
					unwrapRequest: func(w http.ResponseWriter, req *http.Request) {
						w.Header().Add("unwrap-status", "0")
						w.Header().Add("unwrap-count", "1")
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write([]byte(inlineXMLAdM))
					},
				},
				payload: hookstage.RawBidderResponsePayload{
					Bids: []*adapters.TypedBid{
						{
							Bid: &openrtb2.Bid{
								ID:    "Bid-123",
								ImpID: fmt.Sprintf("div-adunit-%d", 123),
								Price: 2.1,
								AdM:   vastXMLAdM,
								CrID:  "Cr-234",
								W:     100,
								H:     50,
							},
							BidType: "video",
						},
					},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: true}}},
				url:                 UnwrapURL,
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordUnwrapRequestStatus("5890", "pubmatic", "0").AnyTimes()
				mockMetricsEngine.EXPECT().RecordUnwrapWrapperCount("5890", "pubmatic", "1").AnyTimes()
				mockMetricsEngine.EXPECT().RecordUnwrapRequestTime("5890", "pubmatic", gomock.Any()).AnyTimes()
				mockMetricsEngine.EXPECT().RecordUnwrapRespTime("5890", "1", gomock.Any()).AnyTimes()
			},
			wantErr: false,
		},
		{
			name: "Set Vast Unwrapper to true in request context for multiple bids with type video",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{
							VASTUnwrapStatsPecent: 2,
							VASTUnwrapPecent:      50,
						},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
					unwrapRequest: func(w http.ResponseWriter, req *http.Request) {
						w.Header().Add("unwrap-status", "0")
						w.Header().Add("unwrap-count", "1")
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write([]byte(inlineXMLAdM))
					},
				},
				payload: hookstage.RawBidderResponsePayload{
					Bids: []*adapters.TypedBid{
						{
							Bid: &openrtb2.Bid{
								ID:    "Bid-123",
								ImpID: fmt.Sprintf("div-adunit-%d", 123),
								Price: 2.1,
								AdM:   vastXMLAdM,
								CrID:  "Cr-234",
								W:     100,
								H:     50,
							},
							BidType: "video",
						},
						{
							Bid: &openrtb2.Bid{
								ID:    "Bid-456",
								ImpID: fmt.Sprintf("div-adunit-%d", 123),
								Price: 2.1,
								AdM:   vastXMLAdM,
								CrID:  "Cr-789",
								W:     100,
								H:     50,
							},
							BidType: "video",
						}},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: true}}},
				url:                 UnwrapURL,
				wantAdM:             true,
				randomNumber:        10,
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordUnwrapRequestStatus("5890", "pubmatic", "0").AnyTimes()
				mockMetricsEngine.EXPECT().RecordUnwrapWrapperCount("5890", "pubmatic", "1").AnyTimes()
				mockMetricsEngine.EXPECT().RecordUnwrapRequestTime("5890", "pubmatic", gomock.Any()).AnyTimes()
				mockMetricsEngine.EXPECT().RecordUnwrapRespTime("5890", "1", gomock.Any()).AnyTimes()
			},
			wantErr: false,
		},

		{
			name: "Set Vast Unwrapper to true in request context for multiple bids with different type",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{
							VASTUnwrapStatsPecent: 2,
							VASTUnwrapPecent:      50,
						},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
					unwrapRequest: func(w http.ResponseWriter, req *http.Request) {
						w.Header().Add("unwrap-status", "0")
						w.Header().Add("unwrap-count", "0")
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write([]byte(inlineXMLAdM))
					},
				},
				payload: hookstage.RawBidderResponsePayload{
					Bids: []*adapters.TypedBid{
						{
							Bid: &openrtb2.Bid{
								ID:    "Bid-123",
								ImpID: fmt.Sprintf("div-adunit-%d", 123),
								Price: 2.1,
								AdM:   vastXMLAdM,
								CrID:  "Cr-234",
								W:     100,
								H:     50,
							},
							BidType: "video",
						},
						{
							Bid: &openrtb2.Bid{
								ID:    "Bid-456",
								ImpID: fmt.Sprintf("div-adunit-%d", 123),
								Price: 2.1,
								AdM:   vastXMLAdM,
								CrID:  "Cr-789",
								W:     100,
								H:     50,
							},
							BidType: "banner",
						}},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: true}}},
				url:                 UnwrapURL,
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			expectedBids: []*adapters.TypedBid{{
				Bid: &openrtb2.Bid{
					ID:    "Bid-123",
					ImpID: fmt.Sprintf("div-adunit-%d", 123),
					Price: 2.1,
					AdM:   inlineXMLAdM,
					CrID:  "Cr-234",
					W:     100,
					H:     50,
				},
				BidType: "video",
			},
				{
					Bid: &openrtb2.Bid{
						ID:    "Bid-456",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   vastXMLAdM,
						CrID:  "Cr-789",
						W:     100,
						H:     50,
					},
					BidType: "banner",
				},
			},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordUnwrapRequestStatus("5890", "pubmatic", "0").AnyTimes()
				mockMetricsEngine.EXPECT().RecordUnwrapWrapperCount("5890", "pubmatic", "0").AnyTimes()
				mockMetricsEngine.EXPECT().RecordUnwrapRequestTime("5890", "pubmatic", gomock.Any()).AnyTimes()
				mockMetricsEngine.EXPECT().RecordUnwrapRespTime("5890", "0", gomock.Any()).AnyTimes()
			},
			wantErr: false,
		},
		{
			name: "Set Vast Unwrapper to true in request context with type video and source owsdk",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{
							VASTUnwrapStatsPecent: 2,
							VASTUnwrapPecent:      50,
						},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
					unwrapRequest: func(w http.ResponseWriter, req *http.Request) {
						w.Header().Add("unwrap-status", "0")
						w.Header().Add("unwrap-count", "1")
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write([]byte(inlineXMLAdM))
					},
				},
				payload: hookstage.RawBidderResponsePayload{
					Bids: []*adapters.TypedBid{
						{
							Bid: &openrtb2.Bid{
								ID:    "Bid-123",
								ImpID: fmt.Sprintf("div-adunit-%d", 123),
								Price: 2.1,
								AdM:   vastXMLAdM,
								CrID:  "Cr-234",
								W:     100,
								H:     50,
							},
							BidType: "video",
						},
					},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{ProfileID: 5890, DisplayID: 1, Endpoint: "/openrtb/video"}}},
				url:                 UnwrapURL,
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordUnwrapRequestStatus("5890", "pubmatic", "0").AnyTimes()
				mockMetricsEngine.EXPECT().RecordUnwrapWrapperCount("5890", "pubmatic", "1").AnyTimes()
				mockMetricsEngine.EXPECT().RecordUnwrapRequestTime("5890", "pubmatic", gomock.Any()).AnyTimes()
				mockMetricsEngine.EXPECT().RecordUnwrapRespTime("5890", "1", gomock.Any()).AnyTimes()
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			GetRandomNumberIn1To100 = func() int {
				return tt.args.randomNumber
			}

			m := tt.args.module
			_, err := m.handleRawBidderResponseHook(tt.args.moduleInvocationCtx, tt.args.payload, tt.args.url)
			if !assert.NoError(t, err, tt.wantErr) {
				return
			}
			if tt.args.moduleInvocationCtx.ModuleContext != nil && tt.args.wantAdM {
				assert.Equal(t, inlineXMLAdM, tt.args.payload.Bids[0].Bid.AdM, "AdM is not updated correctly after executing RawBidderResponse hook.")
			}
		})
	}
}
