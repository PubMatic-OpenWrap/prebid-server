package tracker

import (
	"testing"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func Test_getFloorsDetails(t *testing.T) {
	type args struct {
		bidResponseExt openrtb_ext.ExtBidResponse
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
			name:        "no_responseExt",
			args:        args{},
			skipfloors:  nil,
			floorSource: nil,
		},
		{
			name: "empty_responseExt",
			args: args{
				bidResponseExt: openrtb_ext.ExtBidResponse{},
			},
			skipfloors:  nil,
			floorSource: nil,
		},
		{
			name: "empty_prebid_in_responseExt",
			args: args{
				bidResponseExt: openrtb_ext.ExtBidResponse{
					Prebid: &openrtb_ext.ExtResponsePrebid{},
				},
			},
			skipfloors:  nil,
			floorSource: nil,
		},
		{
			name: "empty_prebidfloors_in_responseExt",
			args: args{
				bidResponseExt: openrtb_ext.ExtBidResponse{
					Prebid: &openrtb_ext.ExtResponsePrebid{
						Floors: &openrtb_ext.PriceFloorRules{},
					},
				},
			},
			skipfloors:  nil,
			floorSource: nil,
		},
		{
			name: "no_enforced_floors_data_in_responseExt",
			args: args{
				bidResponseExt: openrtb_ext.ExtBidResponse{
					Prebid: &openrtb_ext.ExtResponsePrebid{
						Floors: &openrtb_ext.PriceFloorRules{
							Data:               &openrtb_ext.PriceFloorData{},
							PriceFloorLocation: openrtb_ext.FetchLocation,
						},
					},
				},
			},
			skipfloors:        nil,
			floorType:         models.SoftFloor,
			floorSource:       ptrutil.ToPtr(2),
			floorModelVersion: "",
		},
		{
			name: "no_modelsgroups_floors_data_in_responseExt",
			args: args{
				bidResponseExt: openrtb_ext.ExtBidResponse{
					Prebid: &openrtb_ext.ExtResponsePrebid{
						Floors: &openrtb_ext.PriceFloorRules{
							Data:               &openrtb_ext.PriceFloorData{},
							PriceFloorLocation: openrtb_ext.FetchLocation,
							Enforcement: &openrtb_ext.PriceFloorEnforcement{
								EnforcePBS: ptrutil.ToPtr(true),
							},
						},
					},
				},
			},
			skipfloors:        nil,
			floorType:         models.HardFloor,
			floorSource:       ptrutil.ToPtr(2),
			floorModelVersion: "",
		},
		{
			name: "no_skipped_floors_data_in_responseExt",
			args: args{
				bidResponseExt: openrtb_ext.ExtBidResponse{
					Prebid: &openrtb_ext.ExtResponsePrebid{
						Floors: &openrtb_ext.PriceFloorRules{
							Data: &openrtb_ext.PriceFloorData{
								ModelGroups: []openrtb_ext.PriceFloorModelGroup{
									{
										ModelVersion: "version 1",
									},
								},
							},
							PriceFloorLocation: openrtb_ext.FetchLocation,
							Enforcement: &openrtb_ext.PriceFloorEnforcement{
								EnforcePBS: ptrutil.ToPtr(true),
							},
						},
					},
				},
			},
			skipfloors:        nil,
			floorType:         models.HardFloor,
			floorSource:       ptrutil.ToPtr(2),
			floorModelVersion: "version 1",
		},
		{
			name: "all_floors_data_in_responseExt",
			args: args{
				bidResponseExt: openrtb_ext.ExtBidResponse{
					Prebid: &openrtb_ext.ExtResponsePrebid{
						Floors: &openrtb_ext.PriceFloorRules{
							Skipped: ptrutil.ToPtr(true),
							Data: &openrtb_ext.PriceFloorData{
								ModelGroups: []openrtb_ext.PriceFloorModelGroup{
									{
										ModelVersion: "version 1",
									},
								},
							},
							PriceFloorLocation: openrtb_ext.FetchLocation,
							Enforcement: &openrtb_ext.PriceFloorEnforcement{
								EnforcePBS: ptrutil.ToPtr(true),
							},
						},
					},
				},
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
