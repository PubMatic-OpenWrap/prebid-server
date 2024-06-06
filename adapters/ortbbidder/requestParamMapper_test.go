package ortbbidder

import (
	"testing"

	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/stretchr/testify/assert"
)

func TestSetRequestParams(t *testing.T) {
	type args struct {
		request      map[string]any
		bidderParams map[string]any
		paramsMapper map[string]bidderparams.BidderParamMapper
		paramIndices []int
	}
	type want struct {
		request      map[string]any
		bidderParams map[string]any
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "bidder_param_missing",
			args: args{
				request: map[string]any{
					"id": "req_1",
				},
				bidderParams: map[string]any{
					"param": "value",
				},
				paramsMapper: nil,
			},
			want: want{
				request: map[string]any{
					"id": "req_1",
				},
				bidderParams: map[string]any{
					"param": "value",
				},
			},
		},
		{
			name: "request_level_param_set_successfully",
			args: args{
				request: map[string]any{
					"id": "req_1",
				},
				bidderParams: map[string]any{
					"param": "value",
				},
				paramsMapper: func() map[string]bidderparams.BidderParamMapper {
					mapper := bidderparams.BidderParamMapper{}
					mapper.SetLocation("param")
					return map[string]bidderparams.BidderParamMapper{
						"param": mapper,
					}
				}(),
				paramIndices: nil,
			},
			want: want{
				request: map[string]any{
					"param": "value",
					"id":    "req_1",
				},
				bidderParams: map[string]any{},
			},
		},
		{
			name: "imp_level_param_set_successfully",
			args: args{
				request: map[string]any{
					"id": "req_1",
					"imp": []any{
						map[string]any{},
					},
				},
				bidderParams: map[string]any{
					"param": "value",
				},
				paramsMapper: func() map[string]bidderparams.BidderParamMapper {
					mapper := bidderparams.BidderParamMapper{}
					mapper.SetLocation("imp.#.param")
					return map[string]bidderparams.BidderParamMapper{
						"param": mapper,
					}
				}(),
				paramIndices: []int{0},
			},
			want: want{
				request: map[string]any{
					"id": "req_1",
					"imp": []any{
						map[string]any{
							"param": "value",
						},
					},
				},
				bidderParams: map[string]any{},
			},
		},
		{
			name: "attempt_to_set_imp_level_param_in_invalid_index_position",
			args: args{
				request: map[string]any{
					"id": "req_1",
					"imp": []any{
						map[string]any{},
					},
				},
				bidderParams: map[string]any{
					"param": "value",
				},
				paramsMapper: func() map[string]bidderparams.BidderParamMapper {
					mapper := bidderparams.BidderParamMapper{}
					mapper.SetLocation("imp.#.param")
					return map[string]bidderparams.BidderParamMapper{
						"param": mapper,
					}
				}(),
				paramIndices: []int{1},
			},
			want: want{
				request: map[string]any{
					"id": "req_1",
					"imp": []any{
						map[string]any{},
					},
				},
				bidderParams: map[string]any{
					"param": "value",
				},
			},
		},
		{
			name: "attempt_to_set_imp_level_param_when_no_index_is_given",
			args: args{
				request: map[string]any{
					"id": "req_1",
					"imp": []any{
						map[string]any{},
					},
				},
				bidderParams: map[string]any{
					"param": "value",
				},
				paramsMapper: func() map[string]bidderparams.BidderParamMapper {
					mapper := bidderparams.BidderParamMapper{}
					mapper.SetLocation("imp.#.param")
					return map[string]bidderparams.BidderParamMapper{
						"param": mapper,
					}
				}(),
				paramIndices: []int{},
			},
			want: want{
				request: map[string]any{
					"id": "req_1",
					"imp": []any{
						map[string]any{},
					},
				},
				bidderParams: map[string]any{
					"param": "value",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setRequestParams(tt.args.request, tt.args.bidderParams, tt.args.paramsMapper, tt.args.paramIndices)
			assert.Equal(t, tt.want.bidderParams, tt.args.bidderParams, "mismatched bidderparams")
			assert.Equal(t, tt.want.request, tt.args.request, "mismatched request")
		})
	}
}

func TestGetImpExtBidderParams(t *testing.T) {
	type args struct {
		imp map[string]any
	}
	tests := []struct {
		name string
		args args
		want map[string]any
	}{
		{
			name: "ext_key_absent_in_imp",
			args: args{
				imp: map[string]any{},
			},
			want: nil,
		},
		{
			name: "invalid_ext_key_in_imp",
			args: args{
				imp: map[string]any{
					"ext": "invalid",
				},
			},
			want: nil,
		},
		{
			name: "bidder_key_absent_in_imp_ext",
			args: args{
				imp: map[string]any{
					"ext": map[string]any{},
				},
			},
			want: nil,
		},
		{
			name: "bidder_key_present_in_imp_ext",
			args: args{
				imp: map[string]any{
					"ext": map[string]any{
						"bidder": map[string]any{
							"param": "value",
						},
					},
				},
			},
			want: map[string]any{
				"param": "value",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getImpExtBidderParams(tt.args.imp)
			assert.Equal(t, tt.want, got, "mismatched bidder-params")
		})
	}
}
