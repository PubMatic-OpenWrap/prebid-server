package vastunwrap

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/prebid/prebid-server/v2/hooks/hookstage"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/vastunwrap/models"
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
						r, _ := http.NewRequestWithContext(ctx, "", "", nil)
						return r
					}(),
				},
				config: VastUnwrapModule{
					TrafficPercentage: 2,
					Enabled:           false,
				},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: false}}},
		},
		{
			name: "Enable Vast Unwrapper with random number less than traffic percentage",
			args: args{
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						ctx := context.WithValue(context.Background(), VastUnwrapEnabled, "1")
						r, _ := http.NewRequestWithContext(ctx, "", "", nil)
						return r
					}(),
				},
				config: VastUnwrapModule{
					TrafficPercentage: 2,
				},
			},
			randomNum: 1,
			want:      hookstage.HookResult[hookstage.EntrypointPayload]{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: true}}},
		},
		{
			name: "Enable Vast Unwrapper with random number equal to traffic percenatge",
			args: args{
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						ctx := context.WithValue(context.Background(), VastUnwrapEnabled, "1")
						r, _ := http.NewRequestWithContext(ctx, "", "", nil)
						return r
					}(),
				},
				config: VastUnwrapModule{
					TrafficPercentage: 2,
				},
			},
			randomNum: 2,
			want:      hookstage.HookResult[hookstage.EntrypointPayload]{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: false}}},
		},
		{
			name: "Enable Vast Unwrapper with random number greater than traffic percenatge",
			args: args{
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						ctx := context.WithValue(context.Background(), VastUnwrapEnabled, "1")
						r, _ := http.NewRequestWithContext(ctx, "", "", nil)
						return r
					}(),
				},
				config: VastUnwrapModule{
					TrafficPercentage: 2,
				},
			},
			randomNum: 5,
			want:      hookstage.HookResult[hookstage.EntrypointPayload]{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapEnabled: false}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldRandomNumberGen := getRandomNumber
			getRandomNumber = func() int { return tt.randomNum }
			defer func() {
				getRandomNumber = oldRandomNumberGen
			}()
			got, _ := handleEntrypointHook(nil, hookstage.ModuleInvocationContext{}, tt.args.payload, tt.args.config)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleEntrypointHook() = %v, want %v", got, tt.want)
			}
		})
	}
}
