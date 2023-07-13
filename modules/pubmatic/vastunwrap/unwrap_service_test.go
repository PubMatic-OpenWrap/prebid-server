package vastunwrap

import (
	"fmt"
	"testing"

	"git.pubmatic.com/vastunwrap/config"
	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	mock_stats "github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/stats/mock"

	"github.com/stretchr/testify/assert"
)

func TestDoUnwrap(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMetricsEngine := mock_stats.NewMockMetricsEngine(ctrl)
	type args struct {
		module               VastUnwrapModule
		bid                  *adapters.TypedBid
		userAgent            string
		unwrapDefaultTimeout int
		url                  string
	}
	tests := []struct {
		name        string
		args        args
		expectedBid *adapters.TypedBid
		setup       func()
	}{
		{
			name: "doUnwrap for adtype video with Empty Bid",
			args: args{
				module: VastUnwrapModule{Cfg: config.VastUnWrapCfg{APPConfig: config.AppConfig{UnwrapDefaultTimeout: 1000}}, MetricsEngine: mockMetricsEngine},
				bid: &adapters.TypedBid{
					Bid: &openrtb2.Bid{},
				},
				userAgent: "testUA",
				url:       UnwrapURL,
			},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordRequestStatus(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		{
			name: "doUnwrap for adtype video with Empty ADM",
			args: args{
				module: VastUnwrapModule{Cfg: config.VastUnWrapCfg{APPConfig: config.AppConfig{UnwrapDefaultTimeout: 1000}}, MetricsEngine: mockMetricsEngine},
				bid: &adapters.TypedBid{
					Bid: &openrtb2.Bid{
						ID:    "Bid-123",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						CrID:  "Cr-234",
						W:     100,
						H:     50,
					},
					BidType: "video",
				},
				userAgent: "testUA",
				url:       UnwrapURL,
			},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordRequestStatus(gomock.Any(), gomock.Any(), gomock.Any())
			},
			expectedBid: &adapters.TypedBid{
				Bid: &openrtb2.Bid{
					ID:    "Bid-123",
					ImpID: fmt.Sprintf("div-adunit-%d", 123),
					Price: 2.1,
					CrID:  "Cr-234",
					AdM:   "",
					W:     100,
					H:     50,
				},
				BidType: "video",
			},
		},
		{
			name: "doUnwrap for adtype video with invalid URL and timeout",
			args: args{
				module: VastUnwrapModule{Cfg: config.VastUnWrapCfg{APPConfig: config.AppConfig{UnwrapDefaultTimeout: 100}}, MetricsEngine: mockMetricsEngine},
				bid: &adapters.TypedBid{
					Bid: &openrtb2.Bid{
						ID:    "Bid-123",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   vastXMLAdM,
						CrID:  "Cr-234",
						W:     100,
						H:     50,
					},
					BidType: "video",
				},
				userAgent: "testUA",
				url:       "testURL",
			},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordRequestStatus(gomock.Any(), gomock.Any(), gomock.Any())
			},
			expectedBid: &adapters.TypedBid{
				Bid: &openrtb2.Bid{
					ID:    "Bid-123",
					ImpID: fmt.Sprintf("div-adunit-%d", 123),
					Price: 2.1,
					CrID:  "Cr-234",
					AdM:   vastXMLAdM,
					W:     100,
					H:     50,
				},
				BidType: "video",
			},
		},
		{
			name: "doUnwrap for adtype video",
			args: args{
				module: VastUnwrapModule{Cfg: config.VastUnWrapCfg{APPConfig: config.AppConfig{UnwrapDefaultTimeout: 1000}}, MetricsEngine: mockMetricsEngine},
				bid: &adapters.TypedBid{
					Bid: &openrtb2.Bid{
						ID:    "Bid-123",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   vastXMLAdM,
						CrID:  "Cr-234",
						W:     100,
						H:     50,
					},
					BidType: "video",
				},
				userAgent: "testUA",
				url:       UnwrapURL,
			},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordRequestStatus(gomock.Any(), gomock.Any(), gomock.Any())
			},
			expectedBid: &adapters.TypedBid{
				Bid: &openrtb2.Bid{
					ID:    "Bid-123",
					ImpID: fmt.Sprintf("div-adunit-%d", 123),
					Price: 2.1,
					CrID:  "Cr-234",
					AdM:   inlineXMLAdM,
					W:     100,
					H:     50,
				},
				BidType: "video",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			doUnwrap(tt.args.module, tt.args.bid, tt.args.userAgent, tt.args.url, "5890", "pubmatic")
			if tt.args.bid.Bid.AdM != "" {
				assert.Equal(t, tt.expectedBid.Bid.AdM, tt.args.bid.Bid.AdM, "AdM is not updated correctly after executing RawBidderResponse hook.")
			}

		})
	}
}
