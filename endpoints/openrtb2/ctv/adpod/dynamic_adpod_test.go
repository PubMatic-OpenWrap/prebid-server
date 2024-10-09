package adpod

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/v2/exchange"
	"github.com/prebid/prebid-server/v2/metrics"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDurationBasedOnDurationMatchingPolicy(t *testing.T) {
	type args struct {
		duration int64
		policy   openrtb_ext.OWVideoAdDurationMatchingPolicy
		config   []*types.ImpAdPodConfig
	}
	type want struct {
		duration int64
		nbr      *openrtb3.NoBidReason
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "empty_duration_policy",
			args: args{
				duration: 10,
				policy:   "",
				config: []*types.ImpAdPodConfig{
					{MaxDuration: 10},
					{MaxDuration: 20},
					{MaxDuration: 30},
					{MaxDuration: 40},
				},
			},
			want: want{
				duration: 10,
				nbr:      nil,
			},
		},
		{
			name: "policy_exact",
			args: args{
				duration: 10,
				policy:   openrtb_ext.OWExactVideoAdDurationMatching,
				config: []*types.ImpAdPodConfig{
					{MaxDuration: 10},
					{MaxDuration: 20},
					{MaxDuration: 30},
					{MaxDuration: 40},
				},
			},
			want: want{
				duration: 10,
				nbr:      nil,
			},
		},
		{
			name: "policy_exact_didnot_match",
			args: args{
				duration: 15,
				policy:   openrtb_ext.OWExactVideoAdDurationMatching,
				config: []*types.ImpAdPodConfig{
					{MaxDuration: 10},
					{MaxDuration: 20},
					{MaxDuration: 30},
					{MaxDuration: 40},
				},
			},
			want: want{
				duration: 15,
				nbr:      exchange.ResponseRejectedInvalidCreative.Ptr(),
			},
		},
		{
			name: "policy_roundup_exact",
			args: args{
				duration: 20,
				policy:   openrtb_ext.OWRoundupVideoAdDurationMatching,
				config: []*types.ImpAdPodConfig{
					{MaxDuration: 10},
					{MaxDuration: 20},
					{MaxDuration: 30},
					{MaxDuration: 40},
				},
			},
			want: want{
				duration: 20,
				nbr:      nil,
			},
		},
		{
			name: "policy_roundup",
			args: args{
				duration: 25,
				policy:   openrtb_ext.OWRoundupVideoAdDurationMatching,
				config: []*types.ImpAdPodConfig{
					{MaxDuration: 10},
					{MaxDuration: 20},
					{MaxDuration: 30},
					{MaxDuration: 40},
				},
			},
			want: want{
				duration: 30,
				nbr:      nil,
			},
		},
		{
			name: "policy_roundup_didnot_match",
			args: args{
				duration: 45,
				policy:   openrtb_ext.OWRoundupVideoAdDurationMatching,
				config: []*types.ImpAdPodConfig{
					{MaxDuration: 10},
					{MaxDuration: 20},
					{MaxDuration: 30},
					{MaxDuration: 40},
				},
			},
			want: want{
				duration: 45,
				nbr:      exchange.ResponseRejectedInvalidCreative.Ptr(),
			},
		},

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, nbr := getDurationBasedOnDurationMatchingPolicy(tt.args.duration, tt.args.policy, tt.args.config)
			assert.Equal(t, tt.want.duration, duration)
			assert.Equal(t, tt.want.nbr, nbr)
		})
	}
}

func TestGetBidDuration(t *testing.T) {
	type args struct {
		bid             *openrtb2.Bid
		reqExt          *openrtb_ext.ExtRequestAdPod
		config          []*types.ImpAdPodConfig
		defaultDuration int64
	}
	type want struct {
		duration int64
		nbr      *openrtb3.NoBidReason
	}
	var tests = []struct {
		name   string
		args   args
		want   want
		expect int
	}{
		{
			name: "nil_bid_ext",
			args: args{
				bid:             &openrtb2.Bid{},
				reqExt:          nil,
				config:          nil,
				defaultDuration: 100,
			},
			want: want{
				duration: 100,
				nbr:      nil,
			},
		},
		{
			name: "use_default_duration",
			args: args{
				bid: &openrtb2.Bid{
					Ext: json.RawMessage(`{"tmp":123}`),
				},
				reqExt:          nil,
				config:          nil,
				defaultDuration: 100,
			},
			want: want{
				duration: 100,
				nbr:      nil,
			},
		},
		{
			name: "invalid_duration_in_bid_ext",
			args: args{
				bid: &openrtb2.Bid{
					Ext: json.RawMessage(`{"prebid":{"video":{"duration":"invalid"}}}`),
				},
				reqExt:          nil,
				config:          nil,
				defaultDuration: 100,
			},
			want: want{
				duration: 100,
				nbr:      nil,
			},
		},
		{
			name: "0sec_duration_in_bid_ext",
			args: args{
				bid: &openrtb2.Bid{
					Ext: json.RawMessage(`{"prebid":{"video":{"duration":0}}}`),
				},
				reqExt:          nil,
				config:          nil,
				defaultDuration: 100,
			},
			want: want{
				duration: 100,
				nbr:      nil,
			},
		},
		{
			name: "negative_duration_in_bid_ext",
			args: args{
				bid: &openrtb2.Bid{
					Ext: json.RawMessage(`{"prebid":{"video":{"duration":-30}}}`),
				},
				reqExt:          nil,
				config:          nil,
				defaultDuration: 100,
			},
			want: want{
				duration: 100,
				nbr:      nil,
			},
		},
		{
			name: "30sec_duration_in_bid_ext",
			args: args{
				bid: &openrtb2.Bid{
					Ext: json.RawMessage(`{"prebid":{"video":{"duration":30}}}`),
				},
				reqExt:          nil,
				config:          nil,
				defaultDuration: 100,
			},
			want: want{
				duration: 30,
				nbr:      nil,
			},
		},
		{
			name: "duration_matching_empty",
			args: args{
				bid: &openrtb2.Bid{
					Ext: json.RawMessage(`{"prebid":{"video":{"duration":30}}}`),
				},
				reqExt: &openrtb_ext.ExtRequestAdPod{
					VideoAdDurationMatching: "",
				},
				config:          nil,
				defaultDuration: 100,
			},
			want: want{
				duration: 30,
				nbr:      nil,
			},
		},
		{
			name: "duration_matching_exact",
			args: args{
				bid: &openrtb2.Bid{
					Ext: json.RawMessage(`{"prebid":{"video":{"duration":30}}}`),
				},
				reqExt: &openrtb_ext.ExtRequestAdPod{
					VideoAdDurationMatching: openrtb_ext.OWExactVideoAdDurationMatching,
				},
				config: []*types.ImpAdPodConfig{
					{MaxDuration: 10},
					{MaxDuration: 20},
					{MaxDuration: 30},
					{MaxDuration: 40},
				},
				defaultDuration: 100,
			},
			want: want{
				duration: 30,
				nbr:      nil,
			},
		},
		{
			name: "duration_matching_exact_not_present",
			args: args{
				bid: &openrtb2.Bid{
					Ext: json.RawMessage(`{"prebid":{"video":{"duration":35}}}`),
				},
				reqExt: &openrtb_ext.ExtRequestAdPod{
					VideoAdDurationMatching: openrtb_ext.OWExactVideoAdDurationMatching,
				},
				config: []*types.ImpAdPodConfig{
					{MaxDuration: 10},
					{MaxDuration: 20},
					{MaxDuration: 30},
					{MaxDuration: 40},
				},
				defaultDuration: 100,
			},
			want: want{
				duration: 35,
				nbr:      exchange.ResponseRejectedInvalidCreative.Ptr(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, nbr := getBidDuration(tt.args.bid, tt.args.reqExt, tt.args.config, tt.args.defaultDuration)
			assert.Equal(t, tt.want.duration, duration)
			assert.Equal(t, tt.want.nbr, nbr)
		})
	}
}

func TestRecordAdPodRejectedBids(t *testing.T) {
	type args struct {
		bids types.AdPodBid
	}

	type want struct {
		expectedCalls int
	}

	tests := []struct {
		description string
		args        args
		want        want
	}{
		{
			description: "multiple rejected bids",
			args: args{
				bids: types.AdPodBid{
					Bids: []*types.Bid{
						{
							Bid:  &openrtb2.Bid{},
							Nbr:  exchange.ResponseRejectedCreativeCategoryExclusions.Ptr(),
							Seat: "pubmatic",
						},
						{
							Bid:  &openrtb2.Bid{},
							Seat: "pubmatic",
						},
						{
							Bid:  &openrtb2.Bid{},
							Nbr:  nbr.LossBidLostToHigherBid.Ptr(),
							Seat: "pubmatic",
						},
						{
							Bid:  &openrtb2.Bid{},
							Nbr:  ptrutil.ToPtr[openrtb3.NoBidReason](100),
							Seat: "pubmatic",
						},
					},
				},
			},
			want: want{
				expectedCalls: 3,
			},
		},
	}

	for _, test := range tests {
		me := &metrics.MetricsEngineMock{}
		me.On("RecordRejectedBids", mock.Anything, mock.Anything, mock.Anything).Return()

		deps := dynamicAdpod{
			AdpodCtx: AdpodCtx{
				MetricsEngine: me,
			},
			AdpodBid: &test.args.bids,
		}
		deps.recordRejectedAdPodBids("pub_001")
		me.AssertNumberOfCalls(t, "RecordRejectedBids", test.want.expectedCalls)
	}
}

func TestSetBidExtParams(t *testing.T) {
	type args struct {
		adpodBid *types.AdPodBid
	}
	type want struct {
		adpodBid *types.AdPodBid
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "sample",
			args: args{
				adpodBid: &types.AdPodBid{
					Bids: []*types.Bid{
						{
							Bid: &openrtb2.Bid{
								Ext: json.RawMessage(`{"prebid": {"video": {} }}`),
							},
							Duration: 10,
						},
					},
				},
			},
			want: want{
				adpodBid: &types.AdPodBid{
					Bids: []*types.Bid{
						{
							Bid: &openrtb2.Bid{
								Ext: json.RawMessage(`{"prebid": {"video": {"duration":10} }}`),
							},
							Duration: 10,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adpod := dynamicAdpod{
				AdpodBid: tt.args.adpodBid,
			}

			adpod.setBidExtParams()
			assert.Equal(t, tt.want.adpodBid.Bids[0].Ext, adpod.AdpodBid.Bids[0].Ext)
		})
	}
}

func TestGetAdPodBidCreative(t *testing.T) {
	type args struct {
		adpod          *types.AdPodBid
		generatedBidID bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "VAST_element_missing_in_adm",
			args: args{
				adpod: &types.AdPodBid{
					Bids: []*types.Bid{
						{
							Bid: &openrtb2.Bid{
								AdM: "<xml>any_creative_without_vast</xml>",
							},
						},
					},
				},
				generatedBidID: false,
			},
			want: "<VAST version=\"2.0\"/>",
		},
		{
			name: "VAST_element_present_in_adm",
			args: args{
				adpod: &types.AdPodBid{
					Bids: []*types.Bid{
						{
							Bid: &openrtb2.Bid{
								AdM: "<VAST><Ad>url_creative</Ad></VAST>",
							},
						},
					},
				},
				generatedBidID: false,
			},
			want: "<VAST version=\"2.0\"><Ad sequence=\"1\"><![CDATA[url_creative]]></Ad></VAST>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getAdPodBidCreative(tt.args.adpod, tt.args.generatedBidID)
			assert.Equalf(t, tt.want, *got, "found incorrect creative")
		})
	}
}

func TestDynamicAdpodCollectSeatNonBids(t *testing.T) {
	type fields struct {
		AdpodBid *types.AdPodBid
	}
	tests := []struct {
		name   string
		fields fields
		want   openrtb_ext.NonBidCollection
	}{
		{
			name: "Test Get seat non bid- winning and non winning bids",
			fields: fields{
				AdpodBid: &types.AdPodBid{
					Bids: []*types.Bid{
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
							Nbr:               nbr.LossBidLostToHigherBid.Ptr(),
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
							Nbr:               nil,
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
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
			da := &dynamicAdpod{
				AdpodBid: tt.fields.AdpodBid,
			}
			snb := da.CollectSeatNonBids()
			assert.Equal(t, snb, tt.want)
		})
	}
}

func TestDynamicAdpodGetWinningBids(t *testing.T) {
	type fields struct {
		WinningBids *types.AdPodBid
	}
	tests := []struct {
		name   string
		fields fields
		want   []openrtb2.SeatBid
	}{
		{
			name: "Test Empty Bids in WinningBids",
			fields: fields{
				WinningBids: &types.AdPodBid{
					Bids: []*types.Bid{},
				},
			},
			want: nil,
		},
		{
			name: "Test GetWinningBids",
			fields: fields{
				WinningBids: &types.AdPodBid{
					Bids: []*types.Bid{
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
							Nbr:               nil,
							DealTierSatisfied: false,
							Seat:              "pubmatic",
						},
						{
							Bid: &openrtb2.Bid{
								ID:    "BID-3",
								Price: 25,
							},
							ExtBid: openrtb_ext.ExtBid{
								Prebid: &openrtb_ext.ExtBidPrebid{
									Meta: &openrtb_ext.ExtBidPrebidMeta{
										AdapterCode: "appnexus",
									},
								},
							},
							Nbr:               nil,
							DealTierSatisfied: false,
							Seat:              "appnexus",
						},
					},
				},
			},
			want: []openrtb2.SeatBid{
				{
					Seat: "pubmatic",
					Bid: []openrtb2.Bid{
						{
							ID:    "BID-2",
							Price: 15,
						},
					},
				},
				{
					Seat: "appnexus",
					Bid: []openrtb2.Bid{
						{
							ID:    "BID-3",
							Price: 25,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			da := &dynamicAdpod{
				WinningBids: tt.fields.WinningBids,
			}
			if got := da.GetWinningBids(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dynamicAdpod.GetWinningBids() = %v, want %v", got, tt.want)
			}
		})
	}
}
