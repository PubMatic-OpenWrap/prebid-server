package openwrap

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/hooks/hookanalytics"
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	mock_cache "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache/mock"
	endpointmanager "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestOpenWrap_HandleProcessedAuctionHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCache := mock_cache.NewMockCache(ctrl)
	defer ctrl.Finish()

	type args struct {
		ctx       context.Context
		moduleCtx hookstage.ModuleInvocationContext
		payload   hookstage.ProcessedAuctionRequestPayload
	}
	tests := []struct {
		name            string
		args            args
		mutationApplied bool
		want            hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload]
		wantErr         bool
		wantBidRequest  *openrtb2.BidRequest
	}{
		{
			name: "empty module context",
			args: args{
				ctx:       nil,
				moduleCtx: hookstage.ModuleInvocationContext{},
				payload:   hookstage.ProcessedAuctionRequestPayload{},
			},
			want: hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload]{
				Reject:        false,
				ChangeSet:     hookstage.ChangeSet[hookstage.ProcessedAuctionRequestPayload]{},
				DebugMessages: []string{"error: module-ctx not found in handleProcessedAuctionHook()"},
				AnalyticsTags: hookanalytics.Analytics{},
			},
			wantErr: false,
		},
		{
			name: "empty request context",
			args: args{
				ctx: nil,
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: func() *hookstage.ModuleContext {
						moduleCtx := hookstage.NewModuleContext()
						moduleCtx.Set("rctx", nil)
						return moduleCtx
					}(),
				},
			},
			want: hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload]{
				Reject:        false,
				ChangeSet:     hookstage.ChangeSet[hookstage.ProcessedAuctionRequestPayload]{},
				DebugMessages: []string{"error: request-ctx not found in handleProcessedAuctionHook()"},
				AnalyticsTags: hookanalytics.Analytics{},
			},
			wantErr: false,
		},
		{
			name: "SSHb request",
			args: args{
				ctx: nil,
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: func() *hookstage.ModuleContext {
						moduleCtx := hookstage.NewModuleContext()
						moduleCtx.Set("rctx", models.RequestCtx{
							Sshb: "1",
						})
						moduleCtx.Set("endpointhookmanager", &endpointmanager.NilEndpointManager{})
						return moduleCtx
					}(),
				},
			},

			want: hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload]{
				Reject:        false,
				ChangeSet:     hookstage.ChangeSet[hookstage.ProcessedAuctionRequestPayload]{},
				DebugMessages: nil,
				AnalyticsTags: hookanalytics.Analytics{},
			},
		},
		{
			name: "Hybrid request",
			args: args{
				ctx: nil,
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: func() *hookstage.ModuleContext {
						moduleCtx := hookstage.NewModuleContext()
						moduleCtx.Set("rctx", models.RequestCtx{
							Endpoint: models.EndpointHybrid,
						})
						moduleCtx.Set("endpointhookmanager", &endpointmanager.NilEndpointManager{})
						return moduleCtx
					}(),
				},
			},
			want: hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload]{
				Reject:        false,
				ChangeSet:     hookstage.ChangeSet[hookstage.ProcessedAuctionRequestPayload]{},
				DebugMessages: nil,
				AnalyticsTags: hookanalytics.Analytics{},
			},
		},
		{
			name: "empty device ip updated with request ip",
			args: args{
				ctx: nil,
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: func() *hookstage.ModuleContext {
						moduleCtx := hookstage.NewModuleContext()
						moduleCtx.Set("rctx", models.RequestCtx{
							Endpoint:  models.EndpointV25,
							DeviceCtx: models.DeviceCtx{IP: "10.20.30.40"},
						})
						moduleCtx.Set("endpointhookmanager", &endpointmanager.NilEndpointManager{})
						return moduleCtx
					}(),
				},
				payload: hookstage.ProcessedAuctionRequestPayload{
					Request: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{
							Device: &openrtb2.Device{
								IP: "",
							},
						},
					},
				},
			},
			mutationApplied: true,
			want: hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload]{
				Reject:        false,
				ChangeSet:     hookstage.ChangeSet[hookstage.ProcessedAuctionRequestPayload]{},
				DebugMessages: nil,
				AnalyticsTags: hookanalytics.Analytics{},
			},
			wantBidRequest: &openrtb2.BidRequest{
				Device: &openrtb2.Device{
					IP: "10.20.30.40",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := OpenWrap{
				cache: mockCache,
			}
			got, err := m.HandleProcessedAuctionHook(tt.args.ctx, tt.args.moduleCtx, tt.args.payload)
			assert.Equal(t, tt.wantErr, err != nil, "handleAllProcessedBidResponsesHook() error = %v, wantErr %v", err, tt.wantErr)
			if tt.mutationApplied {
				mutations := got.ChangeSet.Mutations()
				assert.NotEmpty(t, mutations, tt.name)
				for _, mut := range mutations {
					result, err := mut.Apply(tt.args.payload)
					assert.Nil(t, err, tt.name)
					assert.Equal(t, tt.wantBidRequest, result.Request.BidRequest, tt.name)
				}
			}
			assert.Equal(t, tt.want.DebugMessages, got.DebugMessages, "Debug messages should be equal")
			assert.Equal(t, tt.want.Reject, false, "Reject should be equal")
		})
	}
}
