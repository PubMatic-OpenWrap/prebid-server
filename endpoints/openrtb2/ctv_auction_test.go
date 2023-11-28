package openrtb2

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/constant"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/v2/metrics"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
						{ID: "imp1", Ext: []byte(`{"prebid": {"bidder": {"openx_vast_bidder": {"tags": [{"dur": 35,"tagid": "openx_35"}, {"dur": 25,"tagid": "openx_25"}, {"dur": 20,"tagid": "openx_20"}]}}}}`)},
					},
				},
				generatedRequest: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}, Ext: []byte(`{"prebid": {"bidder": {"openx_vast_bidder": {"tags": [{"dur": 35,"tagid": "openx_35"}, {"dur": 25,"tagid": "openx_25"}, {"dur": 20,"tagid": "openx_20"}]}}}}`)},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}, Ext: []byte(`{"prebid": {"bidder": {"openx_vast_bidder": {"tags": [{"dur": 35,"tagid": "openx_35"}, {"dur": 25,"tagid": "openx_25"}, {"dur": 20,"tagid": "openx_20"}]}}}}`)},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 35}, Ext: []byte(`{"prebid": {"bidder": {"openx_vast_bidder": {"tags": [{"dur": 35,"tagid": "openx_35"}, {"dur": 25,"tagid": "openx_25"}, {"dur": 20,"tagid": "openx_20"}]}}}}`)},
					},
				},
				impData: []*types.ImpData{},
			},
			expectedOutput: output{
				reqs: openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}, Ext: []byte(`{"prebid": {"bidder": {}}}`)},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}, Ext: []byte(`{"prebid": {"bidder": {"openx_vast_bidder":{"tags":[{"dur":20,"tagid":"openx_20"}]}}}}`)},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 35}, Ext: []byte(`{"prebid": {"bidder": {"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"}]}}}}`)},
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
						{ID: "imp1", Ext: []byte(`{"prebid": {"bidder": {"openx_vast_bidder": {"tags": [{"dur": 35,"tagid": "openx_35"}, {"dur": 25,"tagid": "openx_25"}, {"dur": 20,"tagid": "openx_20"}]}}}}`)},
					},
				},
				generatedRequest: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}, Ext: []byte(`{"prebid": {"bidder": {"openx_vast_bidder": {"tags": [{"dur": 35,"tagid": "openx_35"}, {"dur": 25,"tagid": "openx_25"}, {"dur": 20,"tagid": "openx_20"}]}}}}`)},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}, Ext: []byte(`{"prebid": {"bidder": {"openx_vast_bidder": {"tags": [{"dur": 35,"tagid": "openx_35"}, {"dur": 25,"tagid": "openx_25"}, {"dur": 20,"tagid": "openx_20"}]}}}}`)},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}, Ext: []byte(`{"prebid": {"bidder": {"openx_vast_bidder": {"tags": [{"dur": 35,"tagid": "openx_35"}, {"dur": 25,"tagid": "openx_25"}, {"dur": 20,"tagid": "openx_20"}]}}}}`)},
					},
				},
				impData: []*types.ImpData{
					{ImpID: "imp1"},
				},
			},
			expectedOutput: output{
				reqs: openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}, Ext: []byte(`{"prebid": {"bidder": {}}}`)},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}, Ext: []byte(`{"prebid": {"bidder": {"openx_vast_bidder":{"tags":[{"dur":20,"tagid":"openx_20"}]}}}}`)},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}, Ext: []byte(`{"prebid": {"bidder": {"openx_vast_bidder":{"tags":[{"dur":25,"tagid":"openx_25"}]}}}}`)},
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
						{ID: "imp1", Ext: []byte(`{"prebid":{"bidder":{"spotx_vast_bidder":{"tags":[{"dur":15,"tagid":"spotx_15"},{"dur":25,"tagid":"spotx_25"},{"dur":30,"tagid":"spotx_30"}]},"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":20,"tagid":"openx_20"}]}}}}`)},
					},
				},
				generatedRequest: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}, Ext: []byte(`{"prebid":{"bidder":{"spotx_vast_bidder":{"tags":[{"dur":15,"tagid":"spotx_15"},{"dur":25,"tagid":"spotx_25"},{"dur":30,"tagid":"spotx_30"}]},"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":20,"tagid":"openx_20"}]}}}}`)},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}, Ext: []byte(`{"prebid":{"bidder":{"spotx_vast_bidder":{"tags":[{"dur":15,"tagid":"spotx_15"},{"dur":25,"tagid":"spotx_25"},{"dur":30,"tagid":"spotx_30"}]},"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":20,"tagid":"openx_20"}]}}}}`)},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}, Ext: []byte(`{"prebid":{"bidder":{"spotx_vast_bidder":{"tags":[{"dur":15,"tagid":"spotx_15"},{"dur":25,"tagid":"spotx_25"},{"dur":30,"tagid":"spotx_30"}]},"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":20,"tagid":"openx_20"}]}}}}`)},
					},
				},
				impData: []*types.ImpData{},
			},
			expectedOutput: output{
				reqs: openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}, Ext: []byte(`{"prebid":{"bidder":{}}}`)},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}, Ext: []byte(`{"prebid":{"bidder":{"openx_vast_bidder":{"tags":[{"dur":20,"tagid":"openx_20"}]},"spotx_vast_bidder":{"tags":[{"dur":15,"tagid":"spotx_15"}]}}}}`)},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}, Ext: []byte(`{"prebid":{"bidder":{"openx_vast_bidder":{"tags":[{"dur":25,"tagid":"openx_25"}]},"spotx_vast_bidder":{"tags":[{"dur":25,"tagid":"spotx_25"},{"dur":30,"tagid":"spotx_30"}]}}}}`)},
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
						{ID: "imp1", Ext: []byte(`{"prebid": { "bidder": { "spotx_vast_bidder":{"tags":[{"dur":15,"tagid":"spotx_15"},{"dur":25,"tagid":"spotx_25"},{"dur":35,"tagid":"spotx_35"}]},"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":40,"tagid":"openx_40"}]}}}}`)},
					},
				},
				generatedRequest: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}, Ext: []byte(`{"prebid":{"bidder":{"spotx_vast_bidder":{"tags":[{"dur":15,"tagid":"spotx_15"},{"dur":25,"tagid":"spotx_25"},{"dur":35,"tagid":"spotx_35"}]},"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":40,"tagid":"openx_40"}]}}}}`)},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}, Ext: []byte(`{"prebid":{"bidder":{"spotx_vast_bidder":{"tags":[{"dur":15,"tagid":"spotx_15"},{"dur":25,"tagid":"spotx_25"},{"dur":35,"tagid":"spotx_35"}]},"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":40,"tagid":"openx_40"}]}}}}`)},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}, Ext: []byte(`{"prebid":{"bidder":{"spotx_vast_bidder":{"tags":[{"dur":15,"tagid":"spotx_15"},{"dur":25,"tagid":"spotx_25"},{"dur":35,"tagid":"spotx_35"}]},"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":40,"tagid":"openx_40"}]}}}}`)},
					},
				},
				impData: []*types.ImpData{
					{ImpID: "imp1"},
				},
			},
			expectedOutput: output{
				reqs: openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}, Ext: []byte(`{"prebid":{"bidder":{}}}`)},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}, Ext: []byte(`{"prebid":{"bidder":{"spotx_vast_bidder":{"tags":[{"dur":15,"tagid":"spotx_15"}]}}}}`)},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}, Ext: []byte(`{"prebid":{"bidder":{"openx_vast_bidder":{"tags":[{"dur":25,"tagid":"openx_25"}]},"spotx_vast_bidder":{"tags":[{"dur":25,"tagid":"spotx_25"}]}}}}`)},
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
						{ID: "imp1", Ext: []byte(`{"prebid":{"bidder":{"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":20,"tagid":"openx_20"}]}}}}`)},
						{ID: "imp2", Ext: []byte(`{"prebid":{"bidder":{"spotx_vast_bidder":{"tags":[{"dur":30,"tagid":"spotx_30"},{"dur":40,"tagid":"spotx_40"}]}}}}`)},
					},
				},
				generatedRequest: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}, Ext: []byte(`{"prebid":{"bidder":{"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":20,"tagid":"openx_20"}]}}}}`)},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}, Ext: []byte(`{"prebid":{"bidder":{"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":20,"tagid":"openx_20"}]}}}}`)},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}, Ext: []byte(`{"prebid":{"bidder":{"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":20,"tagid":"openx_20"}]}}}}`)},
						{ID: "imp2_1", Video: &openrtb2.Video{MinDuration: 5, MaxDuration: 30}, Ext: []byte(`{"prebid":{"bidder":{"spotx_vast_bidder":{"tags":[{"dur":30,"tagid":"spotx_30"},{"dur":40,"tagid":"spotx_40"}]}}}}`)},
					},
				},
				impData: nil,
			},
			expectedOutput: output{
				reqs: openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}, Ext: []byte(`{"prebid":{"bidder":{}}}`)},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}, Ext: []byte(`{"prebid":{"bidder":{"openx_vast_bidder":{"tags":[{"dur":20,"tagid":"openx_20"}]}}}}`)},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}, Ext: []byte(`{"prebid":{"bidder":{"openx_vast_bidder":{"tags":[{"dur":25,"tagid":"openx_25"}]}}}}`)},
						{ID: "imp2_1", Video: &openrtb2.Video{MinDuration: 5, MaxDuration: 30}, Ext: []byte(`{"prebid":{"bidder":{"spotx_vast_bidder":{"tags":[{"dur":30,"tagid":"spotx_30"}]}}}}`)},
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
						{ID: "imp1", Ext: []byte(`{"prebid":{"bidder":{"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":20,"tagid":"openx_20"}]}}}}`)},
						{ID: "imp2", Ext: []byte(`{"prebid":{"bidder":{"spotx_vast_bidder":{"tags":[{"dur":30,"tagid":"spotx_30"},{"dur":40,"tagid":"spotx_40"}]}}}}`)},
					},
				},
				generatedRequest: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}, Ext: []byte(`{"prebid":{"bidder":{"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":20,"tagid":"openx_20"}]}}}}`)},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}, Ext: []byte(`{"prebid":{"bidder":{"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":20,"tagid":"openx_20"}]}}}}`)},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}, Ext: []byte(`{"prebid":{"bidder":{"openx_vast_bidder":{"tags":[{"dur":35,"tagid":"openx_35"},{"dur":25,"tagid":"openx_25"},{"dur":20,"tagid":"openx_20"}]}}}}`)},
						{ID: "imp2_1", Video: &openrtb2.Video{MinDuration: 5, MaxDuration: 30}, Ext: []byte(`{"prebid":{"bidder":{"spotx_vast_bidder":{"tags":[{"dur":30,"tagid":"spotx_30"},{"dur":40,"tagid":"spotx_40"}]}}}}`)},
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
						{ID: "imp1_1", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 10}, Ext: []byte(`{"prebid":{"bidder":{}}}`)},
						{ID: "imp1_2", Video: &openrtb2.Video{MinDuration: 10, MaxDuration: 20}, Ext: []byte(`{"prebid":{"bidder":{"openx_vast_bidder":{"tags":[{"dur":20,"tagid":"openx_20"}]}}}}`)},
						{ID: "imp1_3", Video: &openrtb2.Video{MinDuration: 25, MaxDuration: 30}, Ext: []byte(`{"prebid":{"bidder":{"openx_vast_bidder":{"tags":[{"dur":25,"tagid":"openx_25"}]}}}}`)},
						{ID: "imp2_1", Video: &openrtb2.Video{MinDuration: 5, MaxDuration: 30}, Ext: []byte(`{"prebid":{"bidder":{"spotx_vast_bidder":{"tags":[{"dur":30,"tagid":"spotx_30"}]}}}}`)},
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

func Test_getDurationBasedOnDurationMatchingPolicy(t *testing.T) {
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

func TestCreateAdPodBidResponse(t *testing.T) {
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
			actual := deps.createAdPodBidResponse(tt.args.resp, nil)
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
			actual := deps.getBidResponseExt(tt.args.resp)
			assert.Equal(t, string(tt.want.data), string(actual))
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

		deps := ctvEndpointDeps{
			endpointDeps: endpointDeps{
				metricsEngine: me,
			},
			impData: []*types.ImpData{
				{
					Bid: &test.args.bids,
				},
			},
		}

		deps.recordRejectedAdPodBids("pub_001")
		me.AssertNumberOfCalls(t, "RecordRejectedBids", test.want.expectedCalls)
	}
}
