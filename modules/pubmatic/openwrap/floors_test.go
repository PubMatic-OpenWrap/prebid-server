package openwrap

import (
	"testing"

	"github.com/PubMatic-OpenWrap/prebid-server/util/boolutil"
	"github.com/magiconair/properties/assert"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func TestSetPriceFloorFetchURL(t *testing.T) {
	type args struct {
		requestExt       *models.RequestExt
		partnerConfigMap map[int]map[string]string
	}
	tests := []struct {
		name           string
		args           args
		wantRequestExt *models.RequestExt
	}{
		{
			name: "No version config present in partner config map",
			args: args{
				partnerConfigMap: map[int]map[string]string{},
			},
			wantRequestExt: nil,
		},
		{
			name: "RequestExt nil",
			args: args{
				partnerConfigMap: map[int]map[string]string{
					models.VersionLevelConfigID: {},
				},
			},
			wantRequestExt: nil,
		},
		{
			name: "Floors is not requestExt",
			args: args{
				requestExt: &models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: nil,
						},
					},
				},
				partnerConfigMap: map[int]map[string]string{
					models.VersionLevelConfigID: {},
				},
			},
			wantRequestExt: &models.RequestExt{
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Floors: nil,
					},
				},
			},
		},
		{
			name: "Floors is enabled",
			args: args{
				requestExt: &models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled: boolutil.BoolPtr(true),
							},
						},
					},
				},
				partnerConfigMap: map[int]map[string]string{
					models.VersionLevelConfigID: {},
				},
			},
			wantRequestExt: &models.RequestExt{
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Floors: &openrtb_ext.PriceFloorRules{
							Enabled: boolutil.BoolPtr(true),
						},
					},
				},
			},
		},
		{
			name: "Price floor url missing in config",
			args: args{
				requestExt: &models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled: boolutil.BoolPtr(false),
							},
						},
					},
				},
				partnerConfigMap: map[int]map[string]string{
					models.VersionLevelConfigID: {},
				},
			},
			wantRequestExt: &models.RequestExt{
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Floors: &openrtb_ext.PriceFloorRules{
							Enabled: boolutil.BoolPtr(false),
						},
					},
				},
			},
		},
		{
			name: "Floors module disabled",
			args: args{
				requestExt: &models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{},
					},
				},
				partnerConfigMap: map[int]map[string]string{
					models.VersionLevelConfigID: {
						models.FloorModuleEnabled: "0",
						models.PriceFloorURL:      "testurl.com",
					},
				},
			},
			wantRequestExt: &models.RequestExt{
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Floors: &openrtb_ext.PriceFloorRules{
							Enabled: boolutil.BoolPtr(false),
						},
					},
				},
			},
		},
		{
			name: "Set price floor url success",
			args: args{
				requestExt: &models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{},
					},
				},
				partnerConfigMap: map[int]map[string]string{
					models.VersionLevelConfigID: {
						models.FloorModuleEnabled: "1",
						models.PriceFloorURL:      "testurl.com",
					},
				},
			},
			wantRequestExt: &models.RequestExt{
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Floors: &openrtb_ext.PriceFloorRules{
							Enabled: boolutil.BoolPtr(true),
							Location: &openrtb_ext.PriceFloorEndpoint{
								URL: "testurl.com",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setPriceFloorFetchURL(tt.args.requestExt, tt.args.partnerConfigMap)
			assert.Equal(t, tt.args.requestExt, tt.wantRequestExt, tt.name)
		})
	}
}
