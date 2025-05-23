package adpod

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/endpoints/openrtb2/ctv/constant"
	"github.com/prebid/prebid-server/v3/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/v3/metrics"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
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
		status   constant.BidStatus
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
				status:   constant.StatusOK,
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
				status:   constant.StatusOK,
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
				status:   constant.StatusDurationMismatch,
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
				status:   constant.StatusOK,
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
				status:   constant.StatusOK,
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
				status:   constant.StatusDurationMismatch,
			},
		},

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, status := getDurationBasedOnDurationMatchingPolicy(tt.args.duration, tt.args.policy, tt.args.config)
			assert.Equal(t, tt.want.duration, duration)
			assert.Equal(t, tt.want.status, status)
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
		status   constant.BidStatus
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
				status:   constant.StatusOK,
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
				status:   constant.StatusOK,
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
				status:   constant.StatusOK,
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
				status:   constant.StatusOK,
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
				status:   constant.StatusOK,
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
				status:   constant.StatusOK,
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
				status:   constant.StatusOK,
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
				status:   constant.StatusOK,
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
				status:   constant.StatusDurationMismatch,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, status := getBidDuration(tt.args.bid, tt.args.reqExt, tt.args.config, tt.args.defaultDuration)
			assert.Equal(t, tt.want.duration, duration)
			assert.Equal(t, tt.want.status, status)
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
							Bid:    &openrtb2.Bid{},
							Status: constant.StatusCategoryExclusion,
							Seat:   "pubmatic",
						},
						{
							Bid:    &openrtb2.Bid{},
							Status: constant.StatusWinningBid,
							Seat:   "pubmatic",
						},
						{
							Bid:    &openrtb2.Bid{},
							Status: constant.StatusOK,
							Seat:   "pubmatic",
						},
						{
							Bid:    &openrtb2.Bid{},
							Status: 100,
							Seat:   "pubmatic",
						},
					},
				},
			},
			want: want{
				expectedCalls: 2,
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
								Ext: json.RawMessage(`{"prebid": {"video": {} },"adpod": {}}`),
							},
							Duration: 10,
							Status:   1,
						},
					},
				},
			},
			want: want{
				adpodBid: &types.AdPodBid{
					Bids: []*types.Bid{
						{
							Bid: &openrtb2.Bid{
								Ext: json.RawMessage(`{"prebid": {"video": {"duration":10} },"adpod": {"aprc":1}}`),
							},
							Duration: 10,
							Status:   1,
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

func TestMergeAdPodBids(t *testing.T) {
	type args struct {
		adpod *types.AdPodBid
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
								Price: 1.0,
								AdM:   "<xml>any_creative_without_vast</xml>",
							},
						},
					},
				},
			},
			want: "",
		},
		{
			name: "VAST_element_present_in_adm",
			args: args{
				adpod: &types.AdPodBid{
					Bids: []*types.Bid{
						{
							Bid: &openrtb2.Bid{
								Price: 1.0,
								AdM:   "<VAST><Ad>url_creative</Ad></VAST>",
							},
						},
					},
				},
			},
			want: "<VAST version=\"2.0\"><Ad sequence=\"1\"><![CDATA[url_creative]]></Ad></VAST>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := mergeAdPodBids(tt.args.adpod)
			assert.Equalf(t, tt.want, got, "found incorrect creative")
		})
	}
}
