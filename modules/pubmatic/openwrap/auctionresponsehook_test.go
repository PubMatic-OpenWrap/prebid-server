package openwrap

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/hooks/hookstage"
	mock_metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestSeatNonBidsInHandleAuctionResponseHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx       context.Context
		moduleCtx hookstage.ModuleInvocationContext
		payload   hookstage.AuctionResponsePayload
	}

	type want struct {
		bidResponseExt json.RawMessage
		err            error
	}

	tests := []struct {
		name             string
		args             args
		want             want
		getMetricsEngine func() *mock_metrics.MockMetricsEngine
	}{
		{
			name: "returnallbidstatus_true",
			args: args{
				ctx: nil,
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							StartTime:          time.Now().UnixMilli(),
							ReturnAllBidStatus: true,
							ImpBidCtx: map[string]models.ImpCtx{
								"imp1": {},
							},
							AdapterThrottleMap: map[string]struct{}{
								"pubmatic": {},
							},
							PubIDStr: "5890",
						},
					},
				},
				payload: hookstage.AuctionResponsePayload{
					BidResponse: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{},
					},
				},
			},
			getMetricsEngine: func() (me *mock_metrics.MockMetricsEngine) {
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				mockEngine.EXPECT().RecordPublisherResponseTimeStats("5890", gomock.Any())
				mockEngine.EXPECT().RecordNobidErrPrebidServerResponse("5890")
				return mockEngine
			},
			want: want{
				bidResponseExt: json.RawMessage(`{"prebid":{"seatnonbid":[{"nonbid":[{"impid":"imp1","statuscode":504,"ext":{"prebid":{"bid":{"id":""}}}}],"seat":"pubmatic","ext":null}]},"matchedimpression":{}}`),
			},
		},
		{
			name: "returnallbidstatus_false",
			args: args{
				ctx: nil,
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							StartTime:          time.Now().UnixMilli(),
							ReturnAllBidStatus: false,
							ImpBidCtx: map[string]models.ImpCtx{
								"imp1": {},
							},
							AdapterThrottleMap: map[string]struct{}{
								"pubmatic": {},
							},
							PubIDStr: "5890",
						},
					},
				},
				payload: hookstage.AuctionResponsePayload{
					BidResponse: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{},
					},
				},
			},
			getMetricsEngine: func() (me *mock_metrics.MockMetricsEngine) {
				mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
				mockEngine.EXPECT().RecordPublisherResponseTimeStats("5890", gomock.Any())
				mockEngine.EXPECT().RecordNobidErrPrebidServerResponse("5890")
				return mockEngine
			},
			want: want{
				bidResponseExt: json.RawMessage(`{"matchedimpression":{}}`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := OpenWrap{
				metricEngine: tt.getMetricsEngine(),
			}
			hookResult, err := o.handleAuctionResponseHook(tt.args.ctx, tt.args.moduleCtx, tt.args.payload)
			assert.Equal(t, tt.want.err, err, tt.name)
			mutations := hookResult.ChangeSet.Mutations()
			assert.NotEmpty(t, mutations, tt.name)
			for _, mut := range mutations {
				result, err := mut.Apply(tt.args.payload)
				assert.Nil(t, err, tt.name)
				assert.Equal(t, tt.want.bidResponseExt, result.BidResponse.Ext, tt.name)
			}
		})
	}
}
