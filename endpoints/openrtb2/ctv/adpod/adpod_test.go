package adpod

import (
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
)

func TestAddSeatNonBids(t *testing.T) {
	type args struct {
		bids []*types.Bid
	}
	tests := []struct {
		name string
		args args
		want openrtb_ext.NonBidCollection
	}{
		{
			name: "Empty NonBid Collection",
			args: args{
				bids: []*types.Bid{},
			},
			want: openrtb_ext.NonBidCollection{},
		},
		{
			name: "Winning and Nonwinning Bid",
			args: args{
				bids: []*types.Bid{
					{
						Bid: &openrtb2.Bid{
							ID:    "BID-1",
							Price: 10,
						},
						ExtBid: openrtb_ext.ExtBid{
							Prebid: &openrtb_ext.ExtBidPrebid{
								Meta: &openrtb_ext.ExtBidPrebidMeta{
									AdapterCode: "pubmatic",
								},
								Type: "video",
							},
							OriginalBidCPM:    10,
							OriginalBidCur:    "USD",
							OriginalBidCPMUSD: 10,
						},
						DealTierSatisfied: false,
						Seat:              "pubmatic",
						Nbr:               ptrutil.ToPtr(nbr.LossBidLostToHigherBid),
					},
					{
						Bid: &openrtb2.Bid{
							ID:    "BID-2",
							Price: 15,
						},
						ExtBid: openrtb_ext.ExtBid{
							Prebid: &openrtb_ext.ExtBidPrebid{
								Meta: &openrtb_ext.ExtBidPrebidMeta{
									AdapterCode: "pubmatic",
								},
							},
						},
						DealTierSatisfied: false,
						Seat:              "pubmatic",
					},
				},
			},
			want: func() openrtb_ext.NonBidCollection {
				seatNonBid := openrtb_ext.NonBidCollection{}
				nonBid := openrtb_ext.NewNonBid(openrtb_ext.NonBidParams{
					Bid:               &openrtb2.Bid{ID: "BID-1", Price: 10},
					NonBidReason:      501,
					OriginalBidCPM:    10,
					OriginalBidCur:    "USD",
					BidType:           "video",
					OriginalBidCPMUSD: 10,
					BidMeta: &openrtb_ext.ExtBidPrebidMeta{
						AdapterCode: "pubmatic",
					},
				})
				seatNonBid.AddBid(nonBid, "pubmatic")
				return seatNonBid
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snb := addSeatNonBids(tt.args.bids)
			assert.Equal(t, snb, tt.want)
		})
	}
}
