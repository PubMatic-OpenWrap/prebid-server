package fullscreenclickability

import (
	"errors"

	"reflect"
	"testing"
	"time"

	cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	mock_dbcache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache/mock"

	"github.com/golang/mock/gomock"
)

func TestPredictFscValue(t *testing.T) {
	type args struct {
		percentage int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "getting from predict output",
			args: args{
				percentage: 100,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := predictFscValue(tt.args.percentage); got != tt.want {
				t.Errorf("predictFscValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_dbcache.NewMockCache(ctrl)
	var defCpy = initiateReloader
	initiateReloader = func(c cache.Cache, expiryTime int) {}
	defer func() {
		initiateReloader = defCpy
	}()
	type args struct {
		defaultExpiry int
		cache         cache.Cache
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "test Init with valid args",
			args: args{defaultExpiry: 1,
				cache: mockCache,
			},
		},
	}
	for _, tt := range tests {
		Init(tt.args.cache, tt.args.defaultExpiry)
	}

}

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

func TestIsUnderFSCThreshold(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_dbcache.NewMockCache(ctrl)
	type fields struct {
		cache              cache.Cache
		disabledPublishers map[int]struct{}
		thresholdsPerDsp   map[int]int
	}
	type args struct {
		pubid int
		dspid int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
		setup  func()
	}{
		{
			name: "When pubId,dspid and FSC maps are valid, pubID enabled(default) FSC return fsc with prediction algo",
			args: args{
				pubid: 5890,
				dspid: 6,
			},
			setup: func() {

			},
			fields: fields{cache: mockCache,
				disabledPublishers: map[int]struct{}{
					58903: {},
				},
				thresholdsPerDsp: map[int]int{6: 100},
			},

			want: 1,
		},
		{
			name: "When pubId,dspid and FSC maps are valid, pubID disabled FSC return fsc=0",
			args: args{
				pubid: 5890,
				dspid: 6,
			},
			setup: func() {

			},
			fields: fields{cache: mockCache,
				disabledPublishers: map[int]struct{}{
					5890: {},
				},
				thresholdsPerDsp: map[int]int{6: 100}},
			want: 0,
		},
		{
			name: "When pubId,dspid are not present, pubID disabled FSC return fsc=0",
			args: args{
				pubid: 58907,
				dspid: 90,
			},
			setup: func() {

			},
			fields: fields{cache: mockCache,
				disabledPublishers: map[int]struct{}{
					5890: {},
				},
				thresholdsPerDsp: map[int]int{6: 100}},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fsc{
				cache:              tt.fields.cache,
				disabledPublishers: tt.fields.disabledPublishers,
				thresholdsPerDsp:   tt.fields.thresholdsPerDsp,
			}
			tt.setup()
			if got := f.IsUnderFSCThreshold(tt.args.pubid, tt.args.dspid); got != tt.want {
				t.Errorf("fsc.IsUnderFSCThreshold() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_updateFscConfigMapsFromCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_dbcache.NewMockCache(ctrl)
	defer SetAndResetFscWithMockCache(mockCache, nil)()
	type args struct {
		cache cache.Cache
	}
	type wantMaps struct {
		fscDsp map[int]int
		fscPub map[int]struct{}
	}
	tests := []struct {
		name  string
		setup func()
		args  args
		want  wantMaps
	}{
		{
			name: "Cache returns valid values, set in fscConfigs Maps",
			args: args{
				cache: mockCache,
			},
			want: wantMaps{
				fscDsp: map[int]int{6: 70},
				fscPub: map[int]struct{}{
					5890: {}},
			},
			setup: func() {
				mockCache.EXPECT().GetFSCDisabledPublishers().Return(map[int]struct{}{5890: {}}, nil)
				mockCache.EXPECT().GetFSCThresholdPerDSP().Return(map[int]int{6: 70}, nil)
			},
		},
		{
			name: "Cache returns DB error for both DSP and PUB fsc configs",
			args: args{
				cache: mockCache,
			},
			want: wantMaps{
				fscDsp: map[int]int{},
				fscPub: map[int]struct{}{},
			},
			setup: func() {
				mockCache.EXPECT().GetFSCDisabledPublishers().Return(map[int]struct{}{}, errors.New("QUERY FAILED"))
				mockCache.EXPECT().GetFSCThresholdPerDSP().Return(map[int]int{}, errors.New("QUERY FAILED"))
			},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			updateFscConfigMapsFromCache(tt.args.cache)
			if !reflect.DeepEqual(fscConfigs.disabledPublishers, tt.want.fscPub) {
				t.Errorf("updateFscConfigMapsFromCache() for FscDisabledPublishers = %v, %v", fscConfigs.disabledPublishers, tt.want.fscPub)
			}
			if !reflect.DeepEqual(fscConfigs.thresholdsPerDsp, tt.want.fscDsp) {
				t.Errorf("updateFscConfigMapsFromCache() for FscDspThresholds= %v, %v", fscConfigs.disabledPublishers, tt.want.fscDsp)
			}
		})
	}
}

func TestGetFscInstance(t *testing.T) {
	tests := []struct {
		name string
		want *fsc
	}{
		{name: "Return single FSC instance"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFscInstance(); reflect.TypeOf(got) == reflect.TypeOf(fsc{}) {
				t.Errorf("GetFscInstance() gotType = %v, wantedType %v", reflect.TypeOf(got), reflect.TypeOf(fsc{}))
			}
		})
	}
}

func TestIsFscApplicable(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCache := mock_dbcache.NewMockCache(ctrl)
	defer ctrl.Finish()
	defer SetAndResetFscWithMockCache(mockCache, map[int]int{6: 100})()

	type args struct {
		pubId int
		seat  string
		dspId int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Valid Case1: All Params Correct",
			args: args{
				pubId: 5890,
				seat:  "pubmatic",
				dspId: 6,
			},
			want: true,
		},
		{
			name: "Valid Case2: All Params Correct, seat is pubmatic alaias",
			args: args{
				pubId: 5890,
				seat:  "pubmatic2",
				dspId: 6,
			},
			want: true,
		},
		{
			name: "Invalid Case1: DspId is 0",
			args: args{
				pubId: 5890,
				seat:  "pubmatic",
				dspId: 0,
			},
			want: false,
		},
		{
			name: "Invalid Case2: different seat ",
			args: args{
				pubId: 5890,
				seat:  "appnexus",
				dspId: 6,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := IsFscApplicable(tt.args.pubId, tt.args.seat, tt.args.dspId); got != tt.want {
				t.Errorf("isFscApplicable() = %v, want %v", got, tt.want)
			}
		})
	}
}
