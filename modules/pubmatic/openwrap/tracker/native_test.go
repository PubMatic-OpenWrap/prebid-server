package tracker

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
)

func Test_injectNativeCreativeTrackers(t *testing.T) {
	type args struct {
		native   *openrtb2.Native
		adm      string
		tracker  models.OWTracker
		endpoint string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "AppLovinMax_endpoint_request",
			args: args{
				native:   &openrtb2.Native{},
				adm:      `creative`,
				tracker:  models.OWTracker{},
				endpoint: models.EndpointAppLovinMax,
			},
			want:    `creative`,
			wantErr: false,
		},
		{
			name: "native_object_nil",
			args: args{
				adm: `creative`,
				tracker: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
				endpoint: models.EndpointV25,
			},
			want:    `creative`,
			wantErr: true,
		},
		{
			name: "empty_native_object",
			args: args{
				native: &openrtb2.Native{},
				adm:    `creative`,
				tracker: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
				endpoint: models.EndpointV25,
			},
			want:    `creative`,
			wantErr: true,
		},
		{
			name: "error_injecting_impression_tracker",
			args: args{
				native: &openrtb2.Native{
					Request: `request`,
				},
				adm: `creative`,
				tracker: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
				endpoint: models.EndpointV25,
			},
			want:    `creative`,
			wantErr: true,
		},
		{
			name: "no_eventTracker_in_req_but_impTracker_exists_in_adm",
			args: args{
				native: &openrtb2.Native{
					Request: "{\"context\":1,\"plcmttype\":1,\"ver\":\"1.2\",\"assets\":[{\"id\":0,\"required\":0,\"img\":{\"type\":3,\"w\":300,\"h\":250}},{\"id\":1,\"required\":0,\"data\":{\"type\":1,\"len\":2}},{\"id\":2,\"required\":0,\"img\":{\"type\":1,\"w\":50,\"h\":50}},{\"id\":3,\"required\":0,\"title\":{\"len\":80}},{\"id\":4,\"required\":0,\"data\":{\"type\":2,\"len\":2}}]}",
				},
				adm: `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"}]}`,
				tracker: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
				endpoint: models.EndpointV25,
			},
			want:    `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35","Tracking URL"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"}]}`,
			wantErr: false,
		},
		{
			name: "none_of_eventTracker_and_impTracker_in_req",
			args: args{
				native: &openrtb2.Native{
					Request: "{\"context\":1,\"plcmttype\":1,\"ver\":\"1.2\",\"assets\":[{\"id\":0,\"required\":0,\"img\":{\"type\":3,\"w\":300,\"h\":250}},{\"id\":1,\"required\":0,\"data\":{\"type\":1,\"len\":2}},{\"id\":2,\"required\":0,\"img\":{\"type\":1,\"w\":50,\"h\":50}},{\"id\":3,\"required\":0,\"title\":{\"len\":80}},{\"id\":4,\"required\":0,\"data\":{\"type\":2,\"len\":2}}]}",
				},
				adm: `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"}]}`,
				tracker: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
				endpoint: models.EndpointV25,
			},
			want:    `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"}],"imptrackers":["Tracking URL"]}`,
			wantErr: false,
		},
		{
			name: "event_is_1_&_method_is_2",
			args: args{
				native: &openrtb2.Native{
					Request: "{\"context\":1,\"plcmttype\":1,\"eventtrackers\":[{\"event\":1,\"methods\":[2]}],\"ver\":\"1.2\",\"assets\":[{\"id\":0,\"required\":0,\"img\":{\"type\":3,\"w\":300,\"h\":250}},{\"id\":1,\"required\":0,\"data\":{\"type\":1,\"len\":2}},{\"id\":2,\"required\":0,\"img\":{\"type\":1,\"w\":50,\"h\":50}},{\"id\":3,\"required\":0,\"title\":{\"len\":80}},{\"id\":4,\"required\":0,\"data\":{\"type\":2,\"len\":2}}]}",
				},
				adm: `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":2,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet.js"}]}`,
				tracker: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
				endpoint: models.EndpointV25,
			},
			want:    `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":2,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet.js"}],"imptrackers":["Tracking URL"]}`,
			wantErr: false,
		},
		{
			name: "event_is_2_&_method_is_1",
			args: args{
				native: &openrtb2.Native{
					Request: "{\"context\":1,\"plcmttype\":1,\"eventtrackers\":[{\"event\":2,\"methods\":[1]}],\"ver\":\"1.2\",\"assets\":[{\"id\":0,\"required\":0,\"img\":{\"type\":3,\"w\":300,\"h\":250}},{\"id\":1,\"required\":0,\"data\":{\"type\":1,\"len\":2}},{\"id\":2,\"required\":0,\"img\":{\"type\":1,\"w\":50,\"h\":50}},{\"id\":3,\"required\":0,\"title\":{\"len\":80}},{\"id\":4,\"required\":0,\"data\":{\"type\":2,\"len\":2}}]}",
				},
				adm: `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["abc.com"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":2,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet.php"}]}`,
				tracker: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
				endpoint: models.EndpointV25,
			},
			want:    `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["abc.com","Tracking URL"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":2,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet.php"}]}`,
			wantErr: false,
		},
		{
			name: "event_1_&_method_1_but_eventtracker_exist_in_adm",
			args: args{
				native: &openrtb2.Native{
					Request: "{\"context\":1,\"plcmttype\":1,\"eventtrackers\":[{\"event\":1,\"methods\":[1]}],\"ver\":\"1.2\",\"assets\":[{\"id\":0,\"required\":0,\"img\":{\"type\":3,\"w\":300,\"h\":250}},{\"id\":1,\"required\":0,\"data\":{\"type\":1,\"len\":2}},{\"id\":2,\"required\":0,\"img\":{\"type\":1,\"w\":50,\"h\":50}},{\"id\":3,\"required\":0,\"title\":{\"len\":80}},{\"id\":4,\"required\":0,\"data\":{\"type\":2,\"len\":2}}]}",
				},
				adm: `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"}]}`,
				tracker: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
				endpoint: models.EndpointV25,
			},
			want:    `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"},{"event":1,"method":1,"url":"Tracking URL"}]}`,
			wantErr: false,
		},
		{
			name: "event_1_&_method_1&2_but_eventtracker_exist_in_adm",
			args: args{
				native: &openrtb2.Native{
					Request: "{\"context\":1,\"plcmttype\":1,\"eventtrackers\":[{\"event\":1,\"methods\":[1,2]}],\"ver\":\"1.2\",\"assets\":[{\"id\":0,\"required\":0,\"img\":{\"type\":3,\"w\":300,\"h\":250}},{\"id\":1,\"required\":0,\"data\":{\"type\":1,\"len\":2}},{\"id\":2,\"required\":0,\"img\":{\"type\":1,\"w\":50,\"h\":50}},{\"id\":3,\"required\":0,\"title\":{\"len\":80}},{\"id\":4,\"required\":0,\"data\":{\"type\":2,\"len\":2}}]}",
				},
				adm: `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"}]}`,
				tracker: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
				endpoint: models.EndpointV25,
			},
			want:    `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"},{"event":1,"method":1,"url":"Tracking URL"}]}`,
			wantErr: false,
		},
		{
			name: "event_1_&_method_1_but_eventtracker_NOT_exist_in_adm",
			args: args{
				native: &openrtb2.Native{
					Request: "{\"context\":1,\"plcmttype\":1,\"eventtrackers\":[{\"event\":1,\"methods\":[1]}],\"ver\":\"1.2\",\"assets\":[{\"id\":0,\"required\":0,\"img\":{\"type\":3,\"w\":300,\"h\":250}},{\"id\":1,\"required\":0,\"data\":{\"type\":1,\"len\":2}},{\"id\":2,\"required\":0,\"img\":{\"type\":1,\"w\":50,\"h\":50}},{\"id\":3,\"required\":0,\"title\":{\"len\":80}},{\"id\":4,\"required\":0,\"data\":{\"type\":2,\"len\":2}}]}",
				},
				adm: `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e"}`,
				tracker: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
				endpoint: models.EndpointV25,
			},
			want:    `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"Tracking URL"}]}`,
			wantErr: false,
		},
		{
			name: "event_1_method_1&2_and_multiple_eventrackers",
			args: args{
				native: &openrtb2.Native{
					Request: "{\"context\":1,\"plcmttype\":1,\"eventtrackers\":[{\"event\":1,\"methods\":[1,2]}],\"ver\":\"1.2\",\"assets\":[{\"id\":0,\"required\":0,\"img\":{\"type\":3,\"w\":300,\"h\":250}},{\"id\":1,\"required\":0,\"data\":{\"type\":1,\"len\":2}},{\"id\":2,\"required\":0,\"img\":{\"type\":1,\"w\":50,\"h\":50}},{\"id\":3,\"required\":0,\"title\":{\"len\":80}},{\"id\":4,\"required\":0,\"data\":{\"type\":2,\"len\":2}}]}",
				},
				adm: `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"},{"event":1,"method":1,"url":"abc.com"}]}`,
				tracker: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
				endpoint: models.EndpointV25,
			},
			want:    `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"},{"event":1,"method":1,"url":"abc.com"},{"event":1,"method":1,"url":"Tracking URL"}]}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := injectNativeCreativeTrackers(tt.args.native, tt.args.adm, tt.args.tracker, tt.args.endpoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("injectNativeCreativeTrackers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("injectNativeCreativeTrackers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_injectNativeEventTracker(t *testing.T) {
	type args struct {
		adm          *string
		value        []byte
		trackerParam models.OWTracker
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{
			name: "error_injecting_eventTracker_empty_adm",
			args: args{
				value: []byte("{\"event\":1,\"methods\":[1]}"),
				adm:   ptrutil.ToPtr(""),
				trackerParam: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
			},
			want:  "",
			want1: false,
		},
		{
			name: "event is 1 & method 2",
			args: args{
				value: []byte("{\"event\":1,\"methods\":[2]}"),
				adm:   ptrutil.ToPtr(`{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":2,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet.js"}]}`),
				trackerParam: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
			},
			want:  `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":2,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet.js"}]}`,
			want1: false,
		},
		{
			name: "event is 2 & method 1",
			args: args{
				value: []byte("{\"event\":2,\"methods\":[1]}"),
				adm:   ptrutil.ToPtr(`{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":2,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet.php"}]}`),
				trackerParam: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
			},
			want:  `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":2,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet.php"}]}`,
			want1: false,
		},
		{
			name: "event 1 & method 1, eventtracker exist in adm",
			args: args{
				value: []byte("{\"event\":1,\"methods\":[1]}"),
				adm:   ptrutil.ToPtr(`{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"}]}`),
				trackerParam: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
			},
			want:  `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"},{"event":1,"method":1,"url":"Tracking URL"}]}`,
			want1: true,
		},
		{
			name: "event 1 & method 1,2, eventtracker exist in adm",
			args: args{
				value: []byte("{\"event\":1,\"methods\":[1,2]}"),
				adm:   ptrutil.ToPtr(`{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"}]}`),
				trackerParam: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
			},
			want:  `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"},{"event":1,"method":1,"url":"Tracking URL"}]}`,
			want1: true,
		},
		{
			name: "event 1 & method 1, eventtracker NOT exist in adm",
			args: args{
				value: []byte("{\"event\":1,\"methods\":[1]}"),
				adm:   ptrutil.ToPtr(`{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e"}`),
				trackerParam: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
			},
			want:  `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"Tracking URL"}]}`,
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := injectNativeEventTracker(tt.args.adm, tt.args.value, tt.args.trackerParam)
			if got != tt.want {
				t.Errorf("injectNativeEventTracker() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("injectNativeEventTracker() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_injectNativeImpressionTracker(t *testing.T) {
	type args struct {
		adm     *string
		tracker models.OWTracker
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "error_injecting_impression_tracker_empty_adm",
			args: args{
				adm: ptrutil.ToPtr(""),
				tracker: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "no_eventTracker_in_req_but_impTracker_exists_in_adm",
			args: args{
				adm: ptrutil.ToPtr(`{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"}]}`),
				tracker: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
			},
			want:    `{"assets":[{"id":0,"img":{"type":3,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//sample.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.sample.com","clicktrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.sample.com"},"imptrackers":["http://sampletracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35","Tracking URL"],"jstracker":"\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/sample.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://sample.com/AdServer/AdDisplayTrackerServlet"}]}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := injectNativeImpressionTracker(tt.args.adm, tt.args.tracker)
			if (err != nil) != tt.wantErr {
				t.Errorf("injectNativeImpressionTracker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("injectNativeImpressionTracker() = %v, want %v", got, tt.want)
			}
		})
	}
}
