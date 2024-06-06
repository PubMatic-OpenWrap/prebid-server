package bidderparams

import (
	"fmt"
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
			name: "bidderConfig_is_nil",
			fields: fields{
				bidderConfig: nil,
			},
			args: args{
				bidderName: "test",
				requestParams: map[string]BidderParamMapper{
					"adunit": {
						location: "ext.adunit",
					},
				},
			},
			want: want{
				err: fmt.Errorf("BidderConfig is nil"),
			},
		},
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
						location: "ext.adunit",
					},
				},
			},
			want: want{
				bidderCfg: &BidderConfig{
					bidderConfigMap: map[string]*config{
						"test": {
							requestParams: map[string]BidderParamMapper{
								"adunit": {
									location: "ext.adunit",
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
						location: "path",
					},
				},
			},
			want: want{
				bidderCfg: &BidderConfig{
					bidderConfigMap: map[string]*config{
						"test": {
							requestParams: map[string]BidderParamMapper{
								"param-1": {
									location: "path",
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
									location: "path-1",
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
						location: "path-2",
					},
				},
			},
			want: want{
				bidderCfg: &BidderConfig{
					bidderConfigMap: map[string]*config{
						"test": {
							requestParams: map[string]BidderParamMapper{
								"param-2": {
									location: "path-2",
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
			err := tt.fields.bidderConfig.setRequestParams(tt.args.bidderName, tt.args.requestParams)
			assert.Equal(t, tt.want.bidderCfg, tt.fields.bidderConfig, "mismatched bidderConfig")
			assert.Equal(t, tt.want.err, err, "mismatched error")
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
			name: "BidderConfig_is_nil",
			fields: fields{
				biddersConfig: nil,
			},
			args: args{
				bidderName: "test",
			},
			want: want{
				requestParams: nil,
			},
		},
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
									location: "value-1",
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
						location: "value-1",
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

func TestBidderParamMapperGetLocation(t *testing.T) {
	tests := []struct {
		name string
		bpm  BidderParamMapper
		want string
	}{
		{
			name: "location_is_nil",
			bpm: BidderParamMapper{
				location: "",
			},
			want: "",
		},
		{
			name: "location_is_non_empty",
			bpm: BidderParamMapper{
				location: "req.ext",
			},
			want: "req.ext",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.bpm.GetLocation()
			assert.Equal(t, tt.want, got, "mismatched location")
		})
	}
}

func TestBidderParamMapperSetLocation(t *testing.T) {
	type args struct {
		location string
	}
	tests := []struct {
		name string
		bpm  BidderParamMapper
		args args
		want BidderParamMapper
	}{
		{
			name: "set_location",
			bpm:  BidderParamMapper{},
			args: args{
				location: "req.ext",
			},
			want: BidderParamMapper{
				location: "req.ext",
			},
		},
		{
			name: "override_location",
			bpm: BidderParamMapper{
				location: "imp.ext",
			},
			args: args{
				location: "req.ext",
			},
			want: BidderParamMapper{
				location: "req.ext",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.bpm.SetLocation(tt.args.location)
			assert.Equal(t, tt.want, tt.bpm, "mismatched location")
		})
	}
}
