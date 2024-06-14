package profilemetadata

import (
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/magiconair/properties/assert"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache"
	mock_cache "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache/mock"
)

func TestNew(t *testing.T) {
	type args struct {
		config Config
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test",
			args: args{
				config: Config{
					Cache:         nil,
					DefaultExpiry: 0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.config)
			assert.Equal(t, pmd, got)
		})
	}
}

func Test_profileMetaData_Start(t *testing.T) {
	oldInitReloader := initReloader
	defer func() {
		initReloader = oldInitReloader
	}()
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "test",
			setup: func() {
				initReloader = func(pmd *profileMetaData) {}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			pmd := &profileMetaData{
				serviceStop: make(chan struct{}),
			}
			pmd.Start()
			pmd.Stop()
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
				mockCache.EXPECT().GetAppIntegrationPath().Return(map[string]int{}, nil)
				mockCache.EXPECT().GetAppSubIntegrationPath().Return(map[string]int{}, nil)
				mockCache.EXPECT().GetProfileTypePlatform().Return(map[string]int{}, nil)
			},
		},
	}
	for _, tt := range tests {
		tt.setup()
		profileMetaData := &profileMetaData{
			cache:         tt.args.cache,
			defaultExpiry: tt.args.defaultExpiry,
			serviceStop:   make(chan struct{}),
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			initReloader(profileMetaData)
			wg.Done()
		}()
		//closing channel to avoid infinite loop
		profileMetaData.Stop()
		wg.Wait() // wait for initReloader to finish
	}
}

func Test_profileMetaData_updateProfileMetadaMaps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mock_cache.NewMockCache(ctrl)

	type fields struct {
		cache                 cache.Cache
		serviceStop           chan struct{}
		RWMutex               sync.RWMutex
		defaultExpiry         int
		profileTypePlatform   map[string]int
		appIntegrationPath    map[string]int
		appSubIntegrationPath map[string]int
	}
	type want struct {
		profileTypePlatform   map[string]int
		appIntegrationPath    map[string]int
		appSubIntegrationPath map[string]int
	}
	tests := []struct {
		name   string
		fields fields
		setup  func()
		want   want
	}{
		{
			name: "all profile metadata updated from cache",
			fields: fields{
				cache:                 mockCache,
				profileTypePlatform:   map[string]int{},
				appIntegrationPath:    map[string]int{},
				appSubIntegrationPath: map[string]int{},
			},
			setup: func() {
				mockCache.EXPECT().GetProfileTypePlatform().Return(map[string]int{
					"openwrap": 1,
					"identity": 2,
				}, nil)
				mockCache.EXPECT().GetAppIntegrationPath().Return(map[string]int{
					"iOS":     1,
					"Android": 2,
				}, nil)
				mockCache.EXPECT().GetAppSubIntegrationPath().Return(map[string]int{
					"DFP":   1,
					"MoPub": 3,
				}, nil)
			},
			want: want{
				profileTypePlatform: map[string]int{
					"openwrap": 1,
					"identity": 2,
				},
				appIntegrationPath: map[string]int{
					"iOS":     1,
					"Android": 2,
				},
				appSubIntegrationPath: map[string]int{
					"DFP":   1,
					"MoPub": 3,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			pmd := &profileMetaData{
				cache:                 tt.fields.cache,
				serviceStop:           tt.fields.serviceStop,
				RWMutex:               tt.fields.RWMutex,
				defaultExpiry:         tt.fields.defaultExpiry,
				profileTypePlatform:   tt.fields.profileTypePlatform,
				appIntegrationPath:    tt.fields.appIntegrationPath,
				appSubIntegrationPath: tt.fields.appSubIntegrationPath,
			}
			pmd.updateProfileMetadaMaps()
			assert.Equal(t, tt.want.profileTypePlatform, pmd.profileTypePlatform)
			assert.Equal(t, tt.want.appIntegrationPath, pmd.appIntegrationPath)
			assert.Equal(t, tt.want.appSubIntegrationPath, pmd.appSubIntegrationPath)
		})
	}
}