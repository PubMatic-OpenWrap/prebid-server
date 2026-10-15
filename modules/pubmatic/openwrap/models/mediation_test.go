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
			name: "name_and_type",
			path: AppSubIntegrationPath{ID: 16, MediationName: "admob", MediationPath: "bidding"},
			want: &ImpMediation{Name: "admob", Path: "bidding"},
		},
		{
			name: "waterfall_type",
			path: AppSubIntegrationPath{ID: 2, MediationName: "google_ad_manager", MediationPath: "waterfall"},
			want: &ImpMediation{Name: "google_ad_manager", Path: "waterfall"},
		},
		{
			name: "empty_name",
			path: AppSubIntegrationPath{ID: 1, MediationPath: "bidding"},
		},
		{
			name: "empty_type",
			path: AppSubIntegrationPath{ID: 1, MediationName: "admob"},
		},
		{
			name: "empty_name_and_type",
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
			name: "server_side_pubmatic",
			partnerConfigMap: map[int]map[string]string{
				1: {PREBID_PARTNER_NAME: BidderPubMatic, SERVER_SIDE_FLAG: "1"},
			},
			want: true,
		},
		{
			name: "server_side_pubmatic2",
			partnerConfigMap: map[int]map[string]string{
				2: {PREBID_PARTNER_NAME: BidderPubMaticSecondaryAlias, SERVER_SIDE_FLAG: "1"},
			},
			want: true,
		},
		{
			name: "client_side_pubmatic",
			partnerConfigMap: map[int]map[string]string{
				1: {PREBID_PARTNER_NAME: BidderPubMatic, SERVER_SIDE_FLAG: "0"},
			},
		},
		{
			name: "version_level_config_is_ignored",
			partnerConfigMap: map[int]map[string]string{
				VersionLevelConfigID: {PREBID_PARTNER_NAME: BidderPubMatic, SERVER_SIDE_FLAG: "1"},
			},
		},
		{
			name: "other_server_side_partner",
			partnerConfigMap: map[int]map[string]string{
				3: {PREBID_PARTNER_NAME: "appnexus", SERVER_SIDE_FLAG: "1"},
			},
		},
		{
			name: "empty_map",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, HasPubMaticPartner(tt.partnerConfigMap))
		})
	}
}

func TestImpMediationJSON(t *testing.T) {
	raw, err := json.Marshal(ImpMediation{Name: "applovin", Path: "bidding"})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"name":"applovin","path":"bidding"}`, string(raw))

	raw, err = json.Marshal(ImpMediation{})
	assert.NoError(t, err)
	assert.Equal(t, `{}`, string(raw))
}
