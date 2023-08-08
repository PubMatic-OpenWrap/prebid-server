package fullscreenclickability

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	mock_dbcache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache/mock"
)

func TestInitiateReloader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_dbcache.NewMockCache(ctrl)
	defer SetAndResetFscWithMockCache(mockCache, nil)()
	type args struct {
		defaultExpiry int
		cache         cache.Cache
	}

	tests := []struct {
		name      string
		args      args
		runBefore func()
	}{
		{name: "test InitateReloader with valid cache and time",
			args: args{defaultExpiry: 100,
				cache: mockCache,
			},
			runBefore: func() {
				mockCache.EXPECT().GetFSCDisabledPublishers().Return(map[int]struct{}{}, nil)
				mockCache.EXPECT().GetFSCThresholdPerDSP().Return(map[int]int{}, nil)
			},
		},
	}
	for _, tt := range tests {
		tt.runBefore()
		fscConfigs.serviceStop = make(chan bool)
		go initiateReloader(tt.args.cache, tt.args.defaultExpiry)
		//stopService Test
		StopFscReloaderService()
		time.Sleep(250 * time.Millisecond)
	}

}
