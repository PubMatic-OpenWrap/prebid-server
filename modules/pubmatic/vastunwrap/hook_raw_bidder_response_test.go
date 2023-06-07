package vastunwrap

import (
	"fmt"

	reflect "reflect"
	"testing"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestHandleRawBidderResponseHook(t *testing.T) {
	type args struct {
		payload   hookstage.RawBidderResponsePayload
		moduleCtx hookstage.ModuleContext
	}
	tests := []struct {
		name       string
		args       args
		wantResult hookstage.HookResult[hookstage.RawBidderResponsePayload]
		wantErr    bool
	}{
		{
			name: "Empty Request Context",
			args: args{
				moduleCtx: hookstage.ModuleContext{},
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{DebugMessages: []string{"error: request-ctx not found in handleBeforeValidationHook()"}},
			wantErr:    false,
		},
		{
			name: "Set Vast Unwrapper to true in request context",
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
				moduleCtx: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapFlag: true}},
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := handleRawBidderResponseHook(tt.args.payload, tt.args.moduleCtx)
			if !assert.NoError(t, err, tt.wantErr) {
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("handleRawBidderResponseHook() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
