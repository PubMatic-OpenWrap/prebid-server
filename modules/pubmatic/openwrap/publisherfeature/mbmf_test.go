package publisherfeature

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache"
	mock_cache "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestFeatureUpdateMBMF(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	type fields struct {
		cache            cache.Cache
		publisherFeature map[int]map[int]models.FeatureData
		mbmf             mbmf
	}
	tests := []struct {
		name     string
		fields   fields
		setup    func()
		wantMBMF mbmf
	}{
		{
			name: "publisherFeature_map_is_nil",
			fields: fields{
				cache:            nil,
				publisherFeature: nil,
				mbmf:             newMBMF(),
			},
			setup: func() {},
			wantMBMF: mbmf{
				enabledCountries:         [2]models.HashSet{{}, {}},
				enabledPublishers:        [2]map[int]bool{{}, {}},
				profileAdUnitLevelFloors: [2]models.ProfileAdUnitMultiFloors{{}, {}},
				instlFloors:              [2]map[int]*models.MultiFloors{{}, {}},
				rwddFloors:               [2]map[int]*models.MultiFloors{{}, {}},
				index:                    0,
			},
		},
		{
			name: "publisherFeature_map_is_empty",
			fields: fields{
				cache:            mockCache,
				publisherFeature: map[int]map[int]models.FeatureData{},
				mbmf:             newMBMF(),
			},
			setup: func() {
				mockCache.EXPECT().GetProfileAdUnitMultiFloors().Return(models.ProfileAdUnitMultiFloors{}, nil)
			},
			wantMBMF: mbmf{
				enabledCountries:         [2]models.HashSet{{}, {}},
				enabledPublishers:        [2]map[int]bool{{}, {}},
				profileAdUnitLevelFloors: [2]models.ProfileAdUnitMultiFloors{{}, {}},
				instlFloors:              [2]map[int]*models.MultiFloors{{}, {}},
				rwddFloors:               [2]map[int]*models.MultiFloors{{}, {}},
				index:                    1,
			},
		},
		{
			name: "publisherFeature_map_contain_mbmf_country",
			fields: fields{
				cache: mockCache,
				publisherFeature: map[int]map[int]models.FeatureData{
					0: {
						models.FeatureMBMFCountry: {
							Enabled: 1,
							Value:   `US,DE`,
						},
					},
				},
				mbmf: newMBMF(),
			},
			setup: func() {
				mockCache.EXPECT().GetProfileAdUnitMultiFloors().Return(models.ProfileAdUnitMultiFloors{}, nil)
			},
			wantMBMF: mbmf{
				enabledCountries:         [2]models.HashSet{{}, {"US": {}, "DE": {}}},
				enabledPublishers:        [2]map[int]bool{{}, {}},
				profileAdUnitLevelFloors: [2]models.ProfileAdUnitMultiFloors{{}, {}},
				instlFloors:              [2]map[int]*models.MultiFloors{{}, {}},
				rwddFloors:               [2]map[int]*models.MultiFloors{{}, {}},
				index:                    1,
			},
		},
		{
			name: "publisherFeature_map_contain_mbmf_instl_floors",
			fields: fields{
				cache: mockCache,
				publisherFeature: map[int]map[int]models.FeatureData{
					0: {
						models.FeatureMBMFCountry: {
							Enabled: 1,
							Value:   `US,DE`,
						},
					},
					5890: {
						models.FeatureMBMFInstlFloors: {
							Enabled: 1,
							Value:   `{"tier1":1.0,"tier2":2.0,"tier3":3.0}`,
						},
					},
				},
				mbmf: newMBMF(),
			},
			setup: func() {
				mockCache.EXPECT().GetProfileAdUnitMultiFloors().Return(models.ProfileAdUnitMultiFloors{}, nil)
			},
			wantMBMF: mbmf{
				enabledCountries:         [2]models.HashSet{{}, {"US": {}, "DE": {}}},
				enabledPublishers:        [2]map[int]bool{{}, {}},
				profileAdUnitLevelFloors: [2]models.ProfileAdUnitMultiFloors{{}, {}},
				instlFloors:              [2]map[int]*models.MultiFloors{{}, {5890: {IsActive: true, Tier1: 1.0, Tier2: 2.0, Tier3: 3.0}}},
				rwddFloors:               [2]map[int]*models.MultiFloors{{}, {}},
				index:                    1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				cache:            tt.fields.cache,
				publisherFeature: tt.fields.publisherFeature,
				mbmf:             tt.fields.mbmf,
			}
			tt.setup()
			fe.updateMBMF()
			assert.Equal(t, tt.wantMBMF, fe.mbmf)
		})
	}
}

func TestFeatureUpdateMBMFCountries(t *testing.T) {
	type fields struct {
		cache            cache.Cache
		publisherFeature map[int]map[int]models.FeatureData
		mbmf             mbmf
	}
	tests := []struct {
		name                     string
		fields                   fields
		wantMBMFEnabledCountries [2]models.HashSet
	}{
		{
			name: "publisherFeature_map_is_present_but_mbmf_is_not_present_in_DB",
			fields: fields{
				cache:            nil,
				publisherFeature: map[int]map[int]models.FeatureData{},
				mbmf:             newMBMF(),
			},
			wantMBMFEnabledCountries: [2]models.HashSet{
				{},
				{},
			},
		},
		{
			name: "mbmf_enabled_countries",
			fields: fields{
				cache: nil,
				publisherFeature: map[int]map[int]models.FeatureData{
					0: {
						models.FeatureMBMFCountry: {
							Enabled: 1,
							Value:   `US,DE`,
						},
					},
				},
				mbmf: newMBMF(),
			},
			wantMBMFEnabledCountries: [2]models.HashSet{
				{},
				{
					"US": {},
					"DE": {},
				},
			},
		},
		{
			name: "mbmf_enabled_countries_in_flip_map",
			fields: fields{
				cache: nil,
				publisherFeature: map[int]map[int]models.FeatureData{
					0: {
						models.FeatureMBMFCountry: {
							Enabled: 1,
							Value:   `US,DE`,
						},
					},
				},
				mbmf: mbmf{
					enabledCountries: [2]models.HashSet{
						{
							"US": {},
							"DE": {},
						},
						{
							"CH": {},
							"DE": {},
						},
					},
				},
			},
			wantMBMFEnabledCountries: [2]models.HashSet{
				{
					"US": {},
					"DE": {},
				},
				{
					"US": {},
					"DE": {},
				},
			},
		},
		{
			name: "mbmf_enabled_countries_with_space_in_value",
			fields: fields{
				cache: nil,
				publisherFeature: map[int]map[int]models.FeatureData{
					0: {
						models.FeatureMBMFCountry: {
							Enabled: 1,
							Value:   `  US,DE  `,
						},
					},
				},
				mbmf: newMBMF(),
			},
			wantMBMFEnabledCountries: [2]models.HashSet{
				{},
				{
					"US": {},
					"DE": {},
				},
			},
		},
		{
			name: "mbmf_enabled_countries_with_feature_disabled",
			fields: fields{
				cache: nil,
				publisherFeature: map[int]map[int]models.FeatureData{
					0: {
						models.FeatureMBMFCountry: {
							Enabled: 0,
							Value:   `US,DE`,
						},
					},
				},
				mbmf: newMBMF(),
			},
			wantMBMFEnabledCountries: [2]models.HashSet{
				{},
				{},
			},
		},
		{
			name: "mbmf_enabled_countries_with_feature_enabled_but_empty_value",
			fields: fields{
				cache: nil,
				publisherFeature: map[int]map[int]models.FeatureData{
					0: {
						models.FeatureMBMFCountry: {
							Enabled: 1,
							Value:   ``,
						},
					},
				},
				mbmf: newMBMF(),
			},
			wantMBMFEnabledCountries: [2]models.HashSet{
				{},
				{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fe *feature
			fe = &feature{
				publisherFeature: tt.fields.publisherFeature,
				mbmf:             tt.fields.mbmf,
			}
			defer func() {
				fe = nil
			}()
			fe.updateMBMFCountries()
			assert.Equal(t, tt.wantMBMFEnabledCountries, fe.mbmf.enabledCountries)
		})
	}
}

func TestFeatureUpdateMBMFPublishers(t *testing.T) {
	type fields struct {
		cache            cache.Cache
		publisherFeature map[int]map[int]models.FeatureData
		mbmf             mbmf
	}
	tests := []struct {
		name   string
		fields fields
		want   [2]map[int]bool
	}{
		{
			name: "empty publisherFeature map",
			fields: fields{
				cache:            nil,
				publisherFeature: map[int]map[int]models.FeatureData{},
				mbmf:             newMBMF(),
			},
			want: [2]map[int]bool{
				make(map[int]bool),
				make(map[int]bool),
			},
		},
		{
			name: "publisher feature enabled and disabled",
			fields: fields{
				cache: nil,
				publisherFeature: map[int]map[int]models.FeatureData{
					123: {
						models.FeatureMBMFPublisher: {
							Enabled: 1,
						},
					},
					456: {
						models.FeatureMBMFPublisher: {
							Enabled: 0,
						},
					},
				},
				mbmf: newMBMF(),
			},
			want: [2]map[int]bool{
				make(map[int]bool),
				{
					123: true,
					456: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				publisherFeature: tt.fields.publisherFeature,
				mbmf:             tt.fields.mbmf,
			}
			fe.updateMBMFPublishers()
			assert.Equal(t, tt.want, fe.mbmf.enabledPublishers)
		})
	}
}

func TestFeatureUpdateProfileAdUnitLevelFloors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	type fields struct {
		cache cache.Cache
		mbmf  mbmf
	}
	tests := []struct {
		name               string
		fields             fields
		setup              func()
		expectedMBMFFloors [2]models.ProfileAdUnitMultiFloors
	}{
		{
			name: "query failed",
			fields: fields{
				cache: mockCache,
				mbmf:  newMBMF(),
			},
			setup: func() {
				mockCache.EXPECT().GetProfileAdUnitMultiFloors().Return(nil, errors.New("QUERY FAILED"))
			},
			expectedMBMFFloors: [2]models.ProfileAdUnitMultiFloors{
				{},
				{},
			},
		},
		{
			name: "query success",
			fields: fields{
				cache: mockCache,
				mbmf:  newMBMF(),
			},
			setup: func() {
				mockCache.EXPECT().GetProfileAdUnitMultiFloors().Return(models.ProfileAdUnitMultiFloors{
					123: {
						"adunit1": &models.MultiFloors{
							IsActive: true,
							Tier1:    1.0,
							Tier2:    2.0,
							Tier3:    3.0,
						},
					},
				}, nil)
			},
			expectedMBMFFloors: [2]models.ProfileAdUnitMultiFloors{
				{},
				{
					123: {
						"adunit1": &models.MultiFloors{
							IsActive: true,
							Tier1:    1.0,
							Tier2:    2.0,
							Tier3:    3.0,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			fe := &feature{
				cache: tt.fields.cache,
				mbmf:  tt.fields.mbmf,
			}
			defer func() {
				fe = nil
			}()
			fe.updateProfileAdUnitLevelFloors()
			assert.Equal(t, tt.expectedMBMFFloors, fe.mbmf.profileAdUnitLevelFloors, tt.name)
		})
	}
}

func TestFeatureUpdateMBMFInstlFloors(t *testing.T) {
	type fields struct {
		cache            cache.Cache
		publisherFeature map[int]map[int]models.FeatureData
		mbmf             mbmf
	}
	tests := []struct {
		name   string
		fields fields
		want   [2]map[int]*models.MultiFloors
	}{
		{
			name: "empty publisherFeature map",
			fields: fields{
				cache:            nil,
				publisherFeature: map[int]map[int]models.FeatureData{},
				mbmf:             newMBMF(),
			},
			want: [2]map[int]*models.MultiFloors{
				make(map[int]*models.MultiFloors),
				make(map[int]*models.MultiFloors),
			},
		},
		{
			name: "instl floors enabled and disabled",
			fields: fields{
				cache: nil,
				publisherFeature: map[int]map[int]models.FeatureData{
					123: {
						models.FeatureMBMFInstlFloors: {
							Enabled: 1,
							Value:   `{"isActive":true,"tier1":1.5,"tier2":2.0,"tier3":2.5}`,
						},
					},
					456: {
						models.FeatureMBMFInstlFloors: {
							Enabled: 0,
						},
					},
				},
				mbmf: newMBMF(),
			},
			want: [2]map[int]*models.MultiFloors{
				make(map[int]*models.MultiFloors),
				{
					123: {
						IsActive: true,
						Tier1:    1.5,
						Tier2:    2.0,
						Tier3:    2.5,
					},
					456: {
						IsActive: false,
					},
				},
			},
		},
		{
			name: "invalid json in floors value",
			fields: fields{
				cache: nil,
				publisherFeature: map[int]map[int]models.FeatureData{
					123: {
						models.FeatureMBMFInstlFloors: {
							Enabled: 1,
							Value:   `invalid json`,
						},
					},
				},
				mbmf: newMBMF(),
			},
			want: [2]map[int]*models.MultiFloors{
				make(map[int]*models.MultiFloors),
				make(map[int]*models.MultiFloors),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				publisherFeature: tt.fields.publisherFeature,
				mbmf:             tt.fields.mbmf,
			}
			fe.updateMBMFInstlFloors()
			assert.Equal(t, tt.want, fe.mbmf.instlFloors)
		})
	}
}

func TestFeatureUpdateMBMFRwddFloors(t *testing.T) {
	type fields struct {
		cache            cache.Cache
		publisherFeature map[int]map[int]models.FeatureData
		mbmf             mbmf
	}
	tests := []struct {
		name   string
		fields fields
		want   [2]map[int]*models.MultiFloors
	}{
		{
			name: "empty publisherFeature map",
			fields: fields{
				cache:            nil,
				publisherFeature: map[int]map[int]models.FeatureData{},
				mbmf:             newMBMF(),
			},
			want: [2]map[int]*models.MultiFloors{
				make(map[int]*models.MultiFloors),
				make(map[int]*models.MultiFloors),
			},
		},
		{
			name: "rwdd floors enabled and disabled",
			fields: fields{
				cache: nil,
				publisherFeature: map[int]map[int]models.FeatureData{
					123: {
						models.FeatureMBMFRwddFloors: {
							Enabled: 1,
							Value:   `{"isActive":true,"tier1":2.5,"tier2":3.0,"tier3":3.5}`,
						},
					},
					456: {
						models.FeatureMBMFRwddFloors: {
							Enabled: 0,
						},
					},
				},
				mbmf: newMBMF(),
			},
			want: [2]map[int]*models.MultiFloors{
				make(map[int]*models.MultiFloors),
				{
					123: {
						IsActive: true,
						Tier1:    2.5,
						Tier2:    3.0,
						Tier3:    3.5,
					},
					456: {
						IsActive: false,
					},
				},
			},
		},
		{
			name: "invalid json in floors value",
			fields: fields{
				cache: nil,
				publisherFeature: map[int]map[int]models.FeatureData{
					123: {
						models.FeatureMBMFRwddFloors: {
							Enabled: 1,
							Value:   `invalid json`,
						},
					},
				},
				mbmf: newMBMF(),
			},
			want: [2]map[int]*models.MultiFloors{
				make(map[int]*models.MultiFloors),
				make(map[int]*models.MultiFloors),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				publisherFeature: tt.fields.publisherFeature,
				mbmf:             tt.fields.mbmf,
			}
			fe.updateMBMFRwddFloors()
			assert.Equal(t, tt.want, fe.mbmf.rwddFloors)
		})
	}
}

func TestFeatureIsMBMFCountry(t *testing.T) {
	type fields struct {
		mbmf mbmf
	}
	type args struct {
		countryCode string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "mbmf country enabled pub",
			args: args{
				countryCode: "IN",
			},
			fields: fields{
				mbmf: mbmf{
					enabledCountries: [2]models.HashSet{
						{
							"IN": {},
						},
						make(models.HashSet),
					},
				},
			},
			want: true,
		},
		{
			name: "mbmf country disabled pub",
			args: args{
				countryCode: "US",
			},
			fields: fields{
				mbmf: mbmf{
					enabledCountries: [2]models.HashSet{
						make(models.HashSet),
						{
							"IN": {},
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				mbmf: tt.fields.mbmf,
			}
			got := fe.IsMBMFCountry(tt.args.countryCode)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestFeatureIsMBMFPublisherEnabled(t *testing.T) {
	type fields struct {
		mbmf mbmf
	}
	type args struct {
		pubID int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "mbmf publisher enabled pub",
			args: args{
				pubID: 123,
			},
			fields: fields{
				mbmf: mbmf{
					enabledPublishers: [2]map[int]bool{
						{
							123: true,
						},
						make(map[int]bool),
					},
				},
			},
			want: true,
		},
		{
			name: "mbmf publisher disabled pub",
			args: args{
				pubID: 456,
			},
			fields: fields{
				mbmf: mbmf{
					enabledPublishers: [2]map[int]bool{
						{
							456: false,
						},
						make(map[int]bool),
					},
				},
			},
			want: false,
		},
		{
			name: "publisher not present in DB",
			args: args{
				pubID: 789,
			},
			fields: fields{
				mbmf: mbmf{
					enabledPublishers: [2]map[int]bool{
						{
							123: true,
						},
						make(map[int]bool),
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				mbmf: tt.fields.mbmf,
			}
			got := fe.IsMBMFPublisherEnabled(tt.args.pubID)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestFeatureIsMBMFEnabledForAdUnitFormat(t *testing.T) {
	type fields struct {
		mbmf mbmf
	}
	type args struct {
		pubID        int
		adunitFormat string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "empty adunitformat",
			args: args{
				pubID:        123,
				adunitFormat: "",
			},
			fields: fields{
				mbmf: mbmf{
					instlFloors: [2]map[int]*models.MultiFloors{
						{
							123: {
								IsActive: true,
							},
						},
						make(map[int]*models.MultiFloors),
					},
				},
			},
			want: false,
		},
		{
			name: "mbmf publisher enabled for instl",
			args: args{
				pubID:        123,
				adunitFormat: models.AdUnitFormatInstl,
			},
			fields: fields{
				mbmf: mbmf{
					instlFloors: [2]map[int]*models.MultiFloors{
						{
							123: {
								IsActive: true,
							},
						},
						make(map[int]*models.MultiFloors),
					},
				},
			},
			want: true,
		},
		{
			name: "mbmf publisher enabled for rwdd",
			args: args{
				pubID:        1234,
				adunitFormat: models.AdUnitFormatRwddVideo,
			},
			fields: fields{
				mbmf: mbmf{
					rwddFloors: [2]map[int]*models.MultiFloors{
						{
							1234: {
								IsActive: true,
							},
						},
						make(map[int]*models.MultiFloors),
					},
				},
			},
			want: true,
		},
		{
			name: "mbmf publisher disabled for instl",
			args: args{
				pubID:        456,
				adunitFormat: models.AdUnitFormatInstl,
			},
			fields: fields{
				mbmf: mbmf{
					instlFloors: [2]map[int]*models.MultiFloors{
						{
							456: {
								IsActive: false,
							},
						},
						make(map[int]*models.MultiFloors),
					},
				},
			},
			want: false,
		},
		{
			name: "publisher not present in DB",
			args: args{
				pubID:        789,
				adunitFormat: models.AdUnitFormatInstl,
			},
			fields: fields{
				mbmf: mbmf{
					instlFloors: [2]map[int]*models.MultiFloors{
						{
							123: {
								IsActive: true,
							},
						},
						make(map[int]*models.MultiFloors),
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				mbmf: tt.fields.mbmf,
			}
			got := fe.IsMBMFEnabledForAdUnitFormat(tt.args.pubID, tt.args.adunitFormat)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestFeatureGetMBMFFloorsForAdUnitFormat(t *testing.T) {
	type fields struct {
		mbmf mbmf
	}
	type args struct {
		pubID        int
		adunitFormat string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *models.MultiFloors
	}{
		{
			name: "empty adunitformat",
			args: args{
				pubID:        123,
				adunitFormat: "",
			},
			fields: fields{
				mbmf: mbmf{
					instlFloors: [2]map[int]*models.MultiFloors{
						{
							123: {
								IsActive: true,
							},
						},
						make(map[int]*models.MultiFloors),
					},
				},
			},
			want: nil,
		},
		{
			name: "mbmf publisher for instl return as is",
			args: args{
				pubID:        123,
				adunitFormat: models.AdUnitFormatInstl,
			},
			fields: fields{
				mbmf: mbmf{
					instlFloors: [2]map[int]*models.MultiFloors{
						{
							123: {
								IsActive: true,
								Tier1:    1,
								Tier2:    3,
								Tier3:    2,
							},
						},
						make(map[int]*models.MultiFloors),
					},
				},
			},
			want: &models.MultiFloors{
				IsActive: true,
				Tier1:    1,
				Tier2:    3,
				Tier3:    2,
			},
		},
		{
			name: "mbmf publisher enabled for rwdd return as is",
			args: args{
				pubID:        1234,
				adunitFormat: models.AdUnitFormatRwddVideo,
			},
			fields: fields{
				mbmf: mbmf{
					rwddFloors: [2]map[int]*models.MultiFloors{
						{
							1234: {
								IsActive: true,
								Tier1:    1,
								Tier2:    3,
								Tier3:    2,
							},
						},
						make(map[int]*models.MultiFloors),
					},
				},
			},
			want: &models.MultiFloors{
				IsActive: true,
				Tier1:    1,
				Tier2:    3,
				Tier3:    2,
			},
		},
		{
			name: "mbmf publisher disabled for instl return as is",
			args: args{
				pubID:        456,
				adunitFormat: models.AdUnitFormatInstl,
			},
			fields: fields{
				mbmf: mbmf{
					instlFloors: [2]map[int]*models.MultiFloors{
						{
							456: {
								IsActive: false,
								Tier1:    1,
								Tier2:    3,
								Tier3:    2,
							},
						},
						make(map[int]*models.MultiFloors),
					},
				},
			},
			want: &models.MultiFloors{
				IsActive: false,
				Tier1:    1,
				Tier2:    3,
				Tier3:    2,
			},
		},
		{
			name: "publisher not present in DB,return default",
			args: args{
				pubID:        789,
				adunitFormat: models.AdUnitFormatInstl,
			},
			fields: fields{
				mbmf: mbmf{
					instlFloors: [2]map[int]*models.MultiFloors{
						{
							models.DefaultAdUnitFormatFloors: {
								IsActive: true,
								Tier1:    1,
								Tier2:    2,
								Tier3:    3,
							},
							123: {
								IsActive: true,
								Tier1:    1,
								Tier2:    3,
								Tier3:    2,
							},
						},
						make(map[int]*models.MultiFloors),
					},
				},
			},
			want: &models.MultiFloors{
				IsActive: true,
				Tier1:    1,
				Tier2:    2,
				Tier3:    3,
			},
		},
		{
			name: "publisher not present and default floors missing",
			args: args{
				pubID:        789,
				adunitFormat: models.AdUnitFormatInstl,
			},
			fields: fields{
				mbmf: mbmf{
					instlFloors: [2]map[int]*models.MultiFloors{
						{
							123: {
								IsActive: true,
								Tier1:    1,
								Tier2:    3,
								Tier3:    2,
							},
						},
						make(map[int]*models.MultiFloors),
					},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				mbmf: tt.fields.mbmf,
			}
			got := fe.GetMBMFFloorsForAdUnitFormat(tt.args.pubID, tt.args.adunitFormat)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestFeatureGetProfileAdUnitMultiFloors(t *testing.T) {
	type fields struct {
		mbmf mbmf
	}
	type args struct {
		profileID int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]*models.MultiFloors
	}{
		{
			name: "profileID present",
			fields: fields{
				mbmf: mbmf{
					profileAdUnitLevelFloors: [2]models.ProfileAdUnitMultiFloors{
						{
							123: {
								"adunit1": {
									IsActive: true,
									Tier1:    1,
									Tier2:    2,
									Tier3:    3,
								},
							},
						},
						make(models.ProfileAdUnitMultiFloors),
					},
				},
			},
			args: args{
				profileID: 123,
			},
			want: map[string]*models.MultiFloors{
				"adunit1": {
					IsActive: true,
					Tier1:    1,
					Tier2:    2,
					Tier3:    3,
				},
			},
		},
		{
			name: "profileID not present",
			fields: fields{
				mbmf: mbmf{
					profileAdUnitLevelFloors: [2]models.ProfileAdUnitMultiFloors{
						{
							123: {
								"adunit1": {
									IsActive: true,
									Tier1:    1,
									Tier2:    2,
									Tier3:    3,
								},
							},
						},
						make(models.ProfileAdUnitMultiFloors),
					},
				},
			},
			args: args{
				profileID: 456,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				mbmf: tt.fields.mbmf,
			}
			got := fe.GetProfileAdUnitMultiFloors(tt.args.profileID)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestExtractMultiFloors(t *testing.T) {
	type args struct {
		feature    map[int]models.FeatureData
		featureKey int
		pubID      int
	}
	tests := []struct {
		name string
		args args
		want *models.MultiFloors
	}{
		{
			name: "empty feature map",
			args: args{
				feature:    map[int]models.FeatureData{},
				featureKey: models.FeatureMBMFInstlFloors,
				pubID:      123,
			},
			want: nil,
		},
		{
			name: "feature name not present",
			args: args{
				feature: map[int]models.FeatureData{
					12345: {
						Value: "somevalue",
					},
				},
				featureKey: models.FeatureMBMFInstlFloors,
				pubID:      123,
			},
			want: nil,
		},
		{
			name: "value not present",
			args: args{
				feature: map[int]models.FeatureData{
					models.FeatureMBMFInstlFloors: {
						Enabled: 1,
					},
				},
				featureKey: models.FeatureMBMFInstlFloors,
				pubID:      123,
			},
			want: nil,
		},
		{
			name: "invalid json",
			args: args{
				feature: map[int]models.FeatureData{
					models.FeatureMBMFInstlFloors: {
						Enabled: 1,
						Value:   "invalidjson",
					},
				},
				featureKey: models.FeatureMBMFInstlFloors,
				pubID:      123,
			},
			want: nil,
		},
		{
			name: "valid jso for instl",
			args: args{
				feature: map[int]models.FeatureData{
					models.FeatureMBMFInstlFloors: {
						Enabled: 1,
						Value:   `{"isActive":true,"tier1":1.5,"tier2":2.0,"tier3":2.5}`,
					},
				},
				featureKey: models.FeatureMBMFInstlFloors,
				pubID:      123,
			},
			want: &models.MultiFloors{
				IsActive: true,
				Tier1:    1.5,
				Tier2:    2.0,
				Tier3:    2.5,
			},
		},
		{
			name: "valid json for rwdd",
			args: args{
				feature: map[int]models.FeatureData{
					models.FeatureMBMFRwddFloors: {
						Enabled: 1,
						Value:   `{"isActive":true,"tier1":1.5,"tier2":2.0,"tier3":2.5}`,
					},
				},
				featureKey: models.FeatureMBMFRwddFloors,
				pubID:      123,
			},
			want: &models.MultiFloors{
				IsActive: true,
				Tier1:    1.5,
				Tier2:    2.0,
				Tier3:    2.5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractMultiFloors(tt.args.feature, tt.args.featureKey, tt.args.pubID)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
