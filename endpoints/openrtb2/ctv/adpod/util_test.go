package adpod

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
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
			assert.Nil(t, AddTargetingKey(bid, key, test.value))
			extBid := openrtb_ext.ExtBid{}
			json.Unmarshal(bid.Ext, &extBid)
			assert.Equal(t, test.expect, extBid.Prebid.Targeting)
		})
	}
	assert.Equal(t, "Invalid bid", AddTargetingKey(nil, openrtb_ext.HbCategoryDurationKey, "some value").Error())
}

func TestConvertToV25VideoRequest(t *testing.T) {
	type args struct {
		request *openrtb2.BidRequest
	}
	tests := []struct {
		name string
		args args
		want *openrtb2.BidRequest
	}{
		{
			name: "Remove adpod parameters",
			args: args{
				request: &openrtb2.BidRequest{
					ID: "request",
					Imp: []openrtb2.Imp{
						{
							ID: "imp1",
							Video: &openrtb2.Video{
								MinDuration: 10,
								MaxDuration: 30,
								PodDur:      90,
								PodID:       "pod1",
								MaxSeq:      3,
							},
						},
					},
				},
			},
			want: &openrtb2.BidRequest{
				ID: "request",
				Imp: []openrtb2.Imp{
					{
						ID: "imp1",
						Video: &openrtb2.Video{
							MinDuration: 10,
							MaxDuration: 30,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ConvertToV25VideoRequest(tt.args.request)
			assert.Equal(t, tt.want, tt.args.request, "Failed to remove adpod paramaters")
		})
	}
}
