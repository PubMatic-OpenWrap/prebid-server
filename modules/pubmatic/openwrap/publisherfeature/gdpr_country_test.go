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

func TestFeatureUpdateGDPRCountryCodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	type fields struct {
		cache            cache.Cache
		gdprCountryCodes gdprCountryCodes
	}
	tests := []struct {
		name                     string
		fields                   fields
		setup                    func()
		expectedGDPRCountryCodes gdprCountryCodes
	}{
		{
			name: "query failed",
			fields: fields{
				cache:            mockCache,
				gdprCountryCodes: newGDPRCountryCodes(),
			},
			setup: func() {
				mockCache.EXPECT().GetGDPRCountryCodes().Return(nil, errors.New("QUERY FAILED"))
			},
			expectedGDPRCountryCodes: gdprCountryCodes{
				codes: [2]models.HashSet{
					make(models.HashSet),
					make(models.HashSet),
				},
				index: 0,
			},
		},
		{
			name: "query success",
			fields: fields{
				cache:            mockCache,
				gdprCountryCodes: newGDPRCountryCodes(),
			},
			setup: func() {
				mockCache.EXPECT().GetGDPRCountryCodes().Return(models.HashSet{
					"US": {},
					"DE": {},
				}, nil)
			},
			expectedGDPRCountryCodes: gdprCountryCodes{
				codes: [2]models.HashSet{
					{},
					{
						"US": {},
						"DE": {},
					},
				},
				index: 1,
			},
		},
		{
			name: "query success toggled",
			fields: fields{
				cache: mockCache,
				gdprCountryCodes: gdprCountryCodes{
					codes: [2]models.HashSet{
						{},
						{
							"US": {},
							"DE": {},
						},
					},
					index: 1,
				},
			},
			setup: func() {
				mockCache.EXPECT().GetGDPRCountryCodes().Return(models.HashSet{
					"US": {},
					"DE": {},
				}, nil)
			},
			expectedGDPRCountryCodes: gdprCountryCodes{
				codes: [2]models.HashSet{
					{
						"US": {},
						"DE": {},
					},
					{
						"US": {},
						"DE": {},
					},
				},
				index: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			fe := &feature{
				cache:            tt.fields.cache,
				gdprCountryCodes: tt.fields.gdprCountryCodes,
			}
			defer func() {
				fe = nil
			}()
			fe.updateGDPRCountryCodes()
			assert.Equal(t, tt.expectedGDPRCountryCodes, fe.gdprCountryCodes, tt.name)
		})
	}
}

func TestFeature_IsCountryGDPREnabled(t *testing.T) {
	type fields struct {
		gdprCountryCodes gdprCountryCodes
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
			name: "gdpr enabled for countrycode",
			args: args{
				countryCode: "LV",
			},
			fields: fields{
				gdprCountryCodes: gdprCountryCodes{
					codes: [2]models.HashSet{
						{},
						{
							"LV": {},
							"DE": {},
						},
					},
					index: 1,
				},
			},
			want: true,
		},
		{
			name: "gdpr disabled for countrycode",
			args: args{
				countryCode: "IN",
			},
			fields: fields{
				gdprCountryCodes: gdprCountryCodes{
					codes: [2]models.HashSet{
						{},
						{
							"LV": {},
							"DE": {},
						},
					},
					index: 1,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				gdprCountryCodes: tt.fields.gdprCountryCodes,
			}
			got := fe.IsCountryGDPREnabled(tt.args.countryCode)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
