package vastunwrap

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	unWrapCfg "git.pubmatic.com/vastunwrap/config"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/models"
)

func TestHandleEntrypointHook(t *testing.T) {
	type args struct {
		payload hookstage.EntrypointPayload
		config  VastUnwrapModule
	}
	tests := []struct {
		name string
		args args
		want hookstage.HookResult[hookstage.EntrypointPayload]
	}{
		{
			name: "Disable Vast Unwrapper",
			args: args{
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						ctx := context.WithValue(context.Background(), isVastUnWrapEnabled, "0")
						r, _ := http.NewRequestWithContext(ctx, "", "", nil)
						return r
					}(),
				},
				config: VastUnwrapModule{
					Cfg: unWrapCfg.VastUnWrapCfg{
						HTTPConfig:   unWrapCfg.HttpConfig{MaxIdleConns: 100, MaxIdleConnsPerHost: 1, IdleConnTimeout: 300},
						APPConfig:    unWrapCfg.AppConfig{Host: "", Port: 0, UnwrapDefaultTimeout: 100, Debug: 1},
						StatConfig:   unWrapCfg.StatConfig{Host: "10.172.141.13", Port: 8080, RefershIntervalInSec: 1},
						ServerConfig: unWrapCfg.ServerConfig{ServerName: "", DCName: "OW_DC"},
						LogConfig:    unWrapCfg.LogConfig{ErrorLogFile: "/home/test/PBSlogs/unwrap/error.log", DebugLogFile: "/home/test/PBSlogs/unwrap/debug.log"},
					},
					TrafficPercentage: 2,
					Enabled:           false,
				},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{IsVastUnwrapEnabled: false}}},
		},
		{
			name: "Enable Vast Unwrapper",
			args: args{
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						ctx := context.WithValue(context.Background(), isVastUnWrapEnabled, "1")
						r, _ := http.NewRequestWithContext(ctx, "", "", nil)
						return r
					}(),
				},
				config: VastUnwrapModule{
					Cfg: unWrapCfg.VastUnWrapCfg{
						HTTPConfig:   unWrapCfg.HttpConfig{MaxIdleConns: 100, MaxIdleConnsPerHost: 1, IdleConnTimeout: 300},
						APPConfig:    unWrapCfg.AppConfig{Host: "", Port: 0, UnwrapDefaultTimeout: 100, Debug: 1},
						StatConfig:   unWrapCfg.StatConfig{Host: "10.172.141.13", Port: 8080, RefershIntervalInSec: 1},
						ServerConfig: unWrapCfg.ServerConfig{ServerName: "", DCName: "OW_DC"},
						LogConfig:    unWrapCfg.LogConfig{ErrorLogFile: "/home/test/PBSlogs/unwrap/error.log", DebugLogFile: "/home/test/PBSlogs/unwrap/debug.log"},
					},
					TrafficPercentage: 2,
					Enabled:           false,
				},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{IsVastUnwrapEnabled: true}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := handleEntrypointHook(nil, hookstage.ModuleInvocationContext{}, tt.args.payload, tt.args.config)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleEntrypointHook() = %v, want %v", got, tt.want)
			}
		})
	}
}
