package profilemetadata

import (
	"fmt"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache"
	mock_cache "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache/mock"
	"github.com/stretchr/testify/assert"
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
					Cache:                 nil,
					ProfileMetaDataExpiry: 0,
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
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "successfull start data loaded from db",
			setup: func() {
				initReloader = func(pmd *profileMetaData) {
					pmd.failToLoadDBData <- false
				}
			},
			wantErr: false,
		},
		{
			name: "failed to load data from db do not start service",
			setup: func() {
				initReloader = func(pmd *profileMetaData) {
					pmd.failToLoadDBData <- true
				}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			pmd := &profileMetaData{
				serviceStop:      make(chan struct{}),
				failToLoadDBData: make(chan bool),
			}
			err := pmd.Start()
			assert.Equal(t, tt.wantErr, err != nil)
			pmd.Stop()
		})
	}
}

func TestInitiateReloader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	type args struct {
		profileMetaDataExpiry int
		cache                 cache.Cache
	}

	tests := []struct {
		name  string
		args  args
		want  bool
		setup func()
	}{
		{
			name: "test InitateReloader with valid cache and invalid time, exit",
			args: args{
				profileMetaDataExpiry: 0,
				cache:                 mockCache,
			},
			setup: nil,
		},
		{
			name: "test InitateReloader with valid cache and time, call once and exit",
			args: args{
				profileMetaDataExpiry: 1000,
				cache:                 mockCache,
			},
			setup: func() {
				mockCache.EXPECT().GetAppIntegrationPaths().Return(map[string]int{}, nil)
				mockCache.EXPECT().GetAppSubIntegrationPaths().Return(map[string]int{}, nil)
				mockCache.EXPECT().GetProfileTypePlatforms().Return(map[string]int{}, nil)
			},
			want: false,
		},
		{
			name: "test InitateReloader with valid cache and time, failed to load data from cache",
			args: args{
				profileMetaDataExpiry: 1000,
				cache:                 mockCache,
			},
			setup: func() {
				mockCache.EXPECT().GetAppIntegrationPaths().Return(nil, fmt.Errorf("error"))
				mockCache.EXPECT().GetAppSubIntegrationPaths().Return(map[string]int{}, nil)
				mockCache.EXPECT().GetProfileTypePlatforms().Return(map[string]int{}, nil)
			},
			want: true,
		},
	}
	for _, tt := range tests {
		if tt.setup != nil {
			tt.setup()
		}
		profileMetaData := &profileMetaData{
			cache:                 tt.args.cache,
			profileMetaDataExpiry: tt.args.profileMetaDataExpiry,
			failToLoadDBData:      make(chan bool),
			serviceStop:           make(chan struct{}),
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			initReloader(profileMetaData)
			wg.Done()
		}()
		if tt.setup != nil {
			got := <-profileMetaData.failToLoadDBData
			assert.Equal(t, tt.want, got)
		}
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
		profileMetaDataExpiry int
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
		name    string
		fields  fields
		setup   func()
		want    want
		wantErr bool
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
				mockCache.EXPECT().GetProfileTypePlatforms().Return(map[string]int{
					"display": 1,
					"in-app":  2,
				}, nil)
				mockCache.EXPECT().GetAppIntegrationPaths().Return(map[string]int{
					"iOS":     1,
					"Android": 2,
				}, nil)
				mockCache.EXPECT().GetAppSubIntegrationPaths().Return(map[string]int{
					"DFP":   1,
					"MoPub": 3,
				}, nil)
			},
			want: want{
				profileTypePlatform: map[string]int{
					"display": 1,
					"in-app":  2,
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
			wantErr: false,
		},
		{
			name: "profileTypePlatform, appIntegrationPath and appSubIntegrationPath not updated from cache",
			fields: fields{
				cache:                 mockCache,
				profileTypePlatform:   map[string]int{},
				appIntegrationPath:    map[string]int{},
				appSubIntegrationPath: map[string]int{},
			},
			setup: func() {
				mockCache.EXPECT().GetProfileTypePlatforms().Return(nil, fmt.Errorf("error"))
				mockCache.EXPECT().GetAppIntegrationPaths().Return(nil, fmt.Errorf("error"))
				mockCache.EXPECT().GetAppSubIntegrationPaths().Return(nil, fmt.Errorf("error"))

			},
			want: want{
				profileTypePlatform:   map[string]int{},
				appIntegrationPath:    map[string]int{},
				appSubIntegrationPath: map[string]int{},
			},
			wantErr: true,
		},
	}
	for ind := range tests {
		tt := &tests[ind]
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			pmd := &profileMetaData{
				cache:                 tt.fields.cache,
				serviceStop:           tt.fields.serviceStop,
				profileMetaDataExpiry: tt.fields.profileMetaDataExpiry,
				profileTypePlatform:   tt.fields.profileTypePlatform,
				appIntegrationPath:    tt.fields.appIntegrationPath,
				appSubIntegrationPath: tt.fields.appSubIntegrationPath,
			}
			err := pmd.updateProfileMetaDataMaps()
			assert.Equal(t, tt.want.profileTypePlatform, pmd.profileTypePlatform)
			assert.Equal(t, tt.want.appIntegrationPath, pmd.appIntegrationPath)
			assert.Equal(t, tt.want.appSubIntegrationPath, pmd.appSubIntegrationPath)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
