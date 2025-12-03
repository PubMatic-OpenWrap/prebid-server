package tracker

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v20/openrtb2"
	mock_metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
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
			name: "tracker_disabled",
			args: args{
				rctx: models.RequestCtx{
					Platform:        "video",
					TrackerDisabled: true,
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
							BidType:     "banner",
							TrackerURL:  `Tracking URL`,
							ErrorURL:    `Error URL`,
							IsOMEnabled: false,
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
							BidType:     "banner",
							TrackerURL:  `Tracking URL`,
							ErrorURL:    `Error URL`,
							IsOMEnabled: true,
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
									Ext: []byte(`{"key":"value"}`),
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
								Ext: []byte(`{"key":"value","imp_ct_mthd":1}`),
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
		{
			name: "adformat_is_banner_and_AppLovinMax_request",
			args: args{
				rctx: models.RequestCtx{
					Endpoint: models.EndpointAppLovinMax,
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
									ID:   "12345",
									BURL: `http://burl.com`,
									AdM:  `<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="sample.com"></div>`,
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
								ID:   "12345",
								BURL: `Tracking URL&owsspburl=http%3A%2F%2Fburl.com`,
								AdM:  `<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="sample.com"></div>`,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "adformat_is_video_and_AppLovinMax_request",
			args: args{
				rctx: models.RequestCtx{
					Endpoint: models.EndpointAppLovinMax,
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
									ID:   "12345",
									BURL: `http://burl.com`,
									AdM:  `<VAST version="3.0"><Ad><Wrapper></Wrapper></Ad></VAST>`,
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
								ID:   "12345",
								BURL: `Tracking URL&owsspburl=http%3A%2F%2Fburl.com`,
								AdM:  `<VAST version="3.0"><Ad><Wrapper><Error><![CDATA[Error URL]]></Error></Wrapper></Ad></VAST>`,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "adformat_is_native_and_AppLovinMax_request",
			args: args{
				rctx: models.RequestCtx{
					Endpoint: models.EndpointAppLovinMax,
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
									BURL:  `http://burl.com`,
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
								BURL:  `Tracking URL&owsspburl=http%3A%2F%2Fburl.com`,
								AdM:   `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"}]}`,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "VAST_3_0_Wrapper",
			args: args{
				rctx: models.RequestCtx{
					Platform: "",
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:    "video",
							TrackerURL: `Tracking URL`,
							ErrorURL:   `Error URL`,
							Price:      1.2,
						},
					},
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:    "12345",
									AdM:   `<VAST version="3.0"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[www.test.com]]></VASTAdTagURI><Impression></Impression><Creatives></Creatives></Wrapper></Ad></VAST>`,
									Price: 1.2,
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
								AdM:   `<VAST version="3.0"><Ad><Wrapper><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[www.test.com]]></VASTAdTagURI><Impression><![CDATA[Tracking URL]]></Impression><Impression/><Creatives/><Error><![CDATA[Error URL]]></Error><Extensions><Extension><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing></Extension></Extensions></Wrapper></Ad></VAST>`,
								Price: 1.2,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "VAST_3_0_Inline",
			args: args{
				rctx: models.RequestCtx{
					Platform: "",
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:    "video",
							TrackerURL: `Tracking URL`,
							ErrorURL:   `Error URL`,
							Price:      1.5,
						},
					},
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:    "12345",
									AdM:   `<VAST version="3.0" xmlns:xs="http://www.w3.org/2001/XMLSchema"><Ad id="20001"><InLine><AdSystem version="4.0">iabtechlab</AdSystem><AdTitle>iabtechlab video ad</AdTitle><Error>http://example.com/error</Error><Impression id="Impression-ID">http://example.com/track/impression</Impression><Creatives><Creative id="5480" sequence="1"><Linear><Duration>00:00:16</Duration><TrackingEvents><Tracking event="start">http://example.com/tracking/start</Tracking><Tracking event="firstQuartile">http://example.com/tracking/firstQuartile</Tracking><Tracking event="midpoint">http://example.com/tracking/midpoint</Tracking><Tracking event="thirdQuartile">http://example.com/tracking/thirdQuartile</Tracking><Tracking event="complete">http://example.com/tracking/complete</Tracking><Tracking event="progress" offset="00:00:10">http://example.com/tracking/progress-10</Tracking></TrackingEvents><VideoClicks><ClickTracking id="blog"><![CDATA[https://iabtechlab.com]]></ClickTracking><CustomClick>http://iabtechlab.com</CustomClick></VideoClicks><MediaFiles><MediaFile id="5241" delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" minBitrate="360" maxBitrate="1080" scalable="1" maintainAspectRatio="1" codec="0"><![CDATA[https://iab-publicfiles.s3.amazonaws.com/vast/VAST-4.0-Short-Intro.mp4]]></MediaFile></MediaFiles></Linear></Creative></Creatives><Extensions><Extension type="iab-Count"><total_available><![CDATA[ 2 ]]></total_available></Extension></Extensions></InLine></Ad></VAST>`,
									Price: 1.5,
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
								AdM:   `<VAST version="3.0" xmlns:xs="http://www.w3.org/2001/XMLSchema"><Ad id="20001"><InLine><AdSystem version="4.0"><![CDATA[iabtechlab]]></AdSystem><AdTitle><![CDATA[iabtechlab video ad]]></AdTitle><Error><![CDATA[Error URL]]></Error><Error><![CDATA[http://example.com/error]]></Error><Impression><![CDATA[Tracking URL]]></Impression><Impression id="Impression-ID"><![CDATA[http://example.com/track/impression]]></Impression><Creatives><Creative id="5480" sequence="1"><Linear><Duration><![CDATA[00:00:16]]></Duration><TrackingEvents><Tracking event="start"><![CDATA[http://example.com/tracking/start]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://example.com/tracking/firstQuartile]]></Tracking><Tracking event="midpoint"><![CDATA[http://example.com/tracking/midpoint]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://example.com/tracking/thirdQuartile]]></Tracking><Tracking event="complete"><![CDATA[http://example.com/tracking/complete]]></Tracking><Tracking event="progress" offset="00:00:10"><![CDATA[http://example.com/tracking/progress-10]]></Tracking></TrackingEvents><VideoClicks><ClickTracking id="blog"><![CDATA[https://iabtechlab.com]]></ClickTracking><CustomClick><![CDATA[http://iabtechlab.com]]></CustomClick></VideoClicks><MediaFiles><MediaFile id="5241" delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" minBitrate="360" maxBitrate="1080" scalable="1" maintainAspectRatio="1" codec="0"><![CDATA[https://iab-publicfiles.s3.amazonaws.com/vast/VAST-4.0-Short-Intro.mp4]]></MediaFile></MediaFiles></Linear></Creative></Creatives><Extensions><Extension type="iab-Count"><total_available><![CDATA[2]]></total_available></Extension></Extensions><Pricing model="CPM" currency="USD"><![CDATA[1.5]]></Pricing></InLine></Ad></VAST>`,
								Price: 1.5,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "VAST_2_0_Wrapper",
			args: args{
				rctx: models.RequestCtx{
					Platform: "",
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:    "video",
							TrackerURL: `Tracking URL`,
							ErrorURL:   `Error URL`,
							Price:      1.9,
						},
					},
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:    "12345",
									AdM:   `<VAST version="2.0"><Ad id="14568"><Wrapper><AdSystem>TestSystem</AdSystem><VASTAdTagURI>http://demo.test.com/proddev/vast/vast_inline_linear.xml </VASTAdTagURI><Error>http://test.com/wrapper/error</Error><Impression>http://test.com/trackingurl/wrapper/impression</Impression><Creatives><Creative AdID="602833"><Linear><TrackingEvents><Tracking event="creativeView">http://test.com/trackingurl/wrapper/creativeView</Tracking><Tracking event="start">http://test.com/trackingurl/wrapper/start</Tracking><Tracking event="midpoint">http://test.com/trackingurl/wrapper/midpoint</Tracking><Tracking event="firstQuartile">http://test.com/trackingurl/wrapper/firstQuartile</Tracking><Tracking event="thirdQuartile">http://test.com/trackingurl/wrapper/thirdQuartile</Tracking><Tracking event="complete">http://test.com/trackingurl/wrapper/complete</Tracking><Tracking event="mute">http://test.com/trackingurl/wrapper/mute</Tracking><Tracking event="unmute">http://test.com/trackingurl/wrapper/unmute</Tracking><Tracking event="pause">http://test.com/trackingurl/wrapper/pause</Tracking><Tracking event="resume">http://test.com/trackingurl/wrapper/resume</Tracking><Tracking event="fullscreen">http://test.com/trackingurl/wrapper/fullscreen</Tracking></TrackingEvents></Linear></Creative><Creative><Linear><VideoClicks><ClickTracking>http://test.com/trackingurl/wrapper/click</ClickTracking></VideoClicks></Linear></Creative><Creative AdID="1234-NonLinearTracking"><NonLinearAds><TrackingEvents></TrackingEvents></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>`,
									Price: 1.9,
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
								AdM:   `<VAST version="2.0"><Ad id="14568"><Wrapper><AdSystem><![CDATA[TestSystem]]></AdSystem><VASTAdTagURI><![CDATA[http://demo.test.com/proddev/vast/vast_inline_linear.xml]]></VASTAdTagURI><Error><![CDATA[Error URL]]></Error><Error><![CDATA[http://test.com/wrapper/error]]></Error><Impression><![CDATA[Tracking URL]]></Impression><Impression><![CDATA[http://test.com/trackingurl/wrapper/impression]]></Impression><Creatives><Creative AdID="602833"><Linear><TrackingEvents><Tracking event="creativeView"><![CDATA[http://test.com/trackingurl/wrapper/creativeView]]></Tracking><Tracking event="start"><![CDATA[http://test.com/trackingurl/wrapper/start]]></Tracking><Tracking event="midpoint"><![CDATA[http://test.com/trackingurl/wrapper/midpoint]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://test.com/trackingurl/wrapper/firstQuartile]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://test.com/trackingurl/wrapper/thirdQuartile]]></Tracking><Tracking event="complete"><![CDATA[http://test.com/trackingurl/wrapper/complete]]></Tracking><Tracking event="mute"><![CDATA[http://test.com/trackingurl/wrapper/mute]]></Tracking><Tracking event="unmute"><![CDATA[http://test.com/trackingurl/wrapper/unmute]]></Tracking><Tracking event="pause"><![CDATA[http://test.com/trackingurl/wrapper/pause]]></Tracking><Tracking event="resume"><![CDATA[http://test.com/trackingurl/wrapper/resume]]></Tracking><Tracking event="fullscreen"><![CDATA[http://test.com/trackingurl/wrapper/fullscreen]]></Tracking></TrackingEvents></Linear></Creative><Creative><Linear><VideoClicks><ClickTracking><![CDATA[http://test.com/trackingurl/wrapper/click]]></ClickTracking></VideoClicks></Linear></Creative><Creative AdID="1234-NonLinearTracking"><NonLinearAds><TrackingEvents/></NonLinearAds></Creative></Creatives><Extensions><Extension><Pricing model="CPM" currency="USD"><![CDATA[1.9]]></Pricing></Extension></Extensions></Wrapper></Ad></VAST>`,
								Price: 1.9,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "VAST_2_0_Inline",
			args: args{
				rctx: models.RequestCtx{
					Platform: "",
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:    "video",
							TrackerURL: `Tracking URL`,
							ErrorURL:   `Error URL`,
							Price:      2.5,
						},
					},
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:    "12345",
									AdM:   `<VAST version="2.0"><Ad id="ad_1"><InLine><AdSystem>2.0</AdSystem><AdTitle>12345</AdTitle><Error><![CDATA[http://test.com/trackingurl/error]]></Error><Impression><![CDATA[http://test.com/trackingurl/impression]]></Impression><Creatives><Creative><Linear><Duration>00:00:15</Duration><MediaFiles><MediaFile id="5241" delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="1" maintainAspectRatio="1" apiFramework="VPAID"><![CDATA[https://iab-publicfiles.s3.amazonaws.com/vast/VAST-4.0-Short-Intro.mp4]]></MediaFile></MediaFiles></Linear></Creative><Creative><CompanionAds><Companion height="250" width="300" id="573242"><HTMLResource><![CDATA[<A onClick="var i= new Image(1,1); i.src='http://app.scanscout.com/ssframework/log/log.png?a=logitemaction&RI=573242&CbC=1&CbF=true&EC=0&RC=0&SmC=2&CbM=1.0E-5&VI=44cfc3b2382300cb751ba129fe51f46a&admode=preroll&PRI=7496075541100999745&RprC=5&ADsn=20&VcaI=192,197&RrC=1&VgI=44cfc3b2382300cb751ba129fe51f46a&AVI=142&Ust=ma&Uctry=us&CI=1247549&AC=4&PI=567&Udma=506&ADI=5773100&VclF=true';" HREF="http://vaseline.com" target="_blank"><IMG SRC="http://media.scanscout.com/ads/vaseline300x250Companion.jpg" BORDER=0 WIDTH=300 HEIGHT=250 ALT="Click Here"></A><img src="http://app.scanscout.com/ssframework/log/log.png?a=logitemaction&RI=573242&CbC=1&CbF=true&EC=1&RC=0&SmC=2&CbM=1.0E-5&VI=44cfc3b2382300cb751ba129fe51f46a&admode=preroll&PRI=7496075541100999745&RprC=5&ADsn=20&VcaI=192,197&RrC=1&VgI=44cfc3b2382300cb751ba129fe51f46a&AVI=142&Ust=ma&Uctry=us&CI=1247549&AC=4&PI=567&Udma=506&ADI=5773100&VclF=true" height="1" width="1">]]></HTMLResource></Companion></CompanionAds></Creative></Creatives></InLine></Ad></VAST>`,
									Price: 2.5,
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
								AdM:   `<VAST version="2.0"><Ad id="ad_1"><InLine><AdSystem><![CDATA[2.0]]></AdSystem><AdTitle><![CDATA[12345]]></AdTitle><Error><![CDATA[Error URL]]></Error><Error><![CDATA[http://test.com/trackingurl/error]]></Error><Impression><![CDATA[Tracking URL]]></Impression><Impression><![CDATA[http://test.com/trackingurl/impression]]></Impression><Creatives><Creative><Linear><Duration><![CDATA[00:00:15]]></Duration><MediaFiles><MediaFile id="5241" delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="1" maintainAspectRatio="1" apiFramework="VPAID"><![CDATA[https://iab-publicfiles.s3.amazonaws.com/vast/VAST-4.0-Short-Intro.mp4]]></MediaFile></MediaFiles></Linear></Creative><Creative><CompanionAds><Companion height="250" width="300" id="573242"><HTMLResource><![CDATA[<A onClick="var i= new Image(1,1); i.src='http://app.scanscout.com/ssframework/log/log.png?a=logitemaction&RI=573242&CbC=1&CbF=true&EC=0&RC=0&SmC=2&CbM=1.0E-5&VI=44cfc3b2382300cb751ba129fe51f46a&admode=preroll&PRI=7496075541100999745&RprC=5&ADsn=20&VcaI=192,197&RrC=1&VgI=44cfc3b2382300cb751ba129fe51f46a&AVI=142&Ust=ma&Uctry=us&CI=1247549&AC=4&PI=567&Udma=506&ADI=5773100&VclF=true';" HREF="http://vaseline.com" target="_blank"><IMG SRC="http://media.scanscout.com/ads/vaseline300x250Companion.jpg" BORDER=0 WIDTH=300 HEIGHT=250 ALT="Click Here"></A><img src="http://app.scanscout.com/ssframework/log/log.png?a=logitemaction&RI=573242&CbC=1&CbF=true&EC=1&RC=0&SmC=2&CbM=1.0E-5&VI=44cfc3b2382300cb751ba129fe51f46a&admode=preroll&PRI=7496075541100999745&RprC=5&ADsn=20&VcaI=192,197&RrC=1&VgI=44cfc3b2382300cb751ba129fe51f46a&AVI=142&Ust=ma&Uctry=us&CI=1247549&AC=4&PI=567&Udma=506&ADI=5773100&VclF=true" height="1" width="1">]]></HTMLResource></Companion></CompanionAds></Creative></Creatives><Extensions><Extension><Pricing model="CPM" currency="USD"><![CDATA[2.5]]></Pricing></Extension></Extensions></InLine></Ad></VAST>`,
								Price: 2.5,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "VAST_4_0_Wrapper",
			args: args{
				rctx: models.RequestCtx{
					Platform: "",
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:    "video",
							TrackerURL: `Tracking URL`,
							ErrorURL:   `Error URL`,
							Price:      12.5,
						},
					},
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:    "12345",
									AdM:   `<VAST version='4.0' xmlns='http://www.iab.com/VAST'><Ad id='20011' sequence='1' conditionalAd='false'><Wrapper followAdditionalWrappers='0' allowMultipleAds='1' fallbackOnNoAd='0'><AdSystem version='4.0'>iabtechlab</AdSystem><Error>http://example.com/error</Error><Impression id='Impression-ID'>http://example.com/track/impression</Impression><Creatives><Creative id='5480' sequence='1' adId='2447226'><CompanionAds><Companion id='1232' width='100' height='150' assetWidth='250' assetHeight='200' expandedWidth='350' expandedHeight='250' apiFramework='VPAID' adSlotID='3214' pxratio='1400'><StaticResource creativeType='image/png'><![CDATA[https://www.iab.com/wp-content/uploads/2014/09/iab-tech-lab-6-644x290.png]]></StaticResource><CompanionClickThrough><![CDATA[https://iabtechlab.com]]></CompanionClickThrough></Companion></CompanionAds></Creative></Creatives><VASTAdTagURI><![CDATA[https://raw.githubusercontent.com/InteractiveAdvertisingBureau/VAST_Samples/master/VAST%204.0%20Samples/Inline_Companion_Tag-test.xml]]></VASTAdTagURI></Wrapper></Ad></VAST>`,
									Price: 12.5,
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
								AdM:   `<VAST version="4.0" xmlns="http://www.iab.com/VAST"><Ad id="20011" sequence="1" conditionalAd="false"><Wrapper followAdditionalWrappers="0" allowMultipleAds="1" fallbackOnNoAd="0"><AdSystem version="4.0"><![CDATA[iabtechlab]]></AdSystem><Error><![CDATA[Error URL]]></Error><Error><![CDATA[http://example.com/error]]></Error><Impression><![CDATA[Tracking URL]]></Impression><Impression id="Impression-ID"><![CDATA[http://example.com/track/impression]]></Impression><Creatives><Creative id="5480" sequence="1" adId="2447226"><CompanionAds><Companion id="1232" width="100" height="150" assetWidth="250" assetHeight="200" expandedWidth="350" expandedHeight="250" apiFramework="VPAID" adSlotID="3214" pxratio="1400"><StaticResource creativeType="image/png"><![CDATA[https://www.iab.com/wp-content/uploads/2014/09/iab-tech-lab-6-644x290.png]]></StaticResource><CompanionClickThrough><![CDATA[https://iabtechlab.com]]></CompanionClickThrough></Companion></CompanionAds></Creative></Creatives><VASTAdTagURI><![CDATA[https://raw.githubusercontent.com/InteractiveAdvertisingBureau/VAST_Samples/master/VAST%204.0%20Samples/Inline_Companion_Tag-test.xml]]></VASTAdTagURI><Pricing model="CPM" currency="USD"><![CDATA[12.5]]></Pricing></Wrapper></Ad></VAST>`,
								Price: 12.5,
							},
						},
					},
				},
			},
			wantErr: false,
		},

		{
			name: "VAST_4_0_Inline",
			args: args{
				rctx: models.RequestCtx{
					Platform: "",
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:    "video",
							TrackerURL: `Tracking URL`,
							ErrorURL:   `Error URL`,
							Price:      15.7,
						},
					},
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:    "12345",
									AdM:   `<VAST version="4.0" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns="http://www.iab.com/VAST"><Ad id="20008" sequence="1" conditionalAd="false"><InLine><AdSystem version="4.0">iabtechlab</AdSystem><Error>http://example.com/error</Error><Extensions><Extension type="iab-Count"><total_available><![CDATA[ 2 ]]></total_available></Extension></Extensions><Impression id="Impression-ID">http://example.com/track/impression</Impression><Pricing model="cpm" currency="USD"><![CDATA[ 25.00 ]]></Pricing><AdTitle>iabtechlab video ad</AdTitle><Category authority="http://www.iabtechlab.com/categoryauthority">AD CONTENT description category</Category><Creatives><Creative id="5480" sequence="1" adId="2447226"><UniversalAdId idRegistry="Ad-ID" idValue="8465">8465</UniversalAdId><Linear><TrackingEvents><Tracking event="start">http://example.com/tracking/start</Tracking><Tracking event="firstQuartile">http://example.com/tracking/firstQuartile</Tracking><Tracking event="midpoint">http://example.com/tracking/midpoint</Tracking><Tracking event="thirdQuartile">http://example.com/tracking/thirdQuartile</Tracking><Tracking event="complete">http://example.com/tracking/complete</Tracking><Tracking event="progress" offset="00:00:10">http://example.com/tracking/progress-10</Tracking></TrackingEvents><Duration>00:00:16</Duration><MediaFiles><MediaFile id="5241" delivery="progressive" type="video/mp4" bitrate="2000" width="1280" height="720" minBitrate="1500" maxBitrate="2500" scalable="1" maintainAspectRatio="1" codec="H.264"><![CDATA[https://iab-publicfiles.s3.amazonaws.com/vast/VAST-4.0-Short-Intro.mp4]]></MediaFile><MediaFile id="5244" delivery="progressive" type="video/mp4" bitrate="1000" width="854" height="480" minBitrate="700" maxBitrate="1500" scalable="1" maintainAspectRatio="1" codec="H.264"><![CDATA[https://iab-publicfiles.s3.amazonaws.com/vast/VAST-4.0-Short-Intro-mid-resolution.mp4]]></MediaFile><MediaFile id="5246" delivery="progressive" type="video/mp4" bitrate="600" width="640" height="360" minBitrate="500" maxBitrate="700" scalable="1" maintainAspectRatio="1" codec="H.264"><![CDATA[https://iab-publicfiles.s3.amazonaws.com/vast/VAST-4.0-Short-Intro-low-resolution.mp4]]></MediaFile></MediaFiles><VideoClicks><ClickThrough id="blog"><![CDATA[https://iabtechlab.com]]></ClickThrough></VideoClicks></Linear></Creative></Creatives></InLine></Ad></VAST>`,
									Price: 15.7,
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
								AdM:   `<VAST version="4.0" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns="http://www.iab.com/VAST"><Ad id="20008" sequence="1" conditionalAd="false"><InLine><AdSystem version="4.0"><![CDATA[iabtechlab]]></AdSystem><Error><![CDATA[Error URL]]></Error><Error><![CDATA[http://example.com/error]]></Error><Extensions><Extension type="iab-Count"><total_available><![CDATA[2]]></total_available></Extension></Extensions><Impression><![CDATA[Tracking URL]]></Impression><Impression id="Impression-ID"><![CDATA[http://example.com/track/impression]]></Impression><Pricing model="CPM" currency="USD"><![CDATA[15.7]]></Pricing><AdTitle><![CDATA[iabtechlab video ad]]></AdTitle><Category authority="http://www.iabtechlab.com/categoryauthority"><![CDATA[AD CONTENT description category]]></Category><Creatives><Creative id="5480" sequence="1" adId="2447226"><UniversalAdId idRegistry="Ad-ID" idValue="8465"><![CDATA[8465]]></UniversalAdId><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://example.com/tracking/start]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://example.com/tracking/firstQuartile]]></Tracking><Tracking event="midpoint"><![CDATA[http://example.com/tracking/midpoint]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://example.com/tracking/thirdQuartile]]></Tracking><Tracking event="complete"><![CDATA[http://example.com/tracking/complete]]></Tracking><Tracking event="progress" offset="00:00:10"><![CDATA[http://example.com/tracking/progress-10]]></Tracking></TrackingEvents><Duration><![CDATA[00:00:16]]></Duration><MediaFiles><MediaFile id="5241" delivery="progressive" type="video/mp4" bitrate="2000" width="1280" height="720" minBitrate="1500" maxBitrate="2500" scalable="1" maintainAspectRatio="1" codec="H.264"><![CDATA[https://iab-publicfiles.s3.amazonaws.com/vast/VAST-4.0-Short-Intro.mp4]]></MediaFile><MediaFile id="5244" delivery="progressive" type="video/mp4" bitrate="1000" width="854" height="480" minBitrate="700" maxBitrate="1500" scalable="1" maintainAspectRatio="1" codec="H.264"><![CDATA[https://iab-publicfiles.s3.amazonaws.com/vast/VAST-4.0-Short-Intro-mid-resolution.mp4]]></MediaFile><MediaFile id="5246" delivery="progressive" type="video/mp4" bitrate="600" width="640" height="360" minBitrate="500" maxBitrate="700" scalable="1" maintainAspectRatio="1" codec="H.264"><![CDATA[https://iab-publicfiles.s3.amazonaws.com/vast/VAST-4.0-Short-Intro-low-resolution.mp4]]></MediaFile></MediaFiles><VideoClicks><ClickThrough id="blog"><![CDATA[https://iabtechlab.com]]></ClickThrough></VideoClicks></Linear></Creative></Creatives></InLine></Ad></VAST>`,
								Price: 15.7,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "platform_is_app_with_imp_counting_method_enabled_ix",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:     "banner",
							TrackerURL:  `Tracking URL`,
							ErrorURL:    `Error URL`,
							IsOMEnabled: true,
						},
					},
					ImpCountingMethodEnabledBidders: map[string]struct{}{
						string(openrtb_ext.BidderIx): {},
					},
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "12345",
									AdM: `sample_creative`,
									Ext: []byte(`{"key":"value"}`),
								},
							},
							Seat: string(openrtb_ext.BidderIx),
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
								Ext: []byte(`{"key":"value","imp_ct_mthd":1}`),
							},
						},
						Seat: string(openrtb_ext.BidderIx),
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
			assert.Equal(t, got, tt.want, tt.name)
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

func Test_getBurlAppLovinMax(t *testing.T) {
	type args struct {
		burl    string
		tracker models.OWTracker
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty_burl",
			args: args{
				burl:    "",
				tracker: models.OWTracker{TrackerURL: `sample.com`},
			},
			want: `sample.com`,
		},
		{
			name: "empty_tracker_url",
			args: args{
				burl:    `sample.com`,
				tracker: models.OWTracker{TrackerURL: ""},
			},
			want: `sample.com`,
		},
		{
			name: "valid_burl_and_tracker_url",
			args: args{
				burl:    `https://abc.xyz.com/AdServer/AdDisplayTrackerServlet?operId=1&pubId=161527&siteId=991727&adId=4695996&imprId=B430AE6F-4768-41D0-BC55-8CF9D5DD4DA6&cksum=41C0F6460C2ACF7F&adType=10&adServerId=243&kefact=0.095500&kaxefact=0.095500&kadNetFrequecy=0&kadwidth=300&kadheight=250&kadsizeid=9&kltstamp=1721827593&indirectAdId=0`,
				tracker: models.OWTracker{TrackerURL: `sampleTracker.com?id=123`},
			},
			want: `sampleTracker.com?id=123&owsspburl=https%3A%2F%2Fabc.xyz.com%2FAdServer%2FAdDisplayTrackerServlet%3FoperId%3D1%26pubId%3D161527%26siteId%3D991727%26adId%3D4695996%26imprId%3DB430AE6F-4768-41D0-BC55-8CF9D5DD4DA6%26cksum%3D41C0F6460C2ACF7F%26adType%3D10%26adServerId%3D243%26kefact%3D0.095500%26kaxefact%3D0.095500%26kadNetFrequecy%3D0%26kadwidth%3D300%26kadheight%3D250%26kadsizeid%3D9%26kltstamp%3D1721827593%26indirectAdId%3D0`,
		},
		{
			name: "send_tracker_url_if_om_enabled",
			args: args{
				burl:    `https://abc.xyz.com/AdServer/AdDisplayTrackerServlet?operId=1&pubId=161527&siteId=991727&adId=4695996&imprId=B430AE6F-4768-41D0-BC55-8CF9D5DD4DA6&cksum=41C0F6460C2ACF7F&adType=10&adServerId=243&kefact=0.095500&kaxefact=0.095500&kadNetFrequecy=0&kadwidth=300&kadheight=250&kadsizeid=9&kltstamp=1721827593&indirectAdId=0`,
				tracker: models.OWTracker{TrackerURL: `sampleTracker.com?id=123`, IsOMEnabled: true, BidType: models.Banner, Tracker: models.Tracker{PartnerInfo: models.Partner{PartnerID: models.BidderPubMatic}}},
			},
			want: `sampleTracker.com?id=123`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getBURL(tt.args.burl, tt.args.tracker)
			assert.Equal(t, got, tt.want)
		})
	}
}
