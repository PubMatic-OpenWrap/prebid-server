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

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *BidderConfig
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
						location: []string{"ext", "adunit"},
					},
				},
			},
			want: nil,
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
						location: []string{"ext", "adunit"},
					},
				},
			},
			want: &BidderConfig{
				bidderConfigMap: map[string]*config{
					"test": {
						requestParams: map[string]BidderParamMapper{
							"adunit": {
								location: []string{"ext", "adunit"},
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
						location: []string{"path"},
					},
				},
			},
			want: &BidderConfig{
				bidderConfigMap: map[string]*config{
					"test": {
						requestParams: map[string]BidderParamMapper{
							"param-1": {
								location: []string{"path"},
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
									location: []string{"path-1"},
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
						location: []string{"path-2"},
					},
				},
			},
			want: &BidderConfig{
				bidderConfigMap: map[string]*config{
					"test": {
						requestParams: map[string]BidderParamMapper{
							"param-2": {
								location: []string{"path-2"},
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
			assert.Equal(t, tt.want, tt.fields.bidderConfig, "mismatched bidderConfig")
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
		found         bool
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
				found:         false,
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
				found:         false,
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
				found:         false,
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
				found:         false,
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
									location: []string{"value-1"},
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
						location: []string{"value-1"},
					},
				},
				found: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, found := tt.fields.biddersConfig.GetRequestParams(tt.args.bidderName)
			assert.Equal(t, tt.want.requestParams, params, "mismatched requestParams")
			assert.Equal(t, tt.want.found, found, "mismatched found value")
		})
	}
}

func TestBidderParamMapperGetLocation(t *testing.T) {
	tests := []struct {
		name string
		bpm  BidderParamMapper
		want []string
	}{
		{
			name: "location_is_nil",
			bpm: BidderParamMapper{
				location: nil,
			},
			want: nil,
		},
		{
			name: "location_is_non_empty",
			bpm: BidderParamMapper{
				location: []string{"req", "ext"},
			},
			want: []string{"req", "ext"},
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
		location []string
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
				location: []string{"req", "ext"},
			},
			want: BidderParamMapper{
				location: []string{"req", "ext"},
			},
		},
		{
			name: "override_location",
			bpm: BidderParamMapper{
				location: []string{"imp", "ext"},
			},
			args: args{
				location: []string{"req", "ext"},
			},
			want: BidderParamMapper{
				location: []string{"req", "ext"},
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
