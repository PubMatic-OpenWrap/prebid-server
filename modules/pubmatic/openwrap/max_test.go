package openwrap

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v19/adcom1"
	"github.com/prebid/openrtb/v19/openrtb2"
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
				sdkImpression: openrtb2.Imp{ClickBrowser: 0, DisplayManager: "PubMatic_SDK", DisplayManagerVer: "1.2"},
				maxImpression: &openrtb2.Imp{ClickBrowser: 1, DisplayManager: "Applovin_SDK"},
			},
			want: &openrtb2.Imp{ClickBrowser: 0, DisplayManager: "PubMatic_SDK", DisplayManagerVer: "1.2"},
		},
		{
			name: "only maxImp has video",
			args: args{
				sdkImpression: openrtb2.Imp{ClickBrowser: 0, DisplayManager: "PubMatic_SDK", DisplayManagerVer: "1.2"},
				maxImpression: &openrtb2.Imp{ClickBrowser: 1, DisplayManager: "Applovin_SDK", Video: &openrtb2.Video{W: 300, H: 250}},
			},
			want: &openrtb2.Imp{ClickBrowser: 0, DisplayManager: "PubMatic_SDK", DisplayManagerVer: "1.2", Video: &openrtb2.Video{W: 300, H: 250}},
		},
		{
			name: "maxImp and sdkImp has video",
			args: args{
				sdkImpression: openrtb2.Imp{ClickBrowser: 0, DisplayManager: "PubMatic_SDK", DisplayManagerVer: "1.2", Video: &openrtb2.Video{W: 300, H: 250, BAttr: []adcom1.CreativeAttribute{1, 2}}},
				maxImpression: &openrtb2.Imp{ClickBrowser: 1, DisplayManager: "Applovin_SDK", Video: &openrtb2.Video{W: 750, H: 500, BAttr: []adcom1.CreativeAttribute{6, 1, 8, 4}}},
			},
			want: &openrtb2.Imp{ClickBrowser: 0, DisplayManager: "PubMatic_SDK", DisplayManagerVer: "1.2", Video: &openrtb2.Video{W: 300, H: 250, BAttr: []adcom1.CreativeAttribute{6, 1, 8, 4}}},
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
				sdkApp:     &openrtb2.App{Paid: 1, Keywords: "k1=v1", Domain: "abc.com"},
				maxRequest: &openrtb2.BidRequest{},
			},
			want: &openrtb2.App{Paid: 1, Keywords: "k1=v1", Domain: "abc.com"},
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
			want: nil,
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
			name: "sdkRegsExt has gdpr, gpp",
			args: args{
				sdkRegs:    &openrtb2.Regs{Ext: json.RawMessage(`{"gdpr":1,"gpp":"sdfewe3cer"}`)},
				maxRequest: &openrtb2.BidRequest{},
			},
			want: &openrtb2.Regs{Ext: json.RawMessage(`{"gdpr":1,"gpp":sdfewe3cer}`)},
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
			want: &openrtb2.Source{Ext: json.RawMessage(`{"omidpn":MyIntegrationPartner,"omidpv":7.1}`)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateSource(tt.args.sdkSource, tt.args.maxRequest)
			assert.Equal(t, tt.want, tt.args.maxRequest.Source, tt.name)
		})
	}
}
