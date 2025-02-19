package publisherfeature

import (
	"testing"

	"github.com/golang/mock/gomock"
	mock_cache "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestUpdateTBFConfigMapsFromCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	type feilds struct {
		publisherFeature map[int]map[int]models.FeatureData
		tbf              tbf
	}
	type want struct {
		err               error
		pubProfileTraffic map[int]map[int]int
	}

	tests := []struct {
		name   string
		want   want
		feilds feilds
	}{
		{
			name: "publisherFeature_map_is_nil",
			feilds: feilds{
				publisherFeature: nil,
			},
			want: want{
				err:               nil,
				pubProfileTraffic: nil,
			},
		},
		{
			name: "successfully_update_tbf_config_map",
			feilds: feilds{
				publisherFeature: map[int]map[int]models.FeatureData{
					5890: {
						2: {
							Enabled: 1,
							Value:   `{"1234": 100}`,
						},
					},
				},
				tbf: tbf{
					pubProfileTraffic: make(map[int]map[int]int),
				},
			},
			want: want{
				pubProfileTraffic: map[int]map[int]int{5890: {1234: 100}},
				err:               nil,
			},
		},
		{
			name: "failed_to_unmarshal_profile_traffic_rate",
			feilds: feilds{
				publisherFeature: map[int]map[int]models.FeatureData{
					5890: {
						2: {
							Enabled: 1,
							Value:   `"1234": 100}`,
						},
					},
				},
				tbf: tbf{
					pubProfileTraffic: make(map[int]map[int]int),
				},
			},
			want: want{
				pubProfileTraffic: map[int]map[int]int{},
				err:               nil,
			},
		},
		{
			name: "empty_profile_traffic_rate",
			feilds: feilds{
				publisherFeature: map[int]map[int]models.FeatureData{
					5890: {
						2: {
							Enabled: 1,
							Value:   "",
						},
					},
				},
				tbf: tbf{
					pubProfileTraffic: make(map[int]map[int]int),
				},
			},
			want: want{
				pubProfileTraffic: map[int]map[int]int{},
				err:               nil,
			},
		},
		{
			name: "limit_traffic_values",
			feilds: feilds{
				publisherFeature: map[int]map[int]models.FeatureData{
					5890: {
						2: {
							Enabled: 1,
							Value:   `{"1234": 200}`,
						},
					},
					5891: {
						2: {
							Enabled: 1,
							Value:   `{"222": -5}`,
						},
					},
				},
				tbf: tbf{
					pubProfileTraffic: make(map[int]map[int]int),
				},
			},
			want: want{
				pubProfileTraffic: map[int]map[int]int{5890: {1234: 0}, 5891: {222: 0}},
				err:               nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := feature{
				cache:            mockCache,
				publisherFeature: tt.feilds.publisherFeature,
				tbf:              tt.feilds.tbf,
			}
			fe.updateTBFConfigMap()
			assert.Equal(t, tt.want.pubProfileTraffic, fe.tbf.pubProfileTraffic, tt.name)
		})
	}
}

func TestPredictTBFValue(t *testing.T) {
	type args struct {
		percentage int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "100_pct_traffic",
			args: args{
				percentage: 100,
			},
			want: true,
		},
		{
			name: "0_pct_traffic",
			args: args{
				percentage: 0,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := predictTBFValue(tt.args.percentage)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestIsEnabledTBFFeature(t *testing.T) {
	type feilds struct {
		tbf tbf
	}
	type args struct {
		pubidstr int
		profid   int
	}
	tests := []struct {
		name   string
		args   args
		feilds feilds
		want   bool
	}{
		{
			name: "pubProfileTraffic_map_is_nil",
			args: args{
				pubidstr: 5890,
				profid:   1234,
			},
			feilds: feilds{
				tbf: tbf{
					pubProfileTraffic: nil,
				},
			},
			want: false,
		},
		{
			name: "pub_prof_absent_in_map",
			args: args{
				pubidstr: 5890,
				profid:   1234,
			},
			feilds: feilds{
				tbf: tbf{
					pubProfileTraffic: map[int]map[int]int{
						5891: {1234: 100},
					},
				},
			},
			want: false,
		},
		{
			name: "pub_prof_present_in_map",
			args: args{
				pubidstr: 5890,
				profid:   1234,
			},
			feilds: feilds{
				tbf: tbf{
					pubProfileTraffic: map[int]map[int]int{
						5890: {1234: 100},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				tbf: tt.feilds.tbf,
			}
			got := fe.IsTBFFeatureEnabled(tt.args.pubidstr, tt.args.profid)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestLimitTBFTrafficValues(t *testing.T) {

	tests := []struct {
		name               string
		pubProfTraffic     map[int]map[int]int
		wantpubProfTraffic map[int]map[int]int
	}{
		{
			name:               "pubProfTraffic_map_is_nil",
			pubProfTraffic:     nil,
			wantpubProfTraffic: nil,
		},
		{
			name: "nil_prof_traffic_map",
			pubProfTraffic: map[int]map[int]int{
				1: nil,
			},
			wantpubProfTraffic: map[int]map[int]int{
				1: nil,
			},
		},
		{
			name: "negative_and_higher_than_100_values",
			pubProfTraffic: map[int]map[int]int{
				5890: {123: -100},
				5891: {123: 50},
				5892: {123: 200},
			},
			wantpubProfTraffic: map[int]map[int]int{
				5890: {123: 0},
				5891: {123: 50},
				5892: {123: 0},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limitTBFTrafficValues(tt.pubProfTraffic)
			assert.Equal(t, tt.wantpubProfTraffic, tt.pubProfTraffic, tt.name)
		})
	}
}
