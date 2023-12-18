package gocache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	mock_database "github.com/prebid/prebid-server/modules/pubmatic/openwrap/database/mock"
	mock_metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/stretchr/testify/assert"
)

func TestCacheGetBidderFilterConditions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDatabase := mock_database.NewMockDatabase(ctrl)
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)

	type args struct {
		rCtx models.RequestCtx
	}
	type fields struct {
		key  string
		data map[string]*bytes.Reader
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    map[string]*bytes.Reader
		setDate func()
	}{
		{
			name: "Empty bidder condition found in cache",
			args: args{
				rCtx: models.RequestCtx{
					PubID:     123,
					ProfileID: 1234,
					DisplayID: 12345,
				},
			},
			fields: fields{
				key:  fmt.Sprintf("%d_%d_%d_bidding_conditions", 123, 1234, 12345),
				data: map[string]*bytes.Reader{},
			},
			want: map[string]*bytes.Reader{},
		},
		{
			name: "Valid bidder condition found in cache",
			args: args{
				rCtx: models.RequestCtx{
					PubID:     223,
					ProfileID: 134,
					DisplayID: 12345,
				},
			},
			fields: fields{
				key:  fmt.Sprintf("%d_%d_%d_bidding_conditions", 223, 134, 12345),
				data: map[string]*bytes.Reader{"bidderA": bytes.NewReader([]byte(`{"or":[{"and":[{"in":[{"var":"country"},["JPN","KOR"]]},{"==":[{"var":"buyeruidAvailable"},true]}]},{"and":[{"==":[{"var":"testScenario"},"a-jpn-kor-no-uid"]},{"in":[{"var":"country"},["JPN","KOR"]]}]}]}`))},
			},
			want: map[string]*bytes.Reader{"bidderA": bytes.NewReader([]byte(`{"or":[{"and":[{"in":[{"var":"country"},["JPN","KOR"]]},{"==":[{"var":"buyeruidAvailable"},true]}]},{"and":[{"==":[{"var":"testScenario"},"a-jpn-kor-no-uid"]},{"in":[{"var":"country"},["JPN","KOR"]]}]}]}`))},
		},
		{
			name: "No bidding condition found in cache, no bidding conditions in adunit",
			args: args{
				rCtx: models.RequestCtx{
					PubID:     12,
					ProfileID: 123,
					DisplayID: 1234,
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						Config: make(map[string]*adunitconfig.AdConfig),
					},
				},
			},
			fields: fields{
				key:  "",
				data: map[string]*bytes.Reader{},
			},
			want: map[string]*bytes.Reader{},
		},
		{
			name: "No bidding condition found in cache, bidding conditions present in adunit",
			args: args{
				rCtx: models.RequestCtx{
					PubID:     12,
					ProfileID: 13,
					DisplayID: 1234,
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						Config: map[string]*adunitconfig.AdConfig{
							"default": &adunitconfig.AdConfig{
								BidderFilter: &adunitconfig.BidderFilter{
									FilterConfig: []adunitconfig.FilterConfig{
										{
											Bidders:           []string{"bidderA"},
											BiddingConditions: json.RawMessage(`{"or":[{"and":[{"in":[{"var":"country"},["JPN","KOR"]]},{"==":[{"var":"buyeruidAvailable"},true]}]},{"and":[{"==":[{"var":"testScenario"},"a-jpn-kor-no-uid"]},{"in":[{"var":"country"},["JPN","KOR"]]}]}]}`),
										},
									},
								},
							},
						},
					},
				},
			},
			fields: fields{
				key:  "",
				data: map[string]*bytes.Reader{},
			},
			want: map[string]*bytes.Reader{"bidderA": bytes.NewReader([]byte(`{"or":[{"and":[{"in":[{"var":"country"},["JPN","KOR"]]},{"==":[{"var":"buyeruidAvailable"},true]}]},{"and":[{"==":[{"var":"testScenario"},"a-jpn-kor-no-uid"]},{"in":[{"var":"country"},["JPN","KOR"]]}]}]}`))},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := New(gocache.New(1000, 10000), mockDatabase, config.Cache{
				CacheDefaultExpiry: 1000,
			}, mockEngine)
			if tt.fields.key != "" {
				ch.Set(tt.fields.key, tt.fields.data)
			}
			got := ch.GetBidderFilterConditions(tt.args.rCtx)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
