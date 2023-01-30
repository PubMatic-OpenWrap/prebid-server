package openrtb2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"header-bidding/openrtb"
	"net/http/httptest"
	"testing"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/openrtb/v17/openrtb3"
	"github.com/prebid/prebid-server/analytics"
	analyticsConf "github.com/prebid/prebid-server/analytics/config"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/constant"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/exchange"
	metricsConfig "github.com/prebid/prebid-server/metrics/config"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/stored_requests/backends/empty_fetcher"
	"github.com/stretchr/testify/assert"
)

func TestAddTargetingKeys(t *testing.T) {
	var tests = []struct {
		scenario string // Testcase scenario
		key      string
		value    string
		bidExt   string
		expect   map[string]string
	}{
		{scenario: "key_not_exists", key: "hb_pb_cat_dur", value: "some_value", bidExt: `{"prebid":{"targeting":{}}}`, expect: map[string]string{"hb_pb_cat_dur": "some_value"}},
		{scenario: "key_already_exists", key: "hb_pb_cat_dur", value: "new_value", bidExt: `{"prebid":{"targeting":{"hb_pb_cat_dur":"old_value"}}}`, expect: map[string]string{"hb_pb_cat_dur": "new_value"}},
	}
	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			bid := new(openrtb2.Bid)
			bid.Ext = []byte(test.bidExt)
			key := openrtb_ext.TargetingKey(test.key)
			assert.Nil(t, addTargetingKey(bid, key, test.value))
			extBid := openrtb_ext.ExtBid{}
			json.Unmarshal(bid.Ext, &extBid)
			assert.Equal(t, test.expect, extBid.Prebid.Targeting)
		})
	}
	assert.Equal(t, "Invalid bid", addTargetingKey(nil, openrtb_ext.HbCategoryDurationKey, "some value").Error())
}

func TestFilterImpsVastTagsByDuration(t *testing.T) {
	type inputParams struct {
		request          *openrtb2.BidRequest
		generatedRequest *openrtb2.BidRequest
		impData          []*types.ImpData
	}

	type output struct {
		reqs        openrtb2.BidRequest
		blockedTags []map[string][]string
	}

	tt := []struct {
		testName       string
		input          inputParams
		expectedOutput output
	}{
		{
			testName: "test_single_impression_single_vast_partner_with_no_excluded_tags",
			input: inputParams{
				request: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1", Ext: []byte(`{"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":20,"tagid":"openx_20"}]}}`)},
					},
				},
				generatedRequest: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 35}},
					},
				},
				impData: []*types.ImpData{},
			},
			expectedOutput: output{
				reqs: openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}, Ext: []byte(`{}`)},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}, Ext: []byte(`{"openx_vast_bidder":{"tags":[{"dur":20,"tagid":"openx_20"}]}}`)},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 35}, Ext: []byte(`{"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"}]}}`)},
					},
				},
				blockedTags: []map[string][]string{},
			},
		},
		{
			testName: "test_single_impression_single_vast_partner_with_excluded_tags",
			input: inputParams{
				request: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1", Ext: []byte(`{"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":20,"tagid":"openx_20"}]}}`)},
					},
				},
				generatedRequest: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}},
					},
				},
				impData: []*types.ImpData{
					{ImpID: "imp1"},
				},
			},
			expectedOutput: output{
				reqs: openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}, Ext: []byte(`{}`)},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}, Ext: []byte(`{"openx_vast_bidder":{"tags":[{"dur":20,"tagid":"openx_20"}]}}`)},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}, Ext: []byte(`{"openx_vast_bidder":{"tags":[{"dur":25,"tagid":"openx_25"}]}}`)},
					},
				},
				blockedTags: []map[string][]string{
					{"openx_vast_bidder": []string{"openx_35"}},
				},
			},
		},
		{
			testName: "test_single_impression_multiple_vast_partners_no_exclusions",
			input: inputParams{
				request: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1", Ext: []byte(`{"spotx_vast_bidder":{"tags":[{"dur":15,"tagid":"spotx_15"},{"dur":25,"tagid":"spotx_25"},{"dur":30,"tagid":"spotx_30"}]},"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":20,"tagid":"openx_20"}]}}`)},
					},
				},
				generatedRequest: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}},
					},
				},
				impData: []*types.ImpData{},
			},
			expectedOutput: output{
				reqs: openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}, Ext: []byte(`{}`)},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}, Ext: []byte(`{"openx_vast_bidder":{"tags":[{"dur":20,"tagid":"openx_20"}]},"spotx_vast_bidder":{"tags":[{"dur":15,"tagid":"spotx_15"}]}}`)},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}, Ext: []byte(`{"openx_vast_bidder":{"tags":[{"dur":25,"tagid":"openx_25"}]},"spotx_vast_bidder":{"tags":[{"dur":25,"tagid":"spotx_25"},{"dur":30,"tagid":"spotx_30"}]}}`)},
					},
				},
				blockedTags: []map[string][]string{},
			},
		},
		{
			testName: "test_single_impression_multiple_vast_partners_with_exclusions",
			input: inputParams{
				request: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1", Ext: []byte(`{"spotx_vast_bidder":{"tags":[{"dur":15,"tagid":"spotx_15"},{"dur":25,"tagid":"spotx_25"},{"dur":35,"tagid":"spotx_35"}]},"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":40,"tagid":"openx_40"}]}}`)},
					},
				},
				generatedRequest: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}},
					},
				},
				impData: []*types.ImpData{
					{ImpID: "imp1"},
				},
			},
			expectedOutput: output{
				reqs: openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}, Ext: []byte(`{}`)},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}, Ext: []byte(`{"spotx_vast_bidder":{"tags":[{"dur":15,"tagid":"spotx_15"}]}}`)},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}, Ext: []byte(`{"openx_vast_bidder":{"tags":[{"dur":25,"tagid":"openx_25"}]},"spotx_vast_bidder":{"tags":[{"dur":25,"tagid":"spotx_25"}]}}`)},
					},
				},
				blockedTags: []map[string][]string{
					{"openx_vast_bidder": []string{"openx_35", "openx_40"}, "spotx_vast_bidder": []string{"spotx_35"}},
				},
			},
		},
		{
			testName: "test_multi_impression_multi_partner_no_exclusions",
			input: inputParams{
				request: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1", Ext: []byte(`{"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":20,"tagid":"openx_20"}]}}`)},
						{ID: "imp2", Ext: []byte(`{"spotx_vast_bidder":{"tags":[{"dur":30,"tagid":"spotx_30"},{"dur":40,"tagid":"spotx_40"}]}}`)},
					},
				},
				generatedRequest: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}},
						{ID: "imp2_1", Video: &openrtb2.Video{MinDuration: 5, MaxDuration: 30}},
					},
				},
				impData: nil,
			},
			expectedOutput: output{
				reqs: openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}, Ext: []byte(`{}`)},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}, Ext: []byte(`{"openx_vast_bidder":{"tags":[{"dur":20,"tagid":"openx_20"}]}}`)},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}, Ext: []byte(`{"openx_vast_bidder":{"tags":[{"dur":25,"tagid":"openx_25"}]}}`)},
						{ID: "imp2_1", Video: &openrtb2.Video{MinDuration: 5, MaxDuration: 30}, Ext: []byte(`{"spotx_vast_bidder":{"tags":[{"dur":30,"tagid":"spotx_30"}]}}`)},
					},
				},
				blockedTags: nil,
			},
		},
		{
			testName: "test_multi_impression_multi_partner_with_exclusions",
			input: inputParams{
				request: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1", Ext: []byte(`{"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":20,"tagid":"openx_20"}]}}`)},
						{ID: "imp2", Ext: []byte(`{"spotx_vast_bidder":{"tags":[{"dur":30,"tagid":"spotx_30"},{"dur":40,"tagid":"spotx_40"}]}}`)},
					},
				},
				generatedRequest: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}},
						{ID: "imp2_1", Video: &openrtb2.Video{MinDuration: 5, MaxDuration: 30}},
					},
				},
				impData: []*types.ImpData{
					{ImpID: "imp1"},
					{ImpID: "imp2"},
				},
			},
			expectedOutput: output{
				reqs: openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}, Ext: []byte(`{}`)},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}, Ext: []byte(`{"openx_vast_bidder":{"tags":[{"dur":20,"tagid":"openx_20"}]}}`)},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}, Ext: []byte(`{"openx_vast_bidder":{"tags":[{"dur":25,"tagid":"openx_25"}]}}`)},
						{ID: "imp2_1", Video: &openrtb2.Video{MinDuration: 5, MaxDuration: 30}, Ext: []byte(`{"spotx_vast_bidder":{"tags":[{"dur":30,"tagid":"spotx_30"}]}}`)},
					},
				},
				blockedTags: []map[string][]string{
					{"openx_vast_bidder": []string{"openx_35"}},
					{"spotx_vast_bidder": []string{"spotx_40"}},
				},
			},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			deps := ctvEndpointDeps{request: tc.input.request, impData: tc.input.impData}
			deps.readImpExtensionsAndTags()

			outputBids := tc.input.generatedRequest
			deps.filterImpsVastTagsByDuration(outputBids)

			assert.Equal(t, tc.expectedOutput.reqs, *outputBids, "Expected length of impressions array was %d but actual was %d", tc.expectedOutput.reqs, outputBids)

			for i, datum := range deps.impData {
				assert.Equal(t, tc.expectedOutput.blockedTags[i], datum.BlockedVASTTags, "Expected and actual impData was different")
			}
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
					VideoLengthMatching: "",
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
					VideoLengthMatching: openrtb_ext.OWExactVideoLengthsMatching,
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
					VideoLengthMatching: openrtb_ext.OWExactVideoLengthsMatching,
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

func Test_getDurationBasedOnDurationMatchingPolicy(t *testing.T) {
	type args struct {
		duration int64
		policy   openrtb_ext.OWVideoLengthMatchingPolicy
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
				policy:   openrtb_ext.OWExactVideoLengthsMatching,
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
				policy:   openrtb_ext.OWExactVideoLengthsMatching,
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
				policy:   openrtb_ext.OWRoundupVideoLengthMatching,
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
				policy:   openrtb_ext.OWRoundupVideoLengthMatching,
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
				policy:   openrtb_ext.OWRoundupVideoLengthMatching,
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

func TestCreateBidResponse(t *testing.T) {
	type args struct {
		resp *openrtb2.BidResponse
	}
	type want struct {
		resp *openrtb2.BidResponse
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "sample bidresponse",
			args: args{
				resp: &openrtb2.BidResponse{
					ID:         "id1",
					Cur:        "USD",
					CustomData: "custom",
				},
			},
			want: want{
				resp: &openrtb2.BidResponse{
					ID:         "id1",
					Cur:        "USD",
					CustomData: "custom",
					SeatBid:    make([]openrtb2.SeatBid, 0),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := ctvEndpointDeps{
				request: &openrtb2.BidRequest{
					ID: "1",
				},
			}
			actual := deps.createBidResponse(tt.args.resp, nil)
			assert.Equal(t, tt.want.resp, actual)
		})

	}
}

func TestSetBidExtParams(t *testing.T) {
	type args struct {
		impData []*types.ImpData
	}
	type want struct {
		impData []*types.ImpData
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "sample",
			args: args{
				impData: []*types.ImpData{
					{
						Bid: &types.AdPodBid{
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
				},
			},
			want: want{
				impData: []*types.ImpData{
					{
						Bid: &types.AdPodBid{
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			deps := ctvEndpointDeps{
				impData: tt.args.impData,
			}
			deps.setBidExtParams()
			assert.Equal(t, tt.want.impData[0].Bid.Bids[0].Ext, deps.impData[0].Bid.Bids[0].Ext)
		})
	}
}

func TestGetAdPodExt(t *testing.T) {
	type args struct {
		resp *openrtb2.BidResponse
	}
	type want struct {
		data json.RawMessage
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "nil-ext",
			args: args{
				resp: &openrtb2.BidResponse{
					ID: "resp1",
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID: "b1",
								},
								{
									ID: "b2",
								},
							},
							Seat: "pubmatic",
						},
					},
				},
			},
			want: want{
				data: json.RawMessage(`{"adpod":{"bidresponse":{"id":"resp1","seatbid":[{"bid":[{"id":"b1","impid":"","price":0},{"id":"b2","impid":"","price":0}],"seat":"pubmatic"}]},"config":{"imp1":{"vidext":{"adpod":{}}}}}}`),
			},
		},
		{
			name: "non-nil-ext",
			args: args{
				resp: &openrtb2.BidResponse{
					ID: "resp1",
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID: "b1",
								},
								{
									ID: "b2",
								},
							},
							Seat: "pubmatic",
						},
					},
					Ext: json.RawMessage(`{"xyz":10}`),
				},
			},
			want: want{
				data: json.RawMessage(`{"xyz":10,"adpod":{"bidresponse":{"id":"resp1","seatbid":[{"bid":[{"id":"b1","impid":"","price":0},{"id":"b2","impid":"","price":0}],"seat":"pubmatic"}]},"config":{"imp1":{"vidext":{"adpod":{}}}}}}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			deps := ctvEndpointDeps{
				impData: []*types.ImpData{
					{
						ImpID: "imp1",
						VideoExt: &openrtb_ext.ExtVideoAdPod{
							AdPod: &openrtb_ext.VideoAdPod{},
						},
						Bid: &types.AdPodBid{
							Bids: []*types.Bid{},
						},
					},
				},
				request: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1"},
					},
				},
			}
			actual := deps.getAdPodExt(tt.args.resp)
			assert.Equal(t, string(tt.want.data), string(actual))
		})
	}
}

func TestFilterRejectedBids(t *testing.T) {
	type args struct {
		resp           *openrtb2.BidResponse
		loggableObject *analytics.LoggableAuctionObject
	}
	type want struct {
		RejectedBids []analytics.RejectedBid
		SeatBids     []openrtb2.SeatBid
	}
	tests := []struct {
		name string
		args args
		want want
	}{

		{
			name: "single-bidder",
			args: args{
				loggableObject: &analytics.LoggableAuctionObject{},
				resp: &openrtb2.BidResponse{
					ID: "resp1",
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "b1",
									Ext: json.RawMessage(`{"adpod": {"aprc":1}}`),
								},
								{
									ID:  "b2",
									Ext: json.RawMessage(`{"adpod": {"aprc":0}}`),
								},
							},
							Seat: "pubmatic",
						},
					},
				},
			},
			want: want{
				RejectedBids: []analytics.RejectedBid{
					{
						RejectionReason: openrtb3.LossLostToHigherBid,
						Seat:            "pubmatic",
						Bid: &openrtb2.Bid{
							ID:  "b2",
							Ext: json.RawMessage(`{"adpod": {"aprc":0}}`),
						},
					},
				},
				SeatBids: []openrtb2.SeatBid{
					{
						Seat: "pubmatic",
						Bid: []openrtb2.Bid{
							{
								ID:  "b1",
								Ext: json.RawMessage(`{"adpod": {"aprc":1}}`),
							},
						},
					},
				},
			},
		},
		{
			name: "bidder-without-aprc",
			args: args{
				loggableObject: &analytics.LoggableAuctionObject{},
				resp: &openrtb2.BidResponse{
					ID: "resp1",
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "b1",
									Ext: json.RawMessage(`{"adpod": {"noaprc":1}}`),
								},
							},
							Seat: "pubmatic",
						},
					},
				},
			},
			want: want{
				RejectedBids: nil,
				SeatBids: []openrtb2.SeatBid{
					{
						Bid:  []openrtb2.Bid{}, //empty-bid-array
						Seat: "pubmatic",
					},
				},
			},
		},
		{
			name: "multiple-bidders",
			args: args{
				loggableObject: &analytics.LoggableAuctionObject{},
				resp: &openrtb2.BidResponse{
					ID: "resp1",
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "b1",
									Ext: json.RawMessage(`{"adpod": {"aprc":1}}`),
								},
								{
									ID:  "b2",
									Ext: json.RawMessage(`{"adpod": {"aprc":0}}`),
								},
							},
							Seat: "pubmatic",
						},
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "b3",
									Ext: json.RawMessage(`{"adpod": {"aprc":3}}`),
								},
								{
									ID:  "b4",
									Ext: json.RawMessage(`{"adpod": {"aprc":1}}`),
								},
							},
							Seat: "appnexus",
						},
					},
				},
			},
			want: want{
				RejectedBids: []analytics.RejectedBid{
					{
						RejectionReason: openrtb3.LossLostToHigherBid,
						Seat:            "pubmatic",
						Bid: &openrtb2.Bid{
							ID:  "b2",
							Ext: json.RawMessage(`{"adpod": {"aprc":0}}`),
						},
					},
					{
						RejectionReason: openrtb3.LossAdvertiserExclusions,
						Seat:            "appnexus",
						Bid: &openrtb2.Bid{
							ID:  "b3",
							Ext: json.RawMessage(`{"adpod": {"aprc":3}}`),
						},
					},
				},
				SeatBids: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:  "b1",
								Ext: json.RawMessage(`{"adpod": {"aprc":1}}`),
							},
						},
						Seat: "pubmatic",
					},
					{
						Bid: []openrtb2.Bid{
							{
								ID:  "b4",
								Ext: json.RawMessage(`{"adpod": {"aprc":1}}`),
							},
						},
						Seat: "appnexus",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			oldRespSeatBid := make([]openrtb2.SeatBid, 0)
			oldRespSeatBid = append(oldRespSeatBid, tt.args.resp.SeatBid...)

			filterRejectedBids(tt.args.resp, tt.args.loggableObject)
			assert.Equal(t, tt.want.RejectedBids, tt.args.loggableObject.RejectedBids)
			assert.Equal(t, tt.want.SeatBids, tt.args.resp.SeatBid)
		})
	}
}
func formORtbV25Request(formatFlag bool, videoFlag bool) *openrtb.BidRequest {
	request := new(openrtb.BidRequest)
	banner := new(openrtb.Banner)
	if formatFlag == true {
		formatObj1 := new(openrtb.Format) // openrtb.Format{728, 90, nil}
		formatObj1.W = new(int)
		*formatObj1.W = 728
		formatObj1.H = new(int)
		*formatObj1.H = 90

		formatObj2 := new(openrtb.Format) // openrtb.Format{728, 90, nil}
		formatObj2.W = new(int)
		*formatObj2.W = 300
		formatObj2.H = new(int)
		*formatObj2.H = 250

		formatArray := []*openrtb.Format{formatObj1, formatObj2}
		banner.Format = formatArray

		banner.W = new(int)
		*banner.W = 700
		banner.H = new(int)
		*banner.H = 900

	} else {
		banner.W = new(int)
		*banner.W = 728
		banner.H = new(int)
		*banner.H = 90
	}

	imp := new(openrtb.Imp)
	if videoFlag == true {
		video := formVideoObject()
		imp.Video = video
	}

	imp.Id = new(string)
	*imp.Id = "1"
	imp.Banner = banner
	imp.TagId = new(string)
	*imp.TagId = "adunit"

	impWrapExt := new(openrtb.ExtImpWrapper)
	impWrapExt.Div = new(string)
	*impWrapExt.Div = "div"

	inImpExt := new(openrtb.ImpExtension)

	//inImpExt.Wrapper = impWrapExt

	// bidderExt := map[string]*openrtb.BidderExtension{
	// 	"appnexus": &openrtb.BidderExtension{
	// 		KeyWords: []openrtb.KeyVal{
	// 			{
	// 				Key:    "pmzoneid",
	// 				Values: []string{"val1", "val2"},
	// 			},
	// 		},
	// 	},
	// }
	//inImpExt.Bidder = bidderExt

	imp.Ext = inImpExt
	impArr := make([]*openrtb.Imp, 0)
	impArr = append(impArr, imp)
	request.Id = new(string)
	*request.Id = "123-456-789"
	request.Imp = impArr

	inImpExt.Prebid = new(openrtb_ext.ExtImpPrebid)
	inImpExt.Prebid.Bidder = map[string]json.RawMessage{
		"pubmatic": json.RawMessage(`""`),
	}

	len := 2
	request.Wseat = make([]string, len)
	for i := 0; i < len; i++ {
		request.Wseat[i] = fmt.Sprintf("Wseat_%d", i)
	}

	request.Cur = make([]string, len)
	for i := 0; i < len; i++ {
		request.Cur[i] = fmt.Sprintf("cur_%d", i)
	}

	request.Badv = make([]string, len)
	for i := 0; i < len; i++ {
		request.Badv[i] = fmt.Sprintf("badv_%d", i)
	}

	request.Bapp = make([]string, len)
	for i := 0; i < len; i++ {
		request.Bapp[i] = fmt.Sprintf("bapp_%d", i)
	}

	request.Bcat = make([]string, len)
	for i := 0; i < len; i++ {
		request.Bcat[i] = fmt.Sprintf("bcat_%d", i)
	}

	request.Wlang = make([]string, len)
	for i := 0; i < len; i++ {
		request.Wlang[i] = fmt.Sprintf("Wlang_%d", i)
	}

	request.Bseat = make([]string, len)
	for i := 0; i < len; i++ {
		request.Bseat[i] = fmt.Sprintf("Bseat_%d", i)
	}

	site := new(openrtb.Site)
	publisher := new(openrtb.Publisher)
	publisher.Id = new(string)
	*publisher.Id = "5890"
	site.Publisher = publisher
	site.Page = new(string)
	*site.Page = "www.test.com"

	site.Domain = new(string)
	*site.Domain = "test.com"

	request.Site = site

	request.Device = new(openrtb.Device)
	request.Device.IP = new(string)
	*request.Device.IP = "123.145.167.10"
	request.Device.Ua = new(string)
	*request.Device.Ua = "Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36"

	request.User = new(openrtb.User)
	request.User.ID = new(string)
	*request.User.ID = "119208432"

	request.User.BuyerUID = new(string)
	*request.User.BuyerUID = "1rwe432"

	request.User.Yob = new(int)
	*request.User.Yob = 1980

	request.User.Gender = new(string)
	*request.User.Gender = "F"

	request.User.Geo = new(openrtb.Geo)
	request.User.Geo.Country = new(string)
	*request.User.Geo.Country = "US"

	request.User.Geo.Region = new(string)
	*request.User.Geo.Region = "CA"

	request.User.Geo.Metro = new(string)
	*request.User.Geo.Metro = "90001"

	request.User.Geo.City = new(string)
	*request.User.Geo.City = "Alamo"

	request.Source = new(openrtb.Source)
	request.Source.Ext = map[string]interface{}{
		"omidpn": "MyIntegrationPartner",
		"omidpv": "7.1",
	}

	wExt := new(openrtb.ExtRequest)
	dmExt := new(openrtb.ExtRequestWrapper)
	dmExt.ProfileId = new(int)
	*dmExt.ProfileId = 123
	dmExt.VersionId = new(int)
	*dmExt.VersionId = 1
	dmExt.LoggerImpressionID = new(string)
	*dmExt.LoggerImpressionID = "test_display_wiid"
	wExt.Wrapper = dmExt

	request.Ext = wExt

	request.Test = new(int)
	*request.Test = 0
	return request

}

func formVideoObject() *openrtb.Video {
	video := new(openrtb.Video)
	video.Mimes = []string{"video/mp4", "video/mpeg"}
	video.W = new(int)
	*video.W = 640
	video.H = new(int)
	*video.H = 480

	video.Ext = map[string]interface{}{
		"adpod": map[string]int{
			"minads":        1,
			"adminduration": 5,
			"excladv":       50,
			"maxads":        3,
			"excliabcat":    100,
			"admaxduration": 40,
		},
		"offset": 20,
	}
	video.MaxDuration = new(int)
	video.MinDuration = new(int)
	*video.MaxDuration = 50
	*video.MinDuration = 5

	return video
}

type mockExchangeCTV struct {
	lastRequest *openrtb2.BidRequest
}

func (m *mockExchangeCTV) HoldAuction(ctx context.Context, auctionRequest exchange.AuctionRequest, debugLog *exchange.DebugLog) (*openrtb2.BidResponse, error) {

	ext := []byte(`{"prebid":{"targeting":{"hb_bidder_appnexus":"appnexus","hb_pb_appnexus":"20.00","hb_pb_cat_dur_appnex":"20.00_395_30s","hb_size":"1x1", "hb_uuid_appnexus":"837ea3b7-5598-4958-8c45-8e9ef2bf7cc1"},"type":"video","dealpriority":0,"dealtiersatisfied":false},"bidder":{"appnexus":{"brand_id":1,"auction_id":7840037870526938650,"bidder_id":2,"bid_ad_type":1,"creative_info":{"video":{"duration":30,"mimes":["video\/mp4"]}}}}}`)
	return &openrtb2.BidResponse{
		SeatBid: []openrtb2.SeatBid{
			{
				Seat: "appnexus",
				Bid: []openrtb2.Bid{
					{ID: "01", ImpID: "1_0", Price: 10, AdM: "<VAST></VAST>", Ext: ext},
					{ID: "02", ImpID: "1_1", Price: 10, AdM: "<VAST></VAST>", Ext: ext},
					{ID: "03", ImpID: "1_2", Price: 10, AdM: "<VAST></VAST>", Ext: ext},
					{ID: "04", ImpID: "1_3", Price: 10, AdM: "<VAST></VAST>", Ext: ext},
					{ID: "05", ImpID: "2_0", Price: 10, AdM: "<VAST></VAST>", Ext: ext},
				},
			},
			{
				Seat: "pubmatic",
				Bid: []openrtb2.Bid{
					{ID: "01", ImpID: "1_0", Price: 20, AdM: "<VAST></VAST>", Ext: ext},
					{ID: "02", ImpID: "1_1", Price: 20, AdM: "<VAST></VAST>", Ext: ext},
					{ID: "03", ImpID: "1_2", Price: 20, AdM: "<VAST></VAST>", Ext: ext},
					{ID: "04", ImpID: "1_3", Price: 20, AdM: "<VAST></VAST>", Ext: ext},
					{ID: "05", ImpID: "2_0", Price: 20, AdM: "<VAST></VAST>", Ext: ext},
				},
			},
		},
	}, nil
}

func TestCTVRequests(t *testing.T) {

	mockExchange := mockExchangeCTV{}
	endpoint, _ := NewCTVEndpoint(
		&mockExchange,
		mockBidderParamValidator{},
		&mockVideoStoredReqFetcher{},
		&mockVideoStoredReqFetcher{},
		empty_fetcher.EmptyFetcher{},
		&config.Configuration{MaxRequestSize: maxSize},
		&metricsConfig.NilMetricsEngine{},
		analyticsConf.NewPBSAnalytics(&config.Analytics{}),
		map[string]string{},
		[]byte{},
		openrtb_ext.BuildBidderMap(),
	)

	pbReq := formORtbV25Request(false, true)
	body := new(bytes.Buffer)
	_ = json.NewEncoder(body).Encode(pbReq)

	request := httptest.NewRequest("POST", "/openrtb2/video", body)
	//request := httptest.NewRequest("GET", fmt.Sprintf("/openrtb2/auction/amp", requestID), nil)
	recorder := httptest.NewRecorder()

	endpoint(recorder, request, nil)

	if recorder.Code != 200 {
		t.Errorf("Expected status")
	}

}
