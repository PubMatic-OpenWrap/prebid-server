package publisherfeature

// func TestInitiateReloader(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	mockCache := mock_dbcache.NewMockCache(ctrl)
// 	defer SetAndResetFscWithMockCache(mockCache, nil)()
// 	currentChan := reloaderConfig.serviceStop
// 	defer func() {
// 		ctrl.Finish()
// 		reloaderConfig.serviceStop = currentChan
// 	}()
// 	reloaderConfig.serviceStop = make(chan struct{})
// 	type args struct {
// 		defaultExpiry int
// 		cache         cache.Cache
// 	}

// 	tests := []struct {
// 		name  string
// 		args  args
// 		setup func()
// 	}{
// 		{
// 			name: "test InitateReloader with valid cache and invalid time, exit",
// 			args: args{defaultExpiry: 0,
// 				cache: mockCache,
// 			},
// 			setup: func() {},
// 		},
// 		{
// 			name: "test InitateReloader with valid cache and time, call once and exit",
// 			args: args{defaultExpiry: 1000,
// 				cache: mockCache,
// 			},
// 			setup: func() {
// 				mockCache.EXPECT().GetPublisherFeatureMap().Return(map[int]int{}, nil)
// 				mockCache.EXPECT().GetFSCThresholdPerDSP().Return(map[int]int{}, nil)
// 				mockCache.EXPECT().GetTBFTrafficForPublishers().Return(map[int]map[int]int{}, nil)
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		tt.setup()
// 		reloaderConfig.serviceStop = make(chan struct{})
// 		go initiateReloader(tt.args.cache, tt.args.defaultExpiry)
// 		//closing channel to avoid infinite loop
// 		StopReloaderService()
// 		time.Sleep(1 * time.Millisecond)
// 	}
// }

// func TestInit(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()
// 	mockCache := mock_dbcache.NewMockCache(ctrl)
// 	var defCpy = initiateReloader
// 	initiateReloader = func(c cache.Cache, expiryTime int) {}
// 	defer func() {
// 		initiateReloader = defCpy
// 	}()
// 	type args struct {
// 		defaultExpiry int
// 		cache         cache.Cache
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 	}{
// 		{name: "test Init with valid args",
// 			args: args{defaultExpiry: 1,
// 				cache: mockCache,
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		Init(tt.args.cache, tt.args.defaultExpiry)
// 	}

// }
