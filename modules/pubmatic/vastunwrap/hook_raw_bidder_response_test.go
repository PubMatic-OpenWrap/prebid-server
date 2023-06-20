package vastunwrap

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"testing"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestHandleRawBidderResponseHook(t *testing.T) {
	type args struct {
		payload       hookstage.RawBidderResponsePayload
		moduleCtx     hookstage.ModuleContext
		unwrapTimeout int
		url           string
	}
	tests := []struct {
		name         string
		args         args
		wantResult   hookstage.HookResult[hookstage.RawBidderResponsePayload]
		expectedBids []*adapters.TypedBid
		wantErr      bool
	}{
		{
			name: "Empty Request Context",
			args: args{
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
						}}},
				moduleCtx: nil,
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{DebugMessages: []string{"error: request-ctx not found in handleRawBidderResponseHook()"}},
			wantErr:    false,
		},
		{
			name: "Set Vast Unwrapper to false in request context with type video",
			args: args{
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
				moduleCtx:     hookstage.ModuleContext{"rctx": models.RequestCtx{IsVastUnwrapEnabled: false}},
				unwrapTimeout: 1000,
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
			name: "Set Vast Unwrapper to true in request context with type video",
			args: args{
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
						}}},
				moduleCtx:     hookstage.ModuleContext{"rctx": models.RequestCtx{IsVastUnwrapEnabled: true}},
				unwrapTimeout: 1000,
				url:           UnwrapURL,
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			expectedBids: []*adapters.TypedBid{{
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
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				data, _ := json.Marshal(tt.expectedBids)
				_, _ = w.Write([]byte(data))
			}))
			defer server.Close()
			server.URL = tt.args.url
			doUnwrap(tt.args.payload.Bids[0], "test", tt.args.unwrapTimeout, server.URL)
			_, err := handleRawBidderResponseHook(tt.args.payload, tt.args.moduleCtx, tt.args.unwrapTimeout, server.URL)

			if !assert.NoError(t, err, tt.wantErr) {
				return
			}
			if tt.args.moduleCtx != nil {
				assert.Equal(t, tt.expectedBids[0].Bid.AdM, tt.args.payload.Bids[0].Bid.AdM, "AdM is not updated correctly after executing RawBidderResponse hook.")
			}
		})
	}
}
