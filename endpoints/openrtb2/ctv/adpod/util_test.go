package adpod

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
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

func TestGetExclusionConfigs(t *testing.T) {
	tests := []struct {
		name     string
		podId    string
		adpodExt *openrtb_ext.ExtRequestAdPod
		expected Exclusion
	}{
		{
			name:     "Nil_adpodExt",
			podId:    "testPodId",
			adpodExt: nil,
			expected: Exclusion{},
		},
		{
			name:  "Nil_Exclusion",
			podId: "testPodId",
			adpodExt: &openrtb_ext.ExtRequestAdPod{
				Exclusion: nil,
			},
			expected: Exclusion{},
		},
		{
			name:  "IABCategory_exclusion_present",
			podId: "testPodId",
			adpodExt: &openrtb_ext.ExtRequestAdPod{
				Exclusion: &openrtb_ext.AdpodExclusion{
					IABCategory:      []string{"testPodId"},
					AdvertiserDomain: []string{"otherPodId"},
				},
			},
			expected: Exclusion{
				IABCategoryExclusion:      true,
				AdvertiserDomainExclusion: false,
			},
		},
		{
			name:  "AdvertiserDomain_exclusion_present",
			podId: "testPodId",
			adpodExt: &openrtb_ext.ExtRequestAdPod{
				Exclusion: &openrtb_ext.AdpodExclusion{
					IABCategory:      []string{"otherPodId"},
					AdvertiserDomain: []string{"testPodId"},
				},
			},
			expected: Exclusion{
				IABCategoryExclusion:      false,
				AdvertiserDomainExclusion: true,
			},
		},
		{
			name:  "No_exclusion_config_provided",
			podId: "testPodId",
			adpodExt: &openrtb_ext.ExtRequestAdPod{
				Exclusion: &openrtb_ext.AdpodExclusion{
					IABCategory:      []string{"otherPodId"},
					AdvertiserDomain: []string{"anotherPodId"},
				},
			},
			expected: Exclusion{
				IABCategoryExclusion:      false,
				AdvertiserDomainExclusion: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getExclusionConfigs(tt.podId, tt.adpodExt)
			assert.Equal(t, tt.expected, result)
		})
	}
}
