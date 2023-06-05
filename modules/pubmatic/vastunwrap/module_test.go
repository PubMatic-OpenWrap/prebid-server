package vastunwrap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	unWrapCfg "git.pubmatic.com/vastunwrap/config"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/moduledeps"
	"github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/models"
)

func TestVastUnwrapModuleHandleEntrypointHook(t *testing.T) {
	type fields struct {
		cfg VastUnwrapModuleCfg
	}
	type args struct {
		ctx     context.Context
		miCtx   hookstage.ModuleInvocationContext
		payload hookstage.EntrypointPayload
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    hookstage.HookResult[hookstage.EntrypointPayload]
		wantErr bool
	}{
		{
			name:   "Vast unwrap is enabled in the config",
			fields: fields{cfg: VastUnwrapModuleCfg{VastUnWrapCfg: unWrapCfg.VastUnWrapCfg{Enabled: true}}},
			args: args{
				miCtx: hookstage.ModuleInvocationContext{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapFlag: true}}},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						ctx := context.WithValue(context.Background(), VastUnwrapperEnableKey, "1")
						r, _ := http.NewRequestWithContext(ctx, "", "", nil)
						return r
					}(),
				}},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapFlag: true}}},
		},
		{
			name:   "Vast unwrap is disabled in the config",
			fields: fields{cfg: VastUnwrapModuleCfg{VastUnWrapCfg: unWrapCfg.VastUnWrapCfg{Enabled: false}}},
			args: args{
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						ctx := context.WithValue(context.Background(), VastUnwrapperEnableKey, "1")
						r, _ := http.NewRequestWithContext(ctx, "", "", nil)
						return r
					}(),
				}},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := VastUnwrapModule{
				cfg: tt.fields.cfg,
			}
			got, err := m.HandleEntrypointHook(tt.args.ctx, tt.args.miCtx, tt.args.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("VastUnwrapModule.HandleEntrypointHook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VastUnwrapModule.HandleEntrypointHook() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVastUnwrapModuleHandleRawBidderResponseHook(t *testing.T) {
	type fields struct {
		cfg VastUnwrapModuleCfg
	}
	type args struct {
		in0     context.Context
		miCtx   hookstage.ModuleInvocationContext
		payload hookstage.RawBidderResponsePayload
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    hookstage.HookResult[hookstage.RawBidderResponsePayload]
		wantErr bool
	}{
		{
			name:   "Vast unwrap is enabled in the config",
			fields: fields{cfg: VastUnwrapModuleCfg{VastUnWrapCfg: unWrapCfg.VastUnWrapCfg{Enabled: true}}},
			args: args{
				miCtx: hookstage.ModuleInvocationContext{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapFlag: true}}},
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
						}},
				}},
			want: hookstage.HookResult[hookstage.RawBidderResponsePayload]{},
		},
		{
			name:   "Vast unwrap is disabled in the config",
			fields: fields{cfg: VastUnwrapModuleCfg{VastUnWrapCfg: unWrapCfg.VastUnWrapCfg{Enabled: false}}},
			args: args{
				miCtx: hookstage.ModuleInvocationContext{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{VastUnwrapFlag: false}}},
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
						}},
				}},
			want: hookstage.HookResult[hookstage.RawBidderResponsePayload]{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := VastUnwrapModule{
				cfg: tt.fields.cfg,
			}
			got, err := m.HandleRawBidderResponseHook(tt.args.in0, tt.args.miCtx, tt.args.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("VastUnwrapModule.HandleRawBidderResponseHook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VastUnwrapModule.HandleRawBidderResponseHook() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitVastUnrap(t *testing.T) {
	type args struct {
		rawCfg json.RawMessage
		in1    moduledeps.ModuleDeps
	}
	tests := []struct {
		name    string
		args    args
		want    VastUnwrapModule
		wantErr bool
	}{
		{
			name: "Valid vast unwrap config",
			args: args{
				rawCfg: json.RawMessage(`{"enabled":true,"vastunwrapcfg":{"app_config":{"debug":1,"unwrap_default_timeout":100},"enabled":true,"http_config":{"idle_conn_timeout":300,"max_idle_conns":100,"max_idle_conns_per_host":1},"log_config":{"debug_log_file":"/home/test/PBSlogs/unwrap/debug.log","error_log_file":"/home/test/PBSlogs/unwrap/error.log"},"server_config":{"dc_name":"OW_DC"},"stat_config":{"host":"10.172.141.13","port":8080,"referesh_interval_in_sec":1}}}`),
				in1:    moduledeps.ModuleDeps{},
			},
			want: VastUnwrapModule{
				cfg: VastUnwrapModuleCfg{
					VastUnWrapCfg: unWrapCfg.VastUnWrapCfg{
						Enabled:      true,
						HTTPConfig:   unWrapCfg.HttpConfig{MaxIdleConns: 100, MaxIdleConnsPerHost: 1, IdleConnTimeout: 300},
						APPConfig:    unWrapCfg.AppConfig{Host: "", Port: 0, UnwrapDefaultTimeout: 100, Debug: 1},
						StatConfig:   unWrapCfg.StatConfig{Host: "10.172.141.13", Port: 8080, RefershIntervalInSec: 1},
						ServerConfig: unWrapCfg.ServerConfig{ServerName: "", DCName: "OW_DC"},
						LogConfig:    unWrapCfg.LogConfig{ErrorLogFile: "/home/test/PBSlogs/unwrap/error.log", DebugLogFile: "/home/test/PBSlogs/unwrap/debug.log"},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := initVastUnrap(tt.args.rawCfg, tt.args.in1)
			if (err != nil) != tt.wantErr {
				t.Errorf("initVastUnrap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("initVastUnrap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuilder(t *testing.T) {
	type args struct {
		rawCfg json.RawMessage
		deps   moduledeps.ModuleDeps
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "Valid vast unwrap config",
			args: args{
				rawCfg: json.RawMessage(`{"enabled":true,"vastunwrapcfg":{"app_config":{"debug":1,"unwrap_default_timeout":100},"enabled":true,"http_config":{"idle_conn_timeout":300,"max_idle_conns":100,"max_idle_conns_per_host":1},"log_config":{"debug_log_file":"/home/test/PBSlogs/unwrap/debug.log","error_log_file":"/home/test/PBSlogs/unwrap/error.log"},"server_config":{"dc_name":"OW_DC"},"stat_config":{"host":"10.172.141.13","port":8080,"referesh_interval_in_sec":1}}}`),
				deps:   moduledeps.ModuleDeps{},
			},
			want: VastUnwrapModule{
				cfg: VastUnwrapModuleCfg{
					VastUnWrapCfg: unWrapCfg.VastUnWrapCfg{
						Enabled:      true,
						HTTPConfig:   unWrapCfg.HttpConfig{MaxIdleConns: 100, MaxIdleConnsPerHost: 1, IdleConnTimeout: 300},
						APPConfig:    unWrapCfg.AppConfig{Host: "", Port: 0, UnwrapDefaultTimeout: 100, Debug: 1},
						StatConfig:   unWrapCfg.StatConfig{Host: "10.172.141.13", Port: 8080, RefershIntervalInSec: 1},
						ServerConfig: unWrapCfg.ServerConfig{ServerName: "", DCName: "OW_DC"},
						LogConfig:    unWrapCfg.LogConfig{ErrorLogFile: "/home/test/PBSlogs/unwrap/error.log", DebugLogFile: "/home/test/PBSlogs/unwrap/debug.log"},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Builder(tt.args.rawCfg, tt.args.deps)
			if (err != nil) != tt.wantErr {
				t.Errorf("Builder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Builder() = %v, want %v", got, tt.want)
			}
		})
	}
}
