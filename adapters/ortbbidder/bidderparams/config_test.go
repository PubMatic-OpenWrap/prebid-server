package bidderparams

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetRequestParams(t *testing.T) {
	type fields struct {
		bidderConfig *BidderConfig
	}
	type args struct {
		bidderName    string
		requestParams map[string]BidderParamMapper
	}
	type want struct {
		bidderCfg *BidderConfig
		err       error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "bidderConfigMap_is_nil",
			fields: fields{
				bidderConfig: &BidderConfig{
					bidderConfigMap: nil,
				},
			},
			args: args{
				bidderName: "test",
				requestParams: map[string]BidderParamMapper{
					"adunit": {
						Location: "ext.adunit",
					},
				},
			},
			want: want{
				bidderCfg: &BidderConfig{
					bidderConfigMap: map[string]*config{
						"test": {
							requestParams: map[string]BidderParamMapper{
								"adunit": {
									Location: "ext.adunit",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "bidderName_not_found",
			fields: fields{
				bidderConfig: &BidderConfig{
					bidderConfigMap: map[string]*config{},
				},
			},
			args: args{
				bidderName: "test",
				requestParams: map[string]BidderParamMapper{
					"param-1": {
						Location: "path",
					},
				},
			},
			want: want{
				bidderCfg: &BidderConfig{
					bidderConfigMap: map[string]*config{
						"test": {
							requestParams: map[string]BidderParamMapper{
								"param-1": {
									Location: "path",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "bidderName_found",
			fields: fields{
				bidderConfig: &BidderConfig{
					bidderConfigMap: map[string]*config{
						"test": {
							requestParams: map[string]BidderParamMapper{
								"param-1": {
									Location: "path-1",
								},
							},
						},
					},
				},
			},
			args: args{
				bidderName: "test",
				requestParams: map[string]BidderParamMapper{
					"param-2": {
						Location: "path-2",
					},
				},
			},
			want: want{
				bidderCfg: &BidderConfig{
					bidderConfigMap: map[string]*config{
						"test": {
							requestParams: map[string]BidderParamMapper{
								"param-2": {
									Location: "path-2",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.bidderConfig.setRequestParams(tt.args.bidderName, tt.args.requestParams)
			assert.Equal(t, tt.want.bidderCfg, tt.fields.bidderConfig, "mismatched bidderConfig")
		})
	}
}

func TestGetBidderRequestProperties(t *testing.T) {
	type fields struct {
		biddersConfig *BidderConfig
	}
	type args struct {
		bidderName string
	}
	type want struct {
		requestParams map[string]BidderParamMapper
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "BidderConfigMap_is_nil",
			fields: fields{
				biddersConfig: &BidderConfig{
					bidderConfigMap: nil,
				},
			},
			args: args{
				bidderName: "test",
			},
			want: want{
				requestParams: nil,
			},
		},
		{
			name: "BidderName_absent_in_biddersConfigMap",
			fields: fields{
				biddersConfig: &BidderConfig{
					bidderConfigMap: map[string]*config{
						"ortb": {},
					},
				},
			},
			args: args{
				bidderName: "test",
			},
			want: want{
				requestParams: nil,
			},
		},
		{
			name: "BidderName_present_but_config_is_nil",
			fields: fields{
				biddersConfig: &BidderConfig{
					bidderConfigMap: map[string]*config{
						"ortb": nil,
					},
				},
			},
			args: args{
				bidderName: "test",
			},
			want: want{
				requestParams: nil,
			},
		},
		{
			name: "BidderName_present_in_biddersConfigMap",
			fields: fields{
				biddersConfig: &BidderConfig{
					bidderConfigMap: map[string]*config{
						"test": {
							requestParams: map[string]BidderParamMapper{
								"param-1": {
									Location: "value-1",
								},
							},
						},
					},
				},
			},
			args: args{
				bidderName: "test",
			},
			want: want{
				requestParams: map[string]BidderParamMapper{
					"param-1": {
						Location: "value-1",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := tt.fields.biddersConfig.GetRequestParams(tt.args.bidderName)
			assert.Equal(t, tt.want.requestParams, params, "mismatched requestParams")
		})
	}
}
