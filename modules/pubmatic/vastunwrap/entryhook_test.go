package vastunwrap

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/models"
)

func TestHandleEntrypointHook(t *testing.T) {
	type args struct {
		payload hookstage.EntrypointPayload
		config  VastUnwrapModule
	}
	tests := []struct {
		name      string
		args      args
		randomNum int
		want      hookstage.HookResult[hookstage.EntrypointPayload]
	}{
		{
			name: "Disable Vast Unwrapper",
			args: args{
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						ctx := context.WithValue(context.Background(), VastUnwrapEnabled, "0")
						ctx = context.WithValue(ctx, ProfileId, 0)
						ctx = context.WithValue(ctx, VersionId, 0)
						ctx = context.WithValue(ctx, DisplayId, 0)
						ctx = context.WithValue(ctx, Endpoint, "")
						r, _ := http.NewRequestWithContext(ctx, "POST", "http://localhost/video/openrtb?sshb=1", nil)
						return r
					}(),
				},
				config: VastUnwrapModule{
					TrafficPercentage: 2,
					Enabled:           false,
				},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: false, Redirect: true}}},
		},
		{
			name: "Enable Vast Unwrapper",
			args: args{
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						ctx := context.WithValue(context.Background(), VastUnwrapEnabled, "1")
						ctx = context.WithValue(ctx, ProfileId, 0)
						ctx = context.WithValue(ctx, VersionId, 0)
						ctx = context.WithValue(ctx, DisplayId, 0)
						ctx = context.WithValue(ctx, Endpoint, "")
						r, _ := http.NewRequestWithContext(ctx, "POST", "http://localhost/video/openrtb?sshb=1", nil)
						return r
					}(),
				},
				config: VastUnwrapModule{
					TrafficPercentage: 2,
				},
			},
			randomNum: 1,
			want:      hookstage.HookResult[hookstage.EntrypointPayload]{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: true, Redirect: true}}},
		},
		{
			name: "Disable Vast Unwrapper for owsdk source request",
			args: args{
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, _ := http.NewRequest("POST", "http://localhost/video/openrtb?source=owsdk", nil)
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1}}}`),
				},
				config: VastUnwrapModule{
					TrafficPercentage: 2,
					Enabled:           false,
				},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: false, Redirect: false, ProfileID: 5890, DisplayID: 1, Endpoint: "video"}}},
		},
		{
			name: "Enable Vast Unwrapper for owsdk source request",
			args: args{
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, _ := http.NewRequest("POST", "http://localhost/video/openrtb?source=owsdk", nil)
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1}}}`),
				},
				config: VastUnwrapModule{
					TrafficPercentage: 2,
				},
			},
			randomNum: 1,
			want:      hookstage.HookResult[hookstage.EntrypointPayload]{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: false, Redirect: false, ProfileID: 5890, DisplayID: 1, Endpoint: "video"}}},
		},
		{
			name: "Enable Vast Unwrapper for owsdk source request but missing profileID",
			args: args{
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, _ := http.NewRequest("POST", "http://localhost/video/openrtb?source=owsdk", nil)
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"versionid":1}}}`),
				},
				config: VastUnwrapModule{
					TrafficPercentage: 2,
				},
			},
			randomNum: 1,
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				Reject:  true,
				NbrCode: nbr.InvalidProfileID,
				Errors:  []string{"ErrMissingProfileID"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := handleEntrypointHook(nil, hookstage.ModuleInvocationContext{}, tt.args.payload)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleEntrypointHook() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
