package openwrap

import (
	"context"
	"net/http"
	"testing"

	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestOpenWrap_handleEntrypointHook(t *testing.T) {
	type fields struct {
		cfg   config.Config
		cache cache.Cache
	}
	type args struct {
		in0     context.Context
		miCtx   hookstage.ModuleInvocationContext
		payload hookstage.EntrypointPayload
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    hookstage.HookResult[hookstage.EntrypointPayload]
		wantErr error
	}{
		{
			name: "valid /openrtb/2.5 request",
			fields: fields{
				cfg:   config.Config{},
				cache: nil,
			},
			args: args{
				in0:   context.Background(),
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						r, err := http.NewRequest("POST", "http://localhost/openrtb/2.5", nil)
						if err != nil {
							panic(err)
						}
						return r
					}(),
					Body: []byte(`{"ext":{"wrapper":{"profileid":5890,"versionid":1}}}`),
				},
			},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{
				ModuleContext: hookstage.ModuleContext{
					"rctx": models.RequestCtx{
						ProfileID:                5890,
						DisplayID:                1,
						SSAuction:                -1,
						Aliases:                  make(map[string]string),
						ImpBidCtx:                make(map[string]models.ImpCtx),
						PrebidBidderCode:         make(map[string]string),
						BidderResponseTimeMillis: make(map[string]int),
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := OpenWrap{
				cfg:   tt.fields.cfg,
				cache: tt.fields.cache,
			}
			got, err := m.handleEntrypointHook(tt.args.in0, tt.args.miCtx, tt.args.payload)
			assert.Equal(t, err, tt.wantErr)

			// validate runtime values individually and reset them
			rctx := got.ModuleContext["rctx"].(models.RequestCtx)
			assert.NotEmpty(t, rctx.StartTime)
			assert.Len(t, rctx.LoggerImpressionID, 36)

			rctx.StartTime = 0
			rctx.LoggerImpressionID = ""
			got.ModuleContext["rctx"] = rctx

			assert.Equal(t, got, tt.want)
		})
	}
}
