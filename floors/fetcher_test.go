package floors

import (
	"testing"

	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func TestValidatePriceFloorRules(t *testing.T) {
	type args struct {
		configs     config.AccountFloorFetch
		priceFloors *openrtb_ext.PriceFloorRules
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Price floor data is empty",
			args: args{
				configs: config.AccountFloorFetch{
					Enabled:     true,
					URL:         "abc.com",
					Timeout:     5,
					MaxFileSize: 20,
					MaxRules:    5,
					MaxAge:      20,
					Period:      10,
				},
				priceFloors: &openrtb_ext.PriceFloorRules{},
			},
			wantErr: true,
		},
		{
			name: "Model group array is empty",
			args: args{
				configs: config.AccountFloorFetch{
					Enabled:     true,
					URL:         "abc.com",
					Timeout:     5,
					MaxFileSize: 20,
					MaxRules:    5,
					MaxAge:      20,
					Period:      10,
				},
				priceFloors: &openrtb_ext.PriceFloorRules{
					Data: &openrtb_ext.PriceFloorData{},
				},
			},
			wantErr: true,
		},
		{
			name: "floor rules is empty",
			args: args{
				configs: config.AccountFloorFetch{
					Enabled:     true,
					URL:         "abc.com",
					Timeout:     5,
					MaxFileSize: 20,
					MaxRules:    5,
					MaxAge:      20,
					Period:      10,
				},
				priceFloors: &openrtb_ext.PriceFloorRules{
					Data: &openrtb_ext.PriceFloorData{
						ModelGroups: []openrtb_ext.PriceFloorModelGroup{{
							Values: map[string]float64{},
						}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "floor rules is empty",
			args: args{
				configs: config.AccountFloorFetch{
					Enabled:     true,
					URL:         "abc.com",
					Timeout:     5,
					MaxFileSize: 20,
					MaxRules:    5,
					MaxAge:      20,
					Period:      10,
				},
				priceFloors: &openrtb_ext.PriceFloorRules{
					Data: &openrtb_ext.PriceFloorData{
						ModelGroups: []openrtb_ext.PriceFloorModelGroup{{
							Values: map[string]float64{},
						}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "floor rules is grater than max floor rules",
			args: args{
				configs: config.AccountFloorFetch{
					Enabled:     true,
					URL:         "abc.com",
					Timeout:     5,
					MaxFileSize: 20,
					MaxRules:    0,
					MaxAge:      20,
					Period:      10,
				},
				priceFloors: &openrtb_ext.PriceFloorRules{
					Data: &openrtb_ext.PriceFloorData{
						ModelGroups: []openrtb_ext.PriceFloorModelGroup{{
							Values: map[string]float64{
								"*|*|www.website.com": 15.01,
							},
						}},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validatePriceFloorRules(tt.args.configs, tt.args.priceFloors); (err != nil) != tt.wantErr {
				t.Errorf("validatePriceFloorRules() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
