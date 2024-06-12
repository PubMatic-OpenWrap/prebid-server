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
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
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
			tt.fields.bidderConfig.SetRequestParams(tt.args.bidderName, tt.args.requestParams)
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

func TestSetResponseParams(t *testing.T) {
	tests := []struct {
		name           string
		bidderName     string
		responseParams map[string]BidderParamMapper
		expected       map[string]*config
	}{
		{
			name:       "Set response params for new bidder",
			bidderName: "testBidder",
			responseParams: map[string]BidderParamMapper{
				"param1": {
					Location: "location",
				},
			},
			expected: map[string]*config{
				"testBidder": {
					responseParams: map[string]BidderParamMapper{
						"param1": {
							Location: "location",
						},
					},
				},
			},
		},
		{
			name:       "Set response params for existing bidder",
			bidderName: "existingBidder",
			responseParams: map[string]BidderParamMapper{
				"param2": {
					Location: "location",
				},
			},
			expected: map[string]*config{
				"existingBidder": {
					responseParams: map[string]BidderParamMapper{
						"param2": {
							Location: "location",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bcfg := &BidderConfig{
				bidderConfigMap: make(map[string]*config),
			}
			bcfg.SetResponseParams(tt.bidderName, tt.responseParams)
			assert.Equal(t, tt.expected, bcfg.bidderConfigMap)
		})
	}
}

func TestGetResponseParams(t *testing.T) {
	tests := []struct {
		name            string
		bidderName      string
		bidderConfigMap map[string]*config
		expected        map[string]BidderParamMapper
	}{
		{
			name:       "Get response params for existing bidder",
			bidderName: "existingBidder",
			bidderConfigMap: map[string]*config{
				"existingBidder": {
					responseParams: map[string]BidderParamMapper{
						"param1": {
							Location: "location",
						},
					},
				},
			},
			expected: map[string]BidderParamMapper{
				"param1": {
					Location: "location",
				},
			},
		},
		{
			name:       "Get response params for non-existing bidder",
			bidderName: "nonExistingBidder",
			bidderConfigMap: map[string]*config{
				"existingBidder": {
					responseParams: map[string]BidderParamMapper{
						"param1": {
							Location: "location",
						},
					},
				},
			},
			expected: nil,
		},
		{
			name:            "Get response params for empty bidder config map",
			bidderName:      "anyBidder",
			bidderConfigMap: map[string]*config{},
			expected:        nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bcfg := &BidderConfig{
				bidderConfigMap: tt.bidderConfigMap,
			}
			got := bcfg.GetResponseParams(tt.bidderName)
			assert.Equal(t, tt.expected, got)
		})
	}
}
