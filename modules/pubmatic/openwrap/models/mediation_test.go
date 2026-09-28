package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppSubIntegrationPathImpMediation(t *testing.T) {
	tests := []struct {
		name string
		path AppSubIntegrationPath
		want *ImpMediation
	}{
		{
			name: "name and type",
			path: AppSubIntegrationPath{ID: 16, MediationName: "admob", MediationType: "bidding"},
			want: &ImpMediation{Name: "admob", WaterfallOrBidding: "bidding"},
		},
		{
			name: "waterfall type",
			path: AppSubIntegrationPath{ID: 2, MediationName: "google_ad_manager", MediationType: "waterfall"},
			want: &ImpMediation{Name: "google_ad_manager", WaterfallOrBidding: "waterfall"},
		},
		{
			name: "empty name",
			path: AppSubIntegrationPath{ID: 1, MediationType: "bidding"},
		},
		{
			name: "empty type",
			path: AppSubIntegrationPath{ID: 1, MediationName: "admob"},
		},
		{
			name: "empty name and type",
			path: AppSubIntegrationPath{ID: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.path.ImpMediation())
		})
	}
}

func TestHasPubMaticPartner(t *testing.T) {
	tests := []struct {
		name             string
		partnerConfigMap map[int]map[string]string
		want             bool
	}{
		{
			name: "server side pubmatic",
			partnerConfigMap: map[int]map[string]string{
				1: {PREBID_PARTNER_NAME: BidderPubMatic, SERVER_SIDE_FLAG: "1"},
			},
			want: true,
		},
		{
			name: "server side pubmatic2",
			partnerConfigMap: map[int]map[string]string{
				2: {PREBID_PARTNER_NAME: BidderPubMaticSecondaryAlias, SERVER_SIDE_FLAG: "1"},
			},
			want: true,
		},
		{
			name: "client side pubmatic",
			partnerConfigMap: map[int]map[string]string{
				1: {PREBID_PARTNER_NAME: BidderPubMatic, SERVER_SIDE_FLAG: "0"},
			},
		},
		{
			name: "version level config is ignored",
			partnerConfigMap: map[int]map[string]string{
				VersionLevelConfigID: {PREBID_PARTNER_NAME: BidderPubMatic, SERVER_SIDE_FLAG: "1"},
			},
		},
		{
			name: "other server side partner",
			partnerConfigMap: map[int]map[string]string{
				3: {PREBID_PARTNER_NAME: "appnexus", SERVER_SIDE_FLAG: "1"},
			},
		},
		{
			name: "empty map",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, HasPubMaticPartner(tt.partnerConfigMap))
		})
	}
}

func TestImpMediationJSON(t *testing.T) {
	raw, err := json.Marshal(ImpMediation{Name: "applovin", WaterfallOrBidding: "bidding"})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"applovin","waterfall_or_bidding":"bidding"}`, string(raw))

	raw, err = json.Marshal(ImpMediation{})
	assert.NoError(t, err)
	assert.Equal(t, `{}`, string(raw))
}
