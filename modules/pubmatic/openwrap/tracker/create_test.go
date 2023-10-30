package tracker

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func TestConstructTrackerURL(t *testing.T) {
	type args struct {
		rctx    models.RequestCtx
		tracker models.Tracker
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "trackerEndpoint_parsingError",
			args: args{
				rctx: models.RequestCtx{
					TrackerEndpoint: ":www.example.com",
				},
			},
			want: "",
		},
		{
			name: "empty_tracker_details",
			args: args{
				rctx: models.RequestCtx{
					TrackerEndpoint: "//t.pubmatic.com/wt",
					Platform:        models.PLATFORM_DISPLAY,
				},
				tracker: models.Tracker{},
			},
			want: "http://t.pubmatic.com/wt?adv=&af=&aps=0&au=&bc=&bidid=&di=&eg=0&en=0&ft=0&iid=&kgpv=&orig=&origbidid=&pdvid=&pid=&plt=0&pn=&psz=&pubid=0&purl=&sl=1&slot=&ss=0&tgid=0&tst=0",
		},
		{
			name: "platform_amp_with_tracker_details",
			args: args{
				rctx: models.RequestCtx{
					TrackerEndpoint: "//t.pubmatic.com/wt",
					Platform:        models.PLATFORM_AMP,
				},
				tracker: models.Tracker{
					PubID:     12345,
					PageURL:   "www.abc.com",
					IID:       "98765",
					ProfileID: "123",
					VersionID: "1",
					SlotID:    "1234_1234",
					Adunit:    "adunit",
					Platform:  1,
					Origin:    "www.publisher.com",
					TestGroup: 1,
					AdPodSlot: 0,
					PartnerInfo: models.Partner{
						PartnerID:  "AppNexus",
						BidderCode: "AppNexus1",
						BidID:      "6521",
						OrigBidID:  "6521",
						GrossECPM:  4.3,
						NetECPM:    2.5,
						KGPV:       "adunit@300x250",
						AdDuration: 10,
						Adformat:   models.Banner,
						AdSize:     "300x250",
						ServerSide: 1,
						Advertiser: "fb.com",
						DealID:     "420",
					},
				},
			},
			want: "//t.pubmatic.com/wt?adv=fb.com&af=banner&aps=0&au=adunit&bc=AppNexus1&bidid=6521&di=420&dur=10&eg=4.3&en=2.5&ft=0&iid=98765&kgpv=adunit%40300x250&orig=www.publisher.com&origbidid=6521&pdvid=1&pid=123&plt=1&pn=AppNexus&psz=300x250&pubid=12345&purl=www.abc.com&sl=1&slot=1234_1234&ss=1&tgid=1&tst=0",
		},
		{
			name: "all_details_with_ssai_in_tracker",
			args: args{
				rctx: models.RequestCtx{
					TrackerEndpoint: "//t.pubmatic.com/wt",
					Platform:        models.PLATFORM_DISPLAY,
				},
				tracker: models.Tracker{
					PubID:             12345,
					PageURL:           "www.abc.com",
					IID:               "98765",
					ProfileID:         "123",
					VersionID:         "1",
					SlotID:            "1234_1234",
					Adunit:            "adunit",
					Platform:          1,
					Origin:            "www.publisher.com",
					TestGroup:         1,
					AdPodSlot:         0,
					FloorSkippedFlag:  ptrutil.ToPtr(0),
					FloorModelVersion: "test version",
					FloorSource:       ptrutil.ToPtr(1),
					FloorType:         1,
					RewardedInventory: 1,
					Secure:            1,
					SSAI:              "mediatailor",
					PartnerInfo: models.Partner{
						PartnerID:      "AppNexus",
						BidderCode:     "AppNexus1",
						BidID:          "6521",
						OrigBidID:      "6521",
						GrossECPM:      4.3,
						NetECPM:        2.5,
						KGPV:           "adunit@300x250",
						AdDuration:     10,
						Adformat:       models.Banner,
						AdSize:         "300x250",
						ServerSide:     1,
						Advertiser:     "fb.com",
						DealID:         "420",
						FloorValue:     4.4,
						FloorRuleValue: 2,
					},
				},
			},
			want: "https://t.pubmatic.com/wt?adv=fb.com&af=banner&aps=0&au=adunit&bc=AppNexus1&bidid=6521&di=420&dur=10&eg=4.3&en=2.5&fmv=test+version&frv=2&fskp=0&fsrc=1&ft=1&fv=4.4&iid=98765&kgpv=adunit%40300x250&orig=www.publisher.com&origbidid=6521&pdvid=1&pid=123&plt=1&pn=AppNexus&psz=300x250&pubid=12345&purl=www.abc.com&rwrd=1&sl=1&slot=1234_1234&ss=1&ssai=mediatailor&tgid=1&tst=0",
		},
		{
			name: "all_details_with_secure_enable_in_tracker",
			args: args{
				rctx: models.RequestCtx{
					TrackerEndpoint: "//t.pubmatic.com/wt",
					Platform:        models.PLATFORM_DISPLAY,
				},
				tracker: models.Tracker{
					PubID:             12345,
					PageURL:           "www.abc.com",
					IID:               "98765",
					ProfileID:         "123",
					VersionID:         "1",
					SlotID:            "1234_1234",
					Adunit:            "adunit",
					Platform:          1,
					Origin:            "www.publisher.com",
					TestGroup:         1,
					AdPodSlot:         0,
					FloorSkippedFlag:  ptrutil.ToPtr(0),
					FloorModelVersion: "test version",
					FloorSource:       ptrutil.ToPtr(1),
					FloorType:         1,
					RewardedInventory: 1,
					Secure:            1,
					PartnerInfo: models.Partner{
						PartnerID:      "AppNexus",
						BidderCode:     "AppNexus1",
						BidID:          "6521",
						OrigBidID:      "6521",
						GrossECPM:      4.3,
						NetECPM:        2.5,
						KGPV:           "adunit@300x250",
						AdDuration:     10,
						Adformat:       models.Banner,
						AdSize:         "300x250",
						ServerSide:     1,
						Advertiser:     "fb.com",
						DealID:         "420",
						FloorValue:     4.4,
						FloorRuleValue: 2,
					},
				},
			},
			want: "https://t.pubmatic.com/wt?adv=fb.com&af=banner&aps=0&au=adunit&bc=AppNexus1&bidid=6521&di=420&dur=10&eg=4.3&en=2.5&fmv=test+version&frv=2&fskp=0&fsrc=1&ft=1&fv=4.4&iid=98765&kgpv=adunit%40300x250&orig=www.publisher.com&origbidid=6521&pdvid=1&pid=123&plt=1&pn=AppNexus&psz=300x250&pubid=12345&purl=www.abc.com&rwrd=1&sl=1&slot=1234_1234&ss=1&tgid=1&tst=0",
		},
		{
			name: "all_details_with_RewardInventory_in_tracker",
			args: args{
				rctx: models.RequestCtx{
					TrackerEndpoint: "//t.pubmatic.com/wt",
					Platform:        models.PLATFORM_APP,
				},
				tracker: models.Tracker{
					PubID:             12345,
					PageURL:           "www.abc.com",
					IID:               "98765",
					ProfileID:         "123",
					VersionID:         "1",
					SlotID:            "1234_1234",
					Adunit:            "adunit",
					Platform:          1,
					Origin:            "www.publisher.com",
					TestGroup:         1,
					AdPodSlot:         0,
					FloorSkippedFlag:  ptrutil.ToPtr(0),
					FloorModelVersion: "test version",
					FloorSource:       ptrutil.ToPtr(1),
					FloorType:         1,
					RewardedInventory: 1,
					PartnerInfo: models.Partner{
						PartnerID:      "AppNexus",
						BidderCode:     "AppNexus1",
						BidID:          "6521",
						OrigBidID:      "6521",
						GrossECPM:      4.3,
						NetECPM:        2.5,
						KGPV:           "adunit@300x250",
						AdDuration:     10,
						Adformat:       models.Banner,
						AdSize:         "300x250",
						ServerSide:     1,
						Advertiser:     "fb.com",
						DealID:         "420",
						FloorValue:     4.4,
						FloorRuleValue: 2,
					},
				},
			},
			want: "//t.pubmatic.com/wt?adv=fb.com&af=banner&aps=0&au=adunit&bc=AppNexus1&bidid=6521&di=420&dur=10&eg=4.3&en=2.5&fmv=test+version&frv=2&fskp=0&fsrc=1&ft=1&fv=4.4&iid=98765&kgpv=adunit%40300x250&orig=www.publisher.com&origbidid=6521&pdvid=1&pid=123&plt=1&pn=AppNexus&psz=300x250&pubid=12345&purl=www.abc.com&rwrd=1&sl=1&slot=1234_1234&ss=1&tgid=1&tst=0",
		},
		{
			name: "all_floors_details_in_tracker",
			args: args{
				rctx: models.RequestCtx{
					TrackerEndpoint: "//t.pubmatic.com/wt",
					Platform:        models.PLATFORM_APP,
				},
				tracker: models.Tracker{
					PubID:             12345,
					PageURL:           "www.abc.com",
					IID:               "98765",
					ProfileID:         "123",
					VersionID:         "1",
					SlotID:            "1234_1234",
					Adunit:            "adunit",
					Platform:          1,
					Origin:            "www.publisher.com",
					TestGroup:         1,
					AdPodSlot:         0,
					FloorSkippedFlag:  ptrutil.ToPtr(0),
					FloorModelVersion: "test version",
					FloorSource:       ptrutil.ToPtr(1),
					FloorType:         1,
					PartnerInfo: models.Partner{
						PartnerID:      "AppNexus",
						BidderCode:     "AppNexus1",
						BidID:          "6521",
						OrigBidID:      "6521",
						GrossECPM:      4.3,
						NetECPM:        2.5,
						KGPV:           "adunit@300x250",
						AdDuration:     10,
						Adformat:       models.Banner,
						AdSize:         "300x250",
						ServerSide:     1,
						Advertiser:     "fb.com",
						DealID:         "420",
						FloorValue:     4.4,
						FloorRuleValue: 2,
					},
				},
			},
			want: "//t.pubmatic.com/wt?adv=fb.com&af=banner&aps=0&au=adunit&bc=AppNexus1&bidid=6521&di=420&dur=10&eg=4.3&en=2.5&fmv=test+version&frv=2&fskp=0&fsrc=1&ft=1&fv=4.4&iid=98765&kgpv=adunit%40300x250&orig=www.publisher.com&origbidid=6521&pdvid=1&pid=123&plt=1&pn=AppNexus&psz=300x250&pubid=12345&purl=www.abc.com&sl=1&slot=1234_1234&ss=1&tgid=1&tst=0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConstructTrackerURL(tt.args.rctx, tt.args.tracker); got != tt.want {
				t.Errorf("ConstructTrackerURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstructVideoErrorURL(t *testing.T) {
	type args struct {
		rctx           models.RequestCtx
		errorURLString string
		bid            openrtb2.Bid
		tracker        models.Tracker
	}
	tests := []struct {
		name   string
		args   args
		want   string
		prefix string
	}{
		{
			name: "empty_urlString",
			args: args{
				rctx:           models.RequestCtx{},
				errorURLString: "",
				bid:            openrtb2.Bid{},
				tracker:        models.Tracker{},
			},
			want:   "",
			prefix: "",
		},
		{
			name: "invalid_urlString_with_parsing_error",
			args: args{
				rctx:           models.RequestCtx{},
				errorURLString: `:invalid_url`,
				bid:            openrtb2.Bid{},
				tracker:        models.Tracker{},
			},
			want:   "",
			prefix: "",
		},
		{
			name: "invalid_urlString_with_parsing",
			args: args{
				rctx:           models.RequestCtx{},
				errorURLString: `invalid_url`,
				bid:            openrtb2.Bid{},
				tracker:        models.Tracker{},
			},
			want:   "",
			prefix: "",
		},
		{
			name: "valid_video_errorUrl",
			args: args{
				rctx: models.RequestCtx{
					OriginCookie: "domain.com:8080",
				},
				errorURLString: `//t.pubmatic.com/wt`,
				bid:            openrtb2.Bid{},
				tracker: models.Tracker{
					PubID:     12345,
					PageURL:   "www.abc.com",
					IID:       "98765",
					ProfileID: "123",
					VersionID: "1",
					SlotID:    "1234_1234",
					Adunit:    "adunit",
					Platform:  1,
					Origin:    "www.publisher.com",
					TestGroup: 1,
					AdPodSlot: 0,
					SSAI:      "mediatailor",
					PartnerInfo: models.Partner{
						PartnerID:  "AppNexus",
						BidderCode: "AppNexus1",
						BidID:      "6521",
						OrigBidID:  "6521",
						GrossECPM:  4.3,
						NetECPM:    2.5,
						KGPV:       "adunit@300x250",
						AdDuration: 10,
						Adformat:   models.Banner,
						AdSize:     "300x250",
						ServerSide: 1,
						Advertiser: "fb.com",
						DealID:     "420",
					},
				},
			},
			want:   `https://t.pubmatic.com/wt?operId=8&crId=-1&p=12345&pid=123&pn=AppNexus&ts=0&v=1&ier=[ERRORCODE]&bc=AppNexus1&au=adunit&sURL=domain.com%253A8080&ssai=mediatailor&pfi=1&adv=fb.com`,
			prefix: `https://t.pubmatic.com/wt?operId=8`,
		},
		{
			name: "URL_with_Constant_Parameter",
			args: args{
				rctx: models.RequestCtx{
					OriginCookie: "domain.com:8080",
				},
				errorURLString: `//t.pubmatic.com/wt?p1=v1&p2=v2`,
				bid:            openrtb2.Bid{},
				tracker: models.Tracker{
					PubID:     12345,
					PageURL:   "www.abc.com",
					IID:       "98765",
					ProfileID: "123",
					VersionID: "1",
					SlotID:    "1234_1234",
					Adunit:    "adunit",
					Platform:  1,
					Origin:    "www.publisher.com",
					TestGroup: 1,
					AdPodSlot: 0,
					SSAI:      "mediatailor",
					PartnerInfo: models.Partner{
						PartnerID:  "AppNexus",
						BidderCode: "AppNexus1",
						BidID:      "6521",
						OrigBidID:  "6521",
						GrossECPM:  4.3,
						NetECPM:    2.5,
						KGPV:       "adunit@300x250",
						AdDuration: 10,
						Adformat:   models.Banner,
						AdSize:     "300x250",
						ServerSide: 1,
						Advertiser: "fb.com",
						DealID:     "420",
					},
				},
			},
			want:   `https://t.pubmatic.com/wt?operId=8&p1=v1&p2=v2&crId=-1&p=12345&pid=123&pn=AppNexus&ts=0&v=1&ier=[ERRORCODE]&bc=AppNexus1&au=adunit&sURL=domain.com%253A8080&ssai=mediatailor&pfi=1&adv=fb.com`,
			prefix: `https://t.pubmatic.com/wt?operId=8`,
		},
		{
			name: "Creative_ID_in_bid",
			args: args{
				rctx: models.RequestCtx{
					OriginCookie: "domain.com:8080",
				},
				errorURLString: `//t.pubmatic.com/wt`,
				bid: openrtb2.Bid{
					CrID: "cr123",
				},
				tracker: models.Tracker{
					PubID:     12345,
					PageURL:   "www.abc.com",
					IID:       "98765",
					ProfileID: "123",
					VersionID: "1",
					SlotID:    "1234_1234",
					Adunit:    "adunit",
					Platform:  1,
					Origin:    "www.publisher.com",
					TestGroup: 1,
					AdPodSlot: 0,
					SSAI:      "mediatailor",
					PartnerInfo: models.Partner{
						PartnerID:  "AppNexus",
						BidderCode: "AppNexus1",
						BidID:      "6521",
						OrigBidID:  "6521",
						GrossECPM:  4.3,
						NetECPM:    2.5,
						KGPV:       "adunit@300x250",
						AdDuration: 10,
						Adformat:   models.Banner,
						AdSize:     "300x250",
						ServerSide: 1,
						Advertiser: "fb.com",
						DealID:     "420",
					},
				},
			},
			want:   `https://t.pubmatic.com/wt?operId=8&crId=cr123&p=12345&pid=123&pn=AppNexus&ts=0&v=1&ier=[ERRORCODE]&bc=AppNexus1&au=adunit&sURL=domain.com%253A8080&ssai=mediatailor&pfi=1&adv=fb.com`,
			prefix: `https://t.pubmatic.com/wt?operId=8`,
		},
		{
			name: "URL_with_Schema",
			args: args{
				rctx: models.RequestCtx{
					OriginCookie: "com.myapp.test",
				},
				errorURLString: `http://t.pubmatic.com/wt?p1=v1&p2=v2`,
				bid:            openrtb2.Bid{},
				tracker: models.Tracker{
					PubID:     12345,
					PageURL:   "www.abc.com",
					IID:       "98765",
					ProfileID: "123",
					VersionID: "1",
					SlotID:    "1234_1234",
					Adunit:    "adunit",
					Platform:  1,
					Origin:    "www.publisher.com",
					TestGroup: 1,
					AdPodSlot: 0,
					SSAI:      "mediatailor",
					PartnerInfo: models.Partner{
						PartnerID:  "AppNexus",
						BidderCode: "AppNexus1",
						BidID:      "6521",
						OrigBidID:  "6521",
						GrossECPM:  4.3,
						NetECPM:    2.5,
						KGPV:       "adunit@300x250",
						AdDuration: 10,
						Adformat:   models.Banner,
						AdSize:     "300x250",
						ServerSide: 1,
						Advertiser: "fb.com",
						DealID:     "420",
					},
				},
			},
			want:   `https://t.pubmatic.com/wt?operId=8&p1=v1&p2=v2&crId=-1&p=12345&pid=123&pn=AppNexus&ts=0&v=1&ier=[ERRORCODE]&bc=AppNexus1&au=adunit&sURL=com.myapp.test&ssai=mediatailor&pfi=1&adv=fb.com`,
			prefix: `https://t.pubmatic.com/wt?operId=8`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constructedURL := ConstructVideoErrorURL(tt.args.rctx, tt.args.errorURLString, tt.args.bid, tt.args.tracker)
			if len(constructedURL) > 0 && len(tt.want) > 0 {
				wantURL, _ := url.Parse(constructedURL)
				expectedURL, _ := url.Parse(tt.want)
				if wantURL != nil && expectedURL != nil {
					assert.Equal(t, wantURL.Host, expectedURL.Host)
					assert.Equal(t, wantURL.Query(), expectedURL.Query())
					assert.Contains(t, constructedURL, tt.prefix)
				}
			}
		})
	}
}

func Test_getFloorsDetails(t *testing.T) {
	type args struct {
		bidResponseExt json.RawMessage
	}
	tests := []struct {
		name              string
		args              args
		skipfloors        *int
		floorType         int
		floorSource       *int
		floorModelVersion string
	}{
		{
			name: "invalid_responseExt",
			args: args{
				bidResponseExt: json.RawMessage(``),
			},
			skipfloors:  nil,
			floorSource: nil,
		},
		{
			name: "empty_responseExt",
			args: args{
				bidResponseExt: json.RawMessage(`{}`),
			},
			skipfloors:  nil,
			floorSource: nil,
		},
		{
			name: "empty_prebid_in_responseExt",
			args: args{
				bidResponseExt: json.RawMessage(`{"prebid":{}}`),
			},
			skipfloors:  nil,
			floorSource: nil,
		},
		{
			name: "empty_prebidfloors_in_responseExt",
			args: args{
				bidResponseExt: json.RawMessage(`{"prebid":{"floors":{}}}`),
			},
			skipfloors:  nil,
			floorSource: nil,
		},
		{
			name: "no_enforced_floors_data_in_responseExt",
			args: args{
				bidResponseExt: json.RawMessage(`{"prebid":{"floors":{"data":{},"location":"fetch"}}}`),
			},
			skipfloors:        nil,
			floorType:         models.SoftFloor,
			floorSource:       ptrutil.ToPtr(2),
			floorModelVersion: "",
		},
		{
			name: "no_modelsgroups_floors_data_in_responseExt",
			args: args{
				bidResponseExt: json.RawMessage(`{"prebid":{"floors":{"data":{},"location":"fetch","enforcement":{"enforcepbs":true}}}}`),
			},
			skipfloors:        nil,
			floorType:         models.HardFloor,
			floorSource:       ptrutil.ToPtr(2),
			floorModelVersion: "",
		},
		{
			name: "no_skipped_floors_data_in_responseExt",
			args: args{
				bidResponseExt: json.RawMessage(`{"prebid":{"floors":{"data":{"modelgroups":[{"modelversion":"version 1"}]},"location":"fetch","enforcement":{"enforcepbs":true}}}}`),
			},
			skipfloors:        nil,
			floorType:         models.HardFloor,
			floorSource:       ptrutil.ToPtr(2),
			floorModelVersion: "version 1",
		},
		{
			name: "all_floors_data_in_responseExt",
			args: args{
				bidResponseExt: json.RawMessage(`{"prebid":{"floors":{"skipped":true,"data":{"modelgroups":[{"modelversion":"version 1"}]},"location":"fetch","enforcement":{"enforcepbs":true}}}}`),
			},
			skipfloors:        ptrutil.ToPtr(1),
			floorType:         models.HardFloor,
			floorSource:       ptrutil.ToPtr(2),
			floorModelVersion: "version 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, got3 := getFloorsDetails(tt.args.bidResponseExt)
			assert.Equal(t, got, tt.skipfloors)
			if got1 != tt.floorType {
				t.Errorf("getFloorsDetails() got1 = %v, want %v", got1, tt.floorType)
			}
			assert.Equal(t, got2, tt.floorSource)
			if got3 != tt.floorModelVersion {
				t.Errorf("getFloorsDetails() got3 = %v, want %v", got3, tt.floorModelVersion)
			}
		})
	}
}
