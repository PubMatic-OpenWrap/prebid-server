package openrtb2

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/adpod"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/v2/metrics"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

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

			deps := ctvEndpointDeps{request: tc.input.request}
			deps.readImpExtensionsAndTags()

			outputBids := tc.input.generatedRequest
			deps.filterImpsVastTagsByDuration(outputBids)

			assert.Equal(t, tc.expectedOutput.reqs, *outputBids, "Expected length of impressions array was %d but actual was %d", tc.expectedOutput.reqs, outputBids)
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
			actual := deps.createAdPodBidResponse(tt.args.resp)
			assert.Equal(t, tt.want.resp, actual)
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

			req := &openrtb2.BidRequest{
				Imp: []openrtb2.Imp{
					{
						ID:    "imp1",
						Video: &openrtb2.Video{},
					},
				},
			}

			videoExt := openrtb_ext.ExtVideoAdPod{
				AdPod: &openrtb_ext.VideoAdPod{},
			}
			dynamicAdpod := adpod.NewDynamicAdpod("test-pub", req.Imp[0], videoExt, &metrics.MetricsEngineMock{}, nil)

			deps := ctvEndpointDeps{
				podCtx: map[string]adpod.Adpod{
					"imp1": dynamicAdpod,
				},
				request: req,
			}
			actual := deps.getBidResponseExt(tt.args.resp)
			assert.Equal(t, string(tt.want.data), string(actual))
		})
	}
}

func TestGetAdpodConfigFromExtension(t *testing.T) {
	type fields struct {
		endpointDeps              endpointDeps
		request                   *openrtb2.BidRequest
		reqExt                    *openrtb_ext.ExtRequestAdPod
		videoSeats                []*openrtb2.SeatBid
		impsExtPrebidBidder       map[string]map[string]map[string]interface{}
		impPartnerBlockedTagIDMap map[string]map[string][]string
		podCtx                    map[string]adpod.Adpod
		labels                    metrics.Labels
	}
	type args struct {
		imp openrtb2.Imp
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    openrtb_ext.ExtVideoAdPod
		wantErr bool
	}{
		{
			name: "Adpod_Config_available_in_the_impression_extension",
			fields: fields{
				endpointDeps: endpointDeps{},
				request:      &openrtb2.BidRequest{},
			},
			args: args{
				imp: openrtb2.Imp{
					ID:    "imp1",
					TagID: "/Test/unit",
					Video: &openrtb2.Video{
						MinDuration: 10,
						MaxDuration: 30,
						Ext:         json.RawMessage(`{"offset":20,"adpod":{"minads":2,"maxads":3,"adminduration":30,"admaxduration":40,"excladv":100,"excliabcat":100}}`),
					},
				},
			},
			want: openrtb_ext.ExtVideoAdPod{
				Offset: ptrutil.ToPtr(20),
				AdPod: &openrtb_ext.VideoAdPod{
					MinAds:                      ptrutil.ToPtr(2),
					MaxAds:                      ptrutil.ToPtr(3),
					MinDuration:                 ptrutil.ToPtr(30),
					MaxDuration:                 ptrutil.ToPtr(40),
					AdvertiserExclusionPercent:  ptrutil.ToPtr(100),
					IABCategoryExclusionPercent: ptrutil.ToPtr(100),
				},
			},
		},
		{
			name: "video_extension_contains_values_other_than_adpod",
			fields: fields{
				endpointDeps: endpointDeps{},
				request:      &openrtb2.BidRequest{},
			},
			args: args{
				imp: openrtb2.Imp{
					ID:    "imp1",
					TagID: "/Test/unit",
					Video: &openrtb2.Video{
						MinDuration: 10,
						MaxDuration: 30,
						Ext:         json.RawMessage(`{"random":20}`),
					},
				},
			},
			want: openrtb_ext.ExtVideoAdPod{},
		},
		{
			name: "adpod_configuration_present_in_request_extension",
			fields: fields{
				endpointDeps: endpointDeps{},
				request:      &openrtb2.BidRequest{},
				reqExt: &openrtb_ext.ExtRequestAdPod{
					VideoAdPod: &openrtb_ext.VideoAdPod{
						MinAds:                      ptrutil.ToPtr(1),
						MaxAds:                      ptrutil.ToPtr(3),
						MinDuration:                 ptrutil.ToPtr(10),
						MaxDuration:                 ptrutil.ToPtr(30),
						AdvertiserExclusionPercent:  ptrutil.ToPtr(100),
						IABCategoryExclusionPercent: ptrutil.ToPtr(100),
					},
				},
			},
			args: args{
				imp: openrtb2.Imp{
					ID:    "imp1",
					TagID: "/Test/unit",
					Video: &openrtb2.Video{
						MinDuration: 10,
						MaxDuration: 30,
					},
				},
			},
			want: openrtb_ext.ExtVideoAdPod{
				Offset: ptrutil.ToPtr(0),
				AdPod: &openrtb_ext.VideoAdPod{
					MinAds:                      ptrutil.ToPtr(1),
					MaxAds:                      ptrutil.ToPtr(3),
					MinDuration:                 ptrutil.ToPtr(5),
					MaxDuration:                 ptrutil.ToPtr(15),
					AdvertiserExclusionPercent:  ptrutil.ToPtr(100),
					IABCategoryExclusionPercent: ptrutil.ToPtr(100),
				},
			},
		},
		{
			name: "adpod_configuration_not_availbale_in_any_location",
			fields: fields{
				endpointDeps: endpointDeps{},
				request:      &openrtb2.BidRequest{},
			},
			args: args{
				imp: openrtb2.Imp{
					ID:    "imp1",
					TagID: "/Test/unit",
					Video: &openrtb2.Video{
						MinDuration: 10,
						MaxDuration: 30,
					},
				},
			},
			want: openrtb_ext.ExtVideoAdPod{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := &ctvEndpointDeps{
				endpointDeps:              tt.fields.endpointDeps,
				request:                   tt.fields.request,
				reqExt:                    tt.fields.reqExt,
				videoSeats:                tt.fields.videoSeats,
				impsExtPrebidBidder:       tt.fields.impsExtPrebidBidder,
				impPartnerBlockedTagIDMap: tt.fields.impPartnerBlockedTagIDMap,
				podCtx:                    tt.fields.podCtx,
				labels:                    tt.fields.labels,
			}
			got, err := deps.readVideoAdPodExt(tt.args.imp)
			if (err != nil) != tt.wantErr {
				t.Errorf("ctvEndpointDeps.getAdpodConfigFromExtension() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, "Adpod config does not match")
		})
	}
}

func TestCTVAuctionEndpointAdpod(t *testing.T) {
	type args struct {
		w      http.ResponseWriter
		r      *http.Request
		params httprouter.Params
	}
	tests := []struct {
		name           string
		directory      string
		fileName       string
		args           args
		modifyResponse func(resp1, resp2 json.RawMessage) (json.RawMessage, error)
	}{
		{
			name:      "dynamic_adpod_request",
			args:      args{},
			directory: "sample-requests/ctv/valid-requests/",
			fileName:  "dynamic-adpod.json",
		},
		{
			name:      "structured_adpod_request",
			args:      args{},
			directory: "sample-requests/ctv/valid-requests/",
			fileName:  "structured-adpod.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read test case and unmarshal
			fileJsonData, err := os.ReadFile(tt.directory + tt.fileName)
			assert.NoError(t, err, "Failed to fetch a valid request: %v. Test file: %s", err, tt.fileName)

			test := ctvtestCase{}
			assert.NoError(t, json.Unmarshal(fileJsonData, &test), "Failed to unmarshal data from file: %s. Error: %v", tt.fileName, err)

			tt.args.r = httptest.NewRequest("POST", "/video/json", bytes.NewReader(test.BidRequest))
			recorder := httptest.NewRecorder()

			cfg := &config.Configuration{
				MaxRequestSize: maxSize,
				GDPR:           config.GDPR{Enabled: true},
			}
			if test.Config != nil {
				cfg.BlacklistedApps = test.Config.BlacklistedApps
				cfg.BlacklistedAppMap = test.Config.getBlacklistedAppMap()
				cfg.AccountRequired = test.Config.AccountRequired
			}

			CTVAuctionEndpoint, _, mockBidServers, mockCurrencyRatesServer, err := ctvTestEndpoint(test, cfg)
			assert.NoError(t, err, "Error while calling ctv auction endpoint %v", err)

			CTVAuctionEndpoint(recorder, tt.args.r, tt.args.params)

			// Close servers
			for _, mockBidServer := range mockBidServers {
				mockBidServer.Close()
			}
			mockCurrencyRatesServer.Close()

			// if assert.Equal(t, test.ExpectedReturnCode, recorder.Code, "Expected status %d. Got %d. CTV test file: %s", http.StatusOK, recorder.Code, tt.fileName) {
			// 	if test.ExpectedReturnCode == http.StatusOK {
			// 		assert.JSONEq(t, string(test.ExpectedBidResponse), recorder.Body.String(), "Not the expected response. Test file: %s", tt.fileName)
			// 	} else {
			// 		assert.Equal(t, test.ExpectedErrorMessage, recorder.Body.String(), tt.fileName)
			// 	}
			// }
		})
	}
}
