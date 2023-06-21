package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/adapters/vastbidder"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/exchange/entities"
	"github.com/prebid/prebid-server/metrics"
	metricsConf "github.com/prebid/prebid-server/metrics/config"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

// TestApplyAdvertiserBlocking verifies advertiser blocking
// Currently it is expected to work only with TagBidders and not woth
// normal bidders
func TestApplyAdvertiserBlocking(t *testing.T) {
	type args struct {
		advBlockReq     *AuctionRequest
		adaptorSeatBids map[*bidderAdapter]*entities.PbsOrtbSeatBid // bidder adaptor and its dummy seat bids map
	}
	type want struct {
		rejectedBidIds       []string
		validBidCountPerSeat map[string]int
		expectedRejectedBids []analytics.RejectedBid
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "reject_bid_of_blocked_adv_from_tag_bidder",
			args: args{
				advBlockReq: &AuctionRequest{
					BidRequestWrapper: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{
							BAdv: []string{"a.com"}, // block bids returned by a.com
						},
					},
					LoggableObject: &analytics.LoggableAuctionObject{},
				},
				adaptorSeatBids: map[*bidderAdapter]*entities.PbsOrtbSeatBid{
					newTestTagAdapter("vast_tag_bidder"): { // tag bidder returning 1 bid from blocked advertiser
						Bids: []*entities.PbsOrtbBid{
							{
								Bid: &openrtb2.Bid{
									ID:      "a.com_bid",
									ADomain: []string{"a.com"},
								},
							},
							{
								Bid: &openrtb2.Bid{
									ID:      "b.com_bid",
									ADomain: []string{"b.com"},
								},
							},
							{
								Bid: &openrtb2.Bid{
									ID:      "keep_ba.com",
									ADomain: []string{"ba.com"},
								},
							},
							{
								Bid: &openrtb2.Bid{
									ID:      "keep_ba.com",
									ADomain: []string{"b.a.com.shri.com"},
								},
							},
							{
								Bid: &openrtb2.Bid{
									ID:      "reject_b.a.com.a.com.b.c.d.a.com",
									ADomain: []string{"b.a.com.a.com.b.c.d.a.com"},
								},
							},
						},
						BidderCoreName: openrtb_ext.BidderVASTBidder,
					},
				},
			},
			want: want{
				expectedRejectedBids: []analytics.RejectedBid{
					{
						RejectionReason: openrtb3.LossBidAdvertiserBlocking,
						Bid: &entities.PbsOrtbBid{Bid: &openrtb2.Bid{
							ID:      "a.com_bid",
							ADomain: []string{"a.com"},
						}},
						Seat: "",
					},
					{
						RejectionReason: openrtb3.LossBidAdvertiserBlocking,
						Bid: &entities.PbsOrtbBid{Bid: &openrtb2.Bid{
							ID:      "reject_b.a.com.a.com.b.c.d.a.com",
							ADomain: []string{"b.a.com.a.com.b.c.d.a.com"},
						}},
						Seat: "",
					},
				},
				rejectedBidIds: []string{"a.com_bid", "reject_b.a.com.a.com.b.c.d.a.com"},
				validBidCountPerSeat: map[string]int{
					"vast_tag_bidder": 3,
				},
			},
		},
		{
			name: "Badv_is_not_present", // expect no advertiser blocking
			args: args{
				advBlockReq: &AuctionRequest{
					BidRequestWrapper: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{BAdv: nil},
					},
					LoggableObject: &analytics.LoggableAuctionObject{},
				},
				adaptorSeatBids: map[*bidderAdapter]*entities.PbsOrtbSeatBid{
					newTestTagAdapter("tab_bidder_1"): {
						Bids: []*entities.PbsOrtbBid{
							{Bid: &openrtb2.Bid{ID: "bid_1_adapter_1", ADomain: []string{"a.com"}}},
							{Bid: &openrtb2.Bid{ID: "bid_2_adapter_1"}},
						},
					},
				},
			},
			want: want{
				rejectedBidIds: []string{}, // no bid rejection expected
				validBidCountPerSeat: map[string]int{
					"tab_bidder_1": 2,
				},
				expectedRejectedBids: []analytics.RejectedBid{},
			},
		},
		{
			name: "adomain_is_not_present_but_Badv_is_set", // reject bids without adomain as badv is set
			args: args{
				advBlockReq: &AuctionRequest{
					BidRequestWrapper: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{BAdv: []string{"advertiser_1.com"}},
					},
					LoggableObject: &analytics.LoggableAuctionObject{},
				},
				adaptorSeatBids: map[*bidderAdapter]*entities.PbsOrtbSeatBid{
					newTestTagAdapter("tag_bidder_1"): {
						Bids: []*entities.PbsOrtbBid{ // expect all bids are rejected
							{Bid: &openrtb2.Bid{ID: "bid_1_adapter_1_without_adomain"}},
							{Bid: &openrtb2.Bid{ID: "bid_2_adapter_1_with_empty_adomain", ADomain: []string{"", " "}}},
						},
					},
					newTestRtbAdapter("rtb_bidder_1"): {
						Bids: []*entities.PbsOrtbBid{ // all bids should be present. It belongs to RTB adapator
							{Bid: &openrtb2.Bid{ID: "bid_1_adapter_2_without_adomain"}},
							{Bid: &openrtb2.Bid{ID: "bid_2_adapter_2_with_empty_adomain", ADomain: []string{"", " "}}},
						},
					},
				},
			},
			want: want{
				rejectedBidIds: []string{"bid_1_adapter_1_without_adomain", "bid_2_adapter_1_with_empty_adomain"},
				validBidCountPerSeat: map[string]int{
					"tag_bidder_1": 0, // expect 0 bids. i.e. all bids are rejected
					"rtb_bidder_1": 2, // no bid must be rejected
				},
				expectedRejectedBids: []analytics.RejectedBid{},
			},
		},
		{
			name: "adomain_and_badv_is_not_present", // expect no advertiser blocking
			args: args{
				advBlockReq: &AuctionRequest{
					BidRequestWrapper: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{},
					},
					LoggableObject: &analytics.LoggableAuctionObject{},
				},
				adaptorSeatBids: map[*bidderAdapter]*entities.PbsOrtbSeatBid{
					newTestTagAdapter("tag_adaptor_1"): {
						Bids: []*entities.PbsOrtbBid{
							{Bid: &openrtb2.Bid{ID: "bid_without_adomain"}},
						},
					},
				},
			},
			want: want{
				rejectedBidIds: []string{}, // no rejection expected as badv not present
				validBidCountPerSeat: map[string]int{
					"tag_adaptor_1": 1,
				},
				expectedRejectedBids: []analytics.RejectedBid{},
			},
		},
		{
			name: "empty_badv", // expect no advertiser blocking
			args: args{
				advBlockReq: &AuctionRequest{
					BidRequestWrapper: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{BAdv: []string{}},
					},
					LoggableObject: &analytics.LoggableAuctionObject{},
				},
				adaptorSeatBids: map[*bidderAdapter]*entities.PbsOrtbSeatBid{
					newTestTagAdapter("tag_bidder_1"): {
						Bids: []*entities.PbsOrtbBid{ // expect all bids are rejected
							{Bid: &openrtb2.Bid{ID: "bid_1_adapter_1", ADomain: []string{"a.com"}}},
							{Bid: &openrtb2.Bid{ID: "bid_2_adapter_1"}},
						},
					},
					newTestRtbAdapter("rtb_bidder_1"): {
						Bids: []*entities.PbsOrtbBid{ // all bids should be present. It belongs to RTB adapator
							{Bid: &openrtb2.Bid{ID: "bid_1_adapter_2_without_adomain"}},
							{Bid: &openrtb2.Bid{ID: "bid_2_adapter_2_with_empty_adomain", ADomain: []string{"", " "}}},
						},
					},
				},
			},
			want: want{
				rejectedBidIds: []string{}, // no rejections expect as there is not badv set
				validBidCountPerSeat: map[string]int{
					"tag_bidder_1": 2,
					"rtb_bidder_1": 2,
				},
				expectedRejectedBids: []analytics.RejectedBid{},
			},
		},
		{
			name: "nil_badv", // expect no advertiser blocking
			args: args{
				advBlockReq: &AuctionRequest{
					BidRequestWrapper: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{BAdv: nil},
					},
					LoggableObject: &analytics.LoggableAuctionObject{},
				},
				adaptorSeatBids: map[*bidderAdapter]*entities.PbsOrtbSeatBid{
					newTestTagAdapter("tag_bidder_1"): {
						Bids: []*entities.PbsOrtbBid{ // expect all bids are rejected
							{Bid: &openrtb2.Bid{ID: "bid_1_adapter_1", ADomain: []string{"a.com"}}},
							{Bid: &openrtb2.Bid{ID: "bid_2_adapter_1"}},
						},
					},
					newTestRtbAdapter("rtb_bidder_1"): {
						Bids: []*entities.PbsOrtbBid{ // all bids should be present. It belongs to RTB adapator
							{Bid: &openrtb2.Bid{ID: "bid_1_adapter_2_without_adomain"}},
							{Bid: &openrtb2.Bid{ID: "bid_2_adapter_2_with_empty_adomain", ADomain: []string{"", " "}}},
						},
					},
				},
			},
			want: want{
				rejectedBidIds: []string{}, // no rejections expect as there is not badv set
				validBidCountPerSeat: map[string]int{
					"tag_bidder_1": 2,
					"rtb_bidder_1": 2,
				},
				expectedRejectedBids: []analytics.RejectedBid{},
			},
		},
		{
			name: "ad_domains_normalized_and_checked",
			args: args{
				advBlockReq: &AuctionRequest{
					BidRequestWrapper: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{BAdv: []string{"a.com"}},
					},
					LoggableObject: &analytics.LoggableAuctionObject{},
				},
				adaptorSeatBids: map[*bidderAdapter]*entities.PbsOrtbSeatBid{
					newTestTagAdapter("my_adapter"): {
						Bids: []*entities.PbsOrtbBid{
							{Bid: &openrtb2.Bid{ID: "bid_1_of_blocked_adv", ADomain: []string{"www.a.com"}}},
							// expect a.com is extracted from page url
							{Bid: &openrtb2.Bid{ID: "bid_2_of_blocked_adv", ADomain: []string{"http://a.com/my/page?k1=v1&k2=v2"}}},
							// invalid adomain - will be skipped and the bid will be not be rejected
							{Bid: &openrtb2.Bid{ID: "bid_3_with_domain_abcd1234", ADomain: []string{"abcd1234"}}},
						},
					}},
			},
			want: want{
				rejectedBidIds:       []string{"bid_1_of_blocked_adv", "bid_2_of_blocked_adv"},
				validBidCountPerSeat: map[string]int{"my_adapter": 1},
				expectedRejectedBids: []analytics.RejectedBid{},
			},
		}, {
			name: "multiple_badv",
			args: args{
				advBlockReq: &AuctionRequest{
					BidRequestWrapper: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{BAdv: []string{"advertiser_1.com", "advertiser_2.com", "www.advertiser_3.com"}},
					},
					LoggableObject: &analytics.LoggableAuctionObject{},
				},
				adaptorSeatBids: map[*bidderAdapter]*entities.PbsOrtbSeatBid{
					newTestTagAdapter("tag_adapter_1"): {
						Bids: []*entities.PbsOrtbBid{
							// adomain without www prefix
							{Bid: &openrtb2.Bid{ID: "bid_1_tag_adapter_1", ADomain: []string{"advertiser_3.com"}}},
							{Bid: &openrtb2.Bid{ID: "bid_2_tag_adapter_1", ADomain: []string{"advertiser_2.com"}}},
							{Bid: &openrtb2.Bid{ID: "bid_3_tag_adapter_1", ADomain: []string{"advertiser_4.com"}}},
							{Bid: &openrtb2.Bid{ID: "bid_4_tag_adapter_1", ADomain: []string{"advertiser_100.com"}}},
						},
					},
					newTestTagAdapter("tag_adapter_2"): {
						Bids: []*entities.PbsOrtbBid{
							// adomain has www prefix
							{Bid: &openrtb2.Bid{ID: "bid_1_tag_adapter_2", ADomain: []string{"www.advertiser_1.com"}}},
						},
					},
					newTestRtbAdapter("rtb_adapter_1"): {
						Bids: []*entities.PbsOrtbBid{
							// should not reject following bid though its advertiser is blocked
							// because this bid belongs to RTB Adaptor
							{Bid: &openrtb2.Bid{ID: "bid_1_rtb_adapter_2", ADomain: []string{"advertiser_1.com"}}},
						},
					},
				},
			},
			want: want{
				rejectedBidIds: []string{"bid_1_tag_adapter_1", "bid_2_tag_adapter_1", "bid_1_tag_adapter_2"},
				validBidCountPerSeat: map[string]int{
					"tag_adapter_1": 2,
					"tag_adapter_2": 0,
					"rtb_adapter_1": 1,
				},
				expectedRejectedBids: []analytics.RejectedBid{},
			},
		}, {
			name: "multiple_adomain",
			args: args{
				advBlockReq: &AuctionRequest{
					BidRequestWrapper: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{BAdv: []string{"www.advertiser_3.com"}},
					},
					LoggableObject: &analytics.LoggableAuctionObject{},
				},
				adaptorSeatBids: map[*bidderAdapter]*entities.PbsOrtbSeatBid{
					newTestTagAdapter("tag_adapter_1"): {
						Bids: []*entities.PbsOrtbBid{
							// adomain without www prefix
							{Bid: &openrtb2.Bid{ID: "bid_1_tag_adapter_1", ADomain: []string{"a.com", "b.com", "advertiser_3.com", "d.com"}}},
							{Bid: &openrtb2.Bid{ID: "bid_2_tag_adapter_1", ADomain: []string{"a.com", "https://advertiser_3.com"}}},
							{Bid: &openrtb2.Bid{ID: "bid_3_tag_adapter_1", ADomain: []string{"advertiser_4.com"}}},
							{Bid: &openrtb2.Bid{ID: "bid_4_tag_adapter_1", ADomain: []string{"advertiser_100.com"}}},
						},
					},
					newTestTagAdapter("tag_adapter_2"): {
						Bids: []*entities.PbsOrtbBid{
							// adomain has www prefix
							{Bid: &openrtb2.Bid{ID: "bid_1_tag_adapter_2", ADomain: []string{"a.com", "b.com", "www.advertiser_3.com"}}},
						},
					},
					newTestRtbAdapter("rtb_adapter_1"): {
						Bids: []*entities.PbsOrtbBid{
							// should not reject following bid though its advertiser is blocked
							// because this bid belongs to RTB Adaptor
							{Bid: &openrtb2.Bid{ID: "bid_1_rtb_adapter_2", ADomain: []string{"a.com", "b.com", "advertiser_3.com"}}},
						},
					},
				},
			},
			want: want{
				rejectedBidIds: []string{"bid_1_tag_adapter_1", "bid_2_tag_adapter_1", "bid_1_tag_adapter_2"},
				validBidCountPerSeat: map[string]int{
					"tag_adapter_1": 2,
					"tag_adapter_2": 0,
					"rtb_adapter_1": 1,
				},
				expectedRejectedBids: []analytics.RejectedBid{},
			},
		}, {
			name: "case_insensitive_badv", // case of domain not matters
			args: args{
				advBlockReq: &AuctionRequest{
					BidRequestWrapper: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{BAdv: []string{"ADVERTISER_1.COM"}},
					},
					LoggableObject: &analytics.LoggableAuctionObject{},
				},
				adaptorSeatBids: map[*bidderAdapter]*entities.PbsOrtbSeatBid{
					newTestTagAdapter("tag_adapter_1"): {
						Bids: []*entities.PbsOrtbBid{
							{Bid: &openrtb2.Bid{ID: "bid_1_rtb_adapter_1", ADomain: []string{"advertiser_1.com"}}},
							{Bid: &openrtb2.Bid{ID: "bid_2_rtb_adapter_1", ADomain: []string{"www.advertiser_1.com"}}},
						},
					},
				},
			},
			want: want{
				rejectedBidIds: []string{"bid_1_rtb_adapter_1", "bid_2_rtb_adapter_1"},
				validBidCountPerSeat: map[string]int{
					"tag_adapter_1": 0, // expect all bids are rejected as belongs to blocked advertiser
				},
				expectedRejectedBids: []analytics.RejectedBid{},
			},
		},
		{
			name: "case_insensitive_adomain",
			args: args{
				advBlockReq: &AuctionRequest{
					BidRequestWrapper: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{BAdv: []string{"advertiser_1.com"}},
					},
					LoggableObject: &analytics.LoggableAuctionObject{},
				},
				adaptorSeatBids: map[*bidderAdapter]*entities.PbsOrtbSeatBid{
					newTestTagAdapter("tag_adapter_1"): {
						Bids: []*entities.PbsOrtbBid{
							{Bid: &openrtb2.Bid{ID: "bid_1_rtb_adapter_1", ADomain: []string{"advertiser_1.COM"}}},
							{Bid: &openrtb2.Bid{ID: "bid_2_rtb_adapter_1", ADomain: []string{"wWw.ADVERTISER_1.com"}}},
						},
					},
				},
			},
			want: want{
				rejectedBidIds: []string{"bid_1_rtb_adapter_1", "bid_2_rtb_adapter_1"},
				validBidCountPerSeat: map[string]int{
					"tag_adapter_1": 0, // expect all bids are rejected as belongs to blocked advertiser
				},
				expectedRejectedBids: []analytics.RejectedBid{},
			},
		},
		{
			name: "various_tld_combinations",
			args: args{
				advBlockReq: &AuctionRequest{
					BidRequestWrapper: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{BAdv: []string{"http://blockme.shri"}},
					},
					LoggableObject: &analytics.LoggableAuctionObject{},
				},
				adaptorSeatBids: map[*bidderAdapter]*entities.PbsOrtbSeatBid{
					newTestTagAdapter("block_bidder"): {
						Bids: []*entities.PbsOrtbBid{
							{Bid: &openrtb2.Bid{ADomain: []string{"www.blockme.shri"}, ID: "reject_www.blockme.shri"}},
							{Bid: &openrtb2.Bid{ADomain: []string{"http://www.blockme.shri"}, ID: "rejecthttp://www.blockme.shri"}},
							{Bid: &openrtb2.Bid{ADomain: []string{"https://blockme.shri"}, ID: "reject_https://blockme.shri"}},
							{Bid: &openrtb2.Bid{ADomain: []string{"https://www.blockme.shri"}, ID: "reject_https://www.blockme.shri"}},
						},
					},
					newTestRtbAdapter("rtb_non_block_bidder"): {
						Bids: []*entities.PbsOrtbBid{ // all below bids are eligible and should not be rejected
							{Bid: &openrtb2.Bid{ADomain: []string{"www.blockme.shri"}, ID: "accept_bid_www.blockme.shri"}},
							{Bid: &openrtb2.Bid{ADomain: []string{"http://www.blockme.shri"}, ID: "accept_bid__http://www.blockme.shri"}},
							{Bid: &openrtb2.Bid{ADomain: []string{"https://blockme.shri"}, ID: "accept_bid__https://blockme.shri"}},
							{Bid: &openrtb2.Bid{ADomain: []string{"https://www.blockme.shri"}, ID: "accept_bid__https://www.blockme.shri"}},
						},
					},
				},
			},
			want: want{
				rejectedBidIds: []string{"reject_www.blockme.shri", "reject_http://www.blockme.shri", "reject_https://blockme.shri", "reject_https://www.blockme.shri"},
				validBidCountPerSeat: map[string]int{
					"block_bidder":         0,
					"rtb_non_block_bidder": 4,
				},
				expectedRejectedBids: []analytics.RejectedBid{},
			},
		},
		{
			name: "subdomain_tests",
			args: args{
				advBlockReq: &AuctionRequest{
					BidRequestWrapper: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{BAdv: []string{"10th.college.puneunv.edu"}},
					},
					LoggableObject: &analytics.LoggableAuctionObject{},
				},

				adaptorSeatBids: map[*bidderAdapter]*entities.PbsOrtbSeatBid{
					newTestTagAdapter("block_bidder"): {
						Bids: []*entities.PbsOrtbBid{
							{Bid: &openrtb2.Bid{ADomain: []string{"shri.10th.college.puneunv.edu"}, ID: "reject_shri.10th.college.puneunv.edu"}},
							{Bid: &openrtb2.Bid{ADomain: []string{"puneunv.edu"}, ID: "allow_puneunv.edu"}},
							{Bid: &openrtb2.Bid{ADomain: []string{"http://WWW.123.456.10th.college.PUNEUNV.edu"}, ID: "reject_123.456.10th.college.puneunv.edu"}},
						},
					},
				},
			},
			want: want{
				rejectedBidIds: []string{"reject_shri.10th.college.puneunv.edu", "reject_123.456.10th.college.puneunv.edu"},
				validBidCountPerSeat: map[string]int{
					"block_bidder": 1,
				},
				expectedRejectedBids: []analytics.RejectedBid{},
			},
		}, {
			name: "only_domain_test", // do not expect bid rejection. edu is valid domain
			args: args{
				advBlockReq: &AuctionRequest{
					BidRequestWrapper: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{BAdv: []string{"edu"}},
					},
					LoggableObject: &analytics.LoggableAuctionObject{},
				},

				adaptorSeatBids: map[*bidderAdapter]*entities.PbsOrtbSeatBid{
					newTestTagAdapter("tag_bidder"): {
						Bids: []*entities.PbsOrtbBid{
							{Bid: &openrtb2.Bid{ADomain: []string{"school.edu"}, ID: "keep_bid_school.edu"}},
							{Bid: &openrtb2.Bid{ADomain: []string{"edu"}, ID: "keep_bid_edu"}},
							{Bid: &openrtb2.Bid{ADomain: []string{"..edu"}, ID: "keep_bid_..edu"}},
						},
					},
				},
			},
			want: want{
				rejectedBidIds: []string{},
				validBidCountPerSeat: map[string]int{
					"tag_bidder": 3,
				},
				expectedRejectedBids: []analytics.RejectedBid{},
			},
		},
		{
			name: "public_suffix_in_badv",
			args: args{
				advBlockReq: &AuctionRequest{
					BidRequestWrapper: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{BAdv: []string{"co.in"}},
					},
					LoggableObject: &analytics.LoggableAuctionObject{},
				},
				// co.in is valid public suffix
				adaptorSeatBids: map[*bidderAdapter]*entities.PbsOrtbSeatBid{
					newTestTagAdapter("tag_bidder"): {
						Bids: []*entities.PbsOrtbBid{
							{Bid: &openrtb2.Bid{ADomain: []string{"a.co.in"}, ID: "allow_a.co.in"}},
							{Bid: &openrtb2.Bid{ADomain: []string{"b.com"}, ID: "allow_b.com"}},
						},
					},
				},
			},
			want: want{
				rejectedBidIds: []string{},
				validBidCountPerSeat: map[string]int{
					"tag_bidder": 2,
				},
				expectedRejectedBids: []analytics.RejectedBid{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name != "reject_bid_of_blocked_adv_from_tag_bidder" {
				return
			}
			seatBids := make(map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid)
			tagBidders := make(map[openrtb_ext.BidderName]adapters.Bidder)
			adapterMap := make(map[openrtb_ext.BidderName]AdaptedBidder, 0)
			for adaptor, sbids := range tt.args.adaptorSeatBids {
				adapterMap[adaptor.BidderName] = adaptor
				if tagBidder, ok := adaptor.Bidder.(*vastbidder.TagBidder); ok {
					tagBidders[adaptor.BidderName] = tagBidder
				}
				seatBids[adaptor.BidderName] = sbids
			}

			// applyAdvertiserBlocking internally uses tagBidders from (adapter_map.go)
			// not testing alias here
			seatBids, rejections := applyAdvertiserBlocking(tt.args.advBlockReq, seatBids)
			re := regexp.MustCompile("bid rejected \\[bid ID:(.*?)\\] reason")
			for bidder, sBid := range seatBids {
				// verify only eligible bids are returned
				assert.Equal(t, tt.want.validBidCountPerSeat[string(bidder)], len(sBid.Bids), "Expected eligible bids are %d, but found [%d] ", tt.want.validBidCountPerSeat[string(bidder)], len(sBid.Bids))
				// verify  rejections
				assert.Equal(t, len(tt.want.rejectedBidIds), len(rejections), "Expected bid rejections are %d, but found [%d]", len(tt.want.rejectedBidIds), len(rejections))
				// verify rejected bid ids
				present := false
				for _, expectRejectedBidID := range tt.want.rejectedBidIds {
					for _, rejection := range rejections {
						match := re.FindStringSubmatch(rejection)
						rejectedBidID := strings.Trim(match[1], " ")
						if expectRejectedBidID == rejectedBidID {
							present = true
							break
						}
					}
					if present {
						break
					}
				}
				if len(tt.want.rejectedBidIds) > 0 && !present {
					assert.Fail(t, "Expected Bid ID [%s] as rejected. But bid is not rejected", re)
				}

				if sBid.BidderCoreName != openrtb_ext.BidderVASTBidder {
					continue // advertiser blocking is currently enabled only for tag bidders
				}

				sort.Slice(tt.args.advBlockReq.LoggableObject.RejectedBids, func(i, j int) bool {
					return tt.args.advBlockReq.LoggableObject.RejectedBids[i].Bid.Bid.ID > tt.args.advBlockReq.LoggableObject.RejectedBids[j].Bid.Bid.ID
				})
				sort.Slice(tt.want.expectedRejectedBids, func(i, j int) bool {
					return tt.want.expectedRejectedBids[i].Bid.Bid.ID > tt.want.expectedRejectedBids[j].Bid.Bid.ID
				})
				assert.Equal(t, tt.want.expectedRejectedBids, tt.args.advBlockReq.LoggableObject.RejectedBids, "Rejected Bids not matching")

				for _, bid := range sBid.Bids {
					if nil != bid.Bid.ADomain {
						for _, adomain := range bid.Bid.ADomain {
							for _, blockDomain := range tt.args.advBlockReq.BidRequestWrapper.BidRequest.BAdv {
								nDomain, _ := normalizeDomain(adomain)
								if nDomain == blockDomain {
									assert.Fail(t, "bid %s with ad domain %s is not blocked", bid.Bid.ID, adomain)
								}
							}
						}
					}

					// verify this bid not belongs to rejected list
					for _, rejectedBidID := range tt.want.rejectedBidIds {
						if rejectedBidID == bid.Bid.ID {
							assert.Fail(t, "Bid ID [%s] is not expected in list of rejected bids", bid.Bid.ID)
						}
					}
				}
			}
		})
	}
}

func TestNormalizeDomain(t *testing.T) {
	type args struct {
		domain string
	}
	type want struct {
		domain string
		err    error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{name: "a.com", args: args{domain: "a.com"}, want: want{domain: "a.com"}},
		{name: "http://a.com", args: args{domain: "http://a.com"}, want: want{domain: "a.com"}},
		{name: "https://a.com", args: args{domain: "https://a.com"}, want: want{domain: "a.com"}},
		{name: "https://www.a.com", args: args{domain: "https://www.a.com"}, want: want{domain: "a.com"}},
		{name: "https://www.a.com/my/page?k=1", args: args{domain: "https://www.a.com/my/page?k=1"}, want: want{domain: "a.com"}},
		{name: "empty_domain", args: args{domain: ""}, want: want{domain: ""}},
		{name: "trim_domain", args: args{domain: " trim.me?k=v    "}, want: want{domain: "trim.me"}},
		{name: "trim_domain_with_http_in_it", args: args{domain: " http://trim.me?k=v    "}, want: want{domain: "trim.me"}},
		{name: "https://www.something.a.com/my/page?k=1", args: args{domain: "https://www.something.a.com/my/page?k=1"}, want: want{domain: "something.a.com"}},
		{name: "wWW.something.a.com", args: args{domain: "wWW.something.a.com"}, want: want{domain: "something.a.com"}},
		{name: "2_times_www", args: args{domain: "www.something.www.a.com"}, want: want{domain: "something.www.a.com"}},
		{name: "consecutive_www", args: args{domain: "www.www.something.a.com"}, want: want{domain: "www.something.a.com"}},
		{name: "abchttp.com", args: args{domain: "abchttp.com"}, want: want{domain: "abchttp.com"}},
		{name: "HTTP://CAPS.com", args: args{domain: "HTTP://CAPS.com"}, want: want{domain: "caps.com"}},

		// publicsuffix
		{name: "co.in", args: args{domain: "co.in"}, want: want{domain: "", err: fmt.Errorf("domain [co.in] is public suffix")}},
		{name: ".co.in", args: args{domain: ".co.in"}, want: want{domain: ".co.in"}},
		{name: "amazon.co.in", args: args{domain: "amazon.co.in"}, want: want{domain: "amazon.co.in"}},
		// we wont check if shriprasad belongs to icann
		{name: "shriprasad", args: args{domain: "shriprasad"}, want: want{domain: "", err: fmt.Errorf("domain [shriprasad] is public suffix")}},
		{name: ".shriprasad", args: args{domain: ".shriprasad"}, want: want{domain: ".shriprasad"}},
		{name: "abc.shriprasad", args: args{domain: "abc.shriprasad"}, want: want{domain: "abc.shriprasad"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adjustedDomain, err := normalizeDomain(tt.args.domain)
			actualErr := "nil"
			expectedErr := "nil"
			if nil != err {
				actualErr = err.Error()
			}
			if nil != tt.want.err {
				actualErr = tt.want.err.Error()
			}
			assert.Equal(t, tt.want.err, err, "Expected error is %s, but found [%s]", expectedErr, actualErr)
			assert.Equal(t, tt.want.domain, adjustedDomain, "Expected domain is %s, but found [%s]", tt.want.domain, adjustedDomain)
		})
	}
}

func newTestTagAdapter(name string) *bidderAdapter {
	return &bidderAdapter{
		Bidder:     vastbidder.NewTagBidder(openrtb_ext.BidderName(name), config.Adapter{}),
		BidderName: openrtb_ext.BidderName(name),
	}
}

func newTestRtbAdapter(name string) *bidderAdapter {
	return &bidderAdapter{
		Bidder:     &goodSingleBidder{},
		BidderName: openrtb_ext.BidderName(name),
	}
}

func TestRecordAdaptorDuplicateBidIDs(t *testing.T) {
	type bidderCollisions = map[string]int
	testCases := []struct {
		scenario         string
		bidderCollisions *bidderCollisions // represents no of collisions detected for bid.id at bidder level for given request
		hasCollision     bool
	}{
		{scenario: "invalid collision value", bidderCollisions: &map[string]int{"bidder-1": -1}, hasCollision: false},
		{scenario: "no collision", bidderCollisions: &map[string]int{"bidder-1": 0}, hasCollision: false},
		{scenario: "one collision", bidderCollisions: &map[string]int{"bidder-1": 1}, hasCollision: false},
		{scenario: "multiple collisions", bidderCollisions: &map[string]int{"bidder-1": 2}, hasCollision: true}, // when 2 collisions it counter will be 1
		{scenario: "multiple bidders", bidderCollisions: &map[string]int{"bidder-1": 2, "bidder-2": 4}, hasCollision: true},
		{scenario: "multiple bidders with bidder-1 no collision", bidderCollisions: &map[string]int{"bidder-1": 1, "bidder-2": 4}, hasCollision: true},
		{scenario: "no bidders", bidderCollisions: nil, hasCollision: false},
	}
	testEngine := metricsConf.NewMetricsEngine(&config.Configuration{}, nil, nil, nil)

	for _, testcase := range testCases {
		var adapterBids map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid
		if nil == testcase.bidderCollisions {
			break
		}
		adapterBids = make(map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid)
		for bidder, collisions := range *testcase.bidderCollisions {
			bids := make([]*entities.PbsOrtbBid, 0)
			testBidID := "bid_id_for_bidder_" + bidder
			// add bids as per collisions value
			bidCount := 0
			for ; bidCount < collisions; bidCount++ {
				bids = append(bids, &entities.PbsOrtbBid{
					Bid: &openrtb2.Bid{
						ID: testBidID,
					},
				})
			}
			if nil == adapterBids[openrtb_ext.BidderName(bidder)] {
				adapterBids[openrtb_ext.BidderName(bidder)] = new(entities.PbsOrtbSeatBid)
			}
			adapterBids[openrtb_ext.BidderName(bidder)].Bids = bids
		}
		assert.Equal(t, testcase.hasCollision, recordAdaptorDuplicateBidIDs(testEngine, adapterBids))
	}
}

func TestMakeBidExtJSONOW(t *testing.T) {

	type aTest struct {
		description        string
		ext                json.RawMessage
		extBidPrebid       openrtb_ext.ExtBidPrebid
		impExtInfo         map[string]ImpExtInfo
		origbidcpm         float64
		origbidcur         string
		origbidcpmusd      float64
		expectedBidExt     string
		expectedErrMessage string
	}

	testCases := []aTest{
		{
			description:        "Valid extension with origbidcpmusd = 0",
			ext:                json.RawMessage(`{"video":{"h":100}}`),
			extBidPrebid:       openrtb_ext.ExtBidPrebid{Type: openrtb_ext.BidType("video"), Meta: &openrtb_ext.ExtBidPrebidMeta{BrandName: "foo"}, Passthrough: nil},
			impExtInfo:         map[string]ImpExtInfo{"test_imp_id": {true, []byte(`{"video":{"h":480,"mimes":["video/mp4"]}}`), json.RawMessage(`{"imp_passthrough_val": 1}`)}},
			origbidcpm:         10.0000,
			origbidcur:         "USD",
			expectedBidExt:     `{"prebid":{"meta": {"brandName": "foo"}, "passthrough":{"imp_passthrough_val":1}, "type":"video"}, "storedrequestattributes":{"h":480,"mimes":["video/mp4"]},"video":{"h":100}, "origbidcpm": 10, "origbidcur": "USD"}`,
			expectedErrMessage: "",
		},
		{
			description:        "Valid extension with origbidcpmusd > 0",
			ext:                json.RawMessage(`{"video":{"h":100}}`),
			extBidPrebid:       openrtb_ext.ExtBidPrebid{Type: openrtb_ext.BidType("video"), Meta: &openrtb_ext.ExtBidPrebidMeta{BrandName: "foo"}, Passthrough: nil},
			impExtInfo:         map[string]ImpExtInfo{"test_imp_id": {true, []byte(`{"video":{"h":480,"mimes":["video/mp4"]}}`), json.RawMessage(`{"imp_passthrough_val": 1}`)}},
			origbidcpm:         10.0000,
			origbidcur:         "USD",
			origbidcpmusd:      10.0000,
			expectedBidExt:     `{"prebid":{"meta": {"brandName": "foo"}, "passthrough":{"imp_passthrough_val":1}, "type":"video"}, "storedrequestattributes":{"h":480,"mimes":["video/mp4"]},"video":{"h":100}, "origbidcpm": 10, "origbidcur": "USD", "origbidcpmusd": 10}`,
			expectedErrMessage: "",
		},
	}

	for _, test := range testCases {
		result, err := makeBidExtJSON(test.ext, &test.extBidPrebid, test.impExtInfo, "test_imp_id", test.origbidcpm, test.origbidcur, test.origbidcpmusd)

		if test.expectedErrMessage == "" {
			assert.JSONEq(t, test.expectedBidExt, string(result), "Incorrect result")
			assert.NoError(t, err, "Error should not be returned")
		} else {
			assert.Contains(t, err.Error(), test.expectedErrMessage, "incorrect error message")
		}
	}
}

func TestUpdateRejectedBidExt(t *testing.T) {
	type args struct {
		loggableObject *analytics.LoggableAuctionObject
	}
	type want struct {
		loggableObject *analytics.LoggableAuctionObject
	}
	type test struct {
		name string
		args args
		want want
	}

	testCases := []test{
		{
			name: "nil rejected bid",
			args: args{
				loggableObject: &analytics.LoggableAuctionObject{
					RejectedBids: []analytics.RejectedBid{
						{
							Bid: &entities.PbsOrtbBid{Bid: &openrtb2.Bid{
								ID: "b1",
							},
							},
						},
					},
				},
			},
			want: want{
				loggableObject: &analytics.LoggableAuctionObject{
					RejectedBids: []analytics.RejectedBid{
						{
							Bid: &entities.PbsOrtbBid{Bid: &openrtb2.Bid{
								ID:  "b1",
								Ext: json.RawMessage(`{"origbidcpm":0,"prebid":{}}`),
							},
								OriginalBidCPM: 0,
								BidType:        "",
							},
						},
					},
				},
			},
		},
		{
			name: "malformed bid.ext json",
			args: args{
				loggableObject: &analytics.LoggableAuctionObject{
					RejectedBids: []analytics.RejectedBid{
						{
							Bid: &entities.PbsOrtbBid{Bid: &openrtb2.Bid{
								ID: "b1",
							},
							},
						},
					},
				},
			},
			want: want{
				loggableObject: &analytics.LoggableAuctionObject{
					RejectedBids: []analytics.RejectedBid{
						{
							Bid: &entities.PbsOrtbBid{Bid: &openrtb2.Bid{
								ID:  "b1",
								Ext: json.RawMessage(`{"origbidcpm":0,"prebid":{}}`),
							},
								OriginalBidCPM: 0,
								BidType:        "",
							},
						},
					},
				},
			},
		},
		{
			name: "valid pbsOrtbBid",
			args: args{
				loggableObject: &analytics.LoggableAuctionObject{
					RejectedBids: []analytics.RejectedBid{{
						Bid: &entities.PbsOrtbBid{
							Bid: &openrtb2.Bid{
								ID:  "b2",
								Ext: json.RawMessage(`{"key":"value"}`),
							},
							DealPriority:   10,
							OriginalBidCPM: 10,
							BidType:        openrtb_ext.BidTypeBanner,
						},
						Seat:            "pubmatic",
						RejectionReason: openrtb3.LossBidAdvertiserBlocking}},
				},
			},
			want: want{
				loggableObject: &analytics.LoggableAuctionObject{
					RejectedBids: []analytics.RejectedBid{
						{
							Bid: &entities.PbsOrtbBid{Bid: &openrtb2.Bid{
								ID:  "b2",
								Ext: json.RawMessage(`{"key":"value","origbidcpm":10,"prebid":{"dealpriority":10,"type":"banner"}}`),
							},
								DealPriority:   10,
								OriginalBidCPM: 10,
								BidType:        openrtb_ext.BidTypeBanner,
							},
							Seat:            "pubmatic",
							RejectionReason: openrtb3.LossBidAdvertiserBlocking,
						},
					},
				},
			},
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			UpdateRejectedBidExt(test.args.loggableObject)
			assert.Equal(t, test.want.loggableObject.RejectedBids[0].Bid.Bid.Ext, test.args.loggableObject.RejectedBids[0].Bid.Bid.Ext, "mismatched loggableObject for test-[%+v]", test.name)

		})
	}
}

func TestCallRecordBids(t *testing.T) {

	type args struct {
		ctx              context.Context
		pubID            string
		adapterBids      map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid
		getMetricsEngine func() *metrics.MetricsEngineMock
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "empty context",
			args: args{
				ctx:   context.Background(),
				pubID: "1010",
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					return &metrics.MetricsEngineMock{}
				},
			},
		},
		{
			name: "bidCountMetricEnabled is false",
			args: args{
				ctx:   context.WithValue(context.Background(), bidCountMetricEnabled, false),
				pubID: "1010",
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					return &metrics.MetricsEngineMock{}
				},
			},
		},
		{
			name: "bidCountMetricEnabled is true, owProfileId is non-string",
			args: args{
				ctx:   context.WithValue(context.WithValue(context.Background(), bidCountMetricEnabled, true), owProfileId, 1),
				pubID: "1010",
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					return &metrics.MetricsEngineMock{}
				},
			},
		},
		{
			name: "bidCountMetricEnabled is true, owProfileId is empty",
			args: args{
				ctx:   context.WithValue(context.WithValue(context.Background(), bidCountMetricEnabled, true), owProfileId, ""),
				pubID: "1010",
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					return &metrics.MetricsEngineMock{}
				},
			},
		},
		{
			name: "empty adapterBids",
			args: args{
				ctx:         context.WithValue(context.WithValue(context.Background(), bidCountMetricEnabled, true), owProfileId, "11"),
				pubID:       "1010",
				adapterBids: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{},
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					return &metrics.MetricsEngineMock{}
				},
			},
		},
		{
			name: "empty adapterBids.seat",
			args: args{
				ctx:   context.WithValue(context.WithValue(context.Background(), bidCountMetricEnabled, true), owProfileId, "11"),
				pubID: "1010",
				adapterBids: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
					"pubmatic": {},
				},
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					return &metrics.MetricsEngineMock{}
				},
			},
		},
		{
			name: "empty adapterBids.seat.bids",
			args: args{
				ctx:   context.WithValue(context.WithValue(context.Background(), bidCountMetricEnabled, true), owProfileId, "11"),
				pubID: "1010",
				adapterBids: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
					"pubmatic": {
						Bids: []*entities.PbsOrtbBid{},
					},
				},
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					return &metrics.MetricsEngineMock{}
				},
			},
		},
		{
			name: "multiple non deal bid",
			args: args{
				ctx:   context.WithValue(context.WithValue(context.Background(), bidCountMetricEnabled, true), owProfileId, "11"),
				pubID: "1010",
				adapterBids: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
					"pubmatic": {
						Bids: []*entities.PbsOrtbBid{
							{
								Bid: &openrtb2.Bid{
									ID: "bid1",
								},
							},
							{
								Bid: &openrtb2.Bid{
									ID: "bid2",
								},
							},
						},
						Seat: "pubmatic",
					},
				},
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					metricEngine := &metrics.MetricsEngineMock{}
					metricEngine.Mock.On("RecordBids", "1010", "11", "pubmatic", nodeal).Return()
					metricEngine.Mock.On("RecordBids", "1010", "11", "pubmatic", nodeal).Return()
					return metricEngine
				},
			},
		},
		{
			name: "multiple deal bid",
			args: args{
				ctx:   context.WithValue(context.WithValue(context.Background(), bidCountMetricEnabled, true), owProfileId, "11"),
				pubID: "1010",
				adapterBids: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
					"pubmatic": {
						Bids: []*entities.PbsOrtbBid{
							{
								Bid: &openrtb2.Bid{
									ID:     "bid1",
									DealID: "pubdeal",
								},
							},
							{
								Bid: &openrtb2.Bid{
									ID:     "bid2",
									DealID: "pubdeal",
								},
							},
						},
						Seat: "pubmatic",
					},
				},
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					metricEngine := &metrics.MetricsEngineMock{}
					metricEngine.Mock.On("RecordBids", "1010", "11", "pubmatic", "pubdeal").Return()
					metricEngine.Mock.On("RecordBids", "1010", "11", "pubmatic", "pubdeal").Return()
					return metricEngine
				},
			},
		},
		{
			name: "multiple bidders",
			args: args{
				ctx:   context.WithValue(context.WithValue(context.Background(), bidCountMetricEnabled, true), owProfileId, "11"),
				pubID: "1010",
				adapterBids: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
					"pubmatic": {
						Bids: []*entities.PbsOrtbBid{
							{
								Bid: &openrtb2.Bid{
									ID:     "bid1",
									DealID: "pubdeal",
								},
							},
						},
						Seat: "pubmatic",
					},
					"appnexus": {
						Bids: []*entities.PbsOrtbBid{
							{
								Bid: &openrtb2.Bid{
									ID:     "bid2",
									DealID: "appnxdeal",
								},
							},
							{
								Bid: &openrtb2.Bid{
									ID: "bid3",
								},
							},
						},
						Seat: "appnexus",
					},
				},
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					metricEngine := &metrics.MetricsEngineMock{}
					metricEngine.Mock.On("RecordBids", "1010", "11", "pubmatic", "pubdeal").Return()
					metricEngine.Mock.On("RecordBids", "1010", "11", "appnexus", "appnxdeal").Return()
					metricEngine.Mock.On("RecordBids", "1010", "11", "appnexus", nodeal).Return()
					return metricEngine
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMetricEngine := tt.args.getMetricsEngine()
			recordBids(tt.args.ctx, mockMetricEngine, tt.args.pubID, tt.args.adapterBids)
			mockMetricEngine.AssertExpectations(t)
		})
	}
}

func TestRecordVastVersion(t *testing.T) {
	type args struct {
		metricsEngine    metrics.MetricsEngine
		adapterBids      map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid
		getMetricsEngine func() *metrics.MetricsEngineMock
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "No Bids",
			args: args{
				adapterBids: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{},
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					metricEngine := &metrics.MetricsEngineMock{}
					return metricEngine
				},
			},
		},
		{
			name: "Empty Bids in SeatBid",
			args: args{
				adapterBids: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
					"pubmatic": {
						Bids: []*entities.PbsOrtbBid{},
					},
				},
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					metricEngine := &metrics.MetricsEngineMock{}
					return metricEngine
				},
			},
		},
		{
			name: "Empty Bids in SeatBid",
			args: args{
				adapterBids: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
					"pubmatic": {
						Bids: []*entities.PbsOrtbBid{},
					},
				},
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					metricEngine := &metrics.MetricsEngineMock{}
					return metricEngine
				},
			},
		},
		{
			name: "Invalid Bid Type",
			args: args{
				adapterBids: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
					"pubmatic": {
						Bids: []*entities.PbsOrtbBid{
							{
								BidType: openrtb_ext.BidTypeBanner,
							},
						},
					},
				},
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					metricEngine := &metrics.MetricsEngineMock{}
					return metricEngine
				},
			},
		},
		{
			name: "No Adm in Bids",
			args: args{
				adapterBids: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
					"pubmatic": {
						Bids: []*entities.PbsOrtbBid{
							{
								Bid: &openrtb2.Bid{
									AdM: "",
								},
								BidType: openrtb_ext.BidTypeVideo,
							},
						},
					},
				},
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					metricEngine := &metrics.MetricsEngineMock{}
					return metricEngine
				},
			},
		},
		{
			name: "No version found in Adm",
			args: args{
				adapterBids: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
					"pubmatic": {
						BidderCoreName: "pubmatic",
						Bids: []*entities.PbsOrtbBid{
							{
								Bid: &openrtb2.Bid{
									AdM: "<Vast> <Vast>",
								},
								BidType: openrtb_ext.BidTypeVideo,
							},
						},
					},
				},
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					metricEngine := &metrics.MetricsEngineMock{}
					metricEngine.Mock.On("RecordVastVersion", "pubmatic", vastVersionUndefined).Return()
					return metricEngine
				},
			},
		},
		{
			name: "Version found in Adm",
			args: args{
				adapterBids: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
					"pubmatic": {
						BidderCoreName: "pubmatic",
						Bids: []*entities.PbsOrtbBid{
							{
								BidType: openrtb_ext.BidTypeVideo,
								Bid: &openrtb2.Bid{
									AdM: `<VAST version=\"2.0\">
									  <Ad id="601364">
									    <InLine>
									      <AdSystem>Adsystem Example</AdSystem>
									      <AdTitle>VAST 2.0</AdTitle>
									      <Description>VAST 2.0</Description>
									      <Error>http://myErrorURL/error</Error>
									      <Impression>http://myTrackingURL/impression</Impression>
									      <Creatives>
									        <Creative AdID="12345">
									          <Linear>
									           <Duration>00:00:30</Duration>
									            <TrackingEvents>
									              <Tracking event="creativeView">http://myTrackingURL/creativeView</Tracking>
									              <Tracking event="start">http://myTrackingURL/start</Tracking>
									              <Tracking event="midpoint">http://myTrackingURL/midpoint</Tracking>
									              <Tracking event="firstQuartile">http://myTrackingURL/firstQuartile</Tracking>
									              <Tracking event="thirdQuartile">http://myTrackingURL/thirdQuartile</Tracking>
									              <Tracking event="complete">http://myTrackingURL/complete</Tracking>
									            </TrackingEvents>
									            <VideoClicks>
									              <ClickThrough>http://www.examplemedia.com</ClickThrough>
									              <ClickTracking>http://myTrackingURL/click</ClickTracking>
									            </VideoClicks>
									            <MediaFiles>
									             <MediaFile delivery="progressive" type="video/x-flv" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true">
									        http://demo.examplemedia.com/video/acudeo/Carrot_400x300_500kb.flv
									          </MediaFile>
									         </MediaFiles>
									          </Linear>
									    </Creative>
									    <Creative AdID="601364-Companion">
									      <CompanionAds>
									           <Companion width="300" height="250">
									             <StaticResource creativeType="image/jpeg">
									             http://demo.examplemedia.com/vast/this_is_the_ad.jpg
									             </StaticResource>
									             <TrackingEvents>
									               <Tracking event="creativeView">http://myTrackingURL/tracking</Tracking>
									             </TrackingEvents>
									           <CompanionClickThrough>http://www.examplemedia.com</CompanionClickThrough>
									           </Companion>
									           <Companion width="728" height="90">
									             <StaticResource creativeType="image/jpeg">
									             http://demo.examplemedia.com/vast/trackingbanner
									             </StaticResource>
									           <CompanionClickThrough>http://www.examplemedia.com</CompanionClickThrough>
									           </Companion>
									         </CompanionAds>
									       </Creative>
									     </Creatives>
									   </InLine>
									   </Ad>
									</VAST>`,
								},
							},
						},
					},
				},
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					metricEngine := &metrics.MetricsEngineMock{}
					metricEngine.Mock.On("RecordVastVersion", "pubmatic", "2.0").Return()
					return metricEngine
				},
			},
		},
		{
			name: "Version found in Adm with spaces in tag",
			args: args{
				adapterBids: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
					"pubmatic": {
						BidderCoreName: "pubmatic",
						Bids: []*entities.PbsOrtbBid{
							{
								BidType: openrtb_ext.BidTypeVideo,
								Bid: &openrtb2.Bid{
									AdM: `<VAST version = "4.1">
									</VAST>`,
								},
							},
						},
					},
				},
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					metricEngine := &metrics.MetricsEngineMock{}
					metricEngine.Mock.On("RecordVastVersion", "pubmatic", "4.1").Return()
					return metricEngine
				},
			},
		},
		{
			name: "Version found in Adm with multiple attributes",
			args: args{
				adapterBids: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
					"pubmatic": {
						BidderCoreName: "pubmatic",
						Bids: []*entities.PbsOrtbBid{
							{
								BidType: openrtb_ext.BidTypeVideo,
								Bid: &openrtb2.Bid{
									AdM: `<VAST namespace="test" version = \"2.0\">
									</VAST>`,
								},
							},
						},
					},
				},
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					metricEngine := &metrics.MetricsEngineMock{}
					metricEngine.Mock.On("RecordVastVersion", "pubmatic", "2.0").Return()
					return metricEngine
				},
			},
		},
		{
			name: "Version found xml tag before Vast tag attributes",
			args: args{
				adapterBids: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
					"pubmatic": {
						BidderCoreName: "pubmatic",
						Bids: []*entities.PbsOrtbBid{
							{
								BidType: openrtb_ext.BidTypeVideo,
								Bid: &openrtb2.Bid{
									AdM: `<?xml version="1.0" encoding="UTF-8"?><VAST xmlns:xs="http://www.w3.org/2001/XMLSchema" version="2.0">
									</VAST>`,
								},
							},
						},
					},
				},
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					metricEngine := &metrics.MetricsEngineMock{}
					metricEngine.Mock.On("RecordVastVersion", "pubmatic", "2.0").Return()
					return metricEngine
				},
			},
		},
		{
			name: "Version found in Adm inside single quote",
			args: args{
				adapterBids: map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid{
					"pubmatic": {
						BidderCoreName: "pubmatic",
						Bids: []*entities.PbsOrtbBid{
							{
								BidType: openrtb_ext.BidTypeVideo,
								Bid: &openrtb2.Bid{
									AdM: `<VAST namespace="test" version = \'2.0\'>
									</VAST>`,
								},
							},
						},
					},
				},
				getMetricsEngine: func() *metrics.MetricsEngineMock {
					metricEngine := &metrics.MetricsEngineMock{}
					metricEngine.Mock.On("RecordVastVersion", "pubmatic", "2.0").Return()
					return metricEngine
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMetricEngine := tt.args.getMetricsEngine()
			recordVastVersion(mockMetricEngine, tt.args.adapterBids)
			mockMetricEngine.AssertExpectations(t)
		})
	}
}
