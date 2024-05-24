package publisherfeature

import (
	"errors"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache"
	mock_cache "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		args Config
	}{
		{
			name: "test",
			args: Config{
				Cache:                 nil,
				DefaultExpiry:         0,
				AnalyticsThrottleList: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args)
			assert.Equal(t, fe, got)
		})
	}
}

func TestInitiateReloader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	type args struct {
		defaultExpiry int
		cache         cache.Cache
	}

	tests := []struct {
		name  string
		args  args
		setup func()
	}{
		{
			name: "test InitateReloader with valid cache and invalid time, exit",
			args: args{
				defaultExpiry: 0,
				cache:         mockCache,
			},
			setup: func() {},
		},
		{
			name: "test InitateReloader with valid cache and time, call once and exit",
			args: args{
				defaultExpiry: 1000,
				cache:         mockCache,
			},
			setup: func() {
				mockCache.EXPECT().GetPublisherFeatureMap().Return(map[int]map[int]models.FeatureData{}, nil)
				mockCache.EXPECT().GetFSCThresholdPerDSP().Return(map[int]int{}, nil)
			},
		},
	}
	for _, tt := range tests {
		tt.setup()
		feature := &feature{
			cache:         tt.args.cache,
			defaultExpiry: tt.args.defaultExpiry,
			serviceStop:   make(chan struct{}),
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			initReloader(feature)
			wg.Done()
		}()
		//closing channel to avoid infinite loop
		feature.Stop()
		wg.Wait() // wait for initReloader to finish
	}
}

func Test_feature_Start(t *testing.T) {

	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "test",

			setup: func() {
				initReloader = func(fe *feature) {}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			fe := &feature{}
			fe.Start()
		})
	}
}

func Test_feature_updateFeatureConfigMaps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	type fields struct {
		cache cache.Cache
	}
	type want struct {
		fsc            fsc
		tbf            tbf
		ampMultiformat ampMultiformat
	}
	tests := []struct {
		name   string
		fields fields
		setup  func()
		want   want
	}{
		{
			name: "publisher feature map query failed and fsc threshold per DSP query success",
			fields: fields{
				cache: mockCache,
			},
			setup: func() {
				mockCache.EXPECT().GetPublisherFeatureMap().Return(nil, errors.New("QUERY FAILED"))
				mockCache.EXPECT().GetFSCThresholdPerDSP().Return(map[int]int{
					6: 100,
				}, nil)
			},
			want: want{
				fsc: fsc{
					disabledPublishers: map[int]struct{}{},
					thresholdsPerDsp: map[int]int{
						6: 100,
					},
				},
				ampMultiformat: ampMultiformat{
					enabledPublishers: map[int]struct{}{},
				},
				tbf: tbf{
					pubProfileTraffic: map[int]map[int]int{},
				},
			},
		},
		{
			name: "publisher feature map query success but fsc threshold per DSP query failed",
			fields: fields{
				cache: mockCache,
			},
			setup: func() {
				mockCache.EXPECT().GetPublisherFeatureMap().Return(map[int]map[int]models.FeatureData{
					5890: {
						models.FeatureFSC: {
							Enabled: 0,
						},
						models.FeatureTBF: {
							Enabled: 1,
							Value:   `{"1234": 100}`,
						},
						models.FeatureAMPMultiFormat: {
							Enabled: 1,
						},
					},
				}, nil)
				mockCache.EXPECT().GetFSCThresholdPerDSP().Return(nil, errors.New("QUERY FAILED"))
			},
			want: want{
				fsc: fsc{
					disabledPublishers: map[int]struct{}{},
					thresholdsPerDsp:   map[int]int{},
				},
				ampMultiformat: ampMultiformat{
					enabledPublishers: map[int]struct{}{
						5890: {},
					},
				},
				tbf: tbf{
					pubProfileTraffic: map[int]map[int]int{
						5890: {
							1234: 100,
						},
					},
				},
			},
		},
		{
			name: "both queries success and all feature deatils updated successfully",
			fields: fields{
				cache: mockCache,
			},
			setup: func() {
				mockCache.EXPECT().GetPublisherFeatureMap().Return(map[int]map[int]models.FeatureData{
					5890: {
						models.FeatureFSC: {
							Enabled: 0,
						},
						models.FeatureTBF: {
							Enabled: 1,
							Value:   `{"1234": 100}`,
						},
						models.FeatureAMPMultiFormat: {
							Enabled: 1,
						},
					},
				}, nil)
				mockCache.EXPECT().GetFSCThresholdPerDSP().Return(map[int]int{6: 100}, nil)
			},
			want: want{
				fsc: fsc{
					disabledPublishers: map[int]struct{}{
						5890: {},
					},
					thresholdsPerDsp: map[int]int{
						6: 100,
					},
				},
				ampMultiformat: ampMultiformat{
					enabledPublishers: map[int]struct{}{
						5890: {},
					},
				},
				tbf: tbf{
					pubProfileTraffic: map[int]map[int]int{
						5890: {
							1234: 100,
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
				fsc: fsc{
					disabledPublishers: make(map[int]struct{}),
					thresholdsPerDsp:   make(map[int]int),
				},
				tbf: tbf{
					pubProfileTraffic: make(map[int]map[int]int),
				},
				ampMultiformat: ampMultiformat{
					enabledPublishers: make(map[int]struct{}),
				},
			}
			fe.updateFeatureConfigMaps()
			assert.Equal(t, tt.want.fsc, fe.fsc, tt.name)
			assert.Equal(t, tt.want.tbf, fe.tbf, tt.name)
			assert.Equal(t, tt.want.ampMultiformat, fe.ampMultiformat, tt.name)
		})
	}
}
