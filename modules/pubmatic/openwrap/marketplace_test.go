package openwrap

import (
	"testing"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestGetMarketplaceBidders(t *testing.T) {
	type args struct {
		reqABC           *openrtb_ext.ExtAlternateBidderCodes
		partnerConfigMap map[int]map[string]string
	}
	type want struct {
		alternateBidderCodes *openrtb_ext.ExtAlternateBidderCodes
		marketPlaceBidders   map[string]struct{}
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "happy_path,_marketplace_enabled_in_profile,_alternatebiddercodes_ext_should_be_build_using_profile_version_data",
			args: args{
				reqABC: nil,
				partnerConfigMap: map[int]map[string]string{
					models.VersionLevelConfigID: {
						models.MarketplaceBidders: "pubmatic,groupm",
					},
				},
			},
			want: want{
				alternateBidderCodes: &openrtb_ext.ExtAlternateBidderCodes{
					Enabled: true,
					Bidders: map[string]openrtb_ext.ExtAdapterAlternateBidderCodes{
						models.BidderPubMatic: {
							Enabled:            true,
							AllowedBidderCodes: []string{"pubmatic", "groupm"},
						},
					},
				},
				marketPlaceBidders: map[string]struct{}{
					"pubmatic": {},
					"groupm":   {},
				},
			},
		},
		{
			name: "pubmatic_not_present_in_profile_level_data,_alternatebiddercodes_ext_with_addition_of_pubmatic_bidder_should_be_build_using_profile_version_data",
			args: args{
				reqABC: nil,
				partnerConfigMap: map[int]map[string]string{
					models.VersionLevelConfigID: {
						models.MarketplaceBidders: "groupm",
					},
				},
			},
			want: want{
				alternateBidderCodes: &openrtb_ext.ExtAlternateBidderCodes{
					Enabled: true,
					Bidders: map[string]openrtb_ext.ExtAdapterAlternateBidderCodes{
						models.BidderPubMatic: {
							Enabled:            true,
							AllowedBidderCodes: []string{"groupm"},
						},
					},
				},
				marketPlaceBidders: map[string]struct{}{
					"groupm": {},
				},
			},
		},
		{
			name: "empty_incoming_alternatebiddercodes_ext_priority_should_be_given_to_request",
			args: args{
				reqABC: &openrtb_ext.ExtAlternateBidderCodes{},
				partnerConfigMap: map[int]map[string]string{
					models.VersionLevelConfigID: {
						models.MarketplaceBidders: "pubmatic,groupm",
					},
				},
			},
			want: want{
				alternateBidderCodes: &openrtb_ext.ExtAlternateBidderCodes{},
				marketPlaceBidders:   nil,
			},
		},
		{
			name: "incoming_request_has_alternatebiddercodes,_marketplace_enabled_in_profile,_request_data_has_priority",
			args: args{
				reqABC: &openrtb_ext.ExtAlternateBidderCodes{
					Enabled: true,
					Bidders: map[string]openrtb_ext.ExtAdapterAlternateBidderCodes{
						models.BidderPubMatic: {
							Enabled:            true,
							AllowedBidderCodes: []string{"pubmatic", "appnexus"},
						},
					},
				},
				partnerConfigMap: map[int]map[string]string{
					models.VersionLevelConfigID: {
						models.MarketplaceBidders: "pubmatic,groupm",
					},
				},
			},
			want: want{
				alternateBidderCodes: &openrtb_ext.ExtAlternateBidderCodes{
					Enabled: true,
					Bidders: map[string]openrtb_ext.ExtAdapterAlternateBidderCodes{
						models.BidderPubMatic: {
							Enabled:            true,
							AllowedBidderCodes: []string{"pubmatic", "appnexus"},
						},
					},
				},
				marketPlaceBidders: nil,
			},
		},
		{
			name: "marketplace_not_enabled_in profile,_alternatebiddercodes_ext_should_be_nil",
			args: args{
				reqABC: nil,
				partnerConfigMap: map[int]map[string]string{
					models.VersionLevelConfigID: {"k1": "v1"},
				},
			},
			want: want{
				alternateBidderCodes: nil,
				marketPlaceBidders:   nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alternateBidderCodes, marketPlaceBidders := getMarketplaceBidders(tt.args.reqABC, tt.args.partnerConfigMap)
			assert.Equal(t, tt.want.alternateBidderCodes, alternateBidderCodes)
			assert.Equal(t, tt.want.marketPlaceBidders, marketPlaceBidders)
		})
	}
}
