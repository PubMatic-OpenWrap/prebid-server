package exchange

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/boolutil"
	"github.com/stretchr/testify/assert"
)

func TestFloorsEnabled(t *testing.T) {
	type args struct {
		account           config.Account
		bidRequestWrapper *openrtb_ext.RequestWrapper
	}
	tests := []struct {
		name        string
		args        args
		wantEnabled bool
		wantRules   *openrtb_ext.PriceFloorRules
	}{
		{
			name: "Floors data available in request and its enabled",
			args: args{
				account: config.Account{
					PriceFloors: config.AccountPriceFloors{
						Enabled: true,
					},
				},
				bidRequestWrapper: &openrtb_ext.RequestWrapper{
					BidRequest: &openrtb2.BidRequest{
						Ext: func() json.RawMessage {
							ext := make(map[string]interface{})
							prebidExt := openrtb_ext.ExtRequestPrebid{
								Floors: &openrtb_ext.PriceFloorRules{
									Enabled:     boolutil.BoolPtr(true),
									FloorMin:    2,
									FloorMinCur: "INR",
									Data: &openrtb_ext.PriceFloorData{
										Currency: "INR",
									},
								},
							}
							ext["prebid"] = prebidExt
							data, _ := json.Marshal(ext)
							return data
						}(),
					},
				},
			},
			wantEnabled: true,
			wantRules: func() *openrtb_ext.PriceFloorRules {
				floors := openrtb_ext.PriceFloorRules{
					Enabled:     boolutil.BoolPtr(true),
					FloorMin:    2,
					FloorMinCur: "INR",
					Data: &openrtb_ext.PriceFloorData{
						Currency: "INR",
					},
				}
				return &floors
			}(),
		},
		{
			name: "Floors data available in request and floors is disabled",
			args: args{
				account: config.Account{
					PriceFloors: config.AccountPriceFloors{
						Enabled: false,
					},
				},
				bidRequestWrapper: &openrtb_ext.RequestWrapper{
					BidRequest: &openrtb2.BidRequest{
						Ext: func() json.RawMessage {
							ext := map[string]interface{}{
								"prebid": openrtb_ext.ExtRequestPrebid{
									Floors: &openrtb_ext.PriceFloorRules{
										Enabled:     boolutil.BoolPtr(true),
										FloorMin:    2,
										FloorMinCur: "INR",
										Data: &openrtb_ext.PriceFloorData{
											Currency: "INR",
										},
									},
								},
							}
							data, _ := json.Marshal(ext)
							return data
						}(),
					},
				},
			},
			wantEnabled: false,
			wantRules: func() *openrtb_ext.PriceFloorRules {
				floors := openrtb_ext.PriceFloorRules{
					Enabled:     boolutil.BoolPtr(true),
					FloorMin:    2,
					FloorMinCur: "INR",
					Data: &openrtb_ext.PriceFloorData{
						Currency: "INR",
					},
				}
				return &floors
			}(),
		},
		{
			name: "Floors data is nil in request but floors is enabled in account",
			args: args{
				account: config.Account{
					PriceFloors: config.AccountPriceFloors{
						Enabled: true,
					},
				},
				bidRequestWrapper: &openrtb_ext.RequestWrapper{
					BidRequest: &openrtb2.BidRequest{
						Ext: func() json.RawMessage {
							ext := map[string]interface{}{
								"prebid": openrtb_ext.ExtRequestPrebid{},
							}
							data, _ := json.Marshal(ext)
							return data
						}(),
					},
				},
			},
			wantEnabled: true,
			wantRules:   nil,
		},
		{
			name: "extension is empty but floors is enabled in account",
			args: args{
				account: config.Account{
					PriceFloors: config.AccountPriceFloors{
						Enabled: true,
					},
				},
				bidRequestWrapper: &openrtb_ext.RequestWrapper{
					BidRequest: &openrtb2.BidRequest{},
				},
			},
			wantEnabled: false,
			wantRules:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEnabled, gotRules := floorsEnabled(tt.args.account, tt.args.bidRequestWrapper)
			if gotEnabled != tt.wantEnabled {
				t.Errorf("floorsEnabled() got = %v, want %v", gotEnabled, tt.wantEnabled)
			}
			assert.Equal(t, tt.wantRules, gotRules, "Invalid Floors rules")
		})
	}
}
