package openwrap

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v19/openrtb2"
	cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	mock_cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestEvaluateLogic(t *testing.T) {
	type args struct {
		data  *bytes.Reader
		logic *bytes.Reader
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Invalid data",
			args: args{
				data:  bytes.NewReader([]byte(`{"country":}`)),
				logic: bytes.NewReader([]byte(`{"or":[{"and":[{"in":[{"var":"country"},["JPN","KOR"]]},{"==":[{"var":"buyeruidAvailable"},true]}]},{"and":[{"==":[{"var":"testScenario"},"a-jpn-kor-no-uid"]},{"in":[{"var":"country"},["JPN","KOR"]]}]}]}`)),
			},
		},
		{
			name: "Invalid logic",
			args: args{
				data:  bytes.NewReader([]byte(`{"country":"in"}`)),
				logic: bytes.NewReader([]byte(`"or":[{"and":[{"in":[{"var":"country"},["JPN","KOR"]]},{"==":[{"var":"buyeruidAvailable"},true]}]},{"and":[{"==":[{"var":"testScenario"},"a-jpn-kor-no-uid"]},{"in":[{"var":"country"},["JPN","KOR"]]}]}]}`)),
			},
		},
		{
			name: "logic evaluates false",
			args: args{
				data:  bytes.NewReader([]byte(`{"country":"in"}`)),
				logic: bytes.NewReader([]byte(`{"or":[{"and":[{"in":[{"var":"country"},["JPN","KOR"]]},{"==":[{"var":"buyeruidAvailable"},true]}]},{"and":[{"==":[{"var":"testScenario"},"a-jpn-kor-no-uid"]},{"in":[{"var":"country"},["JPN","KOR"]]}]}]}`)),
			},
		},
		{
			name: "logic evaluates true",
			args: args{
				data:  bytes.NewReader([]byte(`{"country":"in","buyeruidAvailable":true}`)),
				logic: bytes.NewReader([]byte(`{"or":[{"and":[{"in":[{"var":"country"},["in","KOR"]]},{"==":[{"var":"buyeruidAvailable"},true]}]},{"and":[{"==":[{"var":"testScenario"},"a-jpn-kor-no-uid"]},{"in":[{"var":"country"},["JPN","KOR"]]}]}]}`)),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evaluateBiddingCondition(tt.args.data, tt.args.logic)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestGetFilteredBidders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)
	type args struct {
		rCtx             models.RequestCtx
		bidRequest       *openrtb2.BidRequest
		cache            cache.Cache
		partnerConfigMap map[int]map[string]string
	}
	tests := []struct {
		name  string
		args  args
		setup func()
		want  map[string]bool
		want1 bool
	}{
		{
			name: "no bidding condition present in adunit",
			args: args{
				rCtx:       models.RequestCtx{},
				bidRequest: &openrtb2.BidRequest{},
				cache:      mockCache,
			},
			setup: func() {
				mockCache.EXPECT().GetBidderFilterConditions(gomock.Any()).Return(map[string]*bytes.Reader{})
			},
			want:  map[string]bool{},
			want1: false,
		},
		{
			name: "BidderA filtered, bidding condition present in adunit",
			args: args{
				rCtx: models.RequestCtx{},
				bidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						Geo: &openrtb2.Geo{
							Country: "IND",
						},
					},
				},
				cache:            mockCache,
				partnerConfigMap: map[int]map[string]string{1: {models.BidderCode: "bidderA", models.SERVER_SIDE_FLAG: "1"}},
			},
			setup: func() {
				mockCache.EXPECT().GetBidderFilterConditions(gomock.Any()).Return(map[string]*bytes.Reader{"bidderA": bytes.NewReader([]byte(`{"or":[{"in":[{"var":"country"},["IN"]]},{"and":[{"==":[{"var":"testScenario"},"a-jpn-kor-no-uid"]},{"in":[{"var":"country"},["JPN","KOR"]]}]}]}`))})
			},
			want:  map[string]bool{"bidderA": true},
			want1: false,
		},
		{
			name: "All bidders dropped, bidder filter condition present in adunit",
			args: args{
				rCtx: models.RequestCtx{},
				bidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						Geo: &openrtb2.Geo{
							Country: "IND",
						},
					},
				},
				cache:            mockCache,
				partnerConfigMap: map[int]map[string]string{1: {models.BidderCode: "bidderA", models.SERVER_SIDE_FLAG: "1"}},
			},
			setup: func() {
				mockCache.EXPECT().GetBidderFilterConditions(gomock.Any()).Return(map[string]*bytes.Reader{"bidderA": bytes.NewReader([]byte(`{"or":[{"in":[{"var":"country"},["US"]]},{"and":[{"==":[{"var":"testScenario"},"a-jpn-kor-no-uid"]},{"in":[{"var":"country"},["JPN","KOR"]]}]}]}`))})
			},
			want:  map[string]bool{},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got, got1 := GetFilteredBidders(tt.args.rCtx, tt.args.bidRequest, tt.args.cache, tt.args.partnerConfigMap)
			assert.Equal(t, got, tt.want, tt.name)
			assert.Equal(t, got1, tt.want1, tt.name)
		})
	}
}
