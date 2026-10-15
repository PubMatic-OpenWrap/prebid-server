package cloudx

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestApplyCloudXBidExpIfMissing(t *testing.T) {
	tests := []struct {
		name       string
		rctx       models.RequestCtx
		bid        *openrtb2.BidResponse
		wantBidExp int64
	}{
		{
			name: "copies_imp_exp_when_bid_exp_missing",
			rctx: models.RequestCtx{
				ImpBidCtx: map[string]models.ImpCtx{"imp-1": {Exp: 600}},
			},
			bid: &openrtb2.BidResponse{SeatBid: []openrtb2.SeatBid{{
				Bid: []openrtb2.Bid{{ID: "b1", ImpID: "imp-1", Exp: 0}},
			}}},
			wantBidExp: 600,
		},
		{
			name: "preserves_existing_bid_exp",
			rctx: models.RequestCtx{
				ImpBidCtx: map[string]models.ImpCtx{"imp-2": {Exp: 3600}},
			},
			bid: &openrtb2.BidResponse{SeatBid: []openrtb2.SeatBid{{
				Bid: []openrtb2.Bid{{ID: "b2", ImpID: "imp-2", Exp: 120}},
			}}},
			wantBidExp: 120,
		},
		{
			name: "unknown_imp_defaults_to_3000",
			rctx: models.RequestCtx{
				ImpBidCtx: map[string]models.ImpCtx{"imp-1": {Exp: 600}},
			},
			bid: &openrtb2.BidResponse{SeatBid: []openrtb2.SeatBid{{
				Bid: []openrtb2.Bid{{ID: "b3", ImpID: "imp-3", Exp: 0}},
			}}},
			wantBidExp: cloudXDefaultExp,
		},
		{
			name: "missing_imp_exp_defaults_to_3000",
			rctx: models.RequestCtx{
				ImpBidCtx: map[string]models.ImpCtx{"imp-4": {Exp: 0}},
			},
			bid: &openrtb2.BidResponse{SeatBid: []openrtb2.SeatBid{{
				Bid: []openrtb2.Bid{{ID: "b4", ImpID: "imp-4", Exp: 0}},
			}}},
			wantBidExp: cloudXDefaultExp,
		},
		{
			name: "nil_bid_response",
			rctx: models.RequestCtx{
				ImpBidCtx: map[string]models.ImpCtx{"imp-1": {Exp: 600}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyCloudXBidExpIfMissing(tt.rctx, tt.bid)
			if tt.bid == nil {
				return
			}
			assert.Equal(t, tt.wantBidExp, tt.bid.SeatBid[0].Bid[0].Exp)
		})
	}
}
