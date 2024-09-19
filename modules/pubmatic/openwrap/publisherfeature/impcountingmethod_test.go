package publisherfeature

import (
	"testing"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestFeatureUpdateImpCountingMethodEnabledBidders(t *testing.T) {
	type fields struct {
		cache             cache.Cache
		publisherFeature  map[int]map[int]models.FeatureData
		impCountingMethod impCountingMethod
	}
	tests := []struct {
		name                               string
		fields                             fields
		wantImpCoutingMethodEnabledBidders [2]map[string]struct{}
		wantImpCoutingMethodIndex          int32
	}{
		{
			name: "publisherFeature map is nil",
			fields: fields{
				cache:             nil,
				publisherFeature:  nil,
				impCountingMethod: newImpCountingMethod(),
			},
			wantImpCoutingMethodEnabledBidders: [2]map[string]struct{}{
				make(map[string]struct{}),
				make(map[string]struct{}),
			},
			wantImpCoutingMethodIndex: 0,
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
				impCountingMethod: newImpCountingMethod(),
			},
			wantImpCoutingMethodEnabledBidders: [2]map[string]struct{}{
				{},
				{
					"appnexus": {},
					"rubicon":  {},
				},
			},
			wantImpCoutingMethodIndex: 1,
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
				impCountingMethod: newImpCountingMethod(),
			},
			wantImpCoutingMethodEnabledBidders: [2]map[string]struct{}{
				{},
				{
					"appnexus": {},
					"rubicon":  {},
				},
			},
			wantImpCoutingMethodIndex: 1,
		},
		{
			name: "update imp counting method with feature disabled",
			fields: fields{
				cache: nil,
				publisherFeature: map[int]map[int]models.FeatureData{
					0: {
						models.FeatureImpCountingMethod: {
							Enabled: 0,
							Value:   `appnexus,rubicon`,
						},
					},
				},
				impCountingMethod: newImpCountingMethod(),
			},
			wantImpCoutingMethodEnabledBidders: [2]map[string]struct{}{
				{},
				{},
			},
			wantImpCoutingMethodIndex: 1,
		},
		{
			name: "update imp counting method with feature disabled",
			fields: fields{
				cache: nil,
				publisherFeature: map[int]map[int]models.FeatureData{
					0: {
						models.FeatureImpCountingMethod: {
							Enabled: 0,
							Value:   `appnexus,rubicon`,
						},
					},
				},
				impCountingMethod: newImpCountingMethod(),
			},
			wantImpCoutingMethodEnabledBidders: [2]map[string]struct{}{
				{},
				{},
			},
			wantImpCoutingMethodIndex: 1,
		},
		{
			name: "update imp counting method with feature enabled but empty value",
			fields: fields{
				cache: nil,
				publisherFeature: map[int]map[int]models.FeatureData{
					0: {
						models.FeatureImpCountingMethod: {
							Enabled: 1,
							Value:   ``,
						},
					},
				},
				impCountingMethod: newImpCountingMethod(),
			},
			wantImpCoutingMethodEnabledBidders: [2]map[string]struct{}{
				{},
				{},
			},
			wantImpCoutingMethodIndex: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fe *feature
			fe = &feature{
				publisherFeature:  tt.fields.publisherFeature,
				impCountingMethod: tt.fields.impCountingMethod,
			}
			defer func() {
				fe = nil
			}()
			fe.updateImpCountingMethodEnabledBidders()
			assert.Equal(t, tt.wantImpCoutingMethodEnabledBidders, fe.impCountingMethod.enabledBidders)
			assert.Equal(t, tt.wantImpCoutingMethodIndex, fe.impCountingMethod.index)
		})
	}
}

func TestFeatureGetImpCountingMethodEnabledBidders(t *testing.T) {
	type fields struct {
		impCountingMethod impCountingMethod
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]struct{}
	}{
		{
			name: "get imp counting method enabled bidders when index is 0",
			fields: fields{
				impCountingMethod: impCountingMethod{
					enabledBidders: [2]map[string]struct{}{
						{
							"appnexus": {},
							"rubicon":  {},
						},
					},
					index: 0,
				},
			},
			want: map[string]struct{}{
				"appnexus": {},
				"rubicon":  {},
			},
		},
		{
			name: "get imp counting method enabled bidders when index is 1",
			fields: fields{
				impCountingMethod: impCountingMethod{
					enabledBidders: [2]map[string]struct{}{
						{},
						{
							"appnexus": {},
							"rubicon":  {},
						},
					},
					index: 1,
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
