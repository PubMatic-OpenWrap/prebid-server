package openwrap

import (
	"testing"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestSetFloorsExt(t *testing.T) {
	enable := true
	disable := false

	type args struct {
		requestExt               *models.RequestExt
		configMap                map[int]map[string]string
		setMaxFloor              bool
		isDynamicFloorEnabledPub bool
		pubID                    int
		profileID                int
	}
	tests := []struct {
		name string
		args args
		want *models.RequestExt
	}{
		{
			name: "Only JSON URL is present in db",
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
		{
			name: "SetMaxFloor true",
			args: args{
				requestExt: &models.RequestExt{},
				configMap: map[int]map[string]string{
					-1: {
						"floorType": "Hard",
					},
				},
				setMaxFloor: true,
			},
			want: func() *models.RequestExt {
				enable := true
				res := models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled:     &enable,
								SetMaxFloor: enable,
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
			name: "floor_not_present_for_in-app",
			args: args{
				requestExt: &models.RequestExt{},
				configMap: map[int]map[string]string{
					-1: {
						"platform": models.PLATFORM_APP,
					},
				},
			},
			want: func() *models.RequestExt {
				res := models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled: &enable,
								Enforcement: &openrtb_ext.PriceFloorEnforcement{
									FloorDeals: &enable,
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
			name: "floor_is_present_with_floordeal_true_for_in-app",
			args: args{
				requestExt: func() *models.RequestExt {
					res := models.RequestExt{
						ExtRequest: openrtb_ext.ExtRequest{
							Prebid: openrtb_ext.ExtRequestPrebid{
								Floors: &openrtb_ext.PriceFloorRules{
									Enabled: &enable,
									Enforcement: &openrtb_ext.PriceFloorEnforcement{
										FloorDeals: &enable,
									},
								},
							},
						},
					}
					return &res
				}(),
				configMap: map[int]map[string]string{
					-1: {
						"platform": models.PLATFORM_APP,
					},
				},
			},
			want: func() *models.RequestExt {
				res := models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled: &enable,
								Enforcement: &openrtb_ext.PriceFloorEnforcement{
									FloorDeals: &enable,
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
			name: "floor_is_present_with_floordeal_false_for_in-app",
			args: args{
				requestExt: func() *models.RequestExt {
					res := models.RequestExt{
						ExtRequest: openrtb_ext.ExtRequest{
							Prebid: openrtb_ext.ExtRequestPrebid{
								Floors: &openrtb_ext.PriceFloorRules{
									Enabled: &enable,
									Enforcement: &openrtb_ext.PriceFloorEnforcement{
										FloorDeals: &disable,
									},
								},
							},
						},
					}
					return &res
				}(),
				configMap: map[int]map[string]string{
					-1: {
						"platform": models.PLATFORM_APP,
					},
				},
			},
			want: func() *models.RequestExt {
				res := models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled: &enable,
								Enforcement: &openrtb_ext.PriceFloorEnforcement{
									FloorDeals: &disable,
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
			name: "floor_is_disable_for_in-app",
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
						"platform": models.PLATFORM_APP,
					},
				},
			},
			want: func() *models.RequestExt {
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
			name: "floor_not_present_for_ctv",
			args: args{
				requestExt: &models.RequestExt{},
				configMap: map[int]map[string]string{
					-1: {
						"platform": "ctv",
					},
				},
			},
			want: func() *models.RequestExt {
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
			name: "floor_is_present_with_floordeal_true_for_ctv",
			args: args{
				requestExt: func() *models.RequestExt {
					res := models.RequestExt{
						ExtRequest: openrtb_ext.ExtRequest{
							Prebid: openrtb_ext.ExtRequestPrebid{
								Floors: &openrtb_ext.PriceFloorRules{
									Enabled: &enable,
									Enforcement: &openrtb_ext.PriceFloorEnforcement{
										FloorDeals: &enable,
									},
								},
							},
						},
					}
					return &res
				}(),
				configMap: map[int]map[string]string{
					-1: {
						"platform": "ctv",
					},
				},
			},
			want: func() *models.RequestExt {
				res := models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled: &enable,
								Enforcement: &openrtb_ext.PriceFloorEnforcement{
									FloorDeals: &enable,
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
			name: "floor_is_present_with_floordeal_false_for_ctv",
			args: args{
				requestExt: func() *models.RequestExt {
					res := models.RequestExt{
						ExtRequest: openrtb_ext.ExtRequest{
							Prebid: openrtb_ext.ExtRequestPrebid{
								Floors: &openrtb_ext.PriceFloorRules{
									Enabled: &enable,
									Enforcement: &openrtb_ext.PriceFloorEnforcement{
										FloorDeals: &disable,
									},
								},
							},
						},
					}
					return &res
				}(),
				configMap: map[int]map[string]string{
					-1: {
						"platform": "ctv",
					},
				},
			},
			want: func() *models.RequestExt {
				res := models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled: &enable,
								Enforcement: &openrtb_ext.PriceFloorEnforcement{
									FloorDeals: &disable,
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
			name: "floor_is_disable_for_ctv",
			args: args{
				requestExt: func() *models.RequestExt {
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
						"platform": "ctv",
					},
				},
			},
			want: func() *models.RequestExt {
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
			name: "dynamic_floor_enabled_and_present_with_floordeal_false_for_in-app",
			args: args{
				requestExt: func() *models.RequestExt {
					res := models.RequestExt{
						ExtRequest: openrtb_ext.ExtRequest{
							Prebid: openrtb_ext.ExtRequestPrebid{
								Floors: &openrtb_ext.PriceFloorRules{
									Enabled: &enable,
									Enforcement: &openrtb_ext.PriceFloorEnforcement{
										FloorDeals: &disable,
									},
								},
							},
						},
					}
					return &res
				}(),
				configMap: map[int]map[string]string{
					-1: {
						"platform":                models.PLATFORM_APP,
						"jsonUrl":                 "http://test.com/floor",
						"floorPriceModuleEnabled": "1",
					},
				},
				isDynamicFloorEnabledPub: true,
				setMaxFloor:              false,
			},
			want: func() *models.RequestExt {
				res := models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled: &enable,
								Enforcement: &openrtb_ext.PriceFloorEnforcement{
									FloorDeals: &disable,
									EnforcePBS: &enable,
								},
								Location: &openrtb_ext.PriceFloorEndpoint{
									URL: "http://test.com/floor",
								},
								SetMaxFloor: true,
							},
						},
					},
				}
				return &res
			}(),
		},
		{
			name: "dynamic_floor_enabled_with_deal_false_and_floormin_and_deals_enforcement_present_in_db_for_in-app",
			args: args{
				requestExt: func() *models.RequestExt {
					res := models.RequestExt{
						ExtRequest: openrtb_ext.ExtRequest{
							Prebid: openrtb_ext.ExtRequestPrebid{
								Floors: &openrtb_ext.PriceFloorRules{
									Enabled: &enable,
									Enforcement: &openrtb_ext.PriceFloorEnforcement{
										FloorDeals: &disable,
									},
								},
							},
						},
					}
					return &res
				}(),
				configMap: map[int]map[string]string{
					-1: {
						"platform":                models.PLATFORM_APP,
						"jsonUrl":                 "http://test.com/floor",
						"floorPriceModuleEnabled": "1",
						"floorMin":                "1.7",
						"dealsEnforcement":        "1",
					},
				},
				isDynamicFloorEnabledPub: true,
				setMaxFloor:              false,
			},
			want: func() *models.RequestExt {
				res := models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled: &enable,
								Enforcement: &openrtb_ext.PriceFloorEnforcement{
									FloorDeals: &disable,
									EnforcePBS: &enable,
								},
								FloorMin: 1.7,
								Location: &openrtb_ext.PriceFloorEndpoint{
									URL: "http://test.com/floor",
								},
								SetMaxFloor: true,
							},
						},
					},
				}
				return &res
			}(),
		},
		{
			name: "dynamic_floor_enabled_at_start_with_deal_false_and_floormin_and_deals_enforcement_present_in_db_for_in-app",
			args: args{
				requestExt: func() *models.RequestExt {
					res := models.RequestExt{
						ExtRequest: openrtb_ext.ExtRequest{
							Prebid: openrtb_ext.ExtRequestPrebid{
								Floors: &openrtb_ext.PriceFloorRules{
									Enabled: &enable,
									Enforcement: &openrtb_ext.PriceFloorEnforcement{
										FloorDeals: &disable,
									},
								},
							},
						},
					}
					return &res
				}(),
				configMap: map[int]map[string]string{
					-1: {
						"platform": models.PLATFORM_APP,
					},
				},
				isDynamicFloorEnabledPub: true,
				setMaxFloor:              false,
				pubID:                    5890,
				profileID:                12312,
			},
			want: func() *models.RequestExt {
				res := models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled: &enable,
								Enforcement: &openrtb_ext.PriceFloorEnforcement{
									FloorDeals: &disable,
									EnforcePBS: &enable,
								},
								Location: &openrtb_ext.PriceFloorEndpoint{
									URL: "https://ads.pubmatic.com/AdServer/js/pwt/floors/5890/12312/floors.json",
								},
								SetMaxFloor: true,
							},
						},
					},
				}
				return &res
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setFloorsExt(tt.args.requestExt, tt.args.configMap, tt.args.setMaxFloor, tt.args.isDynamicFloorEnabledPub, tt.args.pubID, tt.args.profileID)
			assert.Equal(t, tt.want, tt.args.requestExt)
		})
	}
}

func TestSetFloorsData(t *testing.T) {
	enable := true
	disable := false

	type args struct {
		requestExt       *models.RequestExt
		versionConfigMap map[string]string
		pubID            int
		profileID        int
	}
	tests := []struct {
		name string
		args args
		want *models.RequestExt
	}{
		{
			name: "fetch_dynamic_json_for_amp",
			args: args{
				requestExt: &models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled:     &enable,
								Enforcement: &openrtb_ext.PriceFloorEnforcement{},
							},
						},
					},
				},
				versionConfigMap: map[string]string{
					"jsonUrl":                 "http://test.com/floor",
					"floorPriceModuleEnabled": "1",
					"platform":                "amp",
					"floorMin":                "1.8",
					"floorType":               "soft",
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
								EnforcePBS: &disable,
							},
							FloorMin: 1.8,
						},
					},
				},
			},
		},
		{
			name: "fetch_dynamic_json_for_in-app",
			args: args{
				requestExt: &models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled:     &enable,
								Enforcement: &openrtb_ext.PriceFloorEnforcement{},
							},
						},
					},
				},
				versionConfigMap: map[string]string{
					"jsonUrl":                 "http://test.com/floor",
					"floorPriceModuleEnabled": "1",
					"platform":                "in-app",
					"floorMin":                "1.8",
					"floorType":               "soft",
					"dealsEnforcement":        "0",
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
								EnforcePBS: &disable,
								FloorDeals: &disable,
							},
							FloorMin: 1.8,
						},
					},
				},
			},
		},
		{
			name: "fetch_dynamic_json_for_in-app_with_deals_enforcement_not_present",
			args: args{
				requestExt: &models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled:     &enable,
								Enforcement: &openrtb_ext.PriceFloorEnforcement{},
							},
						},
					},
				},
				versionConfigMap: map[string]string{
					"jsonUrl":                 "http://test.com/floor",
					"floorPriceModuleEnabled": "1",
					"platform":                "in-app",
					"floorMin":                "1.8",
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
								FloorDeals: &enable,
							},
							FloorMin: 1.8,
						},
					},
				},
			},
		},
		{
			name: "fetch_dynamic_json_at_start_for_in-app_with_deals_enforcement_not_present",
			args: args{
				requestExt: &models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled:     &enable,
								Enforcement: &openrtb_ext.PriceFloorEnforcement{},
							},
						},
					},
				},
				versionConfigMap: map[string]string{
					"platform": "in-app",
				},
				pubID:     5890,
				profileID: 12312,
			},
			want: &models.RequestExt{
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Floors: &openrtb_ext.PriceFloorRules{
							Enabled: &enable,
							Location: &openrtb_ext.PriceFloorEndpoint{
								URL: "https://ads.pubmatic.com/AdServer/js/pwt/floors/5890/12312/floors.json",
							},
							Enforcement: &openrtb_ext.PriceFloorEnforcement{
								EnforcePBS: &enable,
								FloorDeals: &enable,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setFloorsData(tt.args.requestExt, tt.args.versionConfigMap, tt.args.pubID, tt.args.profileID)
			assert.Equal(t, tt.want, tt.args.requestExt)
		})
	}
}

func TestSetFloorsDefaultsForApp(t *testing.T) {
	enable := true
	disable := false
	type args struct {
		requestExt  *models.RequestExt
		setMaxFloor bool
	}
	tests := []struct {
		name string
		args args
		want *models.RequestExt
	}{
		{
			name: "no_details_in_floor_object",
			args: args{
				requestExt: &models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enabled:     &enable,
								Enforcement: &openrtb_ext.PriceFloorEnforcement{},
							},
						},
					},
				},
				setMaxFloor: false,
			},
			want: &models.RequestExt{
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Floors: &openrtb_ext.PriceFloorRules{
							Enabled: &enable,
							Enforcement: &openrtb_ext.PriceFloorEnforcement{
								FloorDeals: &enable,
								EnforcePBS: &enable,
							},
							SetMaxFloor: false,
						},
					},
				},
			},
		},
		{
			name: "floors_present_with_floordeal_and_enforcepbs_disabled",
			args: args{
				requestExt: &models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Enforcement: &openrtb_ext.PriceFloorEnforcement{
									FloorDeals: &disable,
									EnforcePBS: &disable,
								},
							},
						},
					},
				},
				setMaxFloor: false,
			},
			want: &models.RequestExt{
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Floors: &openrtb_ext.PriceFloorRules{
							Enforcement: &openrtb_ext.PriceFloorEnforcement{
								FloorDeals: &disable,
								EnforcePBS: &disable,
							},
							SetMaxFloor: false,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setFloorsDefaultsForApp(tt.args.requestExt, tt.args.setMaxFloor)
			assert.Equal(t, tt.want, tt.args.requestExt, tt.name)
		})
	}
}

func TestGetFloorsJSON(t *testing.T) {
	type args struct {
		pubID     int
		profileID int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid url",
			args: args{
				pubID:     5890,
				profileID: 12312,
			},
			want: "https://ads.pubmatic.com/AdServer/js/pwt/floors/5890/12312/floors.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prepareFloorJsonURL(tt.args.pubID, tt.args.profileID); got != tt.want {
				t.Errorf("prepareFloorJsonURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetFloorsJSON(t *testing.T) {
	type args struct {
		requestExt *models.RequestExt
		url        string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "valid url with floor location not present",
			args: args{
				requestExt: &models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{},
						},
					},
				},
				url: "https://ads.pubmatic.com/AdServer/js/pwt/floors/5890/12312/floors.json",
			},
		},
		{
			name: "overwrite location url",
			args: args{
				requestExt: &models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Floors: &openrtb_ext.PriceFloorRules{
								Location: &openrtb_ext.PriceFloorEndpoint{
									URL: "abc.com",
								},
							},
						},
					},
				},
				url: "https://ads.pubmatic.com/AdServer/js/pwt/floors/5890/12312/floors.json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setFloorJsonURL(tt.args.requestExt, tt.args.url)
			assert.Equal(t, tt.args.url, tt.args.requestExt.Prebid.Floors.Location.URL)
		})
	}
}
