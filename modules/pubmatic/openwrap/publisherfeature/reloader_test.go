package publisherfeature

import (
	"errors"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache"
	mock_cache "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
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
				mockCache.EXPECT().GetGDPRCountryCodes().Return(map[string]struct{}{}, nil)
				mockCache.EXPECT().GetProfileAdUnitMultiFloors().Return(models.ProfileAdUnitMultiFloors{}, nil)
				mockCache.EXPECT().GetInViewEnabledPublishers().Return(map[int]struct{}{}, nil)
				mockCache.EXPECT().GetPerformanceDSPs().Return(map[int]struct{}{}, nil)
			},
		},
	}
	for _, tt := range tests {
		tt.setup()
		feature := &feature{
			cache:         tt.args.cache,
			defaultExpiry: tt.args.defaultExpiry,
			serviceStop:   make(chan struct{}),
			mbmf:          newMBMF(),
			dynamicFloor:  newDynamicFloor(),
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
			fe := &feature{
				serviceStop: make(chan struct{}),
			}
			fe.Start()
			fe.Stop()
		})
	}
}

func TestFeatureUpdateFeatureConfigMaps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	type fields struct {
		cache cache.Cache
	}
	type want struct {
		fsc                  fsc
		tbf                  tbf
		ampMultiformat       ampMultiformat
		bidRecovery          bidRecovery
		appLovinMultiFloors  appLovinMultiFloors
		impCountingMethod    impCountingMethod
		appLovinSchainABTest appLovinSchainABTest
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
				mockCache.EXPECT().GetProfileAdUnitMultiFloors().Return(models.ProfileAdUnitMultiFloors{}, nil)
				mockCache.EXPECT().GetInViewEnabledPublishers().Return(map[int]struct{}{}, nil)
				mockCache.EXPECT().GetPerformanceDSPs().Return(map[int]struct{}{}, nil)
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
				impCountingMethod: impCountingMethod{
					enabledBidders: [2]map[string]struct{}{
						{},
						{},
					},
					index: 0,
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
				mockCache.EXPECT().GetProfileAdUnitMultiFloors().Return(models.ProfileAdUnitMultiFloors{}, nil)
				mockCache.EXPECT().GetInViewEnabledPublishers().Return(map[int]struct{}{}, nil)
				mockCache.EXPECT().GetPerformanceDSPs().Return(map[int]struct{}{}, nil)
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
				appLovinMultiFloors: appLovinMultiFloors{
					enabledPublisherProfile: map[int]map[string]models.ApplovinAdUnitFloors{},
				},
				bidRecovery: bidRecovery{
					enabledPublisherProfile: map[int]map[int]struct{}{},
				},
				impCountingMethod: impCountingMethod{
					enabledBidders: [2]map[string]struct{}{
						{},
						{},
					},
					index: 1,
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
				mockCache.EXPECT().GetProfileAdUnitMultiFloors().Return(models.ProfileAdUnitMultiFloors{}, nil)
				mockCache.EXPECT().GetInViewEnabledPublishers().Return(map[int]struct{}{}, nil)
				mockCache.EXPECT().GetPerformanceDSPs().Return(map[int]struct{}{}, nil)
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
				bidRecovery: bidRecovery{
					enabledPublisherProfile: map[int]map[int]struct{}{},
				},
				appLovinMultiFloors: appLovinMultiFloors{
					enabledPublisherProfile: map[int]map[string]models.ApplovinAdUnitFloors{},
				},
				impCountingMethod: impCountingMethod{
					enabledBidders: [2]map[string]struct{}{
						{},
						{},
					},
					index: 1,
				},
			},
		},
		{
			name: "fetch applovin_abtest,bidrecovery feature data",
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
						models.FeatureBidRecovery: {
							Enabled: 1,
							Value:   `[1234,3212]`,
						},
						models.FeatureApplovinMultiFloors: {
							Enabled: 1,
							Value:   `{"1232":{"adunit_123":[4.2,5.6,5.8],"adunit_dmdemo":[4.2,5.6,5.8]},"4322":{"adunit_12323":[4.2,5.6,5.8],"adunit_dmdemo1":[4.2,5.6,5.8]}}`,
						},
					},
				}, nil)
				mockCache.EXPECT().GetFSCThresholdPerDSP().Return(map[int]int{6: 100}, nil)
				mockCache.EXPECT().GetProfileAdUnitMultiFloors().Return(models.ProfileAdUnitMultiFloors{}, nil)
				mockCache.EXPECT().GetInViewEnabledPublishers().Return(map[int]struct{}{}, nil)
				mockCache.EXPECT().GetPerformanceDSPs().Return(map[int]struct{}{}, nil)
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
				bidRecovery: bidRecovery{
					enabledPublisherProfile: map[int]map[int]struct{}{
						5890: {
							1234: {},
							3212: {},
						},
					},
				},
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
				impCountingMethod: impCountingMethod{
					enabledBidders: [2]map[string]struct{}{
						{},
						{},
					},
					index: 1,
				},
			},
		},
		{
			name: "fetch impcountingmethod feature data with multiple bidders",
			fields: fields{
				cache: mockCache,
			},
			setup: func() {
				mockCache.EXPECT().GetPublisherFeatureMap().Return(map[int]map[int]models.FeatureData{
					0: {
						models.FeatureImpCountingMethod: {
							Enabled: 1,
							Value:   "appnexus, rubicon",
						},
					},
				}, nil)
				mockCache.EXPECT().GetFSCThresholdPerDSP().Return(map[int]int{6: 100}, nil)
				mockCache.EXPECT().GetInViewEnabledPublishers().Return(map[int]struct{}{}, nil)
				mockCache.EXPECT().GetPerformanceDSPs().Return(map[int]struct{}{}, nil)
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
				bidRecovery: bidRecovery{
					enabledPublisherProfile: map[int]map[int]struct{}{},
				},
				appLovinMultiFloors: appLovinMultiFloors{
					enabledPublisherProfile: map[int]map[string]models.ApplovinAdUnitFloors{},
				},
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
		},
		{
			name: "fetch applovin_schain_abtest feature data",
			fields: fields{
				cache: mockCache,
			},
			setup: func() {
				mockCache.EXPECT().GetPublisherFeatureMap().Return(map[int]map[int]models.FeatureData{
					0: {
						models.FeatureAppLovinSchainABTest: {
							Enabled: 1,
							Value:   "10",
						},
					},
				}, nil)
				mockCache.EXPECT().GetFSCThresholdPerDSP().Return(map[int]int{
					6: 100,
				}, nil)
				mockCache.EXPECT().GetProfileAdUnitMultiFloors().Return(models.ProfileAdUnitMultiFloors{}, nil)
				mockCache.EXPECT().GetInViewEnabledPublishers().Return(map[int]struct{}{}, nil)
				mockCache.EXPECT().GetPerformanceDSPs().Return(map[int]struct{}{}, nil)
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
				bidRecovery: bidRecovery{
					enabledPublisherProfile: map[int]map[int]struct{}{},
				},
				appLovinMultiFloors: appLovinMultiFloors{
					enabledPublisherProfile: map[int]map[string]models.ApplovinAdUnitFloors{},
				},
				impCountingMethod: impCountingMethod{
					enabledBidders: [2]map[string]struct{}{
						{},
						{},
					},
					index: 1,
				},
				appLovinSchainABTest: appLovinSchainABTest{
					schainABTestPercent: 10,
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
				impCountingMethod: newImpCountingMethod(),
				mbmf:              newMBMF(),
			}
			defer func() {
				fe = nil
			}()
			fe.updateFeatureConfigMaps()
			assert.Equal(t, tt.want.fsc, fe.fsc, tt.name)
			assert.Equal(t, tt.want.tbf, fe.tbf, tt.name)
			assert.Equal(t, tt.want.ampMultiformat, fe.ampMultiformat, tt.name)
			assert.Equal(t, tt.want.bidRecovery, fe.bidRecovery, tt.name)
			assert.Equal(t, tt.want.appLovinMultiFloors, fe.appLovinMultiFloors, tt.name)
			assert.Equal(t, tt.want.impCountingMethod, fe.impCountingMethod, tt.name)
			assert.Equal(t, tt.want.appLovinSchainABTest, fe.appLovinSchainABTest, tt.name)
		})
	}
}
