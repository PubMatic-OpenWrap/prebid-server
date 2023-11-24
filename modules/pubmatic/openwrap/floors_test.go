package openwrap

import (
	"testing"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

func Test_setFloorsExt(t *testing.T) {
	enable := true
	disable := false

	type args struct {
		requestExt *models.RequestExt
		configMap  map[int]map[string]string
	}
	tests := []struct {
		name string
		args args
		want *models.RequestExt
	}{
		{
			name: "JSON URL is present in db",
			args: args{
				requestExt: &models.RequestExt{},
				configMap: map[int]map[string]string{
					-1: {
						"jsonUrl": "http://test.com/floor",
					},
				},
			},
			want: &models.RequestExt{
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Floors: &openrtb_ext.PriceFloorRules{
							Location: &openrtb_ext.PriceFloorEndpoint{
								URL: "http://test.com/floor",
							},
							Enabled: &enable,
							Enforcement: &openrtb_ext.PriceFloorEnforcement{
								EnforcePBS: &enable,
							},
						},
					},
				},
			},
		},
		{
			name: "JSON URL is present in request, but floor module disabed",
			args: args{
				requestExt: &models.RequestExt{},
				configMap: map[int]map[string]string{
					-1: {
						"jsonUrl":                 "http://test.com/floor",
						"floorPriceModuleEnabled": "0",
					},
				},
			},
			want: &models.RequestExt{
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Floors: &openrtb_ext.PriceFloorRules{
							Enabled: &disable,
						},
					},
				},
			},
		},
		{
			name: "JSON URL not present in db, but floor module enabled",
			args: args{
				requestExt: &models.RequestExt{},
				configMap: map[int]map[string]string{
					-1: {
						"floorPriceModuleEnabled": "1",
					},
				},
			},
			want: &models.RequestExt{
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Floors: &openrtb_ext.PriceFloorRules{
							Enabled: &enable,
							Enforcement: &openrtb_ext.PriceFloorEnforcement{
								EnforcePBS: &enable,
							},
						},
					},
				},
			},
		},
		{
			name: "JSON URL not present in request, floor module enabled is not present",
			args: args{
				requestExt: &models.RequestExt{},
				configMap: map[int]map[string]string{
					-1: {},
				},
			},
			want: &models.RequestExt{
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Floors: &openrtb_ext.PriceFloorRules{
							Enabled: &enable,
							Enforcement: &openrtb_ext.PriceFloorEnforcement{
								EnforcePBS: &enable,
							},
						},
					},
				},
			},
		},
		{
			name: "Request has floor disabled, db has fetch url and floor module enabled",
			args: args{
				requestExt: func() *models.RequestExt {
					disable := false
					res := models.RequestExt{
						ExtRequest: openrtb_ext.ExtRequest{
							Prebid: openrtb_ext.ExtRequestPrebid{
								Floors: &openrtb_ext.PriceFloorRules{
									Enabled: &disable,
								},
							},
						},
					}
					return &res
				}(),
				configMap: map[int]map[string]string{
					-1: {
						"jsonUrl":                 "http://test.com/floor",
						"floorPriceModuleEnabled": "1",
					},
				},
			},
			want: func() *models.RequestExt {
				disable := false
				res := models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled: &disable,
							},
						},
					},
				}
				return &res
			}(),
		},
		{
			name: "Request has floor enabled, db has fetch url and floor module enabled",
			args: args{
				requestExt: &models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled: &enable,
							},
						},
					},
				},
				configMap: map[int]map[string]string{
					-1: {
						"jsonUrl":                 "http://test.com/floor",
						"floorPriceModuleEnabled": "1",
					},
				},
			},
			want: &models.RequestExt{
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Floors: &openrtb_ext.PriceFloorRules{
							Enabled: &enable,
							Location: &openrtb_ext.PriceFloorEndpoint{
								URL: "http://test.com/floor",
							},
							Enforcement: &openrtb_ext.PriceFloorEnforcement{
								EnforcePBS: &enable,
							},
						},
					},
				},
			},
		},
		{
			name: "Request has floor enabled, db has fetch url and floor module disabled",
			args: args{
				requestExt: func() *models.RequestExt {
					enable := true
					r := models.RequestExt{
						ExtRequest: openrtb_ext.ExtRequest{
							Prebid: openrtb_ext.ExtRequestPrebid{
								Floors: &openrtb_ext.PriceFloorRules{
									Enabled: &enable,
								},
							},
						},
					}
					return &r
				}(),
				configMap: map[int]map[string]string{
					-1: {
						"jsonUrl":                 "http://test.com/floor",
						"floorPriceModuleEnabled": "0",
					},
				},
			},
			want: func() *models.RequestExt {
				enable := true
				res := models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled: &enable,
								Location: &openrtb_ext.PriceFloorEndpoint{
									URL: "http://test.com/floor",
								},
								Enforcement: &openrtb_ext.PriceFloorEnforcement{
									EnforcePBS: &enable,
								},
							},
						},
					},
				}
				return &res
			}(),
		},
		{
			name: "Request is empty, db has floortype as soft",
			args: args{
				requestExt: &models.RequestExt{},
				configMap: map[int]map[string]string{
					-1: {
						"floorType": "Soft",
					},
				},
			},
			want: func() *models.RequestExt {
				enable := true
				disable := false
				res := models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled: &enable,
								Enforcement: &openrtb_ext.PriceFloorEnforcement{
									EnforcePBS: &disable,
								},
							},
						},
					},
				}
				return &res
			}(),
		},
		{
			name: "Request is empty, db has floortype as hard",
			args: args{
				requestExt: &models.RequestExt{},
				configMap: map[int]map[string]string{
					-1: {
						"floorType": "Hard",
					},
				},
			},
			want: func() *models.RequestExt {
				enable := true
				res := models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled: &enable,
								Enforcement: &openrtb_ext.PriceFloorEnforcement{
									EnforcePBS: &enable,
								},
							},
						},
					},
				}
				return &res
			}(),
		},
		{
			name: "Request has EnforcePBS false, db has floortype as hard",
			args: args{
				requestExt: func() *models.RequestExt {
					disable := false
					res := models.RequestExt{
						ExtRequest: openrtb_ext.ExtRequest{
							Prebid: openrtb_ext.ExtRequestPrebid{
								Floors: &openrtb_ext.PriceFloorRules{
									Enforcement: &openrtb_ext.PriceFloorEnforcement{
										EnforcePBS: &disable,
									},
								},
							},
						},
					}
					return &res
				}(),
				configMap: map[int]map[string]string{
					-1: {
						"floorType": "Hard",
					},
				},
			},
			want: &models.RequestExt{
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Floors: &openrtb_ext.PriceFloorRules{
							Enabled: &enable,
							Enforcement: &openrtb_ext.PriceFloorEnforcement{
								EnforcePBS: &disable,
							},
						},
					},
				},
			},
		},
		{
			name: "Request has floors disabled, db has floortype as hard, enforcepbs will be nil",
			args: args{
				requestExt: func() *models.RequestExt {
					disable := false
					res := models.RequestExt{
						ExtRequest: openrtb_ext.ExtRequest{
							Prebid: openrtb_ext.ExtRequestPrebid{
								Floors: &openrtb_ext.PriceFloorRules{
									Enabled: &disable,
								},
							},
						},
					}
					return &res
				}(),
				configMap: map[int]map[string]string{
					-1: {
						"floorType": "Hard",
					},
				},
			},
			want: &models.RequestExt{
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Floors: &openrtb_ext.PriceFloorRules{
							Enabled: &disable,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setFloorsExt(tt.args.requestExt, tt.args.configMap)
		})
	}
}
