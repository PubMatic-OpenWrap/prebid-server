package publisherfeature

import (
	"testing"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func Test_feature_GetApplovinABTestFloors(t *testing.T) {
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

func Test_feature_updateApplovinABTestFeature(t *testing.T) {
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				publisherFeature:    tt.fields.publisherFeature,
				appLovinMultiFloors: tt.fields.appLovinMultiFloors,
			}
			fe.updateApplovinABTestFeature()
			assert.Equal(t, tt.want, fe.appLovinMultiFloors.enabledPublisherProfile, tt.name)
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
