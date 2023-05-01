package openwrap

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/prebid/prebid-server/hooks/hookstage"
)

func TestHandleEntrypointHook(t *testing.T) {
	type args struct {
		payload             hookstage.EntrypointPayload
		enableVastUnwrapper bool
	}
	tests := []struct {
		name string
		args args
		want hookstage.HookResult[hookstage.EntrypointPayload]
	}{
		{
			name: "Disable Vast Unwrapper",
			args: args{
				payload:             hookstage.EntrypointPayload{Request: &http.Request{}},
				enableVastUnwrapper: false,
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{},
		},
		{
			name: "Enable Vast Unwrapper",
			args: args{
				payload:             hookstage.EntrypointPayload{Request: &http.Request{}},
				enableVastUnwrapper: true,
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vastUnwrapper := "0"
			if tt.args.enableVastUnwrapper {
				vastUnwrapper = "1"
			}
			ctx := context.WithValue(tt.args.payload.Request.Context(), "enableVastUnwrapper", vastUnwrapper)
			tt.args.payload.Request = tt.args.payload.Request.WithContext(ctx)
			rCtx := RequestCtx{
				VastUnwrapFlag: getContextValueForField(tt.args.payload.Request.Context(), "enableVastUnwrapper"),
			}

			tt.want.ModuleContext = make(hookstage.ModuleContext)
			tt.want.ModuleContext["rctx"] = rCtx
			got, _ := handleEntrypointHook(nil, hookstage.ModuleInvocationContext{}, tt.args.payload)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleEntrypointHook() = %v, want %v", got, tt.want)
			}
		})
	}
}
