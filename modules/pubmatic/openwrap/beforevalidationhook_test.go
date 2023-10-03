package openwrap

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"git.pubmatic.com/PubMatic/go-common/util"
	mock_cache "github.com/PubMatic-OpenWrap/prebid-server/modules/pubmatic/openwrap/cache/mock"
	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v19/adcom1"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func formVideoObject() *openrtb2.Video {
	video := new(openrtb2.Video)
	video.MIMEs = []string{"video/mp4", "video/mpeg"}
	video.W = 640
	video.H = 480
	return video
}

func formORtbV25Request(formatFlag bool, videoFlag bool) *openrtb2.BidRequest {
	request := new(openrtb2.BidRequest)
	banner := new(openrtb2.Banner)
	if formatFlag == true {
		formatObj1 := openrtb2.Format{
			W: 728,
			H: 90,
		}

		formatObj2 := openrtb2.Format{
			W: 300,
			H: 250,
		} // openrtb2.Format{728, 90, nil}

		formatArray := []openrtb2.Format{formatObj1, formatObj2}
		banner.Format = formatArray

		banner.W = new(int64)
		*banner.W = 700
		banner.H = new(int64)
		*banner.H = 900

	} else {
		banner.W = new(int64)
		*banner.W = 728
		banner.H = new(int64)
		*banner.H = 90
	}

	var imp openrtb2.Imp
	if videoFlag == true {
		video := formVideoObject()
		imp.Video = video
	}

	imp.ID = "123"
	imp.Banner = banner
	imp.TagID = "adunit"

	impWrapExt := new(models.ExtImpWrapper)
	impWrapExt.Div = "div"

	inImpExt := new(models.ImpExtension)
	inImpExt.Wrapper = impWrapExt

	bidderExt := map[string]*models.BidderExtension{
		models.BidderPubMatic: &models.BidderExtension{
			KeyWords: []models.KeyVal{
				{
					Key:    "pmzoneid",
					Values: []string{"val1", "val2"},
				},
			},
		},
	}
	inImpExt.Bidder = bidderExt
	impExt, _ := json.Marshal(inImpExt)
	imp.Ext = impExt
	impArr := make([]openrtb2.Imp, 0)
	impArr = append(impArr, imp)
	request.ID = "123-456-789"
	request.Imp = impArr

	len := 2
	request.WSeat = make([]string, len)
	for i := 0; i < len; i++ {
		request.WSeat[i] = fmt.Sprintf("Wseat_%d", i)
	}

	request.Cur = make([]string, len)
	for i := 0; i < len; i++ {
		request.Cur[i] = fmt.Sprintf("cur_%d", i)
	}

	request.BAdv = make([]string, len)
	for i := 0; i < len; i++ {
		request.BAdv[i] = fmt.Sprintf("badv_%d", i)
	}

	request.BApp = make([]string, len)
	for i := 0; i < len; i++ {
		request.BApp[i] = fmt.Sprintf("bapp_%d", i)
	}

	request.BCat = make([]string, len)
	for i := 0; i < len; i++ {
		request.BCat[i] = fmt.Sprintf("bcat_%d", i)
	}

	request.WLang = make([]string, len)
	for i := 0; i < len; i++ {
		request.WLang[i] = fmt.Sprintf("Wlang_%d", i)
	}

	request.BSeat = make([]string, len)
	for i := 0; i < len; i++ {
		request.BSeat[i] = fmt.Sprintf("Bseat_%d", i)
	}

	site := new(openrtb2.Site)
	publisher := new(openrtb2.Publisher)
	publisher.ID = "5890"
	site.Publisher = publisher
	site.Page = "www.test.com"

	site.Domain = "test.com"

	request.Site = site

	request.Device = new(openrtb2.Device)
	request.Device.IP = "123.145.167.10"
	request.Device.UA = "Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36"

	request.User = new(openrtb2.User)
	request.User.ID = "119208432"

	request.User.BuyerUID = "1rwe432"

	request.User.Yob = 1980

	request.User.Gender = "F"

	request.User.Geo = new(openrtb2.Geo)
	request.User.Geo.Country = "US"

	request.User.Geo.Region = "CA"

	request.User.Geo.Metro = "90001"

	request.User.Geo.City = "Alamo"

	request.Source = new(openrtb2.Source)
	request.Source.Ext = json.RawMessage(`{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}`)

	wExt := new(models.RequestExt)
	dmExt := new(models.RequestExtWrapper)
	dmExt.ProfileId = 123
	dmExt.VersionId = 1
	dmExt.LoggerImpressionID = "test_display_wiid"
	wExt.Wrapper = dmExt

	reqExt, _ := json.Marshal(wExt)
	request.Ext = reqExt

	request.Test = 0
	return request

}

func getExpectedORTBV25Request(test bool, owAPIBidReq openrtb2.BidRequest, prebidPlatform string) *openrtb2.BidRequest {
	request := new(openrtb2.BidRequest)

	request.TMax = 300

	formatObj1 := openrtb2.Format{
		W: 728,
		H: 90,
	}

	formatObj2 := openrtb2.Format{
		W: 300,
		H: 250,
	} // openrtb2.Format{728, 90, nil}

	formatArray := []openrtb2.Format{formatObj1, formatObj2}

	banner := new(openrtb2.Banner)
	banner.Format = formatArray

	var imp openrtb2.Imp
	imp.ID = "123"
	imp.Banner = banner
	imp.TagID = "adunit"

	imp.Banner.W = new(int64)
	*imp.Banner.W = 700
	imp.Banner.H = new(int64)
	*imp.Banner.H = 900

	// OTT-18 - adding deal tiers
	impExt := `{"prebid":{"bidder":{"pubmatic":{"publisherId":"5890","adSlot":"adunit@728x90","wrapper":{"version":1,"profile":123},"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}}}}`
	imp.Ext = json.RawMessage(impExt)
	impArr := make([]openrtb2.Imp, 0)
	impArr = append(impArr, imp)
	request.ID = "123-456-789"
	request.Imp = impArr

	len := 2
	request.WSeat = make([]string, len)
	for i := 0; i < len; i++ {
		request.WSeat[i] = fmt.Sprintf("Wseat_%d", i)
	}

	/*
		request.Cur = make([]string, len)
		for i := 0; i < len; i++ {
			request.Cur[i] = fmt.Sprintf("cur_%d", i)
		}
	*/
	request.Cur = []string{"USD"}

	request.BAdv = make([]string, len)
	for i := 0; i < len; i++ {
		request.BAdv[i] = fmt.Sprintf("badv_%d", i)
	}

	request.BApp = make([]string, len)
	for i := 0; i < len; i++ {
		request.BApp[i] = fmt.Sprintf("bapp_%d", i)
	}

	request.BCat = make([]string, len)
	for i := 0; i < len; i++ {
		request.BCat[i] = fmt.Sprintf("bcat_%d", i)
	}

	request.WLang = make([]string, len)
	for i := 0; i < len; i++ {
		request.WLang[i] = fmt.Sprintf("Wlang_%d", i)
	}

	request.BSeat = make([]string, len)
	for i := 0; i < len; i++ {
		request.BSeat[i] = fmt.Sprintf("Bseat_%d", i)
	}

	site := new(openrtb2.Site)
	publisher := new(openrtb2.Publisher)
	publisher.ID = "5890"
	site.Publisher = publisher
	site.Page = "www.test.com"

	site.Domain = "test.com"

	request.Site = site

	request.Device = new(openrtb2.Device)
	request.Device.IP = "123.145.167.10"
	request.Device.UA = "Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36"

	request.User = new(openrtb2.User)
	request.User.ID = "119208432"

	request.User.BuyerUID = "1rwe432"

	request.User.Yob = 1980

	request.User.Gender = "F"

	request.User.Geo = new(openrtb2.Geo)
	request.User.Geo.Country = "US"

	request.User.Geo.Region = "CA"

	request.User.Geo.Metro = "90001"

	request.User.Geo.City = "Alamo"

	src := new(openrtb2.Source)
	if request.ID != "" {
		src.TID = request.ID
	}

	src.Ext = json.RawMessage(`{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}`)

	request.Source = src

	var prebidExt openrtb_ext.ExtRequestPrebid

	alias := map[string]string{
		models.BidderAdGenerationAlias:      string(openrtb_ext.BidderAdgeneration),
		models.BidderDistrictmDMXAlias:      string(openrtb_ext.BidderDmx),
		models.BidderPubMaticSecondaryAlias: string(openrtb_ext.BidderPubmatic),
		models.BidderDistrictmAlias:         string(openrtb_ext.BidderAppnexus),
		models.BidderAndBeyondAlias:         string(openrtb_ext.BidderAdkernel),
		models.BidderMediaFuseAlias:         string(openrtb_ext.BidderAppnexus),
	}

	prebidExt.Aliases = alias
	priceGranularity, _ := openrtb_ext.NewPriceGranularityFromLegacyID("auto")
	prebidExt.Targeting = &openrtb_ext.ExtRequestTargeting{
		PriceGranularity:  &priceGranularity,
		IncludeWinners:    util.GetBoolPtr(true),
		IncludeBidderKeys: util.GetBoolPtr(true),
	}
	prebidExt.BidderParams, _ = json.Marshal(map[string]map[string]interface{}{
		"pubmatic": {
			"wiid": "test_wiid",
		},
	})

	prebidExt.Floors = &openrtb_ext.PriceFloorRules{}
	prebidExt.Floors.Enabled = new(bool)
	*prebidExt.Floors.Enabled = true
	prebidExt.Floors.Enforcement = new(openrtb_ext.PriceFloorEnforcement)
	prebidExt.Floors.Enforcement.EnforcePBS = new(bool)
	*prebidExt.Floors.Enforcement.EnforcePBS = true

	prebidExt.ReturnAllBidStatus = true

	//set expected custom macros for video tracking events
	if nil != owAPIBidReq.Imp[0].Video {

		reqExt := models.RequestExt{}
		json.Unmarshal(owAPIBidReq.Ext, reqExt)
		macros := map[string]string{
			"[PROFILE_ID]":            strconv.Itoa(reqExt.Wrapper.ProfileId),
			"[PROFILE_VERSION]":       strconv.Itoa(reqExt.Wrapper.VersionId),
			"[UNIX_TIMESTAMP]":        strconv.Itoa(int(time.Now().Unix())),
			"[PLATFORM]":              fmt.Sprintf("%v", GetDevicePlatform("", &owAPIBidReq, prebidPlatform, nil)),
			"[WRAPPER_IMPRESSION_ID]": "test_wiid",
		}
		prebidExt.Macros = macros
	}

	var pbExt openrtb_ext.ExtOWRequest
	pbExt.Prebid = prebidExt
	request.Ext = pbExt

	if test {
		request.Test = new(int)
		*request.Test = 1
	}

	return request
}

func getTestBidRequest() *openrtb2.BidRequest {

	testReq := &openrtb2.BidRequest{}

	testReq.ID = "testID"

	testReq.Imp = []openrtb2.Imp{
		{
			ID: "testImp1",
			Banner: &openrtb2.Banner{
				W: ptrutil.ToPtr[int64](200),
				H: ptrutil.ToPtr[int64](300),
			},
			Video: &openrtb2.Video{
				W:     200,
				H:     300,
				Plcmt: 1,
			},
		},
	}

	testReq.App = &openrtb2.App{
		Publisher: &openrtb2.Publisher{
			ID: "1010",
		},
		Content: &openrtb2.Content{
			Language: "english",
		},
	}
	testReq.Cur = []string{}

	testReq.Device = &openrtb2.Device{
		DeviceType: 1,
		Language:   "english",
	}
	return testReq
}

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
			got := getVASTEventMacros(tt.args.rctx)
			assert.Equal(t, tt.want, got)

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

func TestOpenWrap_applyProfileChanges(t *testing.T) {
	type fields struct {
		cfg          config.Config
		cache        cache.Cache
		metricEngine metrics.MetricsEngine
	}
	type args struct {
		rctx       models.RequestCtx
		bidRequest *openrtb2.BidRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *openrtb2.BidRequest
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				rctx: models.RequestCtx{
					IsTestRequest: 1,
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							models.AdServerCurrency: "USD",
						},
					},
					TMax:     500,
					IP:       "127.0.0.1",
					Platform: models.PLATFORM_APP,
				},
				bidRequest: getTestBidRequest(),
			},
			want: &openrtb2.BidRequest{
				ID:   "testID",
				Test: 1,
				Cur:  []string{"USD"},
				TMax: 500,
				Source: &openrtb2.Source{
					TID: "testID",
				},
				Imp: []openrtb2.Imp{
					{
						ID: "testImp1",
						Banner: &openrtb2.Banner{
							W: ptrutil.ToPtr[int64](200),
							H: ptrutil.ToPtr[int64](300),
						},
						Video: &openrtb2.Video{
							W:     200,
							H:     300,
							Plcmt: 1,
						},
					},
				},
				Device: &openrtb2.Device{
					IP:         "127.0.0.1",
					Language:   "en",
					DeviceType: 1,
				},
				User: &openrtb2.User{},
				App: &openrtb2.App{
					Publisher: &openrtb2.Publisher{
						ID: "1010",
					},
					Content: &openrtb2.Content{
						Language: "en",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &OpenWrap{
				cfg:          tt.fields.cfg,
				cache:        tt.fields.cache,
				metricEngine: tt.fields.metricEngine,
			}
			got, err := m.applyProfileChanges(tt.args.rctx, tt.args.bidRequest)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenWrap.applyProfileChanges() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOpenWrap_applyVideoAdUnitConfig(t *testing.T) {
	type fields struct {
		cfg          config.Config
		cache        cache.Cache
		metricEngine metrics.MetricsEngine
	}
	type args struct {
		rCtx models.RequestCtx
		imp  *openrtb2.Imp
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *openrtb2.Imp
	}{
		{
			name: "imp.video_is_nil",
			args: args{
				imp: &openrtb2.Imp{
					Video: nil,
				},
			},
			want: &openrtb2.Imp{
				Video: nil,
			},
		},
		{
			name: "empty_adunitCfg",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: nil,
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID:    "testImp",
					Video: &openrtb2.Video{},
				},
			},
			want: &openrtb2.Imp{
				ID:    "testImp",
				Video: &openrtb2.Video{},
			},
		},
		{
			name: "imp.BidFloor_and_BidFloorCur_updated_from_adunit_config",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									BidFloor:    ptrutil.ToPtr(2.0),
									BidFloorCur: ptrutil.ToPtr("USD"),
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID:          "testImp",
					BidFloor:    0,
					BidFloorCur: "",
					Video:       &openrtb2.Video{},
				},
			},
			want: &openrtb2.Imp{
				ID:          "testImp",
				Video:       &openrtb2.Video{},
				BidFloor:    2.0,
				BidFloorCur: "USD",
			},
		},
		{
			name: "imp.Exp_updated_from_adunit_config",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Exp: ptrutil.ToPtr(10),
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID:    "testImp",
					Video: &openrtb2.Video{},
				},
			},
			want: &openrtb2.Imp{
				ID:    "testImp",
				Video: &openrtb2.Video{},
				Exp:   10,
			},
		},
		{
			name: "imp_has_video_object_but_adunitConfig_video_is_nil._imp_video_will_not_be_updated",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Video: nil,
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID: "testImp",
					Video: &openrtb2.Video{
						W: 200,
						H: 300,
					},
				},
			},
			want: &openrtb2.Imp{
				ID: "testImp",
				Video: &openrtb2.Video{
					W: 200,
					H: 300,
				},
			},
		},
		{
			name: "imp_has_video_object_but_video_is_disabled_from_adunitConfig_then_remove_video_object_from_imp",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Video: &adunitconfig.Video{
										Enabled: ptrutil.ToPtr(false),
									},
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID: "testImp",
					Video: &openrtb2.Video{
						W: 200,
						H: 300,
					},
				},
			},
			want: &openrtb2.Imp{
				ID:    "testImp",
				Video: nil,
			},
		},
		{
			name: "imp_has_empty_video_object_and_adunitCofig_for_video_is_enable._all_absent_video_parameters_will_be_updated",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Video: &adunitconfig.Video{
										Enabled: ptrutil.ToPtr(true),
										Config: &adunitconfig.VideoConfig{
											Video: openrtb2.Video{
												MinDuration:    10,
												MaxDuration:    40,
												Skip:           ptrutil.ToPtr(int8(1)),
												SkipMin:        5,
												SkipAfter:      10,
												Plcmt:          1,
												Placement:      1,
												MinBitRate:     100,
												MaxBitRate:     200,
												MaxExtended:    50,
												Linearity:      1,
												Protocol:       1,
												W:              640,
												H:              480,
												Sequence:       2,
												BoxingAllowed:  1,
												PlaybackEnd:    2,
												MIMEs:          []string{"mimes"},
												API:            []adcom1.APIFramework{1, 2},
												Delivery:       []adcom1.DeliveryMethod{1, 2},
												PlaybackMethod: []adcom1.PlaybackMethod{1, 2},
												BAttr:          []adcom1.CreativeAttribute{1, 2},
												StartDelay:     ptrutil.ToPtr(adcom1.StartDelay(2)),
												Protocols:      []adcom1.MediaCreativeSubtype{1, 2},
												Pos:            ptrutil.ToPtr(adcom1.PlacementPosition(1)),
												CompanionType:  []adcom1.CompanionType{1, 2},
											},
										},
									},
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID:    "testImp",
					Video: &openrtb2.Video{},
				},
			},
			want: &openrtb2.Imp{
				ID: "testImp",
				Video: &openrtb2.Video{
					W:              640,
					H:              480,
					MinDuration:    10,
					MaxDuration:    40,
					Skip:           ptrutil.ToPtr(int8(1)),
					SkipMin:        5,
					SkipAfter:      10,
					Plcmt:          1,
					Placement:      1,
					MinBitRate:     100,
					MaxBitRate:     200,
					MaxExtended:    50,
					Linearity:      1,
					Protocol:       1,
					Sequence:       2,
					BoxingAllowed:  1,
					PlaybackEnd:    2,
					MIMEs:          []string{"mimes"},
					API:            []adcom1.APIFramework{1, 2},
					Delivery:       []adcom1.DeliveryMethod{1, 2},
					PlaybackMethod: []adcom1.PlaybackMethod{1, 2},
					BAttr:          []adcom1.CreativeAttribute{1, 2},
					StartDelay:     ptrutil.ToPtr(adcom1.StartDelay(2)),
					Protocols:      []adcom1.MediaCreativeSubtype{1, 2},
					Pos:            ptrutil.ToPtr(adcom1.PlacementPosition(1)),
					CompanionType:  []adcom1.CompanionType{1, 2},
				},
			},
		},
		{
			name: "imp_has_video_object_and_adunitConfig_alos_have_parameter_present_then_priority_to_request",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Video: &adunitconfig.Video{
										Enabled: ptrutil.ToPtr(true),
										Config: &adunitconfig.VideoConfig{
											Video: openrtb2.Video{
												MinDuration: 10,
												MaxDuration: 40,
												Skip:        ptrutil.ToPtr(int8(1)),
												SkipMin:     5,
												SkipAfter:   10,
											},
										},
									},
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID: "testImp",
					Video: &openrtb2.Video{
						W:           640,
						H:           480,
						MinDuration: 20,
						MaxDuration: 60,
						Skip:        ptrutil.ToPtr(int8(2)),
						SkipMin:     10,
						SkipAfter:   20,
					},
				},
			},
			want: &openrtb2.Imp{
				ID: "testImp",
				Video: &openrtb2.Video{
					W:           640,
					H:           480,
					MinDuration: 20,
					MaxDuration: 60,
					Skip:        ptrutil.ToPtr(int8(2)),
					SkipMin:     10,
					SkipAfter:   20,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &OpenWrap{
				cfg:          tt.fields.cfg,
				cache:        tt.fields.cache,
				metricEngine: tt.fields.metricEngine,
			}
			m.applyVideoAdUnitConfig(tt.args.rCtx, tt.args.imp)
			assert.Equal(t, tt.args.imp, tt.want, "Imp video is not upadted as expected from adunit config")
		})
	}
}

func TestOpenWrap_applyBannerAdUnitConfig(t *testing.T) {
	type fields struct {
		cfg          config.Config
		cache        cache.Cache
		metricEngine metrics.MetricsEngine
	}
	type args struct {
		rCtx models.RequestCtx
		imp  *openrtb2.Imp
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *openrtb2.Imp
	}{
		{
			name: "imp.banner_is_nil",
			args: args{
				imp: &openrtb2.Imp{
					Banner: nil,
				},
			},
			want: &openrtb2.Imp{
				Banner: nil,
			},
		},
		{
			name: "empty_adunitCfg",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							BannerAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: nil,
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID:     "testImp",
					Banner: &openrtb2.Banner{},
				},
			},
			want: &openrtb2.Imp{
				ID:     "testImp",
				Banner: &openrtb2.Banner{},
			},
		},
		{
			name: "imp.BidFloor_and_BidFloorCur_updated_from_adunit_config",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							BannerAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									BidFloor:    ptrutil.ToPtr(2.0),
									BidFloorCur: ptrutil.ToPtr("USD"),
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID:          "testImp",
					BidFloor:    0,
					BidFloorCur: "",
					Banner:      &openrtb2.Banner{},
				},
			},
			want: &openrtb2.Imp{
				ID:          "testImp",
				Banner:      &openrtb2.Banner{},
				BidFloor:    2.0,
				BidFloorCur: "USD",
			},
		},
		{
			name: "imp.Exp_updated_from_adunit_config",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							BannerAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Exp: ptrutil.ToPtr(10),
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID:     "testImp",
					Banner: &openrtb2.Banner{},
				},
			},
			want: &openrtb2.Imp{
				ID:     "testImp",
				Banner: &openrtb2.Banner{},
				Exp:    10,
			},
		},
		{
			name: "imp_has_banner_object_but_adunitConfig_banner_is_nil._imp_video_will_not_be_updated",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							BannerAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Banner: nil,
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID: "testImp",
					Banner: &openrtb2.Banner{
						W: ptrutil.ToPtr[int64](200),
						H: ptrutil.ToPtr[int64](300),
					},
				},
			},
			want: &openrtb2.Imp{
				ID: "testImp",
				Banner: &openrtb2.Banner{
					W: ptrutil.ToPtr[int64](200),
					H: ptrutil.ToPtr[int64](300),
				},
			},
		},
		{
			name: "imp_has_banner_object_but_banner_is_disabled_from_adunitConfig_then_remove_banner_object_from_imp",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							BannerAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Banner: &adunitconfig.Banner{
										Enabled: ptrutil.ToPtr(false),
									},
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID: "testImp",
					Banner: &openrtb2.Banner{
						W: ptrutil.ToPtr[int64](200),
						H: ptrutil.ToPtr[int64](300),
					},
				},
			},
			want: &openrtb2.Imp{
				ID:     "testImp",
				Banner: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &OpenWrap{
				cfg:          tt.fields.cfg,
				cache:        tt.fields.cache,
				metricEngine: tt.fields.metricEngine,
			}
			m.applyBannerAdUnitConfig(tt.args.rCtx, tt.args.imp)
			assert.Equal(t, tt.args.imp, tt.want, "Imp banner is not upadted as expected from adunit config")
		})
	}
}

func Test_getDomainFromUrl(t *testing.T) {
	type args struct {
		pageUrl string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test_case_1",
			args: args{
				pageUrl: "http://ebay.com/inte/automation/s2s/pwt_parameter_validation_muti_slot_multi_size.html?pwtvc=1&pwtv=1&profileid=3277",
			},
			want: "ebay.com",
		},
		{
			name: "test_case_2",
			args: args{
				pageUrl: "http://ebay.co.in/inte/automation/s2s/pwt_parameter_validation_muti_slot_multi_size.html?pwtvc=1&pwtv=1&profileid=3277",
			},
			want: "ebay.co.in",
		},
		{
			name: "test_case_3",
			args: args{
				pageUrl: "site@sit.com",
			},
			want: "",
		},
		{
			name: "test_case_4",
			args: args{
				pageUrl: " 12 44",
			},
			want: "",
		},
		{
			name: "test_case_5",
			args: args{
				pageUrl: " ",
			},
			want: "",
		},
		{
			name: "test_case_6",
			args: args{
				pageUrl: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDomainFromUrl(tt.args.pageUrl); got != tt.want {
				t.Errorf("getDomainFromUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_updateRequestExtBidderParamsPubmatic(t *testing.T) {
	type args struct {
		bidderParams json.RawMessage
		cookie       string
		loggerID     string
		bidderCode   string
	}
	tests := []struct {
		name    string
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{
			name: "empty_cookie",
			args: args{
				bidderParams: json.RawMessage(`{"pubmatic":{"pmzoneid":"zone1","adSlot":"38519891"}}`),
				loggerID:     "b441a46e-8c1f-428b-9c29-44e2a408a954",
				bidderCode:   "pubmatic",
			},
			want:    json.RawMessage(`{"pubmatic":{"wiid":"b441a46e-8c1f-428b-9c29-44e2a408a954"}}`),
			wantErr: false,
		},
		{
			name: "empty_loggerID",
			args: args{
				bidderParams: json.RawMessage(`{"pubmatic":{"pmzoneid":"zone1","adSlot":"38519891"}}`),
				cookie:       "test_cookie",
				bidderCode:   "pubmatic",
			},
			want: json.RawMessage(`{"pubmatic":{"Cookie":"test_cookie","wiid":""}}`),
		},
		{
			name: "both_cookie_and_loogerID_are_empty",
			args: args{
				bidderParams: json.RawMessage(`{"pubmatic":{"pmzoneid":"zone1","adSlot":"38519891"}}`),
				cookie:       "",
				loggerID:     "",
				bidderCode:   "pubmatic",
			},
			want: json.RawMessage(`{"pubmatic":{"wiid":""}}`),
		},
		{
			name: "both_cookie_and_loogerID_are_present",
			args: args{
				bidderParams: json.RawMessage(`{"pubmatic":{"pmzoneid":"zone1","adSlot":"38519891"}}`),
				cookie:       "test_cookie",
				loggerID:     "b441a46e-8c1f-428b-9c29-44e2a408a954",
				bidderCode:   "pubmatic",
			},
			want: json.RawMessage(`{"pubmatic":{"Cookie":"test_cookie","wiid":"b441a46e-8c1f-428b-9c29-44e2a408a954"}}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := updateRequestExtBidderParamsPubmatic(tt.args.bidderParams, tt.args.cookie, tt.args.loggerID, tt.args.bidderCode)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateRequestExtBidderParamsPubmatic() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)

		})
	}
}

func TestOpenWrap_handleBeforeValidationHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	type fields struct {
		cfg          config.Config
		cache        cache.Cache
		metricEngine metrics.MetricsEngine
	}
	type args struct {
		ctx       context.Context
		moduleCtx hookstage.ModuleInvocationContext
		payload   hookstage.BeforeValidationRequestPayload
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    hookstage.HookResult[hookstage.BeforeValidationRequestPayload]
		setup   func()
		wantErr bool
	}{
		{
			name: "Pubmatic_AdUnit",
			fields: fields{
				cfg:   config.Config{},
				cache: mockCache,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"div@728x90": {
						SlotName: "div@728x90",
						SlotMappings: map[string]interface{}{
							models.SITE_CACHE_KEY: "12313",
							models.TAG_CACHE_KEY:  "45343",
						},
					},
				})

				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"div@728x90", "div@300x25"},
					HashValueMap: map[string]string{
						"div@728x90": "2aa34b52a9e941c1594af7565e599c8d",
						"div@300x25": "2aa34b52a9e941c1594af7565e599c8d",
					},
				})

				mockCache.EXPECT().GetAdunitConfigFromCache(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&adunitconfig.AdUnitConfig{})
			},
			args: args{
				ctx:       context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.BeforeValidationRequestPayload{
					BidRequest: formORtbV25Request(true, false),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := OpenWrap{
				cfg:          tt.fields.cfg,
				cache:        tt.fields.cache,
				metricEngine: tt.fields.metricEngine,
			}
			got, err := m.handleBeforeValidationHook(tt.args.ctx, tt.args.moduleCtx, tt.args.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenWrap.handleBeforeValidationHook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OpenWrap.handleBeforeValidationHook() = %v, want %v", got, tt.want)
			}
		})
	}
}
