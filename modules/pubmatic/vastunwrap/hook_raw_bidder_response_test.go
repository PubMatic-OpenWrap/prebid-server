package vastunwrap

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"testing"

	"git.pubmatic.com/vastunwrap/config"
	unWrapCfg "git.pubmatic.com/vastunwrap/config"
	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/models"
	mock_stats "github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/stats/mock"
	"github.com/stretchr/testify/assert"
)

func TestHandleRawBidderResponseHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMetricsEngine := mock_stats.NewMockMetricsEngine(ctrl)
	VastUnWrapModule := VastUnwrapModule{Cfg: config.VastUnWrapCfg{MaxWrapperSupport: 5, StatConfig: unWrapCfg.StatConfig{Host: "10.172.141.13", Port: 8080, RefershIntervalInSec: 1}, APPConfig: config.AppConfig{UnwrapDefaultTimeout: 1500}}, MetricsEngine: mockMetricsEngine}
	type args struct {
		module              VastUnwrapModule
		payload             hookstage.RawBidderResponsePayload
		moduleInvocationCtx hookstage.ModuleInvocationContext
		unwrapTimeout       int
		url                 string
		status              string
		wantAdM             bool
	}
	tests := []struct {
		name         string
		args         args
		wantResult   hookstage.HookResult[hookstage.RawBidderResponsePayload]
		expectedBids []*adapters.TypedBid
		setup        func()
		wantErr      bool
	}{
		{
			name: "Empty Request Context",
			args: args{
				module: VastUnWrapModule,
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{DebugMessages: []string{"error: request-ctx not found in handleRawBidderResponseHook()"}},
			wantErr:    false,
		},
		{
			name: "Set Vast Unwrapper to false in request context with type video",
			args: args{
				module: VastUnWrapModule,
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
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			expectedBids: []*adapters.TypedBid{{
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
			}},
			wantErr: false,
		},
		{
			name: "Set Vast Unwrapper to true in request context with invalid vast xml",
			args: args{
				module: VastUnWrapModule,
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
				status:              "1",
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			expectedBids: []*adapters.TypedBid{{
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
			}},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordRequestStatus("pubmatic", "1").AnyTimes()
				mockMetricsEngine.EXPECT().RecordRequestTime("pubmatic", gomock.Any()).AnyTimes()
			},
			wantErr: true,
		},
		{
			name: "Set Vast Unwrapper to true in request context with type video",
			args: args{
				module: VastUnWrapModule,
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
				status:              "0",
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
			}},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordRequestStatus("pubmatic", "0").AnyTimes()
				mockMetricsEngine.EXPECT().RecordWrapperCount("pubmatic", "1").AnyTimes()
				mockMetricsEngine.EXPECT().RecordRequestTime("pubmatic", gomock.Any()).AnyTimes()
			},
			wantErr: false,
		},
		{
			name: "Set Vast Unwrapper to true in request context for multiple bids with type video",
			args: args{
				module: VastUnWrapModule,
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
				status:              "0",
				wantAdM:             true,
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
						AdM:   inlineXMLAdM,
						CrID:  "Cr-789",
						W:     100,
						H:     50,
					},
					BidType: "video",
				},
			},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordRequestStatus("pubmatic", "0").AnyTimes()
				mockMetricsEngine.EXPECT().RecordWrapperCount("pubmatic", "1").AnyTimes()
				mockMetricsEngine.EXPECT().RecordRequestTime("pubmatic", gomock.Any()).AnyTimes()
			},
			wantErr: false,
		},
		{
			name: "Set Vast Unwrapper to true in request context for multiple bids with different type",
			args: args{
				module: VastUnWrapModule,
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
				status:              "0",
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
				mockMetricsEngine.EXPECT().RecordRequestStatus("pubmatic", "0").AnyTimes()
				mockMetricsEngine.EXPECT().RecordWrapperCount("pubmatic", "0").AnyTimes()
				mockMetricsEngine.EXPECT().RecordRequestTime("pubmatic", gomock.Any()).AnyTimes()
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			if tt.args.moduleInvocationCtx.ModuleContext != nil {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Header().Add("unwrap-status", tt.args.status)
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(tt.expectedBids[0].Bid.AdM))
				}))
				url := server.URL
				for _, bid := range tt.args.payload.Bids {
					bid.Bid.AdM = strings.Replace(bid.Bid.AdM, "{URL}", url, 1)
				}
				defer server.Close()
			}
			_, err := handleRawBidderResponseHook(tt.args.module, tt.args.moduleInvocationCtx, tt.args.payload, "test")
			if !assert.NoError(t, err, tt.wantErr) {
				return
			}
			if tt.args.moduleInvocationCtx.ModuleContext != nil && tt.args.wantAdM {
				assert.Equal(t, finalAdM, tt.args.payload.Bids[0].Bid.AdM, "AdM is not updated correctly after executing RawBidderResponse hook.")
			}
		})
	}
}
