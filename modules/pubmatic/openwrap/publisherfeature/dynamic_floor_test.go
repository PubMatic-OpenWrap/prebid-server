package publisherfeature

import (
	"testing"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestUpdateDynamicFloorEnabledPublishers(t *testing.T) {
	testCases := []struct {
		name                              string
		publisherFeature                  map[int]map[int]models.FeatureData
		wantDynamicFloorEnabledPublishers map[int]struct{}
	}{
		{
			name:                              "publisherFeature map is nil",
			publisherFeature:                  nil,
			wantDynamicFloorEnabledPublishers: map[int]struct{}{},
		},
		{
			name: "update dynamic floor feature enabled pub",
			publisherFeature: map[int]map[int]models.FeatureData{
				5890: {
					models.FeatureDynamicFloor: models.FeatureData{
						Enabled: 1,
					},
				},
				5891: {
					models.FeatureDynamicFloor: models.FeatureData{
						Enabled: 1,
					},
				},
				5892: {
					models.FeatureDynamicFloor: models.FeatureData{
						Enabled: 0,
					},
				},
			},
			wantDynamicFloorEnabledPublishers: map[int]struct{}{
				5890: {},
				5891: {},
			},
		},
	}
	for _, tc := range testCases {
		fe := &feature{
			publisherFeature: tc.publisherFeature,
			dynamicFloor:     newDynamicFloor(),
		}
		fe.updateDynamicFloorEnabledPublishers()
		if len(fe.dynamicFloor.enabledPublishers[fe.dynamicFloor.index]) != len(tc.wantDynamicFloorEnabledPublishers) {
			t.Errorf("[%s] Unexpected count of dynamic floor enabled publishers. Got: %d, Want: %d", tc.name, len(fe.dynamicFloor.enabledPublishers[fe.dynamicFloor.index]), len(tc.wantDynamicFloorEnabledPublishers))
		}
		for pubID := range fe.dynamicFloor.enabledPublishers[fe.dynamicFloor.index] {
			if _, ok := tc.wantDynamicFloorEnabledPublishers[pubID]; !ok {
				t.Errorf("[%s] Unexpected dynamic floor enabled publishers. Got: %v, Want: %v", tc.name, fe.dynamicFloor.enabledPublishers[fe.dynamicFloor.index], tc.wantDynamicFloorEnabledPublishers)
			}
		}
	}
}

func TestIsDynamicFloorEnabledPublisher(t *testing.T) {
	type args struct {
		pubID             int
		enabledPublishers [2]map[int]struct{}
		index             int32
	}
	tests := []struct {
		name                          string
		args                          args
		expectedIsDynamicFloorEnabled bool
	}{
		{
			name: "index is nil",
			args: args{
				pubID:             5890,
				index:             0,
				enabledPublishers: [2]map[int]struct{}{},
			},
			expectedIsDynamicFloorEnabled: false,
		},
		{
			name: "no enabled publisher found",
			args: args{
				pubID: 5890,
				index: 0,
				enabledPublishers: [2]map[int]struct{}{
					make(map[int]struct{}),
					make(map[int]struct{}),
				},
			},
			expectedIsDynamicFloorEnabled: false,
		},
		{
			name: "enabled publisher found",
			args: args{
				pubID: 5891,
				index: 1,
				enabledPublishers: [2]map[int]struct{}{
					{
						5890: {},
					},
					{
						5890: {},
						5891: {},
					},
				},
			},
			expectedIsDynamicFloorEnabled: true,
		},
		{
			name: "disabled publisher found",
			args: args{
				pubID: 5891,
				index: 1,
				enabledPublishers: [2]map[int]struct{}{
					make(map[int]struct{}),
					{
						5890: {},
					},
				},
			},
			expectedIsDynamicFloorEnabled: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				dynamicFloor: dynamicFloor{
					enabledPublishers: tt.args.enabledPublishers,
					index:             tt.args.index,
				},
			}
			defer func() {
				fe = nil
			}()
			assert.Equal(t, tt.expectedIsDynamicFloorEnabled, fe.IsDynamicFloorEnabledPublisher(tt.args.pubID), tt.name)
		})
	}
}
