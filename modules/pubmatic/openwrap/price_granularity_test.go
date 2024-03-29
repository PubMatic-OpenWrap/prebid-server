package openwrap

import (
	"fmt"
	"testing"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

var priceGranularityMed = openrtb_ext.PriceGranularity{
	Precision: ptrutil.ToPtr(2),
	Ranges: []openrtb_ext.GranularityRange{{
		Min:       0,
		Max:       20,
		Increment: 0.1}},
}

var priceGranularityAuto = openrtb_ext.PriceGranularity{
	Precision: ptrutil.ToPtr(2),
	Ranges: []openrtb_ext.GranularityRange{
		{
			Min:       0,
			Max:       5,
			Increment: 0.05,
		},
		{
			Min:       5,
			Max:       10,
			Increment: 0.1,
		},
		{
			Min:       10,
			Max:       20,
			Increment: 0.5,
		},
	},
}

var priceGranularityTestPG = openrtb_ext.PriceGranularity{
	Test:      true,
	Precision: ptrutil.ToPtr(2),
	Ranges: []openrtb_ext.GranularityRange{{
		Min:       0,
		Max:       50,
		Increment: 50}},
}

var priceGranularityDense = openrtb_ext.PriceGranularity{
	Precision: ptrutil.ToPtr(2),
	Ranges: []openrtb_ext.GranularityRange{
		{
			Min:       0,
			Max:       3,
			Increment: 0.01,
		},
		{
			Min:       3,
			Max:       8,
			Increment: 0.05,
		},
		{
			Min:       8,
			Max:       20,
			Increment: 0.5,
		},
	},
}

func TestComputePriceGranularity(t *testing.T) {
	type args struct {
		rctx models.RequestCtx
	}
	tests := []struct {
		name    string
		args    args
		want    openrtb_ext.PriceGranularity
		wantErr bool
	}{
		{
			name: "dense_price_granularity_for_in_app_platform",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							models.PriceGranularityKey: "dense",
						},
					},
				},
			},
			want:    priceGranularityDense,
			wantErr: false,
		},
		{
			name: "no_pricegranularity_in_db_defaults_to_auto",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: nil,
				},
			},
			want:    priceGranularityAuto, // auto PG Object
			wantErr: false,
		}, {
			name: "custompg_OpenRTB_V25_API",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							models.PriceGranularityKey:          "custom",
							models.PriceGranularityCustomConfig: `{ "ranges": [{"min": 0, "max":2, "increment" : 1}]}`,
						},
					},
				},
			},
			want: openrtb_ext.PriceGranularity{
				Test:      false,
				Precision: ptrutil.ToPtr(2),
				Ranges: []openrtb_ext.GranularityRange{
					{
						Min: 0, Max: 2, Increment: 1,
					},
				},
			},
			wantErr: false,
		}, {
			name: "testreq_ctv_expect_testpg",
			args: args{
				rctx: models.RequestCtx{
					IsTestRequest: 1,
				},
			},
			want:    priceGranularityTestPG,
			wantErr: false,
		}, {
			name: "custompg_ctvapi",
			args: args{
				rctx: models.RequestCtx{
					IsCTVRequest: true,
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							models.PriceGranularityKey:          "custom",
							models.PriceGranularityCustomConfig: `{ "ranges": [{"min": 0, "max":2, "increment" : 1}]}`,
						},
					},
				},
			},
			want: openrtb_ext.PriceGranularity{
				Test:      false,
				Precision: ptrutil.ToPtr(2),
				Ranges: []openrtb_ext.GranularityRange{
					{
						Min: 0, Max: 2, Increment: 1,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := computePriceGranularity(tt.args.rctx)
			if (err != nil) != tt.wantErr {
				assert.Equal(t, tt.wantErr, err != nil)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewCustomPriceGranuality(t *testing.T) {
	type args struct {
		customPGValue string
	}
	tests := []struct {
		name    string
		args    args
		want    openrtb_ext.PriceGranularity
		wantErr bool
	}{
		{
			name:    "empty_pg_expect_default_medium_pg",
			args:    args{customPGValue: ""},
			want:    openrtb_ext.PriceGranularity{},
			wantErr: true,
		},
		{
			name: "always_have_precision_2",
			args: args{customPGValue: `{ "precision": 3, "ranges":[{ "min" : 0.01, "max" : 0.99, "increment" : 0.10}] }`},
			want: openrtb_ext.PriceGranularity{
				Precision: ptrutil.ToPtr(2),
				Ranges: []openrtb_ext.GranularityRange{{
					Min:       0.01,
					Max:       0.99,
					Increment: 0.10,
				}},
			},
			wantErr: false,
		},
		{
			// not expected case as DB will never contain pg without range
			name: "no_ranges_defaults_to_medium_pg",
			args: args{customPGValue: `{}`},
			want: openrtb_ext.PriceGranularity{
				Precision: ptrutil.ToPtr(2),
			},
			wantErr: false,
		},
		{
			name: "0_precision_overwrite_with_2",
			args: args{customPGValue: `{ "precision": 0, "ranges":[{ "min" : 0.01, "max" : 0.99, "increment" : 0.10}] }`},
			want: openrtb_ext.PriceGranularity{
				Precision: ptrutil.ToPtr(2),
				Ranges: []openrtb_ext.GranularityRange{{
					Min:       0.01,
					Max:       0.99,
					Increment: 0.10,
				}},
			},
			wantErr: false,
		},
		{

			name: "precision_greater_than_max_decimal_figures_expect_2",
			args: args{customPGValue: fmt.Sprintf(`{ "precision": %v, "ranges":[{ "min" : 0.01, "max" : 0.99, "increment" : 0.10}] }`, openrtb_ext.MaxDecimalFigures)},
			want: openrtb_ext.PriceGranularity{
				Precision: ptrutil.ToPtr(2),
				Ranges: []openrtb_ext.GranularityRange{{
					Min:       0.01,
					Max:       0.99,
					Increment: 0.10,
				}},
			},
			wantErr: false,
		},
		{

			name: "increment_less_than_0_error",
			args: args{customPGValue: `{ "ranges":[{ "min" : 0.01, "max" : 0.99, "increment" : -1.0}] }`},
			want: openrtb_ext.PriceGranularity{
				Precision: ptrutil.ToPtr(2),
				Ranges: []openrtb_ext.GranularityRange{{
					Min:       0.01,
					Max:       0.99,
					Increment: -1,
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newCustomPriceGranuality(tt.args.customPGValue)
			if (err != nil) != tt.wantErr {
				assert.Equal(t, tt.wantErr, err != nil)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
