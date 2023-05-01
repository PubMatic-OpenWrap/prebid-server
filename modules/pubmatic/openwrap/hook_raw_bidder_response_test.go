package openwrap

import (
	"reflect"
	"testing"

	"github.com/prebid/prebid-server/hooks/hookstage"
)

func TestHandleRawBidderResponseHook(t *testing.T) {
	type args struct {
		moduleCtx hookstage.ModuleContext
	}
	tests := []struct {
		name       string
		args       args
		wantResult hookstage.HookResult[hookstage.RawBidderResponsePayload]
	}{
		{
			name: "Empty Request Context",
			args: args{
				moduleCtx: hookstage.ModuleContext{},
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{DebugMessages: []string{"error: request-ctx not found in handleBeforeValidationHook()"}},
		},
		{
			name: "Set Vast Unwrapper to true in request context",
			args: args{
				moduleCtx: hookstage.ModuleContext{"rctx": RequestCtx{VastUnwrapFlag: true}},
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{},
		},
		{
			name: "Set Vast Unwrapper to false in request context",
			args: args{
				moduleCtx: hookstage.ModuleContext{"rctx": RequestCtx{VastUnwrapFlag: false}},
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, _ := handleRawBidderResponseHook(hookstage.RawBidderResponsePayload{}, tt.args.moduleCtx)
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("handleRawBidderResponseHook() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
