package vastunwrap

import (
	"context"
	"net/http"
	"testing"

	"github.com/prebid/prebid-server/v2/hooks/hookstage"
	ow_models "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/vastunwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestHandleEntrypointHook(t *testing.T) {
	type args struct {
		payload hookstage.EntrypointPayload
	}
	tests := []struct {
		name string
		args args
		want hookstage.HookResult[hookstage.EntrypointPayload]
	}{
		{
			name: "Disable Vast Unwrapper for CTV video/openrtb request",
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
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: false, Redirect: true}}},
		},
		{
			name: "Enable Vast Unwrapper for CTV video/openrtb request",
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
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: true, Redirect: true}}},
		},
		{
			name: "Vast Unwrapper for IN-APP openrtb2/auction request",
			args: args{
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, _ := http.NewRequest("POST", "http://localhost/openrtb2/auction?source=owsdk", nil)
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1}}}`),
				},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: false, Redirect: false, ProfileID: 5890, DisplayID: 1, Endpoint: ow_models.EndpointV25}}},
		},
		{
			name: "Vast Unwrapper for IN-APP /openrtb/2.5 request coming from SSHB",
			args: args{
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						ctx := context.WithValue(context.Background(), VastUnwrapEnabled, "1")
						r, _ := http.NewRequestWithContext(ctx, "POST", "http://localhost/openrtb2/auction?sshb=1", nil)
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1}}}`),
				},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: true, Redirect: true, ProfileID: 0, DisplayID: 0, Endpoint: ""}}},
		},
		{
			name: "Vast Unwrapper for IN-APP /openrtb/2.5 request directly coming to prebid",
			args: args{
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						ctx := context.WithValue(context.Background(), VastUnwrapEnabled, "1")
						r, _ := http.NewRequestWithContext(ctx, "POST", "http://localhost/openrtb/2.5", nil)
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1}}}`),
				},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: false, Redirect: false, ProfileID: 5890, DisplayID: 1, Endpoint: ow_models.EndpointV25}}},
		},
		{
			name: "Vast Unwrapper for WebS2S activation request",
			args: args{
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						ctx := context.WithValue(context.Background(), VastUnwrapEnabled, "1")
						r, _ := http.NewRequestWithContext(ctx, "POST", "http://localhost/openrtb2/auction?source=pbjs", nil)
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1}}}`),
				},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{},
				DebugMessages: []string{"webs2s endpoint does not support vast-unwrap feature"},
			},
		},
		{
			name: "Vast Unwrapper for Hybrid request",
			args: args{
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						ctx := context.WithValue(context.Background(), VastUnwrapEnabled, "1")
						r, _ := http.NewRequestWithContext(ctx, "POST", "http://localhost:8001/pbs/openrtb2/auction", nil)
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1}}}`),
				},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{},
				DebugMessages: []string{"hybrid endpoint does not support vast-unwrap feature"},
			},
		},
		{
			name: "Vast Unwrapper for AMP request coming from SSHB",
			args: args{
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, _ := http.NewRequest("GET", "http://localhost:8001/amp?sshb=1&v=1&w=300&h=250", nil)
						return r
					}(),
				},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: false, Redirect: true, Endpoint: ""}}},
		},
		{
			name: "Vast Unwrapper for IN-APP OTT request coming from SSHB",
			args: args{
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, _ := http.NewRequest("GET", "http://localhost:8001/openrtb/2.5/video?sshb=1&owLoggerDebug=1&pubId=5890&profId=2543", nil)
						return r
					}(),
				},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: false, Redirect: true, Endpoint: ""}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := handleEntrypointHook(nil, hookstage.ModuleInvocationContext{}, tt.args.payload)
			assert.Equal(t, tt.want, got, "mismatched handleEntrypointHook output")
		})
	}
}
