package tbf

import (
	"fmt"
	"testing"
	"time"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache"
	mock_cache "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestInitAndReloader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)
	defer SetAndResetTBFConfig(mockCache, nil)()

	type args struct {
		defaultExpiry int
		cache         cache.Cache
	}

	tests := []struct {
		name      string
		args      args
		runBefore func()
	}{
		{
			name: "test_cache_call_through_init",
			args: args{
				defaultExpiry: 1,
				cache:         mockCache,
			},
			runBefore: func() {
				mockCache.EXPECT().GetTBFTrafficForPublishers().Return(map[int]map[int]int{}, nil).AnyTimes()
			},
		},
	}
	for _, tt := range tests {
		tt.runBefore()
		tbfConfigs.serviceStop = make(chan struct{})
		Init(tt.args.defaultExpiry, tt.args.cache)
		time.Sleep(2 * time.Second)
		StopTBFReloaderService()
		time.Sleep(250 * time.Millisecond)
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

	type args struct {
		pubidstr       int
		profid         int
		pubProfTraffic map[int]map[int]int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil_map",
			args: args{
				pubidstr:       5890,
				profid:         1234,
				pubProfTraffic: nil,
			},
			want: false,
		},
		{
			name: "pub_prof_absent_in_map",
			args: args{
				pubidstr:       5890,
				profid:         1234,
				pubProfTraffic: make(map[int]map[int]int),
			},
			want: false,
		},
		{
			name: "pub_prof_present_in_map",
			args: args{
				pubidstr: 5890,
				profid:   1234,
				pubProfTraffic: map[int]map[int]int{
					5890: {1234: 100},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetAndResetTBFConfig(nil, tt.args.pubProfTraffic)
			got := IsEnabledTBFFeature(tt.args.pubidstr, tt.args.profid)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestUpdateTBFConfigMapsFromCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	defer SetAndResetTBFConfig(mockCache, map[int]map[int]int{})()
	type want struct {
		err               error
		pubProfileTraffic map[int]map[int]int
	}

	tests := []struct {
		name  string
		setup func()
		want  want
	}{
		{
			name: "cache_returns_error",
			setup: func() {
				mockCache.EXPECT().GetTBFTrafficForPublishers().Return(nil, fmt.Errorf("error"))
			},
			want: want{
				pubProfileTraffic: map[int]map[int]int{},
				err:               fmt.Errorf("error"),
			},
		},
		{
			name: "cache_returns_success",
			setup: func() {
				mockCache.EXPECT().GetTBFTrafficForPublishers().Return(map[int]map[int]int{5890: {1234: 100}}, nil)
			},
			want: want{
				pubProfileTraffic: map[int]map[int]int{5890: {1234: 100}},
				err:               nil,
			},
		},
		{
			name: "limit_traffic_values",
			setup: func() {
				mockCache.EXPECT().GetTBFTrafficForPublishers().Return(map[int]map[int]int{5890: {1234: 200}, 5891: {222: -5}}, nil)
			},
			want: want{
				pubProfileTraffic: map[int]map[int]int{5890: {1234: 0}, 5891: {222: 0}},
				err:               nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := updateTBFConfigMapsFromCache()
			assert.Equal(t, tt.want.err, err, tt.name)
			assert.Equal(t, tt.want.pubProfileTraffic, tbfConfigs.pubProfileTraffic, tt.name)
		})
	}
}

func TestLimitTBFTrafficValues(t *testing.T) {

	tests := []struct {
		name      string
		inputMap  map[int]map[int]int
		outputMap map[int]map[int]int
	}{
		{
			name:      "nil_map",
			inputMap:  nil,
			outputMap: nil,
		},
		{
			name: "nil_prof_traffic_map",
			inputMap: map[int]map[int]int{
				1: nil,
			},
			outputMap: map[int]map[int]int{
				1: nil,
			},
		},
		{
			name: "negative_and_higher_than_100_values",
			inputMap: map[int]map[int]int{
				5890: {123: -100},
				5891: {123: 50},
				5892: {123: 200},
			},
			outputMap: map[int]map[int]int{
				5890: {123: 0},
				5891: {123: 50},
				5892: {123: 0},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limitTBFTrafficValues(tt.inputMap)
			assert.Equal(t, tt.outputMap, tt.inputMap, tt.name)
		})
	}
}
