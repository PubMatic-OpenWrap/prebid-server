package openwrap

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/magiconair/properties/assert"
	mock_cache "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

func TestGetVastUnwrapEnabled(t *testing.T) {
	type args struct {
		rctx              models.RequestCtx
		vastunwraptraffic int
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	tests := []struct {
		name         string
		args         args
		setup        func()
		randomNumber int
		want         bool
	}{
		{
			name: "vastunwrap is enabled and trafficpercent is greater than random number",
			args: args{rctx: models.RequestCtx{
				PubID:     5890,
				ProfileID: 123,
				DisplayID: 1,
			},
				vastunwraptraffic: 10,
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					-1: {
						models.VastUnwrapperEnableKey:      "1",
						models.VastUnwrapTrafficPercentKey: "90",
					},
				}, nil)
			},
			randomNumber: 80,
			want:         true,
		},
		{
			name: "vastunwrap is enabled and trafficpercent is less than random number",
			args: args{rctx: models.RequestCtx{
				PubID:     5890,
				ProfileID: 123,
				DisplayID: 1,
			},
				vastunwraptraffic: 0,
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					-1: {
						models.VastUnwrapperEnableKey:      "1",
						models.VastUnwrapTrafficPercentKey: "90",
					},
				}, nil)
			},
			randomNumber: 91,
			want:         false,
		},
		{
			name: "vastunwrap is dissabled and trafficpercent is less than random number",
			args: args{rctx: models.RequestCtx{
				PubID:     5890,
				ProfileID: 123,
				DisplayID: 1,
			}},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					-1: {
						models.VastUnwrapperEnableKey: "0",
					},
				}, nil)
			},
			randomNumber: 91,
			want:         false,
		},
		{
			name: "partnerconfigmap not found",
			args: args{rctx: models.RequestCtx{
				PubID:     5890,
				ProfileID: 123,
				DisplayID: 1,
			}},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			randomNumber: 91,
			want:         false,
		},
		{
			name: "error while fetching partnerconfigmap ",
			args: args{rctx: models.RequestCtx{
				PubID:     5890,
				ProfileID: 123,
				DisplayID: 1,
			}},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("some error"))
			},
			randomNumber: 91,
			want:         false,
		},
		{
			name: "vastunwrap is enabled and trafficpercent not present in DB ",
			args: args{rctx: models.RequestCtx{
				PubID:     5890,
				ProfileID: 123,
				DisplayID: 1,
			},
				vastunwraptraffic: 10,
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					-1: {
						models.VastUnwrapperEnableKey: "1",
					},
				}, nil)
			},
			randomNumber: 9,
			want:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			GetRandomNumberIn1To100 = func() int {
				return tt.randomNumber
			}
			ow = &OpenWrap{
				cache: mockCache,
			}
			got := GetVastUnwrapEnabled(tt.args.rctx, tt.args.vastunwraptraffic)
			assert.Equal(t, got, tt.want)
		})
	}
}
