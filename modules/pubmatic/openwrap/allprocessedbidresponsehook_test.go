package openwrap

import (
	"testing"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/exchange/entities"
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
