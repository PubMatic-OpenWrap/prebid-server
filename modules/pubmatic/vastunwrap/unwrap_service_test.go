package vastunwrap

// import (
// 	"fmt"
// 	"net/http"
// 	"testing"

// "git.pubmatic.com/vastunwrap/config"
// unWrapCfg "git.pubmatic.com/vastunwrap/config"
// "github.com/golang/mock/gomock"
// "github.com/prebid/openrtb/v20/openrtb2"
// "github.com/prebid/prebid-server/v2/adapters"
// mock_stats "github.com/prebid/prebid-server/v2/modules/pubmatic/vastunwrap/stats/mock"

// 	"github.com/stretchr/testify/assert"
// )

// func TestDoUnwrap(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()
// 	mockMetricsEngine := mock_stats.NewMockMetricsEngine(ctrl)
// 	type args struct {
// 		module               VastUnwrapModule
// 		statsEnabled         bool
// 		bid                  *adapters.TypedBid
// 		userAgent            string
// 		ip                   string
// 		unwrapDefaultTimeout int
// 		url                  string
// 		wantAdM              bool
// 	}
// 	tests := []struct {
// 		name          string
// 		args          args
// 		setup         func()
// 		unwrapRequest func(w http.ResponseWriter, r *http.Request)
// 	}{
// 		{
// 			name: "doUnwrap for adtype video with Empty Bid",
// 			args: args{
// 				module: VastUnwrapModule{Cfg: config.VastUnWrapCfg{MaxWrapperSupport: 5, StatConfig: unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1}, APPConfig: config.AppConfig{UnwrapDefaultTimeout: 1000}}, MetricsEngine: mockMetricsEngine},
// 				bid: &adapters.TypedBid{
// 					Bid: &openrtb2.Bid{},
// 				},
// 				userAgent: "testUA",
// 				url:       UnwrapURL,
// 			},
// 		},
// 		{
// 			name: "doUnwrap for adtype video with Empty ADM",
// 			args: args{
// 				module: VastUnwrapModule{Cfg: config.VastUnWrapCfg{MaxWrapperSupport: 5, StatConfig: unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1}, APPConfig: config.AppConfig{UnwrapDefaultTimeout: 1000}}, MetricsEngine: mockMetricsEngine},
// 				bid: &adapters.TypedBid{
// 					Bid: &openrtb2.Bid{
// 						ID:    "Bid-123",
// 						ImpID: fmt.Sprintf("div-adunit-%d", 123),
// 						Price: 2.1,
// 						CrID:  "Cr-234",
// 						W:     100,
// 						H:     50,
// 					},
// 					BidType: "video",
// 				},
// 				userAgent: "testUA",
// 				url:       UnwrapURL,
// 			},
// 		},
// 		{
// 			name: "doUnwrap for adtype video with invalid URL and timeout",
// 			args: args{
// 				module: VastUnwrapModule{Cfg: config.VastUnWrapCfg{MaxWrapperSupport: 5, StatConfig: unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1}, APPConfig: config.AppConfig{UnwrapDefaultTimeout: 2}}, MetricsEngine: mockMetricsEngine},
// 				bid: &adapters.TypedBid{
// 					Bid: &openrtb2.Bid{
// 						ID:    "Bid-123",
// 						ImpID: fmt.Sprintf("div-adunit-%d", 123),
// 						Price: 2.1,
// 						AdM:   vastXMLAdM,
// 						CrID:  "Cr-234",
// 						W:     100,
// 						H:     50,
// 					},
// 					BidType: "video",
// 				},
// 				userAgent: "testUA",
// 				url:       "testURL",
// 			},
// 			setup: func() {
// 				mockMetricsEngine.EXPECT().RecordRequestStatus("5890", "pubmatic", "2")
// 				mockMetricsEngine.EXPECT().RecordRequestTime("5890", "pubmatic", gomock.Any())
// 			},
// 			unwrapRequest: func(w http.ResponseWriter, req *http.Request) {
// 				w.Header().Add("unwrap-status", "2")
// 				w.WriteHeader(http.StatusOK)
// 				_, _ = w.Write([]byte(vastXMLAdM))
// 			},
// 		},
// 		{
// 			name: "doUnwrap for adtype video",
// 			args: args{
// 				module: VastUnwrapModule{Cfg: config.VastUnWrapCfg{MaxWrapperSupport: 5, StatConfig: unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1}, APPConfig: config.AppConfig{UnwrapDefaultTimeout: 1500}}, MetricsEngine: mockMetricsEngine},
// 				bid: &adapters.TypedBid{
// 					Bid: &openrtb2.Bid{
// 						ID:    "Bid-123",
// 						ImpID: fmt.Sprintf("div-adunit-%d", 123),
// 						Price: 2.1,
// 						AdM:   vastXMLAdM,
// 						CrID:  "Cr-234",
// 						W:     100,
// 						H:     50,
// 					},
// 					BidType: "video",
// 				},
// 				userAgent: "testUA",
// 				url:       UnwrapURL,
// 				wantAdM:   true,
// 			},
// 			setup: func() {
// 				mockMetricsEngine.EXPECT().RecordRequestStatus("5890", "pubmatic", "0")
// 				mockMetricsEngine.EXPECT().RecordWrapperCount("5890", "pubmatic", "1")
// 				mockMetricsEngine.EXPECT().RecordRequestTime("5890", "pubmatic", gomock.Any())
// 				mockMetricsEngine.EXPECT().RecordUnwrapRespTime("5890", "1", gomock.Any())
// 			},
// 			unwrapRequest: func(w http.ResponseWriter, req *http.Request) {
// 				w.Header().Add("unwrap-status", "0")
// 				w.Header().Add("unwrap-count", "1")
// 				w.WriteHeader(http.StatusOK)
// 				_, _ = w.Write([]byte(inlineXMLAdM))
// 			},
// 		},
// 		{
// 			name: "doUnwrap for adtype video with invalid vast xml",
// 			args: args{
// 				module: VastUnwrapModule{Cfg: config.VastUnWrapCfg{MaxWrapperSupport: 5, StatConfig: unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1}, APPConfig: config.AppConfig{UnwrapDefaultTimeout: 1000}}, MetricsEngine: mockMetricsEngine},
// 				bid: &adapters.TypedBid{
// 					Bid: &openrtb2.Bid{
// 						ID:    "Bid-123",
// 						ImpID: fmt.Sprintf("div-adunit-%d", 123),
// 						Price: 2.1,
// 						AdM:   invalidVastXMLAdM,
// 						CrID:  "Cr-234",
// 						W:     100,
// 						H:     50,
// 					},
// 					BidType: "video",
// 				},
// 				userAgent: "testUA",
// 				url:       UnwrapURL,
// 				wantAdM:   false,
// 			},
// 			setup: func() {
// 				mockMetricsEngine.EXPECT().RecordRequestStatus("5890", "pubmatic", "1")
// 				mockMetricsEngine.EXPECT().RecordRequestTime("5890", "pubmatic", gomock.Any())
// 			},
// 			unwrapRequest: func(w http.ResponseWriter, req *http.Request) {
// 				w.Header().Add("unwrap-status", "1")
// 				w.WriteHeader(http.StatusOK)
// 				_, _ = w.Write([]byte(invalidVastXMLAdM))
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if tt.setup != nil {
// 				tt.setup()
// 			}
// 			m := VastUnwrapModule{
// 				Cfg:           tt.args.module.Cfg,
// 				Enabled:       true,
// 				MetricsEngine: mockMetricsEngine,
// 				unwrapRequest: tt.unwrapRequest,
// 			}
// 			m.doUnwrapandUpdateBid(tt.args.statsEnabled, tt.args.bid, tt.args.userAgent, tt.args.ip, tt.args.url, "5890", "pubmatic")
// 			if tt.args.bid.Bid.AdM != "" && tt.args.wantAdM {
// 				assert.Equal(t, inlineXMLAdM, tt.args.bid.Bid.AdM, "AdM is not updated correctly after executing RawBidderResponse hook.")
// 			}
// 		})
// 	}
// }
