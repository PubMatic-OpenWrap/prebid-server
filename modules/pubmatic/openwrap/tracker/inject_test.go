package tracker

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/magiconair/properties/assert"
	"github.com/prebid/openrtb/v19/openrtb2"
	mock_metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

func TestInjectTrackers(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
	defer ctrl.Finish()

	models.TrackerCallWrapOMActive = `<script id="OWPubOMVerification" data-owurl="${escapedUrl}" src="https://sample.com/AdServer/js/owpubomverification.js"></script>`

	type args struct {
		rctx        models.RequestCtx
		bidResponse *openrtb2.BidResponse
	}
	tests := []struct {
		name    string
		args    args
		want    *openrtb2.BidResponse
		wantErr bool
	}{
		{
			name: "no_bidresponse",
			args: args{
				bidResponse: &openrtb2.BidResponse{},
			},
			want:    &openrtb2.BidResponse{},
			wantErr: false,
		},
		{
			name: "tracker_params_missing",
			args: args{
				rctx: models.RequestCtx{
					Platform:      "video",
					Trackers:      map[string]models.OWTracker{},
					MetricsEngine: mockEngine,
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "12345",
									AdM: `creative`,
								},
							},
						},
					},
				},
			},
			want: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:  "12345",
								AdM: `creative`,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid_adformat",
			args: args{
				rctx: models.RequestCtx{
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:    "invalid",
							TrackerURL: `Tracking URL`,
						},
					},
					MetricsEngine: mockEngine,
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "12345",
									AdM: `creative`,
								},
							},
						},
					},
				},
			},
			want: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:  "12345",
								AdM: `creative`,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty_tracker_params",
			args: args{
				rctx: models.RequestCtx{
					MetricsEngine: mockEngine,
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "12345",
									AdM: `creative`,
								},
							},
						},
					},
				},
			},
			want: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:  "12345",
								AdM: `creative`,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "adformat_is_banner",
			args: args{
				rctx: models.RequestCtx{
					Platform: "",
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:    "banner",
							TrackerURL: `Tracking URL`,
						},
					},
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "12345",
									AdM: `sample_creative`,
								},
							},
						},
					},
				},
			},
			want: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:  "12345",
								AdM: `sample_creative<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="Tracking URL"></div>`,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "adformat_is_video",
			args: args{
				rctx: models.RequestCtx{
					Platform: "",
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:    "video",
							TrackerURL: `Tracking URL`,
							ErrorURL:   `Error URL`,
						},
					},
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "12345",
									AdM: `<VAST version="3.0"><Ad><Wrapper></Wrapper></Ad></VAST>`,
								},
							},
						},
					},
				},
			},
			want: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:  "12345",
								AdM: `<VAST version="3.0"><Ad><Wrapper><Impression><![CDATA[Tracking URL]]></Impression><Error><![CDATA[Error URL]]></Error></Wrapper></Ad></VAST>`,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "platform_is_video",
			args: args{
				rctx: models.RequestCtx{
					Platform: "video",
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:    "video",
							TrackerURL: `Tracking URL`,
							ErrorURL:   `Error URL`,
						},
					},
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "12345",
									AdM: `<VAST version="3.0"><Ad><Wrapper></Wrapper></Ad></VAST>`,
								},
							},
						},
					},
				},
			},
			want: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:  "12345",
								AdM: `<VAST version="3.0"><Ad><Wrapper><Impression><![CDATA[Tracking URL]]></Impression><Error><![CDATA[Error URL]]></Error></Wrapper></Ad></VAST>`,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "platform_is_app",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:    "banner",
							TrackerURL: `Tracking URL`,
							ErrorURL:   `Error URL`,
						},
					},
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "12345",
									AdM: `sample_creative`,
								},
							},
						},
					},
				},
			},
			want: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:  "12345",
								AdM: `sample_creative<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="Tracking URL"></div>`,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "platform_is_app_with_OM_Inactive_pubmatic",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:    "banner",
							TrackerURL: `Tracking URL`,
							ErrorURL:   `Error URL`,
							DspId:      -1,
						},
					},
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "12345",
									AdM: `sample_creative`,
								},
							},
							Seat: models.BidderPubMatic,
						},
					},
				},
			},
			want: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:  "12345",
								AdM: `sample_creative<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="Tracking URL"></div>`,
							},
						},
						Seat: models.BidderPubMatic,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "platform_is_app_with_OM_active_pubmatic",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:    "banner",
							TrackerURL: `Tracking URL`,
							ErrorURL:   `Error URL`,
							DspId:      models.DspId_DV360,
						},
					},
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "12345",
									AdM: `sample_creative`,
								},
							},
							Seat: models.BidderPubMatic,
						},
					},
				},
			},
			want: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:  "12345",
								AdM: `sample_creative<script id="OWPubOMVerification" data-owurl="Tracking URL" src="https://sample.com/AdServer/js/owpubomverification.js"></script>`,
							},
						},
						Seat: models.BidderPubMatic,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "native_obj_not_found",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:    "native",
							TrackerURL: `Tracking URL`,
							ErrorURL:   `Error URL`,
						},
					},
					MetricsEngine: mockEngine,
					ImpBidCtx:     map[string]models.ImpCtx{},
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:    "12345",
									ImpID: "imp123",
									AdM:   `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"}]}`,
								},
							},
						},
					},
				},
			},
			want: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:    "12345",
								ImpID: "imp123",
								AdM:   `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"}]}`,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "adformat_is_native",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:    "native",
							TrackerURL: `Tracking URL`,
							ErrorURL:   `Error URL`,
						},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"imp123": {
							Native: &openrtb2.Native{
								Request: "{\"context\":1,\"plcmttype\":1,\"eventtrackers\":[{\"event\":1,\"methods\":[1]}],\"ver\":\"1.2\",\"assets\":[{\"id\":0,\"required\":0,\"img\":{\"type\":3,\"w\":300,\"h\":250}},{\"id\":1,\"required\":0,\"data\":{\"type\":1,\"len\":2}},{\"id\":2,\"required\":0,\"img\":{\"type\":1,\"w\":50,\"h\":50}},{\"id\":3,\"required\":0,\"title\":{\"len\":80}},{\"id\":4,\"required\":0,\"data\":{\"type\":2,\"len\":2}}]}",
							},
						},
					},
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:    "12345",
									ImpID: "imp123",
									AdM:   `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"}]}`,
								},
							},
						},
					},
				},
			},
			want: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:    "12345",
								ImpID: "imp123",
								AdM:   `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"},{"event":1,"method":1,"url":"Tracking URL"}]}`,
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.rctx.MetricsEngine != nil {
				mockEngine.EXPECT().RecordInjectTrackerErrorCount(gomock.Any(), gomock.Any(), gomock.Any())
			}
			got, err := InjectTrackers(tt.args.rctx, tt.args.bidResponse)
			if (err != nil) != tt.wantErr {
				t.Errorf("InjectTrackers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InjectTrackers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getUniversalPixels(t *testing.T) {
	type args struct {
		rctx       models.RequestCtx
		adFormat   string
		bidderCode string
	}
	tests := []struct {
		name string
		args args
		want []adunitconfig.UniversalPixel
	}{
		{
			name: "No default in adunitconfig",
			args: args{
				adFormat:   models.Banner,
				bidderCode: models.BidderPubMatic,
				rctx: models.RequestCtx{
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						Config: map[string]*adunitconfig.AdConfig{},
					},
				},
			},
			want: nil,
		},
		{
			name: "No partners",
			args: args{
				adFormat:   models.Banner,
				bidderCode: models.BidderPubMatic,
				rctx: models.RequestCtx{
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						Config: map[string]*adunitconfig.AdConfig{
							"default": {
								UniversalPixel: []adunitconfig.UniversalPixel{
									{
										Id:        123,
										Pixel:     "sample.com",
										PixelType: models.PixelTypeUrl,
										Pos:       models.PixelPosAbove,
										MediaType: "banner",
									},
								},
							},
						},
					},
				},
			},
			want: []adunitconfig.UniversalPixel{
				{
					Id:        123,
					Pixel:     `<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="sample.com"></div>`,
					PixelType: models.PixelTypeUrl,
					Pos:       models.PixelPosAbove,
					MediaType: "banner",
				},
			},
		},
		{
			name: "partner not present",
			args: args{
				adFormat:   models.Banner,
				bidderCode: models.BidderPubMatic,
				rctx: models.RequestCtx{
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						Config: map[string]*adunitconfig.AdConfig{
							"default": {
								UniversalPixel: []adunitconfig.UniversalPixel{
									{
										Id:        123,
										Pixel:     "sample.com",
										PixelType: models.PixelTypeUrl,
										Pos:       models.PixelPosAbove,
										MediaType: models.Banner,
										Partners:  []string{"appnexus"},
									},
								},
							},
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "mismatch in adformat and mediatype",
			args: args{
				adFormat: models.Banner,
				rctx: models.RequestCtx{
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						Config: map[string]*adunitconfig.AdConfig{
							"default": {
								UniversalPixel: []adunitconfig.UniversalPixel{
									{
										Id:        123,
										Pixel:     "sample.com",
										PixelType: models.PixelTypeJS,
										Pos:       models.PixelPosAbove,
										MediaType: "video",
										Partners:  []string{"pubmatic", "appnexus"},
									},
								},
							},
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "send valid upixel",
			args: args{
				bidderCode: models.BidderPubMatic,
				adFormat:   models.Banner,
				rctx: models.RequestCtx{
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						Config: map[string]*adunitconfig.AdConfig{
							"default": {
								UniversalPixel: []adunitconfig.UniversalPixel{
									{
										Id:        123,
										Pixel:     "sample.com",
										PixelType: "url",
										Pos:       models.PixelPosAbove,
										MediaType: "banner",
										Partners:  []string{"pubmatic", "appnexus"},
									},
									{
										Id:        123,
										Pixel:     "<script>__script__</script>",
										PixelType: models.PixelTypeJS,
										Pos:       models.PixelPosBelow,
										MediaType: "banner",
										Partners:  []string{"pubmatic", "appnexus"},
									},
									{
										Id:        123,
										Pixel:     "sample.com",
										PixelType: "url",
										MediaType: "banner",
										Partners:  []string{"pubmatic", "appnexus"},
									},
								},
							},
						},
					},
				},
			},
			want: []adunitconfig.UniversalPixel{
				{
					Id:        123,
					Pixel:     `<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="sample.com"></div>`,
					PixelType: "url",
					Pos:       models.PixelPosAbove,
					MediaType: "banner",
					Partners:  []string{"pubmatic", "appnexus"},
				},
				{
					Id:        123,
					Pixel:     "<script>__script__</script>",
					PixelType: models.PixelTypeJS,
					Pos:       models.PixelPosBelow,
					MediaType: "banner",
					Partners:  []string{"pubmatic", "appnexus"},
				},
				{
					Id:        123,
					Pixel:     `<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="sample.com"></div>`,
					PixelType: "url",
					MediaType: "banner",
					Partners:  []string{"pubmatic", "appnexus"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getUniversalPixels(tt.args.rctx, tt.args.adFormat, tt.args.bidderCode)
			assert.Equal(t, got, tt.want)
		})
	}
}
