package openwrap

import (
	"context"
	"testing"

	mock_cache "github.com/PubMatic-OpenWrap/prebid-server/modules/pubmatic/openwrap/cache/mock"
	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/exchange/entities"
	"github.com/prebid/prebid-server/hooks/hookanalytics"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestUpdateBidIds(t *testing.T) {
	type args struct {
		bidderResponses map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid
	}
	tests := []struct {
		name string
		args args
		want map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid
	}{
		{
			name: "All bidIds are updated",
			args: args{
				bidderResponses: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
					"pubmatic": {
						Bids: []*entities.PbsOrtbBid{
							{
								Bid: &openrtb2.Bid{
									ID: "bid-1",
								},
								GeneratedBidID: "gen-1",
							},
							{
								Bid: &openrtb2.Bid{
									ID: "bid-2",
								},
								GeneratedBidID: "gen-2",
							},
						},
					},
				},
			},
			want: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
				"pubmatic": {
					Bids: []*entities.PbsOrtbBid{
						{
							Bid: &openrtb2.Bid{
								ID: "bid-1::gen-1",
							},
							GeneratedBidID: "gen-1",
						},
						{
							Bid: &openrtb2.Bid{
								ID: "bid-2::gen-2",
							},
							GeneratedBidID: "gen-2",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateBidIds(tt.args.bidderResponses)
			assert.Equal(t, tt.want, tt.args.bidderResponses, "Bid Id should be equal")
		})
	}
}

func TestOpenWrap_handleAllProcessedBidResponsesHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCache := mock_cache.NewMockCache(ctrl)
	defer ctrl.Finish()

	type args struct {
		ctx       context.Context
		moduleCtx hookstage.ModuleInvocationContext
		payload   hookstage.AllProcessedBidResponsesPayload
	}
	tests := []struct {
		name            string
		args            args
		mutationApplied bool
		want            hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload]
		wantErr         bool
		wantResponse    map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid
	}{
		{
			name: "empty module context",
			args: args{
				ctx:       nil,
				moduleCtx: hookstage.ModuleInvocationContext{},
				payload:   hookstage.AllProcessedBidResponsesPayload{},
			},
			want: hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload]{
				Reject:        false,
				ChangeSet:     hookstage.ChangeSet[hookstage.AllProcessedBidResponsesPayload]{},
				DebugMessages: []string{"error: module-ctx not found in handleAllProcessedBidResponsesHook()"},
				AnalyticsTags: hookanalytics.Analytics{},
			},
			wantErr: false,
		},
		{
			name: "empty request context",
			args: args{
				ctx: nil,
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: map[string]interface{}{
						"rctx": nil,
					},
				},
			},
			want: hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload]{
				Reject:        false,
				ChangeSet:     hookstage.ChangeSet[hookstage.AllProcessedBidResponsesPayload]{},
				DebugMessages: []string{"error: request-ctx not found in handleAllProcessedBidResponsesHook()"},
				AnalyticsTags: hookanalytics.Analytics{},
			},
			wantErr: false,
		},
		{
			name: "SSHb request",
			args: args{
				ctx: nil,
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: map[string]interface{}{
						"rctx": models.RequestCtx{
							Sshb: "1",
						},
					},
				},
			},

			want: hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload]{
				Reject:        false,
				ChangeSet:     hookstage.ChangeSet[hookstage.AllProcessedBidResponsesPayload]{},
				DebugMessages: nil,
				AnalyticsTags: hookanalytics.Analytics{},
			},
		},
		{
			name: "Hybrid request",
			args: args{
				ctx: nil,
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: map[string]interface{}{
						"rctx": models.RequestCtx{
							Endpoint: models.EndpointHybrid,
						},
					},
				},
			},
			want: hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload]{
				Reject:        false,
				ChangeSet:     hookstage.ChangeSet[hookstage.AllProcessedBidResponsesPayload]{},
				DebugMessages: nil,
				AnalyticsTags: hookanalytics.Analytics{},
			},
		},
		{
			name: "All bidIds are updated",
			args: args{
				ctx: nil,
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: map[string]interface{}{
						"rctx": models.RequestCtx{
							Endpoint: models.EndpointV25,
						},
					},
				},
				payload: hookstage.AllProcessedBidResponsesPayload{
					Responses: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
						"pubmatic": {
							Bids: []*entities.PbsOrtbBid{
								{
									Bid: &openrtb2.Bid{
										ID: "bid-1",
									},
									GeneratedBidID: "gen-1",
								},
								{
									Bid: &openrtb2.Bid{
										ID: "bid-2",
									},
									GeneratedBidID: "gen-2",
								},
							},
						},
					},
				},
			},
			mutationApplied: true,
			want: hookstage.HookResult[hookstage.AllProcessedBidResponsesPayload]{
				Reject:        false,
				ChangeSet:     hookstage.ChangeSet[hookstage.AllProcessedBidResponsesPayload]{},
				DebugMessages: nil,
				AnalyticsTags: hookanalytics.Analytics{},
			},
			wantErr: false,
			wantResponse: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
				"pubmatic": {
					Bids: []*entities.PbsOrtbBid{
						{
							Bid: &openrtb2.Bid{
								ID: "bid-1::gen-1",
							},
							GeneratedBidID: "gen-1",
						},
						{
							Bid: &openrtb2.Bid{
								ID: "bid-2::gen-2",
							},
							GeneratedBidID: "gen-2",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := OpenWrap{
				cache: mockCache,
			}
			got, err := m.handleAllProcessedBidResponsesHook(tt.args.ctx, tt.args.moduleCtx, tt.args.payload)
			assert.Equal(t, tt.wantErr, err != nil, "handleAllProcessedBidResponsesHook() error = %v, wantErr %v", err, tt.wantErr)
			if tt.mutationApplied {
				mutations := got.ChangeSet.Mutations()
				assert.NotEmpty(t, mutations, tt.name)
				for _, mut := range mutations {
					result, err := mut.Apply(tt.args.payload)
					assert.Nil(t, err, tt.name)
					assert.Equal(t, tt.wantResponse, result.Responses, tt.name)
				}
			}
			assert.Equal(t, tt.want.DebugMessages, got.DebugMessages, "Debug messages should be equal")
			assert.Equal(t, tt.want.Reject, false, "Reject should be equal")
		})
	}
}
