package adpod

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/constant"
	"github.com/prebid/prebid-server/v2/exchange"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
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

func TestConvertNBRCTOAPRC(t *testing.T) {
	type args struct {
		noBidReason *openrtb3.NoBidReason
	}
	tests := []struct {
		name string
		args args
		want *int64
	}{
		{
			name: "Test convert nbr to aprc for loss bid due to price",
			args: args{
				noBidReason: func() *openrtb3.NoBidReason {
					a := nbr.LossBidLostToHigherBid
					return &a
				}(),
			},
			want: ptrutil.ToPtr(constant.StatusOK),
		},
		{
			name: "Test convert nbr to aprc for category exclusion",
			args: args{
				noBidReason: func() *openrtb3.NoBidReason {
					a := exchange.ResponseRejectedCreativeCategoryExclusions
					return &a
				}(),
			},
			want: ptrutil.ToPtr(constant.StatusCategoryExclusion),
		},
		{
			name: "Test convert nbr to aprc for advertiser exclusion",
			args: args{
				noBidReason: func() *openrtb3.NoBidReason {
					a := exchange.ResponseRejectedCreativeAdvertiserExclusions
					return &a
				}(),
			},
			want: ptrutil.ToPtr(constant.StatusDomainExclusion),
		},
		{
			name: "Test convert nbr to aprc for invalid creative",
			args: args{
				noBidReason: func() *openrtb3.NoBidReason {
					a := exchange.ResponseRejectedInvalidCreative
					return &a
				}(),
			},
			want: ptrutil.ToPtr(constant.StatusDurationMismatch),
		},
		{
			name: "Test convert nbr to aprc for unknown reason",
			args: args{
				noBidReason: func() *openrtb3.NoBidReason {
					a := openrtb3.NoBidReason(999)
					return &a
				}(),
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertNBRCTOAPRC(tt.args.noBidReason)
			if !assert.Equal(t, tt.want, got) {
				t.Errorf("ConvertNBRCTOAPRC() = %v, want %v", got, tt.want)
			}
		})
	}
}
