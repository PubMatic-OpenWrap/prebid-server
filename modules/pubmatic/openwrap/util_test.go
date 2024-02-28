package openwrap

import (
	"net/http"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v19/adcom1"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/macros"
	mock_metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/usersync"
	"github.com/stretchr/testify/assert"
)

func TestRecordPublisherPartnerNoCookieStats(t *testing.T) {

	ctrl := gomock.NewController(t)
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
	defer ctrl.Finish()

	type args struct {
		rctx models.RequestCtx
	}

	tests := []struct {
		name           string
		args           args
		getHttpRequest func() *http.Request
		setup          func(*mock_metrics.MockMetricsEngine)
	}{
		{
			name: "Empty cookies and empty partner config map",
			args: args{
				rctx: models.RequestCtx{},
			},
			setup: func(mme *mock_metrics.MockMetricsEngine) {},
			getHttpRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://anyurl.com", nil)
				return req
			},
		},
		{
			name: "Empty cookie and non-empty partner config map",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						1: {
							models.SERVER_SIDE_FLAG:    "1",
							models.PREBID_PARTNER_NAME: "partner1",
							models.BidderCode:          "bidder1",
						},
					},
					PubIDStr: "5890",
				},
			},
			setup: func(mme *mock_metrics.MockMetricsEngine) {
				models.SyncerMap = make(map[string]usersync.Syncer)
				mme.EXPECT().RecordPublisherPartnerNoCookieStats("5890", "bidder1")
			},
			getHttpRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://anyurl.com", nil)
				return req
			},
		},
		{
			name: "only client side partner in config map",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						1: {
							models.SERVER_SIDE_FLAG:    "0",
							models.PREBID_PARTNER_NAME: "partner1",
							models.BidderCode:          "bidder1",
						},
					},
					PubIDStr: "5890",
				},
			},
			setup: func(mme *mock_metrics.MockMetricsEngine) {
				models.SyncerMap = make(map[string]usersync.Syncer)
			},
			getHttpRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://anyurl.com", nil)
				return req
			},
		},
		{
			name: "GetUID returns empty uid",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						1: {
							models.SERVER_SIDE_FLAG:    "1",
							models.PREBID_PARTNER_NAME: "pubmatic",
							models.BidderCode:          "pubmatic",
						},
					},
					PubIDStr: "5890",
				},
			},
			setup: func(mme *mock_metrics.MockMetricsEngine) {
				models.SyncerMap = map[string]usersync.Syncer{
					"pubmatic": fakeSyncer{
						key: "pubmatic",
					},
				}
				mme.EXPECT().RecordPublisherPartnerNoCookieStats("5890", "pubmatic")
			},
			getHttpRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://anyurl.com", nil)
				return req
			},
		},
		{
			name: "GetUID returns non empty uid",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						1: {
							models.SERVER_SIDE_FLAG:    "1",
							models.PREBID_PARTNER_NAME: "pubmatic",
							models.BidderCode:          "pubmatic",
						},
					},
					PubIDStr: "5890",
				},
			},
			setup: func(mme *mock_metrics.MockMetricsEngine) {
				models.SyncerMap = map[string]usersync.Syncer{
					"pubmatic": fakeSyncer{
						key: "pubmatic",
					},
				}
			},
			getHttpRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://anyurl.com", nil)

				cookie := &http.Cookie{
					Name:  "uids",
					Value: "ewoJInRlbXBVSURzIjogewoJCSJwdWJtYXRpYyI6IHsKCQkJInVpZCI6ICI3RDc1RDI1Ri1GQUM5LTQ0M0QtQjJEMS1CMTdGRUUxMUUwMjciLAoJCQkiZXhwaXJlcyI6ICIyMDIyLTEwLTMxVDA5OjE0OjI1LjczNzI1Njg5OVoiCgkJfQoJfSwKCSJiZGF5IjogIjIwMjItMDUtMTdUMDY6NDg6MzguMDE3OTg4MjA2WiIKfQ==",
				}
				req.AddCookie(cookie)
				return req
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(mockEngine)
			tc.args.rctx.MetricsEngine = mockEngine
			tc.args.rctx.ParsedUidCookie = usersync.ReadCookie(tc.getHttpRequest(), usersync.Base64Decoder{}, &config.HostCookie{})
			RecordPublisherPartnerNoCookieStats(tc.args.rctx)
		})
	}
}

// fakeSyncer implements syncer interface for unit test cases
type fakeSyncer struct {
	key string
}

func (s fakeSyncer) Key() string {
	return s.key
}

func (s fakeSyncer) DefaultSyncType() usersync.SyncType {
	return usersync.SyncType("")
}

func (s fakeSyncer) SupportsType(syncTypes []usersync.SyncType) bool {
	return false
}

func (fakeSyncer) GetSync([]usersync.SyncType, macros.UserSyncPrivacy) (usersync.Sync, error) {
	return usersync.Sync{}, nil
}

func TestGetDevicePlatform(t *testing.T) {
	type args struct {
		rCtx       models.RequestCtx
		bidRequest *openrtb2.BidRequest
	}
	tests := []struct {
		name string
		args args
		want models.DevicePlatform
	}{
		{
			name: "Test_empty_platform",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "",
					Platform: "",
				},
				bidRequest: nil,
			},
			want: models.DevicePlatformNotDefined,
		},
		{
			name: "Test_platform_amp",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "",
					Platform: "amp",
				},
				bidRequest: nil,
			},
			want: models.DevicePlatformMobileWeb,
		},
		{
			name: "Test_platform_in-app_with_iOS_UA",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1",
					Platform: "in-app",
				},
				bidRequest: nil,
			},
			want: models.DevicePlatformMobileAppIos,
		},
		{
			name: "Test_platform_in-app_with_Android_UA",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "Mozilla/5.0 (Linux; Android 7.0) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Focus/1.0 Chrome/59.0.3029.83 Mobile Safari/537.36",
					Platform: "in-app",
				},
				bidRequest: nil,
			},
			want: models.DevicePlatformMobileAppAndroid,
		},
		{
			name: "Test_platform_in-app_with_device.os_android",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "",
					Platform: "in-app",
				},
				bidRequest: getORTBRequest("android", "", 0, false, true),
			},
			want: models.DevicePlatformMobileAppAndroid,
		},
		{
			name: "Test_platform_in-app_with_device.os_ios",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1",
					Platform: "in-app",
				},
				bidRequest: getORTBRequest("ios", "", 0, false, true),
			},
			want: models.DevicePlatformMobileAppIos,
		},
		{
			name: "Test_platform_in-app_with_device.ua_for_ios",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1",
					Platform: "in-app",
				},
				bidRequest: getORTBRequest("", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1", 0, false, true),
			},
			want: models.DevicePlatformMobileAppIos,
		},
		{
			name: "Test_platform_display_with_device.deviceType_for_mobile",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "",
					Platform: "display",
				},
				bidRequest: getORTBRequest("", "", adcom1.DeviceMobile, false, true),
			},
			want: models.DevicePlatformMobileWeb,
		},
		{
			name: "Test_platform_display_with_device.deviceType_for_tablet",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "",
					Platform: "display",
				},
				bidRequest: getORTBRequest("", "", adcom1.DeviceMobile, false, true),
			},
			want: models.DevicePlatformMobileWeb,
		},
		{
			name: "Test_platform_display_with_device.deviceType_for_desktop",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "",
					Platform: "display",
				},
				bidRequest: getORTBRequest("", "", adcom1.DevicePC, true, false),
			},
			want: models.DevicePlatformDesktop,
		},
		{
			name: "Test_platform_display_with_device.ua_for_mobile",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1",
					Platform: "display",
				},
				bidRequest: getORTBRequest("", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1", 0, true, false),
			},
			want: models.DevicePlatformMobileWeb,
		},
		{
			name: "Test_platform_display_without_ua,_os_&_deviceType",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "",
					Platform: "display",
				},
				bidRequest: getORTBRequest("", "", 0, false, true),
			},
			want: models.DevicePlatformDesktop,
		},
		{
			name: "Test_platform_video_with_deviceType_as_CTV",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "",
					Platform: "video",
					PubIDStr: "5890",
				},
				bidRequest: getORTBRequest("", "", adcom1.DeviceTV, true, false),
			},
			want: models.DevicePlatformConnectedTv,
		},
		{
			name: "Test_platform_video_with_deviceType_as_connected_device",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "",
					Platform: "video",
					PubIDStr: "5890",
				},
				bidRequest: getORTBRequest("", "", adcom1.DeviceConnected, true, false),
			},
			want: models.DevicePlatformConnectedTv,
		},
		{
			name: "Test_platform_video_with_deviceType_as_set_top_box",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "",
					Platform: "video",
					PubIDStr: "5890",
				},
				bidRequest: getORTBRequest("", "", adcom1.DeviceSetTopBox, false, true),
			},
			want: models.DevicePlatformConnectedTv,
		},
		{
			name: "Test_platform_video_with_nil_values",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "",
					Platform: "video",
				},
				bidRequest: getORTBRequest("", "", 0, true, false),
			},
			want: models.DevicePlatformDesktop,
		},
		{
			name: "Test_platform_video_with_site_entry",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "",
					Platform: "video",
				},
				bidRequest: getORTBRequest("", "", 0, true, false),
			},
			want: models.DevicePlatformDesktop,
		},
		{
			name: "Test_platform_video_with_site_entry_and_mobile_UA",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1",
					Platform: "video",
				},
				bidRequest: getORTBRequest("", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1", 0, true, false),
			},
			want: models.DevicePlatformMobileWeb,
		},
		{
			name: "Test_platform_video_with_app_entry_and_iOS_mobile_UA",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1",
					Platform: "video",
				},
				bidRequest: getORTBRequest("", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1", 0, false, true),
			},
			want: models.DevicePlatformMobileAppIos,
		},
		{
			name: "Test_platform_video_with_app_entry_and_android_mobile_UA",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "Mozilla/5.0 (Linux; Android 7.0) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Focus/1.0 Chrome/59.0.3029.83 Mobile Safari/537.36",
					Platform: "video",
				},

				bidRequest: getORTBRequest("", "Mozilla/5.0 (Linux; Android 7.0) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Focus/1.0 Chrome/59.0.3029.83 Mobile Safari/537.36", 0, false, true),
			},
			want: models.DevicePlatformMobileAppAndroid,
		},
		{
			name: "Test_platform_video_with_app_entry_and_android_os",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "",
					Platform: "video",
				},
				bidRequest: getORTBRequest("android", "", 0, false, true),
			},
			want: models.DevicePlatformMobileAppAndroid,
		},
		{
			name: "Test_platform_video_with_app_entry_and_ios_os",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "",
					Platform: "video",
				},
				bidRequest: getORTBRequest("ios", "", 0, false, true),
			},
			want: models.DevicePlatformMobileAppIos,
		},
		{
			name: "Test_platform_video_with_CTV_and_device_type",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "Mozilla/5.0 (SMART-TV; Linux; Tizen 4.0) AppleWebKit/538.1 (KHTML, like Gecko) Version/4.0 TV Safari/538.1",
					Platform: "video",
					PubIDStr: "5890",
				},
				bidRequest: getORTBRequest("", "Mozilla/5.0 (SMART-TV; Linux; Tizen 4.0) AppleWebKit/538.1 (KHTML, like Gecko) Version/4.0 TV Safari/538.1", 3, false, true),
			},
			want: models.DevicePlatformConnectedTv,
		},
		{
			name: "Test_platform_video_with_CTV_and_no_device_type",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "AppleCoreMedia/1.0.0.20L498 (Apple TV; U; CPU OS 16_4_1 like Mac OS X; en_us)",
					Platform: "video",
				},
				bidRequest: getORTBRequest("", "AppleCoreMedia/1.0.0.20L498 (Apple TV; U; CPU OS 16_4_1 like Mac OS X; en_us)", 0, true, false),
			},
			want: models.DevicePlatformConnectedTv,
		},
		{
			name: "Test_platform_video_for_non_CTV_User_agent_with_device_type_7",
			args: args{
				rCtx: models.RequestCtx{
					UA:       "AppleCoreMedia/1.0.0.20L498 (iphone ; U; CPU OS 16_4_1 like Mac OS X; en_us)",
					Platform: "video",
					PubIDStr: "5890",
				},
				bidRequest: getORTBRequest("", "AppleCoreMedia/1.0.0.20L498 (iphone ; U; CPU OS 16_4_1 like Mac OS X; en_us)", 7, false, true),
			},
			want: models.DevicePlatformConnectedTv,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDevicePlatform(tt.args.rCtx, tt.args.bidRequest)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsMobile(t *testing.T) {
	type args struct {
		deviceType      adcom1.DeviceType
		userAgentString string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test_for_deviceType_1",
			args: args{
				deviceType:      adcom1.DeviceMobile,
				userAgentString: "",
			},
			want: true,
		},
		{
			name: "Test_for_deviceType_2",
			args: args{
				deviceType:      adcom1.DevicePC,
				userAgentString: "",
			},
			want: false,
		},
		{
			name: "Test_for_deviceType_4:_phone",
			args: args{
				deviceType:      adcom1.DevicePhone,
				userAgentString: "",
			},
			want: true,
		},
		{
			name: "Test_for_deviceType_5:_tablet",
			args: args{
				deviceType:      adcom1.DeviceTablet,
				userAgentString: "",
			},
			want: true,
		},
		{
			name: "Test_for_iPad_User-Agent",
			args: args{
				deviceType:      0,
				userAgentString: "Mozilla/5.0 (iPad; CPU OS 10_3_3 like Mac OS X) AppleWebKit/603.3.8 (KHTML, like Gecko) Mobile/14G60",
			},
			want: true,
		},
		{
			name: "Test_for_iPhone_User-Agent",
			args: args{
				deviceType:      0,
				userAgentString: "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1",
			},
			want: true,
		},
		{
			name: "Test_for_Safari_web-browser_on_mobile_User-Agent",
			args: args{
				deviceType:      0,
				userAgentString: "MobileSafari/602.1 CFNetwork/811.5.4 Darwin/16.7.0",
			},
			want: true,
		},
		{
			name: "Test_for_Outlook_3_application_on_mobile_phone_User-Agent",
			args: args{
				deviceType:      0,
				userAgentString: "Outlook-iOS/709.2144270.prod.iphone (3.23.0)",
			},
			want: true,
		},
		{
			name: "Test_for_firefox_11_tablet_User-Agent",
			args: args{
				deviceType:      0,
				userAgentString: "Mozilla/5.0 (Android 4.4; Tablet; rv:41.0) Gecko/41.0 Firefox/41.0",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isMobile(tt.args.deviceType, tt.args.userAgentString)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_getUserAgent(t *testing.T) {
	type args struct {
		request   *openrtb2.BidRequest
		defaultUA string
	}
	tests := []struct {
		name   string
		args   args
		wantUA string
	}{
		{
			name: "request_is_nil",
			args: args{
				request:   nil,
				defaultUA: "default-ua",
			},

			wantUA: "default-ua",
		},
		{
			name: "req.device_is_nil",
			args: args{
				request: &openrtb2.BidRequest{
					Device: nil,
				},
				defaultUA: "default-ua",
			},
			wantUA: "default-ua",
		},
		{
			name: "req.device.ua_empty",
			args: args{
				request: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						UA: "",
					},
				},
				defaultUA: "default-ua",
			},
			wantUA: "default-ua",
		},
		{
			name: "req.device.ua_valid",
			args: args{
				request: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						UA: "Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36",
					},
				},
				defaultUA: "default-ua",
			},
			wantUA: "Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ua := getUserAgent(tt.args.request, tt.args.defaultUA)
			assert.Equal(t, tt.wantUA, ua, "mismatched UA")
		})
	}
}

func TestIsIos(t *testing.T) {
	type args struct {
		os              string
		userAgentString string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "iOS_test_for_Web_Browser_Mobile-Tablet",
			args: args{
				os:              "",
				userAgentString: "Mozilla/5.0 (iPad; CPU OS 10_3_3 like Mac OS X) AppleWebKit/603.3.8 (KHTML, like Gecko) Mobile/14G60",
			},
			want: true,
		},
		{
			name: "iOS_test_for_Safari_13_Mobile-Phone",
			args: args{
				os:              "",
				userAgentString: "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1",
			},
			want: true,
		},
		{
			name: "Test_for_Safari_web-browser_on_mobile_User-Agent",
			args: args{
				os:              "",
				userAgentString: "MobileSafari/602.1 CFNetwork/811.5.4 Darwin/16.7.0",
			},
			want: true,
		},
		{
			name: "Test_for_iPhone_XR_simulator_User-Agent",
			args: args{
				os:              "",
				userAgentString: "Mozilla/5.0 (iPhone; CPU iPhone OS 12_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) FxiOS/7.0.4 Mobile/16B91 Safari/605.1.15",
			},
			want: true,
		},
		{
			name: "Test_for_Outlook_3_Application_User-Agent",
			args: args{
				os:              "",
				userAgentString: "Outlook-iOS/709.2144270.prod.iphone (3.23.0)",
			},
			want: true,
		},
		{
			name: "iOS_test_for_Safari_12_Mobile-Phone",
			args: args{
				os:              "",
				userAgentString: "Mozilla/5.0 (iPhone; CPU iPhone OS 12_1_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.0 Mobile/15E148 Safari/604.1",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isIos(tt.args.os, tt.args.userAgentString)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsAndroid(t *testing.T) {
	type args struct {
		os              string
		userAgentString string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test_android_with_correct_os_value",
			args: args{
				os:              "android",
				userAgentString: "",
			},
			want: true,
		},
		{
			name: "Test_android_with_invalid_os_value",
			args: args{
				os:              "ios",
				userAgentString: "",
			},
			want: false,
		},
		{
			name: "Test_android_with_invalid_osv_alue",
			args: args{
				os:              "",
				userAgentString: "",
			},
			want: false,
		},
		{
			name: "Test_android_with_UA_value",
			args: args{
				os:              "",
				userAgentString: "Mozilla/5.0 (Linux; Android 7.0) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Focus/1.0 Chrome/59.0.3029.83 Mobile Safari/537.36",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAndroid(tt.args.os, tt.args.userAgentString)
			assert.Equal(t, tt.want, got)
		})
	}
}

func getORTBRequest(os, ua string, deviceType adcom1.DeviceType, withSite, withApp bool) *openrtb2.BidRequest {
	request := new(openrtb2.BidRequest)

	if withSite {
		request.Site = &openrtb2.Site{
			Publisher: &openrtb2.Publisher{
				ID: "1010",
			},
		}
	}

	if withApp {
		request.App = &openrtb2.App{
			Publisher: &openrtb2.Publisher{
				ID: "1010",
			},
		}
	}

	request.Device = new(openrtb2.Device)
	request.Device.UA = ua

	request.Device.OS = os

	request.Device.DeviceType = deviceType

	return request
}

func TestGetSourceAndOrigin(t *testing.T) {
	type args struct {
		bidRequest *openrtb2.BidRequest
	}
	type want struct {
		source string
		origin string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "bidRequest_site_conatins_Domain",
			args: args{
				bidRequest: &openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Page:   "http://www.test.com",
						Domain: "test.com",
					},
				},
			},
			want: want{
				source: "test.com",
				origin: "test.com",
			},
		},
		{
			name: "bidRequest_conatins_Site_Page",
			args: args{
				bidRequest: &openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Page: "http://www.test.com",
					},
				},
			},
			want: want{
				source: "www.test.com",
				origin: "www.test.com",
			},
		},
		{
			name: "bidRequest_conatins_App_Bundle",
			args: args{
				bidRequest: &openrtb2.BidRequest{
					App: &openrtb2.App{
						Bundle: "com.pub.test",
					},
				},
			},
			want: want{
				source: "com.pub.test",
				origin: "com.pub.test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, origin := getSourceAndOrigin(tt.args.bidRequest)
			assert.Equal(t, tt.want.source, source)
			assert.Equal(t, tt.want.origin, origin)
		})
	}
}

func TestGetHostName(t *testing.T) {
	var (
		node string
		pod  string
	)

	saveEnvVarsForServerName := func() {
		node, _ = os.LookupEnv(models.ENV_VAR_NODE_NAME)
		pod, _ = os.LookupEnv(models.ENV_VAR_POD_NAME)
	}

	resetEnvVarsForServerName := func() {
		os.Setenv(models.ENV_VAR_NODE_NAME, node)
		os.Setenv(models.ENV_VAR_POD_NAME, pod)
	}
	type args struct {
		nodeName string
		podName  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "default_value",
			args: args{},
			want: models.DEFAULT_NODENAME + ":" + models.DEFAULT_PODNAME,
		},
		{
			name: "valid_name",
			args: args{
				nodeName: "sfo2hyp084.sfo2.pubmatic.com",
				podName:  "ssheaderbidding-0-0-38-pr-26-2-k8s-5679748b7b-tqh42",
			},
			want: "sfo2hyp084:0-0-38-pr-26-2-k8s-5679748b7b-tqh42",
		},
		{
			name: "special_characters",
			args: args{
				nodeName: "sfo2hyp084.sfo2.pubmatic.com!!!@#$-_^%x090",
				podName:  "ssheaderbidding-0-0-38-pr-26-2-k8s-5679748b7b-tqh42",
			},
			want: "sfo2hyp084:0-0-38-pr-26-2-k8s-5679748b7b-tqh42",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saveEnvVarsForServerName()

			if len(tt.args.nodeName) > 0 {
				os.Setenv(models.ENV_VAR_NODE_NAME, tt.args.nodeName)
			}

			if len(tt.args.podName) > 0 {
				os.Setenv(models.ENV_VAR_POD_NAME, tt.args.podName)
			}

			got := GetHostName()
			assert.Equal(t, tt.want, got)

			resetEnvVarsForServerName()
		})
	}
}

func TestGetPubmaticErrorCode(t *testing.T) {
	type args struct {
		standardNBR openrtb3.NoBidReason
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "ErrMissingPublisherID",
			args: args{
				standardNBR: nbr.InvalidPublisherID,
			},
			want: 604,
		},
		{
			name: "ErrBadRequest",
			args: args{
				standardNBR: nbr.InvalidRequestExt,
			},
			want: 18,
		},
		{
			name: "ErrMissingProfileID",
			args: args{
				standardNBR: nbr.InvalidProfileID,
			},
			want: 700,
		},
		{
			name: "ErrAllPartnerThrottled",
			args: args{
				standardNBR: nbr.AllPartnerThrottled,
			},
			want: 11,
		},
		{
			name: "ErrPrebidInvalidCustomPriceGranularity",
			args: args{
				standardNBR: nbr.InvalidPriceGranularityConfig,
			},
			want: 26,
		},
		{
			name: "ErrMissingTagID",
			args: args{
				standardNBR: nbr.InvalidImpressionTagID,
			},
			want: 605,
		},
		{
			name: "ErrInvalidConfiguration",
			args: args{
				standardNBR: nbr.InvalidProfileConfiguration,
			},
			want: 6,
		},
		{
			name: "ErrInvalidConfiguration_platform",
			args: args{
				standardNBR: nbr.InvalidPlatform,
			},
			want: 6,
		},
		{
			name: "ErrInvalidConfiguration_AllSlotsDisabled",
			args: args{
				standardNBR: nbr.AllSlotsDisabled,
			},
			want: 6,
		},
		{
			name: "ErrInvalidConfiguration_ServerSidePartnerNotConfigured",
			args: args{
				standardNBR: nbr.ServerSidePartnerNotConfigured,
			},
			want: 6,
		},
		{
			name: "ErrInvalidImpression",
			args: args{
				standardNBR: nbr.InternalError,
			},
			want: 17,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPubmaticErrorCode(tt.args.standardNBR)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_getIP(t *testing.T) {
	type args struct {
		bidRequest *openrtb2.BidRequest
		defaultIP  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test_empty_Device_IP",
			args: args{
				bidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						IP: "",
					},
				},
				defaultIP: "10.20.30.40",
			},
			want: "10.20.30.40",
		},
		{
			name: "Test_valid_Device_IP",
			args: args{
				bidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						IP: "10.20.30.40",
					},
				},
				defaultIP: "",
			},
			want: "10.20.30.40",
		},
		{
			name: "Test_valid_Device_IP_with_default_IP",
			args: args{
				bidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						IP: "10.20.30.40",
					},
				},
				defaultIP: "20.30.40.50",
			},
			want: "10.20.30.40",
		},
		{
			name: "Test_empty_Device_IP_with_default_IP",
			args: args{
				bidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						IP: "",
					},
				},
				defaultIP: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getIP(tt.args.bidRequest, tt.args.defaultIP); got != tt.want {
				t.Errorf("getIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
