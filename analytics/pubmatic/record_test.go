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
		uaFromHTTPReq  string
		ortbBidRequest *openrtb2.BidRequest
		platform       string
		rctx           models.RequestCtx
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
				rctx: models.RequestCtx{
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
				uaFromHTTPReq:  `Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36`,
				ortbBidRequest: &openrtb2.BidRequest{},
				platform:       models.PLATFORM_AMP,
				rctx: models.RequestCtx{
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
				uaFromHTTPReq: `Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36`,
				ortbBidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						Ext: json.RawMessage(`invalid ext`),
					},
				},
				platform: models.PLATFORM_AMP,
				rctx: models.RequestCtx{
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
				uaFromHTTPReq: `Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36`,
				ortbBidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						Ext: json.RawMessage(`{"anykey":"anyval"}`),
					},
				},
				rctx: models.RequestCtx{
					DevicePlatform: models.DevicePlatformMobileWeb,
				},
				platform: models.PLATFORM_AMP,
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
				uaFromHTTPReq: `Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36`,
				ortbBidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						Ext: json.RawMessage(`{"ifa_type": 123}`)},
				},
				rctx: models.RequestCtx{
					DevicePlatform: models.DevicePlatformMobileWeb,
				},
				platform: models.PLATFORM_AMP,
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
				uaFromHTTPReq: `Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36`,
				ortbBidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						Ext: json.RawMessage(`{"ifa_type": "anything"}`),
					},
				},
				rctx: models.RequestCtx{
					DevicePlatform: models.DevicePlatformMobileWeb,
				},
				platform: models.PLATFORM_AMP,
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
				uaFromHTTPReq: `Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36`,
				ortbBidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						Ext: json.RawMessage(`{"ifa_type": "DpId"}`),
					},
				},
				rctx: models.RequestCtx{
					DevicePlatform: models.DevicePlatformMobileWeb,
				},
				platform: models.PLATFORM_AMP,
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
				uaFromHTTPReq: `Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36`,
				ortbBidRequest: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						Ext: json.RawMessage(`{"ifa_type": "sessionid"}`),
					},
				},
				rctx: models.RequestCtx{
					DevicePlatform: models.DevicePlatformMobileWeb,
				},
				platform: models.PLATFORM_AMP,
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
			wlog.logDeviceObject(tt.args.rctx, tt.args.uaFromHTTPReq, tt.args.ortbBidRequest, tt.args.platform)
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

func TestLogFloorDetails(t *testing.T) {
	type fields struct {
		record record
	}
	type args struct {
		floors *openrtb_ext.PriceFloorRules
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   record
	}{
		{
			name: "Floor details are nil",
			args: args{
				floors: nil,
			},
			fields: fields{
				record: record{
					PubID: 5890,
				},
			},
			want: record{
				PubID: 5890,
			},
		},
		{
			name: "Floor details are available in prebid extension",
			args: args{
				floors: &openrtb_ext.PriceFloorRules{
					Skipped: func() *bool {
						disable := false
						return &disable
					}(),
					PriceFloorLocation: "fetch",
					Data: &openrtb_ext.PriceFloorData{
						ModelGroups: []openrtb_ext.PriceFloorModelGroup{
							{
								ModelVersion: "version 1",
							},
						},
					},
				},
			},
			fields: fields{
				record: record{
					PubID: 5890,
					Slots: []SlotRecord{
						{
							SlotName: "abc",
						},
					},
				},
			},
			want: record{
				PubID: 5890,
				Slots: []SlotRecord{
					{
						SlotName:         "abc",
						FloorSkippedFlag: ptrutil.ToPtr(0),
					},
				},
				FloorModelVersion: "version 1",
				FloorSource:       ptrutil.ToPtr(2),
			},
		},
		{
			name: "Floor details are available except data in prebid extension",
			args: args{
				floors: &openrtb_ext.PriceFloorRules{
					Skipped: func() *bool {
						disable := false
						return &disable
					}(),
					PriceFloorLocation: "fetch",
				},
			},
			fields: fields{
				record: record{
					PubID: 5890,
					Slots: []SlotRecord{
						{
							SlotName: "abc",
						},
					},
				},
			},
			want: record{
				PubID: 5890,
				Slots: []SlotRecord{
					{
						SlotName:         "abc",
						FloorSkippedFlag: ptrutil.ToPtr(0),
					},
				},
				FloorSource: ptrutil.ToPtr(2),
			},
		},
		{
			name: "Floor details are available except modelgroups in prebid extension",
			args: args{
				floors: &openrtb_ext.PriceFloorRules{
					Skipped: func() *bool {
						disable := false
						return &disable
					}(),
					PriceFloorLocation: "fetch",
					Data: &openrtb_ext.PriceFloorData{
						Currency: "INR",
					},
				},
			},
			fields: fields{
				record: record{
					PubID: 5890,
					Slots: []SlotRecord{
						{
							SlotName: "abc",
						},
					},
				},
			},
			want: record{
				PubID: 5890,
				Slots: []SlotRecord{
					{
						SlotName:         "abc",
						FloorSkippedFlag: ptrutil.ToPtr(0),
					},
				},
				FloorSource: ptrutil.ToPtr(2),
			},
		},
		{
			name: "Floor details are available, source is not availble in map",
			args: args{
				floors: &openrtb_ext.PriceFloorRules{
					Skipped: func() *bool {
						disable := false
						return &disable
					}(),
					PriceFloorLocation: "invalid",
					Data: &openrtb_ext.PriceFloorData{
						Currency: "INR",
					},
				},
			},
			fields: fields{
				record: record{
					PubID: 5890,
					Slots: []SlotRecord{
						{
							SlotName: "abc",
						},
					},
				},
			},
			want: record{
				PubID: 5890,
				Slots: []SlotRecord{
					{
						SlotName:         "abc",
						FloorSkippedFlag: ptrutil.ToPtr(0),
					},
				},
			},
		},
		{
			name: "Floor details are available, Enforcement is nil",
			args: args{
				floors: &openrtb_ext.PriceFloorRules{
					Enforcement: nil,
				},
			},
			fields: fields{
				record: record{
					PubID: 5890,
				},
			},
			want: record{
				PubID:     5890,
				FloorType: models.SoftFloor,
			},
		},
		{
			name: "Floor details are available, Enforcement is present but enforcePBS is nil",
			args: args{
				floors: &openrtb_ext.PriceFloorRules{
					Enforcement: &openrtb_ext.PriceFloorEnforcement{
						EnforcePBS: nil,
					},
				},
			},
			fields: fields{
				record: record{
					PubID: 5890,
				},
			},
			want: record{
				PubID:     5890,
				FloorType: models.SoftFloor,
			},
		},
		{
			name: "Floor details are available, enforcePBS is present with value false",
			args: args{
				floors: &openrtb_ext.PriceFloorRules{
					Enforcement: &openrtb_ext.PriceFloorEnforcement{
						EnforcePBS: ptrutil.ToPtr(false),
					},
				},
			},
			fields: fields{
				record: record{
					PubID: 5890,
				},
			},
			want: record{
				PubID:     5890,
				FloorType: models.SoftFloor,
			},
		},
		{
			name: "Floor details are available, enforcePBS is present with value true",
			args: args{
				floors: &openrtb_ext.PriceFloorRules{
					Enforcement: &openrtb_ext.PriceFloorEnforcement{
						EnforcePBS: ptrutil.ToPtr(true),
					},
				},
			},
			fields: fields{
				record: record{
					PubID: 5890,
				},
			},
			want: record{
				PubID:     5890,
				FloorType: models.HardFloor,
			},
		},
		{
			name: "Floor_source_fetched,success_fetch_status_should_be_logged",
			args: args{
				floors: &openrtb_ext.PriceFloorRules{
					PriceFloorLocation: openrtb_ext.FetchLocation,
					FetchStatus:        openrtb_ext.FetchSuccess,
				},
			},
			fields: fields{
				record: record{
					PubID: 5890,
				},
			},
			want: record{
				PubID:            5890,
				FloorSource:      ptrutil.ToPtr(2),
				FloorFetchStatus: ptrutil.ToPtr(1),
			},
		},
		{
			name: "Floor_source_fetched,error_fetch_status_should_be_logged",
			args: args{
				floors: &openrtb_ext.PriceFloorRules{
					PriceFloorLocation: openrtb_ext.FetchLocation,
					FetchStatus:        openrtb_ext.FetchError,
				},
			},
			fields: fields{
				record: record{
					PubID: 5890,
				},
			},
			want: record{
				PubID:            5890,
				FloorSource:      ptrutil.ToPtr(2),
				FloorFetchStatus: ptrutil.ToPtr(2),
			},
		},
		{
			name: "Floor_source_request,fetch_error_out",
			args: args{
				floors: &openrtb_ext.PriceFloorRules{
					PriceFloorLocation: openrtb_ext.RequestLocation,
					FetchStatus:        openrtb_ext.FetchError,
				},
			},
			fields: fields{
				record: record{
					PubID: 5890,
				},
			},
			want: record{
				PubID:            5890,
				FloorSource:      ptrutil.ToPtr(1),
				FloorFetchStatus: ptrutil.ToPtr(2),
			},
		},
		{
			name: "Floor_source_request,floor_provider_should_be_logged",
			args: args{
				floors: &openrtb_ext.PriceFloorRules{
					PriceFloorLocation: openrtb_ext.RequestLocation,
					FloorProvider:      "test-provider",
				},
			},
			fields: fields{
				record: record{
					PubID: 5890,
				},
			},
			want: record{
				PubID:         5890,
				FloorSource:   ptrutil.ToPtr(1),
				FloorProvider: "test-provider",
			},
		},
		{
			name: "Floor_source_fetched,floor_provider_should_be_logged",
			args: args{
				floors: &openrtb_ext.PriceFloorRules{
					PriceFloorLocation: openrtb_ext.FetchLocation,
					FloorProvider:      "test-provider",
				},
			},
			fields: fields{
				record: record{
					PubID: 5890,
				},
			},
			want: record{
				PubID:         5890,
				FloorSource:   ptrutil.ToPtr(2),
				FloorProvider: "test-provider",
			},
		},
		{
			name: "Floor_source_fetched,floor_provider_available_in_data",
			args: args{
				floors: &openrtb_ext.PriceFloorRules{
					PriceFloorLocation: openrtb_ext.FetchLocation,
					FloorProvider:      "test-provider",
					Data: &openrtb_ext.PriceFloorData{
						FloorProvider: "data-provider",
					},
				},
			},
			fields: fields{
				record: record{
					PubID: 5890,
				},
			},
			want: record{
				PubID:         5890,
				FloorSource:   ptrutil.ToPtr(2),
				FloorProvider: "data-provider",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wlog := &WloggerRecord{
				record: tt.fields.record,
			}
			wlog.setFloorDetails(tt.args.floors)
			assert.Equal(t, tt.want, wlog.record)

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
