package fullscreenclickability

import (
	"testing"

	"github.com/golang/mock/gomock"
	cache "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache"
	mock_dbcache "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache/mock"
)

func TestInitiateReloader(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCache := mock_dbcache.NewMockCache(ctrl)
	defer SetAndResetFscWithMockCache(mockCache, nil)()
	currentChan := fscConfigs.serviceStop
	defer func() {
		ctrl.Finish()
		fscConfigs.serviceStop = currentChan
	}()
	fscConfigs.serviceStop = make(chan struct{})
	type args struct {
		defaultExpiry int
		cache         cache.Cache
	}

	waitChan := make(chan struct{}, 1)
	tests := []struct {
		name  string
		args  args
		setup func(chan struct{})
	}{
		{
			name: "test InitateReloader with valid cache and invalid time, exit",
			args: args{defaultExpiry: 0,
				cache: mockCache,
			},
			setup: func(w chan struct{}) {
				w <- struct{}{}
			},
		},
		{
			name: "test InitateReloader with valid cache and time, call once and exit",
			args: args{defaultExpiry: 1000,
				cache: mockCache,
			},
			setup: func(w chan struct{}) {
				mockCache.EXPECT().GetFSCDisabledPublishers().Return(map[int]struct{}{}, nil)
				mockCache.EXPECT().GetFSCThresholdPerDSP().Do(func() {
					w <- struct{}{}
				}).Return(map[int]int{}, nil)
			},
		},
	}
	for _, tt := range tests {
		tt.setup(waitChan)
		fscConfigs.serviceStop = make(chan struct{})
		go initiateReloader(tt.args.cache, tt.args.defaultExpiry)
		//wait
		<-waitChan
		//closing channel to avoid infinite loop
		StopReloaderService()
	}
}
