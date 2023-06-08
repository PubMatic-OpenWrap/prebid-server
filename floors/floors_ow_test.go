package floors

import (
	"reflect"
	"testing"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/exchange/entities"
)

func TestRequestHasFloors(t *testing.T) {

	tests := []struct {
		name       string
		bidRequest *openrtb2.BidRequest
		want       bool
	}{
		{
			bidRequest: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
				},
				Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
			},
			want: false,
		},
		{
			bidRequest: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
				},
				Imp: []openrtb2.Imp{{ID: "1234", BidFloor: 10, BidFloorCur: "USD", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RequestHasFloors(tt.bidRequest); got != tt.want {
				t.Errorf("RequestHasFloors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPbsOrtbBidToAnalyticsRejectedBid(t *testing.T) {
	type args struct {
		pbsRejSeatBids []*entities.PbsOrtbSeatBid
	}
	tests := []struct {
		name string
		args args
		want []analytics.RejectedBid
	}{
		{
			name: "empty PbsRejSeatBids",
			args: args{
				pbsRejSeatBids: []*entities.PbsOrtbSeatBid{},
			},
			want: nil,
		},
		{
			name: "Multiple Bids with DealId and without DealId",
			args: args{
				pbsRejSeatBids: []*entities.PbsOrtbSeatBid{
					{
						Bids: []*entities.PbsOrtbBid{
							{
								Bid: &openrtb2.Bid{
									DealID: "123",
								},
							},
							{
								Bid: &openrtb2.Bid{},
							},
							{
								Bid: &openrtb2.Bid{
									DealID: "1234",
								},
							},
						},
						Seat: "xandrr",
					},
				},
			},
			want: []analytics.RejectedBid{
				{
					Bid: &entities.PbsOrtbBid{
						Bid: &openrtb2.Bid{
							DealID: "123",
						},
					},
					RejectionReason: openrtb3.LossBidBelowDealFloor,
					Seat:            "xandrr",
				},
				{
					Bid: &entities.PbsOrtbBid{
						Bid: &openrtb2.Bid{},
					},
					RejectionReason: openrtb3.LossBidBelowAuctionFloor,
					Seat:            "xandrr",
				},
				{
					Bid: &entities.PbsOrtbBid{
						Bid: &openrtb2.Bid{
							DealID: "1234",
						},
					},
					RejectionReason: openrtb3.LossBidBelowDealFloor,
					Seat:            "xandrr",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PbsOrtbBidToAnalyticsRejectedBid(tt.args.pbsRejSeatBids); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PbsOrtbBidToAnalyticsRejectedBid() = %v, want %v", got, tt.want)
			}
		})
	}
}
