package fullscreenclickability

import (
	"testing"
	"time"

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

	tests := []struct {
		name  string
		args  args
		setup func()
	}{
		{
			name: "test InitateReloader with valid cache and invalid time, exit",
			args: args{defaultExpiry: 0,
				cache: mockCache,
			},
			setup: func() {},
		},
		{
			name: "test InitateReloader with valid cache and time, call once and exit",
			args: args{defaultExpiry: 1000,
				cache: mockCache,
			},
			setup: func() {
				mockCache.EXPECT().GetFSCDisabledPublishers().Return(map[int]struct{}{}, nil)
				mockCache.EXPECT().GetFSCThresholdPerDSP().Return(map[int]int{}, nil)
			},
		},
	}
	for _, tt := range tests {
		tt.setup()
		fscConfigs.serviceStop = make(chan struct{})
		go initiateReloader(tt.args.cache, tt.args.defaultExpiry)
		//closing channel to avoid infinite loop
		StopReloaderService()
		time.Sleep(1 * time.Millisecond)
	}
}
