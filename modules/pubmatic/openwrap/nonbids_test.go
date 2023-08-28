package openwrap

import (
	"testing"

	"github.com/prebid/prebid-server/exchange"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestPrepareSeatNonBids(t *testing.T) {

	type args struct {
		rctx models.RequestCtx
	}

	tests := []struct {
		name        string
		args        args
		seatNonBids map[string][]openrtb_ext.NonBid
	}{
		{
			name: "empty_impbidctx",
			args: args{
				rctx: models.RequestCtx{
					SeatNonBids: make(map[string][]openrtb_ext.NonBid),
				},
			},
			seatNonBids: make(map[string][]openrtb_ext.NonBid),
		},
		{
			name: "empty_seatnonbids",
			args: args{
				rctx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							ImpID: "imp1",
						},
					},
					SeatNonBids: make(map[string][]openrtb_ext.NonBid),
				},
			},
			seatNonBids: make(map[string][]openrtb_ext.NonBid),
		},
		{
			name: "partner_throttled_nonbids",
			args: args{
				rctx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							ImpID: "imp1",
						},
					},
					AdapterThrottleMap: map[string]struct{}{
						"pubmatic": {},
					},
					SeatNonBids: map[string][]openrtb_ext.NonBid{},
				},
			},
			seatNonBids: map[string][]openrtb_ext.NonBid{
				"pubmatic": {
					openrtb_ext.NonBid{
						ImpId:      "imp1",
						StatusCode: int(exchange.RequestBlockedPartnerThrottle),
					},
				},
			},
		},
		{
			name: "slot_not_mapped_nonbids",
			args: args{
				rctx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							NonMapped: map[string]struct{}{
								"pubmatic": {},
								"appnexus": {},
							},
						},
					},
					SeatNonBids: map[string][]openrtb_ext.NonBid{
						"pubmatic": {
							{
								ImpId:      "imp2",
								StatusCode: 2,
							},
						},
					},
				},
			},
			seatNonBids: map[string][]openrtb_ext.NonBid{
				"pubmatic": {
					{
						ImpId:      "imp2",
						StatusCode: 2,
					},
					{
						ImpId:      "imp1",
						StatusCode: int(exchange.RequestBlockedSlotNotMapped),
					},
				},
				"appnexus": {
					{
						ImpId:      "imp1",
						StatusCode: int(exchange.RequestBlockedSlotNotMapped),
					},
				},
			},
		},
		{
			name: "slot_not_mapped_plus_partner_throttled_nonbids",
			args: args{
				rctx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							NonMapped: map[string]struct{}{
								"pubmatic": {},
							},
						},
						"imp2": {},
					},
					AdapterThrottleMap: map[string]struct{}{
						"appnexus": {},
					},
					SeatNonBids: map[string][]openrtb_ext.NonBid{
						"pubmatic": {
							{
								ImpId:      "imp2",
								StatusCode: 2,
							},
						},
					},
				},
			},
			seatNonBids: map[string][]openrtb_ext.NonBid{
				"pubmatic": {
					{
						ImpId:      "imp2",
						StatusCode: 2,
					},
					{
						ImpId:      "imp1",
						StatusCode: int(exchange.RequestBlockedSlotNotMapped),
					},
				},
				"appnexus": {
					{
						ImpId:      "imp2",
						StatusCode: int(exchange.RequestBlockedPartnerThrottle),
					},
					{
						ImpId:      "imp1",
						StatusCode: int(exchange.RequestBlockedPartnerThrottle),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prepareSeatNonBids(tt.args.rctx)
			assert.Equal(t, tt.seatNonBids, tt.args.rctx.SeatNonBids, tt.name)
		})
	}
}

func TestAddSeatNonBidsInResponseExt(t *testing.T) {

	type args struct {
		rctx        models.RequestCtx
		responseExt *openrtb_ext.ExtBidResponse
	}

	tests := []struct {
		name string
		args args
		want *openrtb_ext.ExtBidResponse
	}{
		{
			name: "empty_rtcx_seatnonbids",
			args: args{
				rctx: models.RequestCtx{},
				responseExt: &openrtb_ext.ExtBidResponse{
					Prebid: nil,
				},
			},
			want: &openrtb_ext.ExtBidResponse{
				Prebid: nil,
			},
		},
		{
			name: "prebid_exist_but_seatnonbid_is_empty_in_ext",
			args: args{
				rctx: models.RequestCtx{
					SeatNonBids: map[string][]openrtb_ext.NonBid{
						"pubmatic": {
							openrtb_ext.NonBid{
								ImpId:      "imp1",
								StatusCode: 1,
							},
							openrtb_ext.NonBid{
								ImpId:      "imp2",
								StatusCode: 2,
							},
						},
					},
				},
				responseExt: &openrtb_ext.ExtBidResponse{
					Prebid: &openrtb_ext.ExtResponsePrebid{
						AuctionTimestamp: 100,
					},
				},
			},
			want: &openrtb_ext.ExtBidResponse{
				Prebid: &openrtb_ext.ExtResponsePrebid{
					AuctionTimestamp: 100,
					SeatNonBid: []openrtb_ext.SeatNonBid{
						{
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId:      "imp1",
									StatusCode: 1,
								},
								{
									ImpId:      "imp2",
									StatusCode: 2,
								},
							},
							Seat: "pubmatic",
						},
					},
				},
			},
		},
		{
			name: "nonbid_exist_in_rctx_and_in_ext_for_specific_bidder",
			args: args{
				rctx: models.RequestCtx{
					SeatNonBids: map[string][]openrtb_ext.NonBid{
						"pubmatic": {
							openrtb_ext.NonBid{
								ImpId:      "imp1",
								StatusCode: 1,
							},
						},
					},
				},
				responseExt: &openrtb_ext.ExtBidResponse{
					Prebid: &openrtb_ext.ExtResponsePrebid{
						AuctionTimestamp: 100,
						SeatNonBid: []openrtb_ext.SeatNonBid{
							{
								Seat: "pubmatic",
								NonBid: []openrtb_ext.NonBid{
									{
										ImpId:      "imp2",
										StatusCode: 2,
									},
								},
							},
							{
								Seat: "appnexus",
								NonBid: []openrtb_ext.NonBid{
									{
										ImpId:      "imp1",
										StatusCode: 1,
									},
								},
							},
						},
					},
				},
			},
			want: &openrtb_ext.ExtBidResponse{
				Prebid: &openrtb_ext.ExtResponsePrebid{
					AuctionTimestamp: 100,
					SeatNonBid: []openrtb_ext.SeatNonBid{
						{
							Seat: "pubmatic",
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId:      "imp2",
									StatusCode: 2,
								},
								{
									ImpId:      "imp1",
									StatusCode: 1,
								},
							},
						},
						{
							Seat: "appnexus",
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId:      "imp1",
									StatusCode: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "nonbid_exist_in_rctx_but_not_in_ext_for_specific_bidder",
			args: args{
				rctx: models.RequestCtx{
					SeatNonBids: map[string][]openrtb_ext.NonBid{
						"pubmatic": {
							openrtb_ext.NonBid{
								ImpId:      "imp1",
								StatusCode: 1,
							},
						},
						"appnexus": {
							openrtb_ext.NonBid{
								ImpId:      "imp1",
								StatusCode: 1,
							},
						},
					},
				},
				responseExt: &openrtb_ext.ExtBidResponse{
					Prebid: &openrtb_ext.ExtResponsePrebid{
						AuctionTimestamp: 100,
						SeatNonBid: []openrtb_ext.SeatNonBid{
							{
								Seat: "pubmatic",
								NonBid: []openrtb_ext.NonBid{
									{
										ImpId:      "imp2",
										StatusCode: 2,
									},
								},
							},
						},
					},
				},
			},
			want: &openrtb_ext.ExtBidResponse{
				Prebid: &openrtb_ext.ExtResponsePrebid{
					AuctionTimestamp: 100,
					SeatNonBid: []openrtb_ext.SeatNonBid{
						{
							Seat: "pubmatic",
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId:      "imp2",
									StatusCode: 2,
								},
								{
									ImpId:      "imp1",
									StatusCode: 1,
								},
							},
						},
						{
							Seat: "appnexus",
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId:      "imp1",
									StatusCode: 1,
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addSeatNonBidsInResponseExt(tt.args.rctx, tt.args.responseExt)
			assert.Equal(t, tt.want, tt.args.responseExt, tt.name)
		})
	}
}
