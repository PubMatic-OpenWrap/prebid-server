package pubmatic

import (
	"testing"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

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
		{
			name:            "ows2s",
			endpoint:        models.EndpointWebS2S,
			integrationType: models.TypeWebS2S,
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

func TestLogDeviceObject(t *testing.T) {
	type args struct {
		dvc *models.DeviceCtx
	}
	tests := []struct {
		name string
		args args
		want Device
	}{
		{
			name: `empty`,
			args: args{
				dvc: nil,
			},
			want: Device{},
		},
		{
			name: `missing_ifatype`,
			args: args{
				dvc: &models.DeviceCtx{
					Platform: models.DevicePlatformDesktop,
				},
			},
			want: Device{
				Platform: models.DevicePlatformDesktop,
			},
		},
		{
			name: `missing_ext`,
			args: args{
				dvc: &models.DeviceCtx{
					Platform:  models.DevicePlatformDesktop,
					IFATypeID: ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeDPID]),
				},
			},
			want: Device{
				Platform: models.DevicePlatformDesktop,
				IFAType:  ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeDPID]),
			},
		},
		{
			name: `missing_device_ext`,
			args: args{
				dvc: &models.DeviceCtx{
					Platform:  models.DevicePlatformDesktop,
					IFATypeID: ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeDPID]),
					Ext:       &models.ExtDevice{},
				},
			},
			want: Device{
				Platform: models.DevicePlatformDesktop,
				IFAType:  ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeDPID]),
			},
		},
		{
			name: `missing_atts`,
			args: args{
				dvc: &models.DeviceCtx{
					Platform:  models.DevicePlatformDesktop,
					IFATypeID: ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeDPID]),
					Ext:       &models.ExtDevice{},
				},
			},
			want: Device{
				Platform: models.DevicePlatformDesktop,
				IFAType:  ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeDPID]),
			},
		},
		{
			name: `valid`,
			args: args{
				dvc: &models.DeviceCtx{
					Platform:  models.DevicePlatformDesktop,
					IFATypeID: ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeDPID]),
					Ext: func() *models.ExtDevice {
						extDevice := models.ExtDevice{}
						extDevice.UnmarshalJSON([]byte(`{"atts":0}`))
						return &extDevice
					}(),
				},
			},
			want: Device{
				Platform: models.DevicePlatformDesktop,
				IFAType:  ptrutil.ToPtr(models.DeviceIFATypeID[models.DeviceIFATypeDPID]),
				ATTS:     ptrutil.ToPtr(float64(openrtb_ext.IOSAppTrackingStatusNotDetermined)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wlog := &WloggerRecord{}
			wlog.logDeviceObject(tt.args.dvc)
			assert.Equal(t, tt.want, wlog.Device)
		})
	}
}
