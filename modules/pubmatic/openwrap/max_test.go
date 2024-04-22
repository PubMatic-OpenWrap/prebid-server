package openwrap

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func Test_updateImpression(t *testing.T) {
	type args struct {
		sdkImpression openrtb2.Imp
		maxImpression *openrtb2.Imp
	}
	tests := []struct {
		name string
		args args
		want *openrtb2.Imp
	}{
		{
			name: "maxImp nil",
			args: args{
				maxImpression: nil,
			},
			want: nil,
		},
		{
			name: "maxImp video not present",
			args: args{
				sdkImpression: openrtb2.Imp{ClickBrowser: openrtb2.Int8Ptr(0), DisplayManager: "PubMatic_SDK", DisplayManagerVer: "1.2"},
				maxImpression: &openrtb2.Imp{ClickBrowser: openrtb2.Int8Ptr(1), DisplayManager: "Applovin_SDK"},
			},
			want: &openrtb2.Imp{ClickBrowser: openrtb2.Int8Ptr(0), DisplayManager: "PubMatic_SDK", DisplayManagerVer: "1.2"},
		},
		{
			name: "only maxImp has video",
			args: args{
				sdkImpression: openrtb2.Imp{ClickBrowser: openrtb2.Int8Ptr(0), DisplayManager: "PubMatic_SDK", DisplayManagerVer: "1.2"},
				maxImpression: &openrtb2.Imp{ClickBrowser: openrtb2.Int8Ptr(1), DisplayManager: "Applovin_SDK", Video: &openrtb2.Video{W: openrtb2.Int64Ptr(300), H: openrtb2.Int64Ptr(250)}},
			},
			want: &openrtb2.Imp{ClickBrowser: openrtb2.Int8Ptr(0), DisplayManager: "PubMatic_SDK", DisplayManagerVer: "1.2", Video: &openrtb2.Video{W: openrtb2.Int64Ptr(300), H: openrtb2.Int64Ptr(250)}},
		},
		{
			name: "maxImp and sdkImp has video",
			args: args{
				sdkImpression: openrtb2.Imp{ClickBrowser: openrtb2.Int8Ptr(0), DisplayManager: "PubMatic_SDK", DisplayManagerVer: "1.2", Video: &openrtb2.Video{W: openrtb2.Int64Ptr(300), H: openrtb2.Int64Ptr(250), BAttr: []adcom1.CreativeAttribute{1, 2}}},
				maxImpression: &openrtb2.Imp{ClickBrowser: openrtb2.Int8Ptr(1), DisplayManager: "Applovin_SDK", Video: &openrtb2.Video{W: openrtb2.Int64Ptr(750), H: openrtb2.Int64Ptr(500), BAttr: []adcom1.CreativeAttribute{6, 1, 8, 4}}},
			},
			want: &openrtb2.Imp{ClickBrowser: openrtb2.Int8Ptr(0), DisplayManager: "PubMatic_SDK", DisplayManagerVer: "1.2", Video: &openrtb2.Video{W: openrtb2.Int64Ptr(300), H: openrtb2.Int64Ptr(250), BAttr: []adcom1.CreativeAttribute{1, 2}}},
		},
		{
			name: "maxImp has and sdkImp has banner with api",
			args: args{
				sdkImpression: openrtb2.Imp{Banner: &openrtb2.Banner{ID: "sdk_banner", API: []adcom1.APIFramework{1, 2, 3, 4}}},
				maxImpression: &openrtb2.Imp{Banner: &openrtb2.Banner{ID: "max_banner"}},
			},
			want: &openrtb2.Imp{Banner: &openrtb2.Banner{ID: "max_banner", API: []adcom1.APIFramework{1, 2, 3, 4}}},
		},
		{
			name: "maxImp has bannertype rewarded",
			args: args{
				sdkImpression: openrtb2.Imp{Banner: &openrtb2.Banner{ID: "sdk_banner", API: []adcom1.APIFramework{1, 2, 3, 4}}},
				maxImpression: &openrtb2.Imp{Banner: &openrtb2.Banner{ID: "max_banner", Ext: json.RawMessage(`{"bannertype":"rewarded"}`)}},
			},
			want: &openrtb2.Imp{},
		},
		{
			name: "maxImp has no ext, sdkImp has reward in ext",
			args: args{
				sdkImpression: openrtb2.Imp{Ext: json.RawMessage(`{"reward":1}`)},
				maxImpression: &openrtb2.Imp{},
			},
			want: &openrtb2.Imp{Ext: json.RawMessage(`{"reward":1}`)},
		},
		{
			name: "maxImp has no ext, sdkImp has reward and skadn in ext",
			args: args{
				sdkImpression: openrtb2.Imp{Ext: json.RawMessage(`{"reward":1,"skadn":{"versions":["2.0","2.1"],"sourceapp":"11111","skadnetids":["424m5254lk.skadnetwork","4fzdc2evr5.skadnetwork"]}}`)},
				maxImpression: &openrtb2.Imp{},
			},
			want: &openrtb2.Imp{Ext: json.RawMessage(`{"reward":1,"skadn":{"versions":["2.0","2.1"],"sourceapp":"11111","skadnetids":["424m5254lk.skadnetwork","4fzdc2evr5.skadnetwork"]}}`)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateImpression(tt.args.sdkImpression, tt.args.maxImpression)
			assert.Equal(t, tt.want, tt.args.maxImpression, tt.name)
		})
	}
}

func Test_updateDevice(t *testing.T) {
	type args struct {
		sdkDevice  *openrtb2.Device
		maxRequest *openrtb2.BidRequest
	}
	tests := []struct {
		name string
		args args
		want *openrtb2.Device
	}{
		{
			name: "sdkDevice nil",
			args: args{
				sdkDevice:  nil,
				maxRequest: &openrtb2.BidRequest{Device: &openrtb2.Device{DeviceType: 5}},
			},
			want: &openrtb2.Device{DeviceType: 5},
		},
		{
			name: "sdkDevice has mccmnc,connectiontype",
			args: args{
				sdkDevice:  &openrtb2.Device{MCCMNC: "mccmnc", ConnectionType: adcom1.Connection2G.Ptr()},
				maxRequest: &openrtb2.BidRequest{Device: &openrtb2.Device{DeviceType: 5}},
			},
			want: &openrtb2.Device{DeviceType: 5, MCCMNC: "mccmnc", ConnectionType: adcom1.Connection2G.Ptr()},
		},
		{
			name: "sdkDeviceExt has atts",
			args: args{
				sdkDevice:  &openrtb2.Device{Ext: json.RawMessage(`{"atts":3}`)},
				maxRequest: &openrtb2.BidRequest{},
			},
			want: &openrtb2.Device{Ext: json.RawMessage(`{"atts":3}`)},
		},
		{
			name: "sdkDevice has geo city and utcoffset",
			args: args{
				sdkDevice:  &openrtb2.Device{Geo: &openrtb2.Geo{City: "Jalgaon", UTCOffset: 3}, Ext: json.RawMessage(`{"atts":3}`)},
				maxRequest: &openrtb2.BidRequest{},
			},
			want: &openrtb2.Device{Geo: &openrtb2.Geo{City: "Jalgaon", UTCOffset: 3}, Ext: json.RawMessage(`{"atts":3}`)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateDevice(tt.args.sdkDevice, tt.args.maxRequest)
			assert.Equal(t, tt.want, tt.args.maxRequest.Device, tt.name)
		})
	}
}

func Test_updateApp(t *testing.T) {
	type args struct {
		sdkApp     *openrtb2.App
		maxRequest *openrtb2.BidRequest
	}
	tests := []struct {
		name string
		args args
		want *openrtb2.App
	}{
		{
			name: "sdkApp is nil",
			args: args{
				sdkApp:     nil,
				maxRequest: &openrtb2.BidRequest{App: &openrtb2.App{ID: "sdkapp"}},
			},
			want: &openrtb2.App{ID: "sdkapp"},
		},
		{
			name: "maxDevice is nil",
			args: args{
				sdkApp:     &openrtb2.App{},
				maxRequest: &openrtb2.BidRequest{},
			},
			want: &openrtb2.App{},
		},
		{
			name: "sdkApp has Paid,Keywords and Domain",
			args: args{
				sdkApp:     &openrtb2.App{Paid: openrtb2.Int8Ptr(1), Keywords: "k1=v1", Domain: "abc.com"},
				maxRequest: &openrtb2.BidRequest{},
			},
			want: &openrtb2.App{Paid: openrtb2.Int8Ptr(1), Keywords: "k1=v1", Domain: "abc.com"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateApp(tt.args.sdkApp, tt.args.maxRequest)
			assert.Equal(t, tt.want, tt.args.maxRequest.App, tt.name)
		})
	}
}

func Test_updateRegs(t *testing.T) {
	type args struct {
		sdkRegs    *openrtb2.Regs
		maxRequest *openrtb2.BidRequest
	}
	tests := []struct {
		name string
		args args
		want *openrtb2.Regs
	}{
		{
			name: "sdkRegs is nil",
			args: args{
				sdkRegs:    nil,
				maxRequest: &openrtb2.BidRequest{},
			},
			want: nil,
		},
		{
			name: "sdkRegsExt is nil",
			args: args{
				sdkRegs:    &openrtb2.Regs{},
				maxRequest: &openrtb2.BidRequest{},
			},
			want: &openrtb2.Regs{},
		},
		{
			name: "maxRegs is nil",
			args: args{
				sdkRegs:    &openrtb2.Regs{Ext: json.RawMessage(`{}`)},
				maxRequest: &openrtb2.BidRequest{},
			},
			want: &openrtb2.Regs{},
		},
		{
			name: "sdkRegs has coppa, sdkRegsExt has gdpr, gpp",
			args: args{
				sdkRegs:    &openrtb2.Regs{COPPA: 1, Ext: json.RawMessage(`{"gdpr":1,"gpp":"sdfewe3cer"}`)},
				maxRequest: &openrtb2.BidRequest{},
			},
			want: &openrtb2.Regs{COPPA: 1, Ext: json.RawMessage(`{"gdpr":1,"gpp":"sdfewe3cer"}`)},
		},
		{
			name: "sdkRegsExt has gdpr, gpp, gpp_sid, us_privacy and maxRegsExt has gpp",
			args: args{
				sdkRegs:    &openrtb2.Regs{Ext: json.RawMessage(`{"gdpr":1,"gpp":"sdfewe3cer","gpp_sid":[6],"us_privacy":"uspConsentString"}`)},
				maxRequest: &openrtb2.BidRequest{Regs: &openrtb2.Regs{Ext: json.RawMessage(`{"gpp":"gpp_string"}`)}},
			},
			want: &openrtb2.Regs{Ext: json.RawMessage(`{"gpp":"sdfewe3cer","gdpr":1,"gpp_sid":[6],"us_privacy":"uspConsentString"}`)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateRegs(tt.args.sdkRegs, tt.args.maxRequest)
			assert.Equal(t, tt.want, tt.args.maxRequest.Regs, tt.name)
		})
	}
}

func Test_updateSource(t *testing.T) {
	type args struct {
		sdkSource  *openrtb2.Source
		maxRequest *openrtb2.BidRequest
	}
	tests := []struct {
		name string
		args args
		want *openrtb2.Source
	}{
		{
			name: "sdkSource is nil",
			args: args{
				sdkSource:  nil,
				maxRequest: &openrtb2.BidRequest{},
			},
			want: nil,
		},
		{
			name: "sdkSourceExt is nil",
			args: args{
				sdkSource:  &openrtb2.Source{},
				maxRequest: &openrtb2.BidRequest{},
			},
			want: nil,
		},
		{
			name: "maxSource is nil",
			args: args{
				sdkSource:  &openrtb2.Source{Ext: json.RawMessage(`{}`)},
				maxRequest: &openrtb2.BidRequest{},
			},
			want: &openrtb2.Source{},
		},
		{
			name: "sdkSourceExt has omidpn, omidpv",
			args: args{
				sdkSource:  &openrtb2.Source{Ext: json.RawMessage(`{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}`)},
				maxRequest: &openrtb2.BidRequest{},
			},
			want: &openrtb2.Source{Ext: json.RawMessage(`{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}`)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateSource(tt.args.sdkSource, tt.args.maxRequest)
			assert.Equal(t, tt.want, tt.args.maxRequest.Source, tt.name)
		})
	}
}

func Test_updateUser(t *testing.T) {
	type args struct {
		sdkUser    *openrtb2.User
		maxRequest *openrtb2.BidRequest
	}
	tests := []struct {
		name string
		args args
		want *openrtb2.User
	}{
		{
			name: "sdkUser is nil",
			args: args{
				sdkUser:    nil,
				maxRequest: &openrtb2.BidRequest{},
			},
			want: nil,
		},
		{
			name: "maxUser is nil",
			args: args{
				sdkUser:    &openrtb2.User{Ext: json.RawMessage(``)},
				maxRequest: &openrtb2.BidRequest{},
			},
			want: &openrtb2.User{},
		},
		{
			name: "sdkUser has yob, gender, keywords",
			args: args{
				sdkUser:    &openrtb2.User{Yob: 1999, Gender: "M", Keywords: "k1=v2;k2=v2", Ext: json.RawMessage(``)},
				maxRequest: &openrtb2.BidRequest{},
			},
			want: &openrtb2.User{Yob: 1999, Gender: "M", Keywords: "k1=v2;k2=v2"},
		},
		{
			name: "sdkUser and maxUser has yob, gender, keywords and data",
			args: args{
				sdkUser:    &openrtb2.User{Data: []openrtb2.Data{{ID: "123", Name: "PubMatic_SDK", Segment: []openrtb2.Segment{{ID: "seg_id", Name: "PubMatic_Seg", Value: "segment_value", Ext: json.RawMessage(`{"segtax":4}`)}}}}, Yob: 1999, Gender: "M", Keywords: "k1=v2;k2=v2", Ext: json.RawMessage(``)},
				maxRequest: &openrtb2.BidRequest{User: &openrtb2.User{Data: []openrtb2.Data{{ID: "max_id", Name: "Publisher Passed", Segment: []openrtb2.Segment{{ID: "max_seg_id", Name: "max_Seg", Value: "max_segment_value"}}}}, Yob: 2000, Gender: "F", Keywords: "k52=v43"}},
			},
			want: &openrtb2.User{Data: []openrtb2.Data{{ID: "123", Name: "PubMatic_SDK", Segment: []openrtb2.Segment{{ID: "seg_id", Name: "PubMatic_Seg", Value: "segment_value", Ext: json.RawMessage(`{"segtax":4}`)}}}}, Yob: 1999, Gender: "M", Keywords: "k1=v2;k2=v2"},
		},
		{
			name: "sdkUserExt has consent",
			args: args{
				sdkUser:    &openrtb2.User{ID: "sdkID", Yob: 1999, Gender: "M", Keywords: "k1=v2;k2=v2", Ext: json.RawMessage(`{"consent":"consent_string"}`)},
				maxRequest: &openrtb2.BidRequest{User: &openrtb2.User{ID: "maxID", Yob: 2000, Gender: "F", Keywords: "k52=v43"}},
			},
			want: &openrtb2.User{ID: "maxID", Yob: 1999, Gender: "M", Keywords: "k1=v2;k2=v2", Ext: json.RawMessage(`{"consent":"consent_string"}`)},
		},
		{
			name: "sdkUserExt has consent and eids",
			args: args{
				sdkUser:    &openrtb2.User{ID: "sdkID", Yob: 1999, Gender: "M", Keywords: "k1=v2;k2=v2", Ext: json.RawMessage(`{"consent":"consent_string","eids":[{"source":"amxid","uids":[{"atype":1,"id":"88de601e-3d98-48e7-81d7-00000000"}]},{"source":"adserver.org","uids":[{"id":"1234567","ext":{"rtiPartner":"TDID"}}]}]}`)},
				maxRequest: &openrtb2.BidRequest{User: &openrtb2.User{ID: "maxID", Yob: 2000, Gender: "F", Keywords: "k52=v43"}},
			},
			want: &openrtb2.User{ID: "maxID", Yob: 1999, Gender: "M", Keywords: "k1=v2;k2=v2", Ext: json.RawMessage(`{"consent":"consent_string","eids":[{"source":"amxid","uids":[{"atype":1,"id":"88de601e-3d98-48e7-81d7-00000000"}]},{"source":"adserver.org","uids":[{"id":"1234567","ext":{"rtiPartner":"TDID"}}]}]}`)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateUser(tt.args.sdkUser, tt.args.maxRequest)
			assert.Equal(t, tt.want, tt.args.maxRequest.User, tt.name)
		})
	}
}

func Test_setIfKeysExists(t *testing.T) {
	type args struct {
		source []byte
		target []byte
		keys   []string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "keys not found in source",
			args: args{
				source: nil,
				target: nil,
				keys:   []string{"key1", "key2"},
			},
			want: nil,
		},
		{
			name: "int value key found out of all keys",
			args: args{
				source: []byte(`{"key1":23,"key40":"v40"}`),
				target: nil,
				keys:   []string{"key1", "key2"},
			},
			want: []byte(`{"key1":23}`),
		},
		{
			name: "string value key found out of all keys",
			args: args{
				source: []byte(`{"key1":23,"key40":"v40"}`),
				target: nil,
				keys:   []string{"key40", "key2"},
			},
			want: []byte(`{"key40":"v40"}`),
		},
		{
			name: "overwrite string value key in target",
			args: args{
				source: []byte(`{"key1":55555,"key40":"v40"}`),
				target: []byte(`{"key1":23,"key40":"will_overwrite"}`),
				keys:   []string{"key40", "key2"},
			},
			want: []byte(`{"key1":23,"key40":"v40"}`),
		},
		{
			name: "error while setting key, return oldTarget",
			args: args{
				source: []byte(`{"key1":555555,"key40":"v40"}`),
				target: []byte(`"key1":23,"key40":"value40"}`),
				keys:   []string{"key40", "key2"},
			},
			want: []byte(`"key1":23,"key40":"value40"}`),
		},
		{
			name: "overwrite key in target with object",
			args: args{
				source: []byte(`{"key1":55555,"key40":{"user":{"id":"1kjh3429kjh295jkl","ext":{"consent":"CONSENT_STRING"}},"regs":{"ext":{"gdpr":1}}}}`),
				target: []byte(`{"key1":23,"key40":[]}`),
				keys:   []string{"key40", "key2"},
			},
			want: []byte(`{"key1":23,"key40":{"user":{"id":"1kjh3429kjh295jkl","ext":{"consent":"CONSENT_STRING"}},"regs":{"ext":{"gdpr":1}}}}`),
		},
		{
			name: "set slice in key",
			args: args{
				source: []byte(`{"key1":555555,"key40":[1,2,3,4,5]}`),
				target: []byte(`{"key1":23,"key40":"value40"}`),
				keys:   []string{"key40", "key2"},
			},
			want: []byte(`{"key1":23,"key40":[1,2,3,4,5]}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := setIfKeysExists(tt.args.source, tt.args.target, tt.args.keys...)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func Test_addSignalDataInRequest(t *testing.T) {
	type args struct {
		signal           string
		maxRequest       json.RawMessage
		clientconfigflag int
	}
	tests := []struct {
		name           string
		args           args
		wantMaxRequest json.RawMessage
	}{
		{
			name: "signal is empty",
			args: args{
				signal:     "",
				maxRequest: json.RawMessage(`{"id":"123"}`),
			},
			wantMaxRequest: json.RawMessage(`{"id":"123"}`),
		},
		{
			name: "error unmarshalling signal",
			args: args{
				signal:     `{":"test-request-id","site":{}}`,
				maxRequest: json.RawMessage(`{"id":"123"}`),
			},
			wantMaxRequest: json.RawMessage(`{"id":"123"}`),
		},
		{
			name: "replace or add from signal",
			args: args{
				signal:           `{"device":{"devicetype":4,"w":393,"h":852,"ifa":"F5BA1637-7156-4369-BA7E-3C45033D9F61","mccmnc":"311-480","js":1,"osv":"17.3.1","connectiontype":5,"os":"iOS","pxratio":3,"geo":{"lastfix":8,"lat":37.48773508935608,"utcoffset":-480,"lon":-122.22855027909678,"type":1},"language":"en","make":"Apple","ext":{"atts":3},"ua":"Mozilla/5.0 (iPhone; CPU iPhone OS 17_3_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148","model":"iPhone15,2","carrier":"Verizon"},"source":{"ext":{"omidpn":"Pubmatic","omidpv":"3.1.0"}},"id":"CE204A0E-31C3-4D7F-A1A0-D34AF5ED1A7F","app":{"id":"406719683","paid":1,"keywords":"k1=v1","domain":"abc.com","bundle":"406719683","storeurl":"https://apps.apple.com/us/app/gasbuddy-find-pay-for-gas/id406719683","name":"GasBuddy","publisher":{"id":"160361"},"ver":"700.89.22927"},"ext":{"wrapper":{"sumry_disable":1,"clientconfig":1,"profileid":3422}},"imp":[{"secure":1,"tagid":"Mobile_iPhone_List_Screen_Bottom","banner":{"pos":0,"format":[{"w":300,"h":250}],"api":[5,6,7]},"id":"98D9318E-5276-402F-BAA4-CDBD8A364957","ext":{"skadn":{"sourceapp":"406719683","versions":["2.0","2.1","2.2","3.0","4.0"],"skadnetids":["cstr6suwn9.skadnetwork","7ug5zh24hu.skadnetwork","uw77j35x4d.skadnetwork","c6k4g5qg8m.skadnetwork","hs6bdukanm.skadnetwork","yclnxrl5pm.skadnetwork","3sh42y64q3.skadnetwork","cj5566h2ga.skadnetwork","klf5c3l5u5.skadnetwork","8s468mfl3y.skadnetwork","2u9pt9hc89.skadnetwork","7rz58n8ntl.skadnetwork","ppxm28t8ap.skadnetwork","mtkv5xtk9e.skadnetwork","cg4yq2srnc.skadnetwork","wzmmz9fp6w.skadnetwork","k674qkevps.skadnetwork","v72qych5uu.skadnetwork","578prtvx9j.skadnetwork","3rd42ekr43.skadnetwork","g28c52eehv.skadnetwork","2fnua5tdw4.skadnetwork","9nlqeag3gk.skadnetwork","5lm9lj6jb7.skadnetwork","97r2b46745.skadnetwork","e5fvkxwrpn.skadnetwork","4pfyvq9l8r.skadnetwork","tl55sbb4fm.skadnetwork","t38b2kh725.skadnetwork","prcb7njmu6.skadnetwork","mlmmfzh3r3.skadnetwork","9t245vhmpl.skadnetwork","9rd848q2bz.skadnetwork","4fzdc2evr5.skadnetwork","4468km3ulz.skadnetwork","m8dbw4sv7c.skadnetwork","ejvt5qm6ak.skadnetwork","5lm9lj6jb7.skadnetwork","44jx6755aq.skadnetwork","6g9af3uyq4.skadnetwork","u679fj5vs4.skadnetwork","rx5hdcabgc.skadnetwork","275upjj5gd.skadnetwork","p78axxw29g.skadnetwork"],"productpage":1,"version":"2.0"}},"displaymanagerver":"3.1.0","clickbrowser":1,"video":{"companionad":[{"pos":0,"format":[{"w":300,"h":250}],"vcm":1}],"protocols":[2,3,5,6,7,8,11,12,13,14],"h":250,"w":300,"linearity":1,"pos":0,"boxingallowed":1,"placement":2,"mimes":["video/3gpp2","video/quicktime","video/mp4","video/x-m4v","video/3gpp"],"companiontype":[1,2,3],"delivery":[2],"startdelay":0,"playbackend":1,"api":[7]},"displaymanager":"PubMatic_OpenWrap_SDK","instl":0}],"at":1,"cur":["USD"],"regs":{"coppa":1,"ext":{"ccpa":0,"gdpr":1,"gpp":"gpp_string","gpp_sid":[7],"us_privacy":"uspConsentString","consent":"0"}}}`,
				maxRequest:       json.RawMessage(`{"id":"{BID_ID}","at":1,"bcat":["IAB26-4","IAB26-2","IAB25-6","IAB25-5","IAB25-4","IAB25-3","IAB25-1","IAB25-7","IAB8-18","IAB26-3","IAB26-1","IAB8-5","IAB25-2","IAB11-4"],"tmax":3000,"app":{"name":"DrawHappyAngel","ver":"0.5.4","bundle":"com.newstory.DrawHappyAngel","cat":["IAB9-30"],"id":"{NETWORK_APP_ID}","publisher":{"name":"New Story Inc.","ext":{"installed_sdk":{"id":"MOLOCO_BIDDING","sdk_version":{"major":1,"minor":0,"micro":0},"adapter_version":{"major":1,"minor":0,"micro":0}}}},"ext":{"orientation":1}},"device":{"ifa":"497a10d6-c4dd-4e04-a986-c32b7180d462","ip":"38.158.207.171","carrier":"MYTEL","language":"en_US","hwv":"ruby","ppi":440,"pxratio":2.75,"devicetype":4,"connectiontype":2,"js":1,"h":2400,"w":1080,"geo":{"type":2,"ipservice":3,"lat":40.7429,"lon":-73.9392,"long":-73.9392,"city":"Queens","country":"USA","region":"ny","dma":"501","metro":"501","zip":"11101","ext":{"org":"Myanmar Broadband Telecom Co.","isp":"Myanmar Broadband Telecom Co."}},"ext":{},"osv":"13.0.0","ua":"Mozilla/5.0 (Linux; Android 13; 22101316C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.230 Mobile Safari/537.36","make":"xiaomi","model":"22101316c","os":"android"},"imp":[{"id":"1","displaymanager":"applovin_mediation","displaymanagerver":"11.8.2","instl":0,"secure":0,"tagid":"{NETWORK_PLACEMENT_ID}","bidfloor":0.01,"bidfloorcur":"USD","exp":14400,"banner":{"id":"1","w":320,"h":50,"btype":[],"battr":[1,2,5,8,9,14,17],"pos":1,"format":[{"w":320,"h":50}]},"rwdd":0}],"user":{"data":[{"id":"1","name":"Publisher Passed","segment":[{"signal":"{BIDDING_SIGNAL}"}]}],"ext":{"gdpr":0}},"regs":{"ext":{"ccpa":0,"gdpr":1,"consent":"0","tcf_consent_string":"{TCF_STRING}"}},"source":{"ext":{"schain":{"ver":"1.0","complete":1,"nodes":[{"asi":"applovin.com","sid":"53bf468f18c5a0e2b7d4e3f748c677c1","rid":"494dbe15a3ce08c54f4e456363f35a022247f997","hp":1}]}}},"ext":{"wrapper":{"profileid":1234}}}`),
				clientconfigflag: 1,
			},
			wantMaxRequest: json.RawMessage(`{"id":"{BID_ID}","at":1,"bcat":["IAB26-4","IAB26-2","IAB25-6","IAB25-5","IAB25-4","IAB25-3","IAB25-1","IAB25-7","IAB8-18","IAB26-3","IAB26-1","IAB8-5","IAB25-2","IAB11-4"],"tmax":3000,"app":{"paid":1,"keywords":"k1=v1","domain":"abc.com","name":"DrawHappyAngel","ver":"0.5.4","bundle":"com.newstory.DrawHappyAngel","cat":["IAB9-30"],"id":"{NETWORK_APP_ID}","publisher":{"name":"New Story Inc.","ext":{"installed_sdk":{"id":"MOLOCO_BIDDING","sdk_version":{"major":1,"minor":0,"micro":0},"adapter_version":{"major":1,"minor":0,"micro":0}}}},"ext":{"orientation":1}},"device":{"ifa":"497a10d6-c4dd-4e04-a986-c32b7180d462","ip":"38.158.207.171","carrier":"MYTEL","language":"en_US","hwv":"ruby","ppi":440,"pxratio":2.75,"devicetype":4,"mccmnc":"311-480","connectiontype":5,"js":1,"h":2400,"w":1080,"geo":{"city":"Queens","type":2,"ipservice":3,"lat":40.7429,"lon":-73.9392,"long":-73.9392,"country":"USA","region":"ny","dma":"501","metro":"501","zip":"11101","utcoffset":-480,"ext":{"org":"Myanmar Broadband Telecom Co.","isp":"Myanmar Broadband Telecom Co."}},"ext":{"atts":3},"osv":"13.0.0","ua":"Mozilla/5.0 (Linux; Android 13; 22101316C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.230 Mobile Safari/537.36","make":"xiaomi","model":"22101316c","os":"android"},"imp":[{"id":"1","displaymanagerver":"3.1.0","clickbrowser":1,"displaymanager":"PubMatic_OpenWrap_SDK","instl":0,"secure":0,"tagid":"{NETWORK_PLACEMENT_ID}","bidfloor":0.01,"bidfloorcur":"USD","exp":14400,"banner":{"id":"1","w":320,"h":50,"btype":[],"api":[5,6,7],"battr":[1,2,5,8,9,14,17],"pos":1,"format":[{"w":320,"h":50}]},"video":{"companionad":[{"pos":0,"format":[{"w":300,"h":250}],"vcm":1}],"protocols":[2,3,5,6,7,8,11,12,13,14],"h":250,"w":300,"linearity":1,"pos":0,"boxingallowed":1,"placement":2,"mimes":["video/3gpp2","video/quicktime","video/mp4","video/x-m4v","video/3gpp"],"companiontype":[1,2,3],"delivery":[2],"startdelay":0,"playbackend":1,"api":[7]},"rwdd":0,"ext":{"skadn":{"sourceapp":"406719683","versions":["2.0","2.1","2.2","3.0","4.0"],"skadnetids":["cstr6suwn9.skadnetwork","7ug5zh24hu.skadnetwork","uw77j35x4d.skadnetwork","c6k4g5qg8m.skadnetwork","hs6bdukanm.skadnetwork","yclnxrl5pm.skadnetwork","3sh42y64q3.skadnetwork","cj5566h2ga.skadnetwork","klf5c3l5u5.skadnetwork","8s468mfl3y.skadnetwork","2u9pt9hc89.skadnetwork","7rz58n8ntl.skadnetwork","ppxm28t8ap.skadnetwork","mtkv5xtk9e.skadnetwork","cg4yq2srnc.skadnetwork","wzmmz9fp6w.skadnetwork","k674qkevps.skadnetwork","v72qych5uu.skadnetwork","578prtvx9j.skadnetwork","3rd42ekr43.skadnetwork","g28c52eehv.skadnetwork","2fnua5tdw4.skadnetwork","9nlqeag3gk.skadnetwork","5lm9lj6jb7.skadnetwork","97r2b46745.skadnetwork","e5fvkxwrpn.skadnetwork","4pfyvq9l8r.skadnetwork","tl55sbb4fm.skadnetwork","t38b2kh725.skadnetwork","prcb7njmu6.skadnetwork","mlmmfzh3r3.skadnetwork","9t245vhmpl.skadnetwork","9rd848q2bz.skadnetwork","4fzdc2evr5.skadnetwork","4468km3ulz.skadnetwork","m8dbw4sv7c.skadnetwork","ejvt5qm6ak.skadnetwork","5lm9lj6jb7.skadnetwork","44jx6755aq.skadnetwork","6g9af3uyq4.skadnetwork","u679fj5vs4.skadnetwork","rx5hdcabgc.skadnetwork","275upjj5gd.skadnetwork","p78axxw29g.skadnetwork"],"productpage":1,"version":"2.0"}}}],"user":{"data":[{"id":"1","name":"Publisher Passed","segment":[{"signal":"{BIDDING_SIGNAL}"}]}],"ext":{"gdpr":0}},"regs":{"coppa":1,"ext":{"ccpa":0,"gdpr":1,"consent":"0","tcf_consent_string":"{TCF_STRING}","gpp":"gpp_string","gpp_sid":[7],"us_privacy":"uspConsentString"}},"source":{"ext":{"schain":{"ver":"1.0","complete":1,"nodes":[{"asi":"applovin.com","sid":"53bf468f18c5a0e2b7d4e3f748c677c1","rid":"494dbe15a3ce08c54f4e456363f35a022247f997","hp":1}]},"omidpn":"Pubmatic","omidpv":"3.1.0"}},"ext":{"wrapper":{"profileid":1234,"clientconfig":1}}}`),
		},
		{
			name: "replace or add from signal,and remove banner as bannertype rewarded",
			args: args{
				signal:     `{"device":{"devicetype":4,"w":393,"h":852,"ifa":"F5BA1637-7156-4369-BA7E-3C45033D9F61","mccmnc":"311-480","js":1,"osv":"17.3.1","connectiontype":5,"os":"iOS","pxratio":3,"geo":{"lastfix":8,"lat":37.48773508935608,"utcoffset":-480,"lon":-122.22855027909678,"type":1},"language":"en","make":"Apple","ext":{"atts":3},"ua":"Mozilla/5.0 (iPhone; CPU iPhone OS 17_3_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148","model":"iPhone15,2","carrier":"Verizon"},"source":{"ext":{"omidpn":"Pubmatic","omidpv":"3.1.0"}},"id":"CE204A0E-31C3-4D7F-A1A0-D34AF5ED1A7F","app":{"id":"406719683","paid":1,"keywords":"k1=v1","domain":"abc.com","bundle":"406719683","storeurl":"https://apps.apple.com/us/app/gasbuddy-find-pay-for-gas/id406719683","name":"GasBuddy","publisher":{"id":"160361"},"ver":"700.89.22927"},"ext":{"wrapper":{"sumry_disable":1,"clientconfig":1,"profileid":3422}},"imp":[{"secure":1,"tagid":"Mobile_iPhone_List_Screen_Bottom","banner":{"pos":0,"format":[{"w":300,"h":250}],"api":[5,6,7]},"id":"98D9318E-5276-402F-BAA4-CDBD8A364957","ext":{"skadn":{"sourceapp":"406719683","versions":["2.0","2.1","2.2","3.0","4.0"],"skadnetids":["cstr6suwn9.skadnetwork","7ug5zh24hu.skadnetwork","uw77j35x4d.skadnetwork","c6k4g5qg8m.skadnetwork","hs6bdukanm.skadnetwork","yclnxrl5pm.skadnetwork","3sh42y64q3.skadnetwork","cj5566h2ga.skadnetwork","klf5c3l5u5.skadnetwork","8s468mfl3y.skadnetwork","2u9pt9hc89.skadnetwork","7rz58n8ntl.skadnetwork","ppxm28t8ap.skadnetwork","mtkv5xtk9e.skadnetwork","cg4yq2srnc.skadnetwork","wzmmz9fp6w.skadnetwork","k674qkevps.skadnetwork","v72qych5uu.skadnetwork","578prtvx9j.skadnetwork","3rd42ekr43.skadnetwork","g28c52eehv.skadnetwork","2fnua5tdw4.skadnetwork","9nlqeag3gk.skadnetwork","5lm9lj6jb7.skadnetwork","97r2b46745.skadnetwork","e5fvkxwrpn.skadnetwork","4pfyvq9l8r.skadnetwork","tl55sbb4fm.skadnetwork","t38b2kh725.skadnetwork","prcb7njmu6.skadnetwork","mlmmfzh3r3.skadnetwork","9t245vhmpl.skadnetwork","9rd848q2bz.skadnetwork","4fzdc2evr5.skadnetwork","4468km3ulz.skadnetwork","m8dbw4sv7c.skadnetwork","ejvt5qm6ak.skadnetwork","5lm9lj6jb7.skadnetwork","44jx6755aq.skadnetwork","6g9af3uyq4.skadnetwork","u679fj5vs4.skadnetwork","rx5hdcabgc.skadnetwork","275upjj5gd.skadnetwork","p78axxw29g.skadnetwork"],"productpage":1,"version":"2.0"}},"displaymanagerver":"3.1.0","clickbrowser":1,"video":{"companionad":[{"pos":0,"format":[{"w":300,"h":250}],"vcm":1}],"protocols":[2,3,5,6,7,8,11,12,13,14],"h":250,"w":300,"linearity":1,"pos":0,"boxingallowed":1,"placement":2,"mimes":["video/3gpp2","video/quicktime","video/mp4","video/x-m4v","video/3gpp"],"companiontype":[1,2,3],"delivery":[2],"startdelay":0,"playbackend":1,"api":[7]},"displaymanager":"PubMatic_OpenWrap_SDK","instl":0}],"at":1,"cur":["USD"],"regs":{"coppa":1,"ext":{"ccpa":0,"gdpr":1,"gpp":"gpp_string","gpp_sid":[7],"us_privacy":"uspConsentString","consent":"0"}}}`,
				maxRequest: json.RawMessage(`{"id":"{BID_ID}","at":1,"bcat":["IAB26-4","IAB26-2","IAB25-6","IAB25-5","IAB25-4","IAB25-3","IAB25-1","IAB25-7","IAB8-18","IAB26-3","IAB26-1","IAB8-5","IAB25-2","IAB11-4"],"tmax":3000,"app":{"name":"DrawHappyAngel","ver":"0.5.4","bundle":"com.newstory.DrawHappyAngel","cat":["IAB9-30"],"id":"{NETWORK_APP_ID}","publisher":{"name":"New Story Inc.","ext":{"installed_sdk":{"id":"MOLOCO_BIDDING","sdk_version":{"major":1,"minor":0,"micro":0},"adapter_version":{"major":1,"minor":0,"micro":0}}}},"ext":{"orientation":1}},"device":{"ifa":"497a10d6-c4dd-4e04-a986-c32b7180d462","ip":"38.158.207.171","carrier":"MYTEL","language":"en_US","hwv":"ruby","ppi":440,"pxratio":2.75,"devicetype":4,"connectiontype":2,"js":1,"h":2400,"w":1080,"geo":{"type":2,"ipservice":3,"lat":40.7429,"lon":-73.9392,"long":-73.9392,"city":"Queens","country":"USA","region":"ny","dma":"501","metro":"501","zip":"11101","ext":{"org":"Myanmar Broadband Telecom Co.","isp":"Myanmar Broadband Telecom Co."}},"ext":{},"osv":"13.0.0","ua":"Mozilla/5.0 (Linux; Android 13; 22101316C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.230 Mobile Safari/537.36","make":"xiaomi","model":"22101316c","os":"android"},"imp":[{"id":"1","displaymanager":"applovin_mediation","displaymanagerver":"11.8.2","instl":0,"secure":0,"tagid":"{NETWORK_PLACEMENT_ID}","bidfloor":0.01,"bidfloorcur":"USD","exp":14400,"banner":{"id":"1","w":320,"h":50,"btype":[],"battr":[1,2,5,8,9,14,17],"pos":1,"format":[{"w":320,"h":50}],"ext":{"bannertype":"rewarded"}},"rwdd":0}],"user":{"data":[{"id":"1","name":"Publisher Passed","segment":[{"signal":"{BIDDING_SIGNAL}"}]}],"ext":{"gdpr":0}},"regs":{"ext":{"ccpa":0,"gdpr":1,"consent":"0","tcf_consent_string":"{TCF_STRING}"}},"source":{"ext":{"schain":{"ver":"1.0","complete":1,"nodes":[{"asi":"applovin.com","sid":"53bf468f18c5a0e2b7d4e3f748c677c1","rid":"494dbe15a3ce08c54f4e456363f35a022247f997","hp":1}]}}}}`),
			},
			wantMaxRequest: json.RawMessage(`{"id":"{BID_ID}","at":1,"bcat":["IAB26-4","IAB26-2","IAB25-6","IAB25-5","IAB25-4","IAB25-3","IAB25-1","IAB25-7","IAB8-18","IAB26-3","IAB26-1","IAB8-5","IAB25-2","IAB11-4"],"tmax":3000,"app":{"paid":1,"keywords":"k1=v1","domain":"abc.com","name":"DrawHappyAngel","ver":"0.5.4","bundle":"com.newstory.DrawHappyAngel","cat":["IAB9-30"],"id":"{NETWORK_APP_ID}","publisher":{"name":"New Story Inc.","ext":{"installed_sdk":{"id":"MOLOCO_BIDDING","sdk_version":{"major":1,"minor":0,"micro":0},"adapter_version":{"major":1,"minor":0,"micro":0}}}},"ext":{"orientation":1}},"device":{"ifa":"497a10d6-c4dd-4e04-a986-c32b7180d462","ip":"38.158.207.171","carrier":"MYTEL","language":"en_US","hwv":"ruby","ppi":440,"pxratio":2.75,"devicetype":4,"mccmnc":"311-480","connectiontype":5,"js":1,"h":2400,"w":1080,"geo":{"city":"Queens","type":2,"ipservice":3,"lat":40.7429,"lon":-73.9392,"long":-73.9392,"country":"USA","region":"ny","dma":"501","metro":"501","zip":"11101","utcoffset":-480,"ext":{"org":"Myanmar Broadband Telecom Co.","isp":"Myanmar Broadband Telecom Co."}},"ext":{"atts":3},"osv":"13.0.0","ua":"Mozilla/5.0 (Linux; Android 13; 22101316C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.230 Mobile Safari/537.36","make":"xiaomi","model":"22101316c","os":"android"},"imp":[{"id":"1","displaymanagerver":"3.1.0","clickbrowser":1,"displaymanager":"PubMatic_OpenWrap_SDK","instl":0,"secure":0,"tagid":"{NETWORK_PLACEMENT_ID}","bidfloor":0.01,"bidfloorcur":"USD","exp":14400,"rwdd":0,"video":{"companionad":[{"pos":0,"format":[{"w":300,"h":250}],"vcm":1}],"protocols":[2,3,5,6,7,8,11,12,13,14],"h":250,"w":300,"linearity":1,"pos":0,"boxingallowed":1,"placement":2,"mimes":["video/3gpp2","video/quicktime","video/mp4","video/x-m4v","video/3gpp"],"companiontype":[1,2,3],"delivery":[2],"startdelay":0,"playbackend":1,"api":[7]},"ext":{"skadn":{"sourceapp":"406719683","versions":["2.0","2.1","2.2","3.0","4.0"],"skadnetids":["cstr6suwn9.skadnetwork","7ug5zh24hu.skadnetwork","uw77j35x4d.skadnetwork","c6k4g5qg8m.skadnetwork","hs6bdukanm.skadnetwork","yclnxrl5pm.skadnetwork","3sh42y64q3.skadnetwork","cj5566h2ga.skadnetwork","klf5c3l5u5.skadnetwork","8s468mfl3y.skadnetwork","2u9pt9hc89.skadnetwork","7rz58n8ntl.skadnetwork","ppxm28t8ap.skadnetwork","mtkv5xtk9e.skadnetwork","cg4yq2srnc.skadnetwork","wzmmz9fp6w.skadnetwork","k674qkevps.skadnetwork","v72qych5uu.skadnetwork","578prtvx9j.skadnetwork","3rd42ekr43.skadnetwork","g28c52eehv.skadnetwork","2fnua5tdw4.skadnetwork","9nlqeag3gk.skadnetwork","5lm9lj6jb7.skadnetwork","97r2b46745.skadnetwork","e5fvkxwrpn.skadnetwork","4pfyvq9l8r.skadnetwork","tl55sbb4fm.skadnetwork","t38b2kh725.skadnetwork","prcb7njmu6.skadnetwork","mlmmfzh3r3.skadnetwork","9t245vhmpl.skadnetwork","9rd848q2bz.skadnetwork","4fzdc2evr5.skadnetwork","4468km3ulz.skadnetwork","m8dbw4sv7c.skadnetwork","ejvt5qm6ak.skadnetwork","5lm9lj6jb7.skadnetwork","44jx6755aq.skadnetwork","6g9af3uyq4.skadnetwork","u679fj5vs4.skadnetwork","rx5hdcabgc.skadnetwork","275upjj5gd.skadnetwork","p78axxw29g.skadnetwork"],"productpage":1,"version":"2.0"}}}],"user":{"data":[{"id":"1","name":"Publisher Passed","segment":[{"signal":"{BIDDING_SIGNAL}"}]}],"ext":{"gdpr":0}},"regs":{"coppa":1,"ext":{"ccpa":0,"gdpr":1,"consent":"0","tcf_consent_string":"{TCF_STRING}","gpp":"gpp_string","gpp_sid":[7],"us_privacy":"uspConsentString"}},"source":{"ext":{"schain":{"ver":"1.0","complete":1,"nodes":[{"asi":"applovin.com","sid":"53bf468f18c5a0e2b7d4e3f748c677c1","rid":"494dbe15a3ce08c54f4e456363f35a022247f997","hp":1}]},"omidpn":"Pubmatic","omidpv":"3.1.0"}}}`),
		},
		{
			name: "replace imp.video from signal",
			args: args{
				signal:     `{"device":{"devicetype":4,"w":393,"h":852,"ifa":"F5BA1637-7156-4369-BA7E-3C45033D9F61","mccmnc":"311-480","js":1,"osv":"17.3.1","connectiontype":5,"os":"iOS","pxratio":3,"geo":{"lastfix":8,"lat":37.48773508935608,"utcoffset":-480,"lon":-122.22855027909678,"type":1},"language":"en","make":"Apple","ext":{"atts":3},"ua":"Mozilla/5.0 (iPhone; CPU iPhone OS 17_3_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148","model":"iPhone15,2","carrier":"Verizon"},"source":{"ext":{"omidpn":"Pubmatic","omidpv":"3.1.0"}},"id":"CE204A0E-31C3-4D7F-A1A0-D34AF5ED1A7F","app":{"id":"406719683","paid":1,"keywords":"k1=v1","domain":"abc.com","bundle":"406719683","storeurl":"https://apps.apple.com/us/app/gasbuddy-find-pay-for-gas/id406719683","name":"GasBuddy","publisher":{"id":"160361"},"ver":"700.89.22927"},"ext":{"wrapper":{"sumry_disable":1,"clientconfig":1,"profileid":3422}},"imp":[{"secure":1,"tagid":"Mobile_iPhone_List_Screen_Bottom","banner":{"pos":0,"format":[{"w":300,"h":250}],"api":[5,6,7]},"id":"98D9318E-5276-402F-BAA4-CDBD8A364957","ext":{"skadn":{"sourceapp":"406719683","versions":["2.0","2.1","2.2","3.0","4.0"],"skadnetids":["cstr6suwn9.skadnetwork","7ug5zh24hu.skadnetwork","uw77j35x4d.skadnetwork","c6k4g5qg8m.skadnetwork","hs6bdukanm.skadnetwork","yclnxrl5pm.skadnetwork","3sh42y64q3.skadnetwork","cj5566h2ga.skadnetwork","klf5c3l5u5.skadnetwork","8s468mfl3y.skadnetwork","2u9pt9hc89.skadnetwork","7rz58n8ntl.skadnetwork","ppxm28t8ap.skadnetwork","mtkv5xtk9e.skadnetwork","cg4yq2srnc.skadnetwork","wzmmz9fp6w.skadnetwork","k674qkevps.skadnetwork","v72qych5uu.skadnetwork","578prtvx9j.skadnetwork","3rd42ekr43.skadnetwork","g28c52eehv.skadnetwork","2fnua5tdw4.skadnetwork","9nlqeag3gk.skadnetwork","5lm9lj6jb7.skadnetwork","97r2b46745.skadnetwork","e5fvkxwrpn.skadnetwork","4pfyvq9l8r.skadnetwork","tl55sbb4fm.skadnetwork","t38b2kh725.skadnetwork","prcb7njmu6.skadnetwork","mlmmfzh3r3.skadnetwork","9t245vhmpl.skadnetwork","9rd848q2bz.skadnetwork","4fzdc2evr5.skadnetwork","4468km3ulz.skadnetwork","m8dbw4sv7c.skadnetwork","ejvt5qm6ak.skadnetwork","5lm9lj6jb7.skadnetwork","44jx6755aq.skadnetwork","6g9af3uyq4.skadnetwork","u679fj5vs4.skadnetwork","rx5hdcabgc.skadnetwork","275upjj5gd.skadnetwork","p78axxw29g.skadnetwork"],"productpage":1,"version":"2.0"}},"displaymanagerver":"3.1.0","clickbrowser":1,"video":{"companionad":[{"pos":0,"format":[{"w":300,"h":250}],"vcm":1}],"protocols":[2,3,5,6,7,8,11,12,13,14],"h":250,"w":300,"linearity":1,"pos":0,"boxingallowed":1,"placement":2,"mimes":["video/3gpp2","video/quicktime","video/mp4","video/x-m4v","video/3gpp"],"companiontype":[1,2,3],"delivery":[2],"startdelay":0,"playbackend":1,"api":[7]},"displaymanager":"PubMatic_OpenWrap_SDK","instl":0}],"at":1,"cur":["USD"],"regs":{"coppa":1,"ext":{"ccpa":0,"gdpr":1,"gpp":"gpp_string","gpp_sid":[7],"us_privacy":"uspConsentString","consent":"0"}}}`,
				maxRequest: json.RawMessage(`{"id":"{BID_ID}","at":1,"bcat":["IAB26-4","IAB26-2","IAB25-6","IAB25-5","IAB25-4","IAB25-3","IAB25-1","IAB25-7","IAB8-18","IAB26-3","IAB26-1","IAB8-5","IAB25-2","IAB11-4"],"tmax":3000,"app":{"name":"DrawHappyAngel","ver":"0.5.4","bundle":"com.newstory.DrawHappyAngel","cat":["IAB9-30"],"id":"{NETWORK_APP_ID}","publisher":{"name":"New Story Inc.","ext":{"installed_sdk":{"id":"MOLOCO_BIDDING","sdk_version":{"major":1,"minor":0,"micro":0},"adapter_version":{"major":1,"minor":0,"micro":0}}}},"ext":{"orientation":1}},"device":{"ifa":"497a10d6-c4dd-4e04-a986-c32b7180d462","ip":"38.158.207.171","carrier":"MYTEL","language":"en_US","hwv":"ruby","ppi":440,"pxratio":2.75,"devicetype":4,"connectiontype":2,"js":1,"h":2400,"w":1080,"geo":{"type":2,"ipservice":3,"lat":40.7429,"lon":-73.9392,"long":-73.9392,"city":"Queens","country":"USA","region":"ny","dma":"501","metro":"501","zip":"11101","ext":{"org":"Myanmar Broadband Telecom Co.","isp":"Myanmar Broadband Telecom Co."}},"ext":{},"osv":"13.0.0","ua":"Mozilla/5.0 (Linux; Android 13; 22101316C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.230 Mobile Safari/537.36","make":"xiaomi","model":"22101316c","os":"android"},"imp":[{"id":"1","displaymanager":"applovin_mediation","displaymanagerver":"11.8.2","instl":1,"secure":0,"tagid":"{NETWORK_PLACEMENT_ID}","exp":14400,"banner":{"id":"1","w":320,"h":480,"btype":[],"battr":[1,2,5,8,9,14,17],"pos":7,"format":[{"w":320,"h":480}]},"video":{"w":320,"h":480,"battr":[1,2,5,8,9,14,17],"mimes":["video/mp4","video/3gpp","video/3gpp2","video/x-m4v"],"placement":5,"pos":7,"minduration":5,"maxduration":60,"skipafter":5,"skipmin":0,"startdelay":0,"playbackmethod":[1],"linearity":1},"rwdd":0}],"user":{"data":[{"id":"1","name":"Publisher Passed","segment":[{"signal":"{BIDDING_SIGNAL}"}]}],"ext":{"gdpr":0}},"regs":{"ext":{"ccpa":0,"gdpr":1,"consent":"0","tcf_consent_string":"{TCF_STRING}"}},"source":{"ext":{"schain":{"ver":"1.0","complete":1,"nodes":[{"asi":"applovin.com","sid":"53bf468f18c5a0e2b7d4e3f748c677c1","rid":"494dbe15a3ce08c54f4e456363f35a022247f997","hp":1}]}}}}`),
			},
			wantMaxRequest: json.RawMessage(`{"id":"{BID_ID}","at":1,"bcat":["IAB26-4","IAB26-2","IAB25-6","IAB25-5","IAB25-4","IAB25-3","IAB25-1","IAB25-7","IAB8-18","IAB26-3","IAB26-1","IAB8-5","IAB25-2","IAB11-4"],"tmax":3000,"app":{"paid":1,"keywords":"k1=v1","domain":"abc.com","name":"DrawHappyAngel","ver":"0.5.4","bundle":"com.newstory.DrawHappyAngel","cat":["IAB9-30"],"id":"{NETWORK_APP_ID}","publisher":{"name":"New Story Inc.","ext":{"installed_sdk":{"id":"MOLOCO_BIDDING","sdk_version":{"major":1,"minor":0,"micro":0},"adapter_version":{"major":1,"minor":0,"micro":0}}}},"ext":{"orientation":1}},"device":{"ifa":"497a10d6-c4dd-4e04-a986-c32b7180d462","ip":"38.158.207.171","carrier":"MYTEL","language":"en_US","hwv":"ruby","ppi":440,"pxratio":2.75,"devicetype":4,"mccmnc":"311-480","connectiontype":5,"js":1,"h":2400,"w":1080,"geo":{"city":"Queens","type":2,"ipservice":3,"lat":40.7429,"lon":-73.9392,"long":-73.9392,"country":"USA","region":"ny","dma":"501","metro":"501","zip":"11101","utcoffset":-480,"ext":{"org":"Myanmar Broadband Telecom Co.","isp":"Myanmar Broadband Telecom Co."}},"ext":{"atts":3},"osv":"13.0.0","ua":"Mozilla/5.0 (Linux; Android 13; 22101316C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.230 Mobile Safari/537.36","make":"xiaomi","model":"22101316c","os":"android"},"imp":[{"id":"1","displaymanagerver":"3.1.0","clickbrowser":1,"displaymanager":"PubMatic_OpenWrap_SDK","instl":1,"secure":0,"tagid":"{NETWORK_PLACEMENT_ID}","exp":14400,"banner":{"id":"1","w":320,"h":480,"btype":[],"api":[5,6,7],"battr":[1,2,5,8,9,14,17],"pos":7,"format":[{"w":320,"h":480}]},"video":{"companionad":[{"pos":0,"format":[{"w":300,"h":250}],"vcm":1}],"protocols":[2,3,5,6,7,8,11,12,13,14],"h":250,"w":300,"linearity":1,"pos":0,"boxingallowed":1,"placement":2,"mimes":["video/3gpp2","video/quicktime","video/mp4","video/x-m4v","video/3gpp"],"companiontype":[1,2,3],"delivery":[2],"startdelay":0,"playbackend":1,"api":[7]},"rwdd":0,"ext":{"skadn":{"sourceapp":"406719683","versions":["2.0","2.1","2.2","3.0","4.0"],"skadnetids":["cstr6suwn9.skadnetwork","7ug5zh24hu.skadnetwork","uw77j35x4d.skadnetwork","c6k4g5qg8m.skadnetwork","hs6bdukanm.skadnetwork","yclnxrl5pm.skadnetwork","3sh42y64q3.skadnetwork","cj5566h2ga.skadnetwork","klf5c3l5u5.skadnetwork","8s468mfl3y.skadnetwork","2u9pt9hc89.skadnetwork","7rz58n8ntl.skadnetwork","ppxm28t8ap.skadnetwork","mtkv5xtk9e.skadnetwork","cg4yq2srnc.skadnetwork","wzmmz9fp6w.skadnetwork","k674qkevps.skadnetwork","v72qych5uu.skadnetwork","578prtvx9j.skadnetwork","3rd42ekr43.skadnetwork","g28c52eehv.skadnetwork","2fnua5tdw4.skadnetwork","9nlqeag3gk.skadnetwork","5lm9lj6jb7.skadnetwork","97r2b46745.skadnetwork","e5fvkxwrpn.skadnetwork","4pfyvq9l8r.skadnetwork","tl55sbb4fm.skadnetwork","t38b2kh725.skadnetwork","prcb7njmu6.skadnetwork","mlmmfzh3r3.skadnetwork","9t245vhmpl.skadnetwork","9rd848q2bz.skadnetwork","4fzdc2evr5.skadnetwork","4468km3ulz.skadnetwork","m8dbw4sv7c.skadnetwork","ejvt5qm6ak.skadnetwork","5lm9lj6jb7.skadnetwork","44jx6755aq.skadnetwork","6g9af3uyq4.skadnetwork","u679fj5vs4.skadnetwork","rx5hdcabgc.skadnetwork","275upjj5gd.skadnetwork","p78axxw29g.skadnetwork"],"productpage":1,"version":"2.0"}}}],"user":{"data":[{"id":"1","name":"Publisher Passed","segment":[{"signal":"{BIDDING_SIGNAL}"}]}],"ext":{"gdpr":0}},"regs":{"coppa":1,"ext":{"ccpa":0,"gdpr":1,"consent":"0","tcf_consent_string":"{TCF_STRING}","gpp":"gpp_string","gpp_sid":[7],"us_privacy":"uspConsentString"}},"source":{"ext":{"schain":{"ver":"1.0","complete":1,"nodes":[{"asi":"applovin.com","sid":"53bf468f18c5a0e2b7d4e3f748c677c1","rid":"494dbe15a3ce08c54f4e456363f35a022247f997","hp":1}]},"omidpn":"Pubmatic","omidpv":"3.1.0"}}}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var maxRequest openrtb2.BidRequest
			if err := json.Unmarshal(tt.args.maxRequest, &maxRequest); err != nil {
				t.Errorf("Unmarshal Faild for Incoming MaxRequest, Error: %s", err)
			}

			var expectedMaxRequest openrtb2.BidRequest
			addSignalDataInRequest(tt.args.signal, &maxRequest, tt.args.clientconfigflag)
			if err := json.Unmarshal(tt.wantMaxRequest, &expectedMaxRequest); err != nil {
				t.Errorf("Unmarshal Faild for Expected MaxRequest, Error: %s", err)
			}
			assert.Equal(t, expectedMaxRequest, maxRequest, tt.name)
		})
	}
}

func Test_getSignalData(t *testing.T) {
	type args struct {
		requestBody []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "incorrect body",
			args: args{
				requestBody: []byte(`{"id":"123","user":Passed","segment":[{"signal":{BIDDING_SIGNA}]}],"ext":{"gdpr":0}}}`),
			},
			want: "",
		},
		{
			name: "single user.data with signal with incorrect signal",
			args: args{
				requestBody: []byte(`{"id":"123","user":{"data":[{"id":"1","name":"Publisher Passed","segment":[{"signal":{BIDDING_SIGNA}]}],"ext":{"gdpr":0}}}`),
			},
			want: "",
		},
		{
			name: "single user.data with signal",
			args: args{
				requestBody: []byte(`{"id":"123","user":{"data":[{"id":"1","name":"Publisher Passed","segment":[{"signal":"{BIDDING_SIGNAL}"}]}],"ext":{"gdpr":0}}}`),
			},
			want: "{BIDDING_SIGNAL}",
		},
		{
			name: "multiple user.data with signal",
			args: args{
				requestBody: []byte(`{"id":"123","user":{"data":[{"id":"1","name":"Publisher Passed","segment":[{"signal":"{BIDDING_SIGNAL}"}]},{"id":"1","name":"PubMatic","segment":[{"signal":"{NEW_SIGNAL}"}]}],"ext":{"gdpr":0}}}`),
			},
			want: "{BIDDING_SIGNAL}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSignalData(tt.args.requestBody); got != tt.want {
				t.Errorf("getSignalData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateMaxResponse(t *testing.T) {
	type args struct {
		rctx        models.RequestCtx
		bidResponse *openrtb2.BidResponse
	}
	tests := []struct {
		name string
		args args
		want *openrtb2.BidResponse
	}{
		{
			name: "bidresponse seatbid is empty",
			args: args{
				rctx: models.RequestCtx{
					Debug: true,
				},
				bidResponse: &openrtb2.BidResponse{
					ID: "123",
				},
			},
			want: &openrtb2.BidResponse{
				ID: models.MaxRejected,
			},
		},
		{
			name: "bidresponse contains NBR and debug is enabled",
			args: args{
				rctx: models.RequestCtx{
					Debug: true,
				},
				bidResponse: &openrtb2.BidResponse{
					ID:  "123",
					NBR: ptrutil.ToPtr(nbr.InvalidPlatform),
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:    "456",
									ImpID: "789",
								},
							},
						},
					},
				},
			},
			want: &openrtb2.BidResponse{
				ID:  "123",
				NBR: ptrutil.ToPtr(nbr.InvalidPlatform),
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:    "456",
								ImpID: "789",
							},
						},
					},
				},
			},
		},
		{
			name: "bidresponse contains NBR and debug is disabled",
			args: args{
				rctx: models.RequestCtx{
					Debug: false,
				},
				bidResponse: &openrtb2.BidResponse{
					ID:  "123",
					NBR: ptrutil.ToPtr(nbr.InvalidPlatform),
				},
			},
			want: &openrtb2.BidResponse{
				ID: models.MaxRejected,
			},
		},
		{
			name: "bidresponse with no seatbid",
			args: args{
				rctx: models.RequestCtx{
					Debug: true,
				},
				bidResponse: &openrtb2.BidResponse{
					ID:  "123",
					Ext: json.RawMessage(`{"key":"value"}`),
				},
			},
			want: &openrtb2.BidResponse{
				ID: models.MaxRejected,
			},
		},
		{
			name: "failed to marshal bidresponse",
			args: args{
				rctx: models.RequestCtx{
					Debug: true,
				},
				bidResponse: &openrtb2.BidResponse{
					ID:    "123",
					BidID: "456",
					Ext:   json.RawMessage(`{`),
				},
			},
			want: &openrtb2.BidResponse{
				ID: "max_rejected",
			},
		},
		{
			name: "valid bidresponse",
			args: args{
				rctx: models.RequestCtx{
					Debug: true,
				},
				bidResponse: &openrtb2.BidResponse{
					ID:    "123",
					BidID: "456",
					Cur:   "USD",
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:    "456",
									ImpID: "789",
									Price: 1.0,
									BURL:  "http://example.com",
									Ext:   json.RawMessage(`{"key":"value"}`),
								},
							},
						},
					},
				},
			},
			want: &openrtb2.BidResponse{
				ID:    "123",
				BidID: "456",
				Cur:   "USD",
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:    "456",
								ImpID: "789",
								Price: 1.0,
								BURL:  "http://example.com",
								Ext:   json.RawMessage(`{"signaldata":"{\"id\":\"123\",\"seatbid\":[{\"bid\":[{\"id\":\"456\",\"impid\":\"789\",\"price\":1,\"burl\":\"http://example.com\",\"ext\":{\"key\":\"value\"}}]}],\"bidid\":\"456\",\"cur\":\"USD\"}"}`),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := updateMaxApplovinResponse(tt.args.rctx, tt.args.bidResponse)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
