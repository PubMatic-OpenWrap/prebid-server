package pubmatic

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func TestLogDeviceObject(t *testing.T) {
	type args struct {
		ortbBidRequest *openrtb2.BidRequest
		rctx           *models.RequestCtx
	}

	type want struct {
		device Device
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: `Nil request`,
			args: args{
				rctx: &models.RequestCtx{
					DevicePlatform: models.DevicePlatformDesktop,
				},
			},
			want: want{
				device: Device{
					Platform: models.DevicePlatformDesktop,
				},
			},
		},
		{
			name: `Empty uaFromHTTPReq`,
			args: args{
				ortbBidRequest: &openrtb2.BidRequest{},
				rctx: &models.RequestCtx{
					DevicePlatform: models.DevicePlatformMobileWeb,
				},
			},
			want: want{
				device: Device{
					Platform: models.DevicePlatformMobileWeb,
				},
			},
		},
		{
			name: `Invalid device ext`,
			args: args{
				ortbBidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						Ext: json.RawMessage(`invalid ext`),
					},
				},
				rctx: &models.RequestCtx{
					DevicePlatform: models.DevicePlatformMobileWeb,
				},
			},
			want: want{
				device: Device{
					Platform: 0,
				},
			},
		},
		{
			name: `IFA Type key absent`,
			args: args{
				ortbBidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						Ext: json.RawMessage(`{"anykey":"anyval"}`),
					},
				},
				rctx: &models.RequestCtx{
					DevicePlatform: models.DevicePlatformMobileWeb,
				},
			},
			want: want{
				device: Device{
					Platform: models.DevicePlatformMobileWeb,
				},
			},
		},
		{
			name: `Invalid data type for ifa_type key`,
			args: args{
				ortbBidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						Ext: json.RawMessage(`{"ifa_type": 123}`)},
				},
				rctx: &models.RequestCtx{
					DevicePlatform: models.DevicePlatformMobileWeb,
				},
			},
			want: want{
				device: Device{
					Platform: models.DevicePlatformMobileWeb,
				},
			},
		},
		{
			name: `ifa_type missing in DeviceIFATypeID mapping`,
			args: args{
				ortbBidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						Ext: json.RawMessage(`{"ifa_type": "anything"}`),
					},
				},
				rctx: &models.RequestCtx{
					DevicePlatform: models.DevicePlatformMobileWeb,
				},
			},
			want: want{
				device: Device{
					Platform: models.DevicePlatformMobileWeb,
					IFAType:  ptrutil.ToPtr(0),
				},
			},
		},
		{
			name: `Case insensitive ifa_type`,
			args: args{
				ortbBidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						Ext: json.RawMessage(`{"ifa_type": "DpId"}`),
					},
				},
				rctx: &models.RequestCtx{
					DevicePlatform: models.DevicePlatformMobileWeb,
				},
			},
			want: want{
				device: Device{
					Platform: models.DevicePlatformMobileWeb,
					IFAType:  ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeDPID]),
				},
			},
		},
		{
			name: `Valid ifa_type`,
			args: args{
				ortbBidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						Ext: json.RawMessage(`{"ifa_type": "sessionid"}`),
					},
				},
				rctx: &models.RequestCtx{
					DevicePlatform: models.DevicePlatformMobileWeb,
				},
			},
			want: want{
				device: Device{
					Platform: models.DevicePlatformMobileWeb,
					IFAType:  ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeSESSIONID]),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wlog := &WloggerRecord{}
			wlog.logDeviceObject(tt.args.rctx, tt.args.ortbBidRequest)
			assert.Equal(t, tt.want.device, wlog.Device)
		})
	}
}

func TestLogIntegrationType(t *testing.T) {

	tests := []struct {
		name            string
		endpoint        string
		integrationType string
	}{
		{
			name:            "sdk",
			endpoint:        models.EndpointV25,
			integrationType: models.TypeSDK,
		},
		{
			name:            "amp",
			endpoint:        models.EndpointAMP,
			integrationType: models.TypeAmp,
		},
		{
			name:            "ctv-vast",
			endpoint:        models.EndpointVAST,
			integrationType: models.TypeTag,
		},
		{
			name:            "ctv-ortb",
			endpoint:        models.EndpointORTB,
			integrationType: models.TypeS2S,
		},
		{
			name:            "ctv-json",
			endpoint:        models.EndpointJson,
			integrationType: models.TypeInline,
		},
		{
			name:            "openrtb-video",
			endpoint:        models.EndpointVideo,
			integrationType: models.TypeInline,
		},
		{
			name:            "invalid",
			endpoint:        "invalid",
			integrationType: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wlog := WloggerRecord{}
			wlog.logIntegrationType(tt.endpoint)
			assert.Equal(t, tt.integrationType, wlog.IntegrationType, tt.name)
		})
	}
}

func TestLogFloorType(t *testing.T) {

	tests := []struct {
		name      string
		prebidExt *openrtb_ext.ExtRequestPrebid
		floorType int
	}{
		{
			name:      "Nil prebidExt",
			prebidExt: nil,
			floorType: models.SoftFloor,
		},
		{
			name:      "Nil prebidExt.Floors",
			prebidExt: &openrtb_ext.ExtRequestPrebid{},
			floorType: models.SoftFloor,
		},
		{
			name: "Nil prebidExt.Floors.Enabled",
			prebidExt: &openrtb_ext.ExtRequestPrebid{
				Floors: &openrtb_ext.PriceFloorRules{},
			},
			floorType: models.SoftFloor,
		},
		{
			name: "false prebidExt.Floors.Enabled",
			prebidExt: &openrtb_ext.ExtRequestPrebid{
				Floors: &openrtb_ext.PriceFloorRules{
					Enabled: ptrutil.ToPtr(false),
				},
			},
			floorType: models.SoftFloor,
		},
		{
			name: "Nil prebidExt.Floors.Enabled.Enforcement",
			prebidExt: &openrtb_ext.ExtRequestPrebid{
				Floors: &openrtb_ext.PriceFloorRules{
					Enabled:     ptrutil.ToPtr(true),
					Enforcement: nil,
				},
			},
			floorType: models.SoftFloor,
		},
		{
			name: "Nil prebidExt.Floors.Enabled.Enforcement.EnforcePBS",
			prebidExt: &openrtb_ext.ExtRequestPrebid{
				Floors: &openrtb_ext.PriceFloorRules{
					Enabled: ptrutil.ToPtr(true),
					Enforcement: &openrtb_ext.PriceFloorEnforcement{
						EnforcePBS: nil,
					},
				},
			},
			floorType: models.SoftFloor,
		},
		{
			name: "false prebidExt.Floors.Enabled.Enforcement.EnforcePBS",
			prebidExt: &openrtb_ext.ExtRequestPrebid{
				Floors: &openrtb_ext.PriceFloorRules{
					Enabled: ptrutil.ToPtr(true),
					Enforcement: &openrtb_ext.PriceFloorEnforcement{
						EnforcePBS: ptrutil.ToPtr(false),
					},
				},
			},
			floorType: models.SoftFloor,
		},
		{
			name: "true prebidExt.Floors.Enabled.Enforcement.EnforcePBS",
			prebidExt: &openrtb_ext.ExtRequestPrebid{
				Floors: &openrtb_ext.PriceFloorRules{
					Enabled: ptrutil.ToPtr(true),
					Enforcement: &openrtb_ext.PriceFloorEnforcement{
						EnforcePBS: ptrutil.ToPtr(true),
					},
				},
			},
			floorType: models.HardFloor,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wlog := WloggerRecord{}
			wlog.logFloorType(tt.prebidExt)
			assert.Equal(t, tt.floorType, wlog.FloorType, tt.name)
		})
	}
}

func TestLogContentObject(t *testing.T) {
	type args struct {
		content *openrtb2.Content
	}
	tests := []struct {
		name string
		args args
		want *Content
	}{
		{
			name: "Empty",
			args: args{},
			want: nil,
		},
		{
			name: "OnlyID",
			args: args{
				content: &openrtb2.Content{
					ID: "ID",
				},
			},
			want: &Content{
				ID: "ID",
			},
		},
		{
			name: "OnlyEpisode",
			args: args{
				content: &openrtb2.Content{
					Episode: 123,
				},
			},
			want: &Content{
				Episode: 123,
			},
		},
		{
			name: "OnlyTitle",
			args: args{
				content: &openrtb2.Content{
					Title: "Title",
				},
			},
			want: &Content{
				Title: "Title",
			},
		},
		{
			name: "OnlySeries",
			args: args{
				content: &openrtb2.Content{
					Series: "Series",
				},
			},
			want: &Content{
				Series: "Series",
			},
		},
		{
			name: "OnlySeason",
			args: args{
				content: &openrtb2.Content{
					Season: "Season",
				},
			},
			want: &Content{
				Season: "Season",
			},
		},
		{
			name: "OnlyCat",
			args: args{
				content: &openrtb2.Content{
					Cat: []string{"CAT-1"},
				},
			},
			want: &Content{
				Cat: []string{"CAT-1"},
			},
		},
		{
			name: "AllPresent",
			args: args{
				content: &openrtb2.Content{
					ID:      "ID",
					Episode: 123,
					Title:   "Title",
					Series:  "Series",
					Season:  "Season",
					Cat:     []string{"CAT-1"},
				},
			},
			want: &Content{
				ID:      "ID",
				Episode: 123,
				Title:   "Title",
				Series:  "Series",
				Season:  "Season",
				Cat:     []string{"CAT-1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wlog := &WloggerRecord{}
			wlog.logContentObject(tt.args.content)
			assert.Equal(t, tt.want, wlog.Content)
		})
	}
}

func TestSetMetaDataObject(t *testing.T) {
	type args struct {
		meta          *openrtb_ext.ExtBidPrebidMeta
		partnerRecord *PartnerRecord
	}
	tests := []struct {
		name          string
		args          args
		partnerRecord *PartnerRecord
	}{
		{
			name: "NetworkID 0, AdvertiserID 0, SecondaryCategoryIDs size 0",
			args: args{
				meta: &openrtb_ext.ExtBidPrebidMeta{
					NetworkID:            0,
					AdvertiserID:         0,
					SecondaryCategoryIDs: []string{},
				},
				partnerRecord: &PartnerRecord{
					PartnerID: "pubmatic",
				},
			},
			partnerRecord: &PartnerRecord{
				PartnerID: "pubmatic",
			},
		},
		{
			name: "NetworkID other than 0",
			args: args{
				meta: &openrtb_ext.ExtBidPrebidMeta{
					NetworkID:    10,
					AdvertiserID: 0,
				},
				partnerRecord: &PartnerRecord{
					PartnerID: "pubmatic",
				},
			},
			partnerRecord: &PartnerRecord{
				PartnerID: "pubmatic",
				MetaData: &MetaData{
					NetworkID: 10,
				},
			},
		},
		{
			name: "AdvertiserID other than 0",
			args: args{
				meta: &openrtb_ext.ExtBidPrebidMeta{
					NetworkID:    0,
					AdvertiserID: 10,
				},
				partnerRecord: &PartnerRecord{
					PartnerID: "pubmatic",
				},
			},
			partnerRecord: &PartnerRecord{
				PartnerID: "pubmatic",
				MetaData: &MetaData{
					AdvertiserID: 10,
				},
			},
		},
		{
			name: "SecondaryCategoryIDs size other than 0",
			args: args{
				meta: &openrtb_ext.ExtBidPrebidMeta{
					NetworkID:            0,
					AdvertiserID:         0,
					SecondaryCategoryIDs: []string{"cat1"},
				},
				partnerRecord: &PartnerRecord{
					PartnerID: "pubmatic",
				},
			},
			partnerRecord: &PartnerRecord{
				PartnerID: "pubmatic",
				MetaData: &MetaData{
					SecondaryCategoryIDs: []string{"cat1"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.partnerRecord.setMetaDataObject(tt.args.meta)
			assert.Equal(t, tt.partnerRecord, tt.args.partnerRecord, tt.name)
		})
	}
}
