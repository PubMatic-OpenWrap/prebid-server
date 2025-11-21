package publisherfeature

import (
	"testing"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func Test_feature_GetApplovinMultiFloors(t *testing.T) {
	type fields struct {
		publisherFeature    map[int]map[int]models.FeatureData
		appLovinMultiFloors appLovinMultiFloors
	}
	type args struct {
		pubID     int
		profileID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   models.ApplovinAdUnitFloors
	}{
		{
			name: "applovin adunitfloors not found",
			fields: fields{
				appLovinMultiFloors: appLovinMultiFloors{},
			},
			args: args{
				pubID:     5890,
				profileID: "1234",
			},
			want: models.ApplovinAdUnitFloors{},
		},
		{
			name: "applovin adunitfloors found",
			fields: fields{
				appLovinMultiFloors: appLovinMultiFloors{
					enabledPublisherProfile: map[int]map[string]models.ApplovinAdUnitFloors{
						5890: {
							"1232": models.ApplovinAdUnitFloors{
								"adunit_123":    {4.2, 5.6, 5.8},
								"adunit_dmdemo": {4.2, 5.6, 5.8},
							},
							"4322": models.ApplovinAdUnitFloors{
								"adunit_12323":   {4.2, 5.6, 5.8},
								"adunit_dmdemo1": {4.2, 5.6, 5.8},
							},
						},
					},
				},
			},
			args: args{
				pubID:     5890,
				profileID: "4322",
			},
			want: models.ApplovinAdUnitFloors{
				"adunit_12323":   {4.2, 5.6, 5.8},
				"adunit_dmdemo1": {4.2, 5.6, 5.8},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				publisherFeature:    tt.fields.publisherFeature,
				appLovinMultiFloors: tt.fields.appLovinMultiFloors,
			}
			got := fe.GetApplovinMultiFloors(tt.args.pubID, tt.args.profileID)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func Test_feature_updateApplovinMultiFloorsFeature(t *testing.T) {
	type fields struct {
		publisherFeature    map[int]map[int]models.FeatureData
		appLovinMultiFloors appLovinMultiFloors
	}
	tests := []struct {
		name   string
		fields fields
		want   map[int]map[string]models.ApplovinAdUnitFloors
	}{
		{
			name: "publisherFeature map is nil",
			fields: fields{
				publisherFeature: nil,
			},
		},
		{
			name: "update applovin_abtest feature enabled pub",
			fields: fields{
				publisherFeature: map[int]map[int]models.FeatureData{
					5890: {
						3: models.FeatureData{
							Enabled: 1,
						},
						1: models.FeatureData{
							Enabled: 1,
						},
						7: models.FeatureData{
							Enabled: 1,
							Value:   `{"1232":{"adunit_123":[4.2,5.6,5.8],"adunit_dmdemo":[4.2,5.6,5.8]}}`,
						},
					},
					162990: {
						1: models.FeatureData{
							Enabled: 1,
						},
						7: models.FeatureData{
							Enabled: 1,
							Value:   `{"4322":{"adunit_12323":[4.2,5.6,5.8],"adunit_dmdemo1":[4.2,5.6,5.8]}}`,
						},
					},
					15123: {
						3: models.FeatureData{
							Enabled: 1,
						},
						1: models.FeatureData{
							Enabled: 1,
						},
						7: models.FeatureData{
							Enabled: 1,
							Value:   `{}`,
						},
					},
				},
				appLovinMultiFloors: appLovinMultiFloors{
					enabledPublisherProfile: make(map[int]map[string]models.ApplovinAdUnitFloors),
				},
			},
			want: map[int]map[string]models.ApplovinAdUnitFloors{
				5890: {
					"1232": models.ApplovinAdUnitFloors{
						"adunit_123":    {4.2, 5.6, 5.8},
						"adunit_dmdemo": {4.2, 5.6, 5.8},
					},
				},
				162990: {
					"4322": models.ApplovinAdUnitFloors{
						"adunit_12323":   {4.2, 5.6, 5.8},
						"adunit_dmdemo1": {4.2, 5.6, 5.8},
					},
				},
				15123: {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				publisherFeature:    tt.fields.publisherFeature,
				appLovinMultiFloors: tt.fields.appLovinMultiFloors,
			}
			fe.updateApplovinMultiFloorsFeature()
			assert.Equal(t, tt.want, fe.appLovinMultiFloors.enabledPublisherProfile, tt.name)
		})
	}
}

func Test_feature_updateApplovinSchainABTestFeature(t *testing.T) {
	type fields struct {
		publisherFeature     map[int]map[int]models.FeatureData
		appLovinSchainABTest appLovinSchainABTest
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "publisherFeature map is nil",
			fields: fields{
				publisherFeature: nil,
			},
			want: 0,
		},
		{
			name: "update applovin_schain_abtest feature enabled pub with valid percentage",
			fields: fields{
				publisherFeature: map[int]map[int]models.FeatureData{
					0: {
						models.FeatureAppLovinSchainABTest: models.FeatureData{
							Enabled: 1,
							Value:   "25",
						},
					},
				},
				appLovinSchainABTest: appLovinSchainABTest{
					schainABTestPercent: 0,
				},
			},
			want: 25, // Should take the percentage from pub 0
		},
		{
			name: "update applovin_schain_abtest feature with invalid percentage",
			fields: fields{
				publisherFeature: map[int]map[int]models.FeatureData{
					0: {
						models.FeatureAppLovinSchainABTest: models.FeatureData{
							Enabled: 1,
							Value:   "invalid",
						},
					},
				},
				appLovinSchainABTest: appLovinSchainABTest{
					schainABTestPercent: 0,
				},
			},
			want: 0, // Should reset to 0 as invalid percentage provided
		},
		{
			name: "update applovin_schain_abtest feature with empty value",
			fields: fields{
				publisherFeature: map[int]map[int]models.FeatureData{
					0: {
						models.FeatureAppLovinSchainABTest: models.FeatureData{
							Enabled: 1,
							Value:   "",
						},
					},
				},
				appLovinSchainABTest: appLovinSchainABTest{
					schainABTestPercent: 0,
				},
			},
			want: 0, // Should reset to 0 as empty value provided
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				publisherFeature:     tt.fields.publisherFeature,
				appLovinSchainABTest: tt.fields.appLovinSchainABTest,
			}
			fe.updateApplovinSchainABTestFeature()
			assert.Equal(t, tt.want, fe.appLovinSchainABTest.schainABTestPercent, tt.name)
		})
	}
}

func Test_feature_GetApplovinSchainABTestPercentage(t *testing.T) {
	type fields struct {
		appLovinSchainABTest appLovinSchainABTest
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "appLovinSchainABTest not found",
			fields: fields{
				appLovinSchainABTest: appLovinSchainABTest{},
			},
			want: 0,
		},
		{
			name: "get schain AB test percentage with disabled feature",
			fields: fields{
				appLovinSchainABTest: appLovinSchainABTest{
					schainABTestPercent: 0,
				},
			},
			want: 0,
		},
		{
			name: "get schain AB test percentage with positive value",
			fields: fields{
				appLovinSchainABTest: appLovinSchainABTest{
					schainABTestPercent: 25,
				},
			},
			want: 25,
		},
		{
			name: "get schain AB test percentage with maximum value",
			fields: fields{
				appLovinSchainABTest: appLovinSchainABTest{
					schainABTestPercent: 100,
				},
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				appLovinSchainABTest: tt.fields.appLovinSchainABTest,
			}
			got := fe.GetApplovinSchainABTestPercentage()
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func Test_feature_IsApplovinMultiFloorsEnabled(t *testing.T) {
	type fields struct {
		appLovinMultiFloors appLovinMultiFloors
	}
	type args struct {
		pubID     int
		profileID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "feature disabled",
			fields: fields{
				appLovinMultiFloors: appLovinMultiFloors{},
			},
			args: args{
				pubID:     5890,
				profileID: "1234",
			},
			want: false,
		},
		{
			name: "feature disabled for profile",
			fields: fields{
				appLovinMultiFloors: appLovinMultiFloors{
					enabledPublisherProfile: map[int]map[string]models.ApplovinAdUnitFloors{
						5890: {
							"1234": models.ApplovinAdUnitFloors{},
						},
					},
				},
			},
			args: args{
				pubID:     5890,
				profileID: "4345",
			},
			want: false,
		},
		{
			name: "feature enabled for pub-profile",
			fields: fields{
				appLovinMultiFloors: appLovinMultiFloors{
					enabledPublisherProfile: map[int]map[string]models.ApplovinAdUnitFloors{
						5890: {
							"4345": models.ApplovinAdUnitFloors{},
						},
					},
				},
			},
			args: args{
				pubID:     5890,
				profileID: "4345",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				appLovinMultiFloors: tt.fields.appLovinMultiFloors,
			}
			got := fe.IsApplovinMultiFloorsEnabled(tt.args.pubID, tt.args.profileID)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
