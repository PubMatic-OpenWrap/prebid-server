package openwrap

import (
	"reflect"
	"testing"

	"github.com/PubMatic-OpenWrap/prebid-server/util/ptrutil"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

func Test_getPageURL(t *testing.T) {
	type args struct {
		bidRequest *openrtb2.BidRequest
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "App_storeurl_is_not_empty",
			args: args{
				bidRequest: &openrtb2.BidRequest{
					App: &openrtb2.App{
						StoreURL: "testurlApp",
					},
				},
			},
			want: "testurlApp",
		},
		{
			name: "Site_page_is_not_empty",
			args: args{
				bidRequest: &openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Page: "testurlSite",
					},
				},
			},
			want: "testurlSite",
		},
		{
			name: "both_app_and_site_are_nil",
			args: args{
				bidRequest: &openrtb2.BidRequest{},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPageURL(tt.args.bidRequest); got != tt.want {
				t.Errorf("getPageURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getVASTEventMacros(t *testing.T) {
	type args struct {
		rctx models.RequestCtx
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "SSAI_is_empty",
			args: args{
				rctx: models.RequestCtx{
					ProfileID:          1234,
					DisplayID:          1234,
					StartTime:          1234,
					DevicePlatform:     1234,
					LoggerImpressionID: "1234",
					SSAI:               "",
				},
			},
			want: map[string]string{
				"[PROFILE_ID]":            "1234",
				"[PROFILE_VERSION]":       "1234",
				"[UNIX_TIMESTAMP]":        "1234",
				"[PLATFORM]":              "1234",
				"[WRAPPER_IMPRESSION_ID]": "1234",
			},
		},
		{
			name: "SSAI_is_not_empty",
			args: args{
				rctx: models.RequestCtx{
					ProfileID:          1234,
					DisplayID:          1234,
					StartTime:          1234,
					DevicePlatform:     1234,
					LoggerImpressionID: "1234",
					SSAI:               "1234",
				},
			},
			want: map[string]string{
				"[PROFILE_ID]":            "1234",
				"[PROFILE_VERSION]":       "1234",
				"[UNIX_TIMESTAMP]":        "1234",
				"[PLATFORM]":              "1234",
				"[WRAPPER_IMPRESSION_ID]": "1234",
				"[SSAI]":                  "1234",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getVASTEventMacros(tt.args.rctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getVASTEventMacros() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_updateAliasGVLIds(t *testing.T) {
	type args struct {
		aliasgvlids   map[string]uint16
		bidderCode    string
		partnerConfig map[string]string
	}
	type want struct {
		aliasgvlids map[string]uint16
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "vendorId not present in config",
			args: args{
				aliasgvlids:   map[string]uint16{},
				bidderCode:    "vastbidder1",
				partnerConfig: map[string]string{},
			},
			want: want{
				aliasgvlids: map[string]uint16{},
			},
		},
		{
			name: "Empty vendorID",
			args: args{
				aliasgvlids:   map[string]uint16{},
				bidderCode:    "vastbidder1",
				partnerConfig: map[string]string{models.VENDORID: ""},
			},
			want: want{
				aliasgvlids: map[string]uint16{},
			},
		},
		{
			name: "Error parsing vendorID",
			args: args{
				aliasgvlids:   map[string]uint16{},
				bidderCode:    "vastbidder1",
				partnerConfig: map[string]string{models.VENDORID: "abc"},
			},
		},
		{
			name: "VendorID is 0",
			args: args{
				aliasgvlids:   map[string]uint16{},
				bidderCode:    "vastbidder1",
				partnerConfig: map[string]string{models.VENDORID: "0"},
			},
			want: want{
				aliasgvlids: map[string]uint16{},
			},
		},
		{
			name: "Negative vendorID",
			args: args{
				aliasgvlids:   map[string]uint16{},
				bidderCode:    "vastbidder1",
				partnerConfig: map[string]string{models.VENDORID: "-76"},
			},
		},
		{
			name: "Valid vendorID",
			args: args{
				aliasgvlids:   map[string]uint16{},
				bidderCode:    "vastbidder1",
				partnerConfig: map[string]string{models.VENDORID: "76"},
			},
			want: want{
				aliasgvlids: map[string]uint16{"vastbidder1": uint16(76)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateAliasGVLIds(tt.args.aliasgvlids, tt.args.bidderCode, tt.args.partnerConfig)
		})
	}
}

func TestOpenWrap_setTimeout(t *testing.T) {
	type fields struct {
		cfg          config.Config
		cache        cache.Cache
		metricEngine metrics.MetricsEngine
	}
	type args struct {
		rCtx models.RequestCtx
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		{
			name: "ssTimeout_greater_than_minTimeout_and_less_than_maxTimeout",
			args: args{
				rCtx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							"ssTimeout": "250",
						},
					},
				},
			},
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						MinTimeout: 200,
						MaxTimeout: 300,
					},
				},
			},
			want: 250,
		},
		{
			name: "ssTimeout_less_than_minTimeout",
			args: args{
				rCtx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							"ssTimeout": "250",
						},
					},
				},
			},
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						MinTimeout: 300,
						MaxTimeout: 400,
					},
				},
			},
			want: 300,
		},
		{
			name: "ssTimeout_greater_than_minTimeout_and_also_greater_than_maxTimeout",
			args: args{
				rCtx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							"ssTimeout": "500",
						},
					},
				},
			},
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						MinTimeout: 300,
						MaxTimeout: 400,
					},
				},
			},
			want: 400,
		},
		{
			name: "ssTimeout_greater_than_minTimeout_and_less_than_maxTimeout",
			args: args{
				rCtx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							"ssTimeout": "400",
						},
					},
				},
			},
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						MinTimeout: 300,
						MaxTimeout: 500,
					},
				},
			},
			want: 400,
		},
		//Below piece of code is applicable for older profiles where ssTimeout is not set
		//Here we will check the partner timeout and select max timeout considering timeout range
		{
			name: "at_lease_one_partner_timeout_greater_than_cofig_maxTimeout",
			args: args{
				rCtx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						1: {
							"timeout": "500",
						},
						2: {
							"timeout": "250",
						},
					},
				},
			},
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						MinTimeout: 200,
						MaxTimeout: 300,
					},
				},
			},
			want: 300,
		},
		{
			name: "all_partner_timeout_less_than_cofig_maxTimeout",
			args: args{
				rCtx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						1: {
							"timeout": "230",
						},
						2: {
							"timeout": "250",
						},
						3: {
							"timeout": "280",
						},
					},
				},
			},
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						MinTimeout: 200,
						MaxTimeout: 300,
					},
				},
			},
			want: 280,
		},
		{
			name: "all_partner_timeout_less_than_cofig_minTimeout",
			args: args{
				rCtx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						1: {
							"timeout": "100",
						},
						2: {
							"timeout": "150",
						},
						3: {
							"timeout": "180",
						},
					},
				},
			},
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						MinTimeout: 200,
						MaxTimeout: 300,
					},
				},
			},
			want: 200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := OpenWrap{
				cfg:          tt.fields.cfg,
				cache:        tt.fields.cache,
				metricEngine: tt.fields.metricEngine,
			}
			if got := m.setTimeout(tt.args.rCtx); got != tt.want {
				t.Errorf("OpenWrap.setTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isSendAllBids(t *testing.T) {
	type args struct {
		rctx models.RequestCtx
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Don't_do_ssauction",
			args: args{
				rctx: models.RequestCtx{
					SSAuction: 0,
				},
			},
			want: true,
		},
		{
			name: "SSAuction_flag_not_set_In-app_sendAllbids_flag_1",
			args: args{
				rctx: models.RequestCtx{
					SSAuction: -1,
					Platform:  models.PLATFORM_APP,
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							"sendAllBids": "1",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "SSAuction_flag_not_set_In-app_sendAllbids_flag_other_than_1",
			args: args{
				rctx: models.RequestCtx{
					SSAuction: -1,
					Platform:  models.PLATFORM_APP,
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							"sendAllBids": "5",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "Random_value_of_ssauctionflag",
			args: args{
				rctx: models.RequestCtx{
					SSAuction: 5,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSendAllBids(tt.args.rctx); got != tt.want {
				t.Errorf("isSendAllBids() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getValidLanguage(t *testing.T) {
	type args struct {
		language string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Language_of_length_less_than_2",
			args: args{
				language: "te",
			},
			want: "te",
		},
		{
			name: "Language_of_length_greater_than_2_and_it_is_valid_code",
			args: args{
				language: "hindi",
			},
			want: "hi",
		},
		{
			name: "Language_of_length_greater_than_2_and_it_is_Invalid_code",
			args: args{
				language: "xyz",
			},
			want: "xyz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getValidLanguage(tt.args.language); got != tt.want {
				t.Errorf("getValidLanguage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isSlotEnabled(t *testing.T) {
	type args struct {
		videoAdUnitCtx  models.AdUnitCtx
		bannerAdUnitCtx models.AdUnitCtx
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Video_enabled_in_Video_adunit_context",
			args: args{
				videoAdUnitCtx: models.AdUnitCtx{
					AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
						Video: &adunitconfig.Video{
							Enabled: ptrutil.ToPtr(true),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Banner_enabled_in_banner_adunit_context",
			args: args{
				bannerAdUnitCtx: models.AdUnitCtx{
					AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
						Banner: &adunitconfig.Banner{
							Enabled: ptrutil.ToPtr(true),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "both_banner_and_video_enabled_in_adunit_context",
			args: args{
				bannerAdUnitCtx: models.AdUnitCtx{
					AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
						Banner: &adunitconfig.Banner{
							Enabled: ptrutil.ToPtr(true),
						},
					},
				},
				videoAdUnitCtx: models.AdUnitCtx{
					AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
						Video: &adunitconfig.Video{
							Enabled: ptrutil.ToPtr(true),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "both_banner_and_video_disabled_in_adunit_context",
			args: args{
				bannerAdUnitCtx: models.AdUnitCtx{
					AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
						Banner: &adunitconfig.Banner{
							Enabled: ptrutil.ToPtr(false),
						},
					},
				},
				videoAdUnitCtx: models.AdUnitCtx{
					AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
						Video: &adunitconfig.Video{
							Enabled: ptrutil.ToPtr(false),
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSlotEnabled(tt.args.videoAdUnitCtx, tt.args.bannerAdUnitCtx); got != tt.want {
				t.Errorf("isSlotEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPubID(t *testing.T) {
	type args struct {
		bidRequest openrtb2.BidRequest
	}
	type want struct {
		wantErr bool
		pubID   int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "publisher_id_present_in_site_object_and_it_is_valid_integer",
			args: args{
				bidRequest: openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{
							ID: "5890",
						},
					},
				},
			},
			want: want{
				wantErr: false,
				pubID:   5890,
			},
		},
		{
			name: "publisher_id_present_in_site_object_but_it_is_not_valid_integer",
			args: args{
				bidRequest: openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{
							ID: "test",
						},
					},
				},
			},
			want: want{
				wantErr: true,
				pubID:   0,
			},
		},
		{
			name: "publisher_id_present_in_App_object_and_it_is_valid_integer",
			args: args{
				bidRequest: openrtb2.BidRequest{
					App: &openrtb2.App{
						Publisher: &openrtb2.Publisher{
							ID: "5890",
						},
					},
				},
			},
			want: want{
				wantErr: false,
				pubID:   5890,
			},
		},
		{
			name: "publisher_id_present_in_App_object_but_it_is_not_valid_integer",
			args: args{
				bidRequest: openrtb2.BidRequest{
					App: &openrtb2.App{
						Publisher: &openrtb2.Publisher{
							ID: "test",
						},
					},
				},
			},
			want: want{
				wantErr: true,
				pubID:   0,
			},
		},
		{
			name: "publisher_id_present_in_both_Site_and_App_object",
			args: args{
				bidRequest: openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{
							ID: "5800",
						},
					},
					App: &openrtb2.App{
						Publisher: &openrtb2.Publisher{
							ID: "5890",
						},
					},
				},
			},
			want: want{
				wantErr: false,
				pubID:   5800,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPubID(tt.args.bidRequest)
			if (err != nil) != tt.want.wantErr {
				t.Errorf("getPubID() error = %v, wantErr %v", err, tt.want.wantErr)
				return
			}
			if got != tt.want.pubID {
				t.Errorf("getPubID() = %v, want %v", got, tt.want)
			}
		})
	}
}
