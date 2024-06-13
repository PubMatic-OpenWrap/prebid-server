package ortbbidder

import (
	"encoding/json"
	"testing"

	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/stretchr/testify/assert"
)

func TestSetRequestParams(t *testing.T) {
	type args struct {
		requestBody []byte
		mapper      map[string]bidderparams.BidderParamMapper
	}
	type want struct {
		err         string
		requestBody []byte
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "empty_mapper",
			args: args{
				requestBody: json.RawMessage(`{"imp":[{"ext":{"bidder":{}}}]}`),
			},
			want: want{
				err:         "",
				requestBody: json.RawMessage(`{"imp":[{"ext":{"bidder":{}}}]}`),
			},
		},
		{
			name: "nil_requestbody",
			args: args{
				requestBody: nil,
				mapper: map[string]bidderparams.BidderParamMapper{
					"adunit": {},
				},
			},
			want: want{
				err: "unexpected end of JSON input",
			},
		},
		{
			name: "requestbody_has_invalid_imps",
			args: args{
				requestBody: json.RawMessage(`{"imp":{"id":"1"}}`),
				mapper: map[string]bidderparams.BidderParamMapper{
					"adunit": {},
				},
			},
			want: want{
				err: "error:[invalid_imp_found_in_requestbody], imp:[map[id:1]]",
			},
		},
		{
			name: "missing_imp_ext",
			args: args{
				requestBody: json.RawMessage(`{"imp":[{}]}`),
				mapper: map[string]bidderparams.BidderParamMapper{
					"adunit": {},
				},
			},
			want: want{
				err:         "",
				requestBody: json.RawMessage(`{"imp":[{}]}`),
			},
		},
		{
			name: "missing_bidder_in_imp_ext",
			args: args{
				requestBody: json.RawMessage(`{"imp":[{"ext":{}}]}`),
				mapper: map[string]bidderparams.BidderParamMapper{
					"adunit": {},
				},
			},
			want: want{
				err:         "",
				requestBody: json.RawMessage(`{"imp":[{"ext":{}}]}`),
			},
		},
		{
			name: "missing_bidderparams_in_imp_ext",
			args: args{
				requestBody: json.RawMessage(`{"imp":[{"ext":{"bidder":{}}}]}`),
				mapper: map[string]bidderparams.BidderParamMapper{
					"adunit": {},
				},
			},
			want: want{
				err:         "",
				requestBody: json.RawMessage(`{"imp":[{"ext":{"bidder":{}}}]}`),
			},
		},
		{
			name: "mapper_not_contains_bidder_param_location",
			args: args{
				requestBody: json.RawMessage(`{"imp":[{"ext":{"bidder":{"adunit":123}}}]}`),
				mapper: func() map[string]bidderparams.BidderParamMapper {
					bpm := bidderparams.BidderParamMapper{}
					bpm.SetLocation([]string{"ext"})
					return map[string]bidderparams.BidderParamMapper{
						"slot": bpm,
					}
				}(),
			},
			want: want{
				err:         "",
				requestBody: json.RawMessage(`{"imp":[{"ext":{"bidder":{"adunit":123}}}]}`),
			},
		},
		{
			name: "mapper_contains_bidder_param_location",
			args: args{
				requestBody: json.RawMessage(`{"imp":[{"ext":{"bidder":{"adunit":123}}}]}`),
				mapper: func() map[string]bidderparams.BidderParamMapper {
					bpm := bidderparams.BidderParamMapper{}
					bpm.SetLocation([]string{"ext", "adunit"})
					return map[string]bidderparams.BidderParamMapper{
						"adunit": bpm,
					}
				}(),
			},
			want: want{
				err:         "",
				requestBody: json.RawMessage(`{"ext":{"adunit":123},"imp":[{"ext":{"bidder":{}}}]}`),
			},
		},
		{
			name: "do_not_delete_bidder_param_if_failed_to_set_value",
			args: args{
				requestBody: json.RawMessage(`{"imp":[{"ext":{"bidder":{"adunit":123}}}]}`),
				mapper: func() map[string]bidderparams.BidderParamMapper {
					bpm := bidderparams.BidderParamMapper{}
					bpm.SetLocation([]string{"req", "", ""})
					return map[string]bidderparams.BidderParamMapper{
						"adunit": bpm,
					}
				}(),
			},
			want: want{
				err:         "",
				requestBody: json.RawMessage(`{"imp":[{"ext":{"bidder":{"adunit":123}}}]}`),
			},
		},
		{
			name: "set_multiple_bidder_params",
			args: args{
				requestBody: json.RawMessage(`{"app":{"name":"sampleapp"},"imp":[{"tagid":"oldtagid","ext":{"bidder":{"paramWithoutLocation":"value","adunit":123,"slot":"test_slot","wrapper":{"pubid":5890,"profile":1}}}}]}`),
				mapper: func() map[string]bidderparams.BidderParamMapper {
					adunit := bidderparams.BidderParamMapper{}
					adunit.SetLocation([]string{"adunit", "id"})
					slot := bidderparams.BidderParamMapper{}
					slot.SetLocation([]string{"imp", "tagid"})
					wrapper := bidderparams.BidderParamMapper{}
					wrapper.SetLocation([]string{"app", "ext"})
					return map[string]bidderparams.BidderParamMapper{
						"adunit":  adunit,
						"slot":    slot,
						"wrapper": wrapper,
					}
				}(),
			},
			want: want{
				err:         "",
				requestBody: json.RawMessage(`{"adunit":{"id":123},"app":{"ext":{"profile":1,"pubid":5890},"name":"sampleapp"},"imp":[{"ext":{"bidder":{"paramWithoutLocation":"value"}},"tagid":"test_slot"}]}`),
			},
		},
		{
			name: "conditional_mapping_set_app_object",
			args: args{
				requestBody: json.RawMessage(`{"app":{"name":"sampleapp"},"imp":[{"tagid":"oldtagid","ext":{"bidder":{"paramWithoutLocation":"value","adunit":123,"slot":"test_slot","wrapper":{"pubid":5890,"profile":1}}}}]}`),
				mapper: func() map[string]bidderparams.BidderParamMapper {
					bpm := bidderparams.BidderParamMapper{}
					bpm.SetLocation([]string{"appsite", "wrapper"})
					return map[string]bidderparams.BidderParamMapper{
						"wrapper": bpm,
					}
				}(),
			},
			want: want{
				err:         "",
				requestBody: json.RawMessage(`{"app":{"name":"sampleapp","wrapper":{"profile":1,"pubid":5890}},"imp":[{"ext":{"bidder":{"adunit":123,"paramWithoutLocation":"value","slot":"test_slot"}},"tagid":"oldtagid"}]}`),
			},
		},
		{
			name: "conditional_mapping_set_site_object",
			args: args{
				requestBody: json.RawMessage(`{"site":{"name":"sampleapp"},"imp":[{"tagid":"oldtagid","ext":{"bidder":{"paramWithoutLocation":"value","adunit":123,"slot":"test_slot","wrapper":{"pubid":5890,"profile":1}}}}]}`),
				mapper: func() map[string]bidderparams.BidderParamMapper {
					bpm := bidderparams.BidderParamMapper{}
					bpm.SetLocation([]string{"appsite", "wrapper"})
					return map[string]bidderparams.BidderParamMapper{
						"wrapper": bpm,
					}
				}(),
			},
			want: want{
				err:         "",
				requestBody: json.RawMessage(`{"imp":[{"ext":{"bidder":{"adunit":123,"paramWithoutLocation":"value","slot":"test_slot"}},"tagid":"oldtagid"}],"site":{"name":"sampleapp","wrapper":{"profile":1,"pubid":5890}}}`),
			},
		},
		{
			name: "multi_imps_bidder_params_mapping",
			args: args{
				requestBody: json.RawMessage(`{"app":{"name":"sampleapp"},"imp":[{"tagid":"tagid_1","ext":{"bidder":{"paramWithoutLocation":"value","adunit":111,"slot":"test_slot_1","wrapper":{"pubid":5890,"profile":1}}}},{"tagid":"tagid_2","ext":{"bidder":{"slot":"test_slot_2","adunit":222}}}]}`),
				mapper: func() map[string]bidderparams.BidderParamMapper {
					adunit := bidderparams.BidderParamMapper{}
					adunit.SetLocation([]string{"adunit", "id"})
					slot := bidderparams.BidderParamMapper{}
					slot.SetLocation([]string{"imp", "tagid"})
					wrapper := bidderparams.BidderParamMapper{}
					wrapper.SetLocation([]string{"app", "ext"})
					return map[string]bidderparams.BidderParamMapper{
						"adunit":  adunit,
						"slot":    slot,
						"wrapper": wrapper,
					}
				}(),
			},
			want: want{
				err:         "",
				requestBody: json.RawMessage(`{"adunit":{"id":222},"app":{"ext":{"profile":1,"pubid":5890},"name":"sampleapp"},"imp":[{"ext":{"bidder":{"paramWithoutLocation":"value"}},"tagid":"test_slot_1"},{"ext":{"bidder":{}},"tagid":"test_slot_2"}]}`),
			},
		},
		{
			name: "multi_imps_bidder_params_mapping_override_if_same_param_present",
			args: args{
				requestBody: json.RawMessage(`{"app":{"name":"sampleapp"},"imp":[{"tagid":"tagid_1","ext":{"bidder":{"paramWithoutLocation":"value","adunit":111}}},{"tagid":"tagid_2","ext":{"bidder":{"adunit":222}}}]}`),
				mapper: func() map[string]bidderparams.BidderParamMapper {
					bpm := bidderparams.BidderParamMapper{}
					bpm.SetLocation([]string{"adunit", "id"})
					return map[string]bidderparams.BidderParamMapper{
						"adunit": bpm,
					}
				}(),
			},
			want: want{
				err:         "",
				requestBody: json.RawMessage(`{"adunit":{"id":222},"app":{"name":"sampleapp"},"imp":[{"ext":{"bidder":{"paramWithoutLocation":"value"}},"tagid":"tagid_1"},{"ext":{"bidder":{}},"tagid":"tagid_2"}]}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := setRequestParams(tt.args.requestBody, tt.args.mapper)
			assert.Equal(t, string(tt.want.requestBody), string(got), "mismatched request body")
			assert.Equal(t, len(tt.want.err) == 0, err == nil, "mismatched error")
			if err != nil {
				assert.Equal(t, err.Error(), tt.want.err, "mismatched error string")
			}
		})
	}
}
