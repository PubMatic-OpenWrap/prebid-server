package bidderparams

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBidderRequestProperties(t *testing.T) {
	type fields struct {
		biddersConfig *BidderConfig
	}
	type args struct {
		bidderName string
	}
	type want struct {
		RequestParams map[string]BidderParamMapper
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
					BidderConfigMap: nil,
				},
			},
			args: args{
				bidderName: "test",
			},
			want: want{
				RequestParams: nil,
			},
		},
		{
			name: "BidderName_absent_in_biddersConfigMap",
			fields: fields{
				biddersConfig: &BidderConfig{
					BidderConfigMap: map[string]*Config{
						"ortb": {},
					},
				},
			},
			args: args{
				bidderName: "test",
			},
			want: want{
				RequestParams: nil,
			},
		},
		{
			name: "BidderName_present_but_Config_is_nil",
			fields: fields{
				biddersConfig: &BidderConfig{
					BidderConfigMap: map[string]*Config{
						"ortb": nil,
					},
				},
			},
			args: args{
				bidderName: "test",
			},
			want: want{
				RequestParams: nil,
			},
		},
		{
			name: "BidderName_present_in_biddersConfigMap",
			fields: fields{
				biddersConfig: &BidderConfig{
					BidderConfigMap: map[string]*Config{
						"test": {
							RequestParams: map[string]BidderParamMapper{
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
				RequestParams: map[string]BidderParamMapper{
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
			assert.Equal(t, tt.want.RequestParams, params, "mismatched RequestParams")
		})
	}
}

func TestGetResponseParams(t *testing.T) {
	tests := []struct {
		name            string
		bidderName      string
		BidderConfigMap map[string]*Config
		expected        map[string]BidderParamMapper
	}{
		{
			name:       "Get response params for existing bidder",
			bidderName: "existingBidder",
			BidderConfigMap: map[string]*Config{
				"existingBidder": {
					ResponseParams: map[string]BidderParamMapper{
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
			BidderConfigMap: map[string]*Config{
				"existingBidder": {
					ResponseParams: map[string]BidderParamMapper{
						"param1": {
							Location: "location",
						},
					},
				},
			},
			expected: map[string]BidderParamMapper{},
		},
		{
			name:            "Get response params for empty bidder Config map",
			bidderName:      "anyBidder",
			BidderConfigMap: map[string]*Config{},
			expected:        nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bcfg := &BidderConfig{
				BidderConfigMap: tt.BidderConfigMap,
			}
			got := bcfg.GetResponseParams(tt.bidderName)
			assert.Equal(t, tt.expected, got)
		})
	}
}
