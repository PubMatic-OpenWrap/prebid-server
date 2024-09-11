package publisherfeature

import (
	"testing"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestFeature_updateImpCountingMethodEnabledBidders(t *testing.T) {
	type fields struct {
		cache             cache.Cache
		publisherFeature  map[int]map[int]models.FeatureData
		impCountingMethod impCountingMethod
	}
	tests := []struct {
		name                               string
		fields                             fields
		wantImpCoutingMethodEnabledBidders map[string]struct{}
	}{
		{
			name: "publisherFeature map is nil",
			fields: fields{
				cache:            nil,
				publisherFeature: nil,
				impCountingMethod: impCountingMethod{
					enabledBidders: map[string]struct{}{},
				},
			},
			wantImpCoutingMethodEnabledBidders: map[string]struct{}{},
		},
		{
			name: "update imp counting method enabled bidders",
			fields: fields{
				cache: nil,
				publisherFeature: map[int]map[int]models.FeatureData{
					0: {
						models.FeatureImpCountingMethod: {
							Enabled: 1,
							Value:   `appnexus,rubicon`,
						},
					},
				},
			},
			wantImpCoutingMethodEnabledBidders: map[string]struct{}{
				"appnexus": {},
				"rubicon":  {},
			},
		},
		{
			name: "update imp counting method enabled bidders with space in value",
			fields: fields{
				cache: nil,
				publisherFeature: map[int]map[int]models.FeatureData{
					0: {
						models.FeatureImpCountingMethod: {
							Enabled: 1,
							Value:   `  appnexus,rubicon  `,
						},
					},
				},
			},
			wantImpCoutingMethodEnabledBidders: map[string]struct{}{
				"appnexus": {},
				"rubicon":  {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				publisherFeature:  tt.fields.publisherFeature,
				impCountingMethod: tt.fields.impCountingMethod,
			}
			fe.updateImpCountingMethodEnabledBidders()
			assert.Equal(t, tt.wantImpCoutingMethodEnabledBidders, fe.impCountingMethod.enabledBidders)
		})
	}
}

func TestFeature_GetImpCountingMethodEnabledBidders(t *testing.T) {
	type fields struct {
		impCountingMethod impCountingMethod
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]struct{}
	}{
		{
			name: "get imp counting method enabled bidders",
			fields: fields{
				impCountingMethod: impCountingMethod{
					enabledBidders: map[string]struct{}{
						"appnexus": {},
						"rubicon":  {},
					},
				},
			},
			want: map[string]struct{}{
				"appnexus": {},
				"rubicon":  {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				impCountingMethod: tt.fields.impCountingMethod,
			}
			got := fe.GetImpCountingMethodEnabledBidders()
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
