package ortbbidder

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetValue(t *testing.T) {
	type args struct {
		node     JSONNode
		location string
		value    any
	}
	type want struct {
		node   JSONNode
		status bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "set_nil_value",
			args: args{
				node:     JSONNode{},
				location: "key",
				value:    nil,
			},
			want: want{
				status: false,
				node:   JSONNode{},
			},
		},
		{
			name: "set_value_in_empty_location",
			args: args{
				node:     JSONNode{},
				location: "",
				value:    123,
			},
			want: want{
				status: false,
				node:   JSONNode{},
			},
		},
		{
			name: "set_value_in_invalid_location",
			args: args{
				node:     JSONNode{},
				location: "......",
				value:    123,
			},
			want: want{
				status: false,
				node:   JSONNode{},
			},
		},
		{
			name: "set_value_in_invalid_location_modifies_node",
			args: args{
				node:     JSONNode{},
				location: "key...",
				value:    123,
			},
			want: want{
				status: false,
				node: JSONNode{
					"key": map[string]interface{}{},
				},
			},
		},
		{
			name: "set_value_at_root_level_in_empty_node",
			args: args{
				node:     JSONNode{},
				location: "key",
				value:    123,
			},
			want: want{
				status: true,
				node:   JSONNode{"key": 123},
			},
		},
		{
			name: "set_value_at_root_level_in_non-empty_node",
			args: args{
				node:     JSONNode{"oldKey": "oldValue"},
				location: "key",
				value:    123,
			},
			want: want{
				status: true,
				node:   JSONNode{"oldKey": "oldValue", "key": 123},
			},
		},
		{
			name: "set_value_at_non-root_level_in_non-json_node",
			args: args{
				node:     JSONNode{"rootKey": "rootValue"},
				location: "rootKey.key",
				value:    123,
			},
			want: want{
				status: false,
				node:   JSONNode{"rootKey": "rootValue"},
			},
		},
		{
			name: "set_value_at_non-root_level_in_json_node",
			args: args{
				node: JSONNode{"rootKey": map[string]interface{}{
					"oldKey": "oldValue",
				}},
				location: "rootKey.newKey",
				value:    123,
			},
			want: want{
				status: true,
				node: JSONNode{"rootKey": map[string]interface{}{
					"oldKey": "oldValue",
					"newKey": 123,
				}},
			},
		},
		{
			name: "set_value_at_non-root_level_in_nested-json_node",
			args: args{
				node: JSONNode{"rootKey": map[string]interface{}{
					"parentKey1": map[string]interface{}{
						"innerKey": "innerValue",
					},
				}},
				location: "rootKey.parentKey2",
				value:    "newKeyValue",
			},
			want: want{
				status: true,
				node: JSONNode{"rootKey": map[string]interface{}{
					"parentKey1": map[string]interface{}{
						"innerKey": "innerValue",
					},
					"parentKey2": "newKeyValue",
				}},
			},
		},
		{
			name: "override_existing_key's_value",
			args: args{
				node: JSONNode{"rootKey": map[string]interface{}{
					"parentKey": map[string]interface{}{
						"innerKey": "innerValue",
					},
				}},
				location: "rootKey.parentKey",
				value:    "newKeyValue",
			},
			want: want{
				status: true,
				node: JSONNode{"rootKey": map[string]interface{}{
					"parentKey": "newKeyValue",
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := setValueAtLocation(tt.args.node, tt.args.location, tt.args.value)
			assert.Equalf(t, tt.want.node, tt.args.node, "SetValue failed to update node object")
			assert.Equalf(t, tt.want.status, got, "SetValue returned invalid status")
		})
	}
}

func Test_updateBidderParamsMapper(t *testing.T) {
	type args struct {
		mapper       bidderParamMapper
		fileBytesMap JSONNode
		bidderName   string
	}
	type want struct {
		mapper bidderParamMapper
		err    error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "properties_missing_from_fileContents",
			args: args{
				mapper: bidderParamMapper{},
				fileBytesMap: map[string]interface{}{
					"title": "test bidder parameters",
				},
				bidderName: "testbidder",
			},
			want: want{
				mapper: bidderParamMapper{},
				err:    nil,
			},
		},
		{
			name: "properties_data_type_invalid",
			args: args{
				mapper: bidderParamMapper{},
				fileBytesMap: map[string]interface{}{
					"title":      "test bidder parameters",
					"properties": "type invalid",
				},
				bidderName: "testbidder",
			},
			want: want{
				mapper: bidderParamMapper{},
				err:    fmt.Errorf("error:[invalid_json_file_content_malformed_properties] bidderName:[testbidder]"),
			},
		},
		{
			name: "bidder-params_data_type_invalid",
			args: args{
				mapper: bidderParamMapper{},
				fileBytesMap: JSONNode{
					"title": "test bidder parameters",
					"properties": JSONNode{
						"adunitid": "invalid-type",
					},
				},
				bidderName: "testbidder",
			},
			want: want{
				mapper: bidderParamMapper{},
				err:    fmt.Errorf("error:[invalid_json_file_content] bidder:[testbidder] bidderParam:[adunitid]"),
			},
		},
		{
			name: "bidder-params_properties_is_not_provided",
			args: args{
				mapper: bidderParamMapper{},
				fileBytesMap: JSONNode{
					"title": "test bidder parameters",
					"properties": JSONNode{
						"adunitid": JSONNode{
							"type": "string",
						},
					},
				},
				bidderName: "testbidder",
			},
			want: want{
				mapper: bidderParamMapper{},
				err:    nil,
			},
		},
		{
			name: "bidder-params_location_is_not_in_string",
			args: args{
				mapper: bidderParamMapper{},
				fileBytesMap: JSONNode{
					"title": "test bidder parameters",
					"properties": JSONNode{
						"adunitid": JSONNode{
							"type":     "string",
							"location": 100,
						},
					},
				},
				bidderName: "testbidder",
			},
			want: want{
				mapper: bidderParamMapper{},
				err:    fmt.Errorf("error:[incorrect_location_in_bidderparam] bidder:[testbidder] bidderParam:[adunitid]"),
			},
		},
		{
			name: "bidder-params_location_not_starts_with_req",
			args: args{
				mapper: bidderParamMapper{},
				fileBytesMap: JSONNode{
					"title": "test bidder parameters",
					"properties": JSONNode{
						"adunitid": JSONNode{
							"type":     "string",
							"location": "imp.ext.adunit",
						},
					},
				},
				bidderName: "testbidder",
			},
			want: want{
				mapper: bidderParamMapper{},
				err:    fmt.Errorf("error:[incorrect_location_in_bidderparam] bidder:[testbidder] bidderParam:[adunitid]"),
			},
		},
		{
			name: "set_bidder-params_location_in_mapper",
			args: args{
				mapper: bidderParamMapper{},
				fileBytesMap: JSONNode{
					"title": "test bidder parameters",
					"properties": JSONNode{
						"adunitid": JSONNode{
							"type":     "string",
							"location": "req.app.adunitid",
						},
					},
				},
				bidderName: "testbidder",
			},
			want: want{
				mapper: bidderParamMapper{
					"testbidder": map[string]string{
						"adunitid": "req.app.adunitid",
					},
				},
				err: nil,
			},
		},
		{
			name: "set_multiple_bidder-params_and_locations_in_mapper",
			args: args{
				mapper: bidderParamMapper{},
				fileBytesMap: JSONNode{
					"title": "test bidder parameters",
					"properties": JSONNode{
						"adunitid": JSONNode{
							"type":     "string",
							"location": "req.app.adunitid",
						},
						"slotname": JSONNode{
							"type":     "string",
							"location": "req.ext.slot",
						},
					},
				},
				bidderName: "testbidder",
			},
			want: want{
				mapper: bidderParamMapper{
					"testbidder": map[string]string{
						"adunitid": "req.app.adunitid",
						"slotname": "req.ext.slot",
					},
				},
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := updateBidderParamsMapper(tt.args.mapper, tt.args.fileBytesMap, tt.args.bidderName)
			assert.Equalf(t, tt.want.err, err, "updateBidderParamsMapper returned unexpected error")
			assert.Equalf(t, tt.want.mapper, got, "updateBidderParamsMapper returned unexpected mapper")
		})
	}
}

func Test_prepareMapperFromFiles(t *testing.T) {
	var cleanup = func() error {
		err := os.RemoveAll("test")
		return err
	}
	type want struct {
		mapper *Mapper
		err    string
	}
	tests := []struct {
		name    string
		dirPath string
		want    want
		setup   func() error
		cleanup func() error
	}{
		{
			name:    "read_directory_fail",
			dirPath: "invalid-directory",
			want: want{
				mapper: nil,
				err:    "error:[open invalid-directory: no such file or directory] dirPath:[invalid-directory]",
			},
			setup:   func() error { return nil },
			cleanup: func() error { return nil },
		},
		{
			name:    "found_file_without_.json_extension",
			dirPath: "test",
			want: want{
				mapper: nil,
				err:    "error:[invalid_json_file_name] filename:[example.txt]",
			},
			setup: func() error {
				err := os.MkdirAll("test", 0755)
				if err != nil {
					return err
				}
				err = os.WriteFile("test/example.txt", []byte("anything"), 0644)
				if err != nil {
					return err
				}
				return nil
			},
			cleanup: cleanup,
		},
		{
			name:    "oRTB_bidder_not_found",
			dirPath: "test",
			want: want{
				mapper: &Mapper{bidderParamMapper: bidderParamMapper{}},
				err:    "",
			},
			setup: func() error {
				err := os.MkdirAll("test", 0755)
				if err != nil {
					return err
				}
				err = os.WriteFile("test/example.json", []byte("anything"), 0644)
				if err != nil {
					return err
				}
				return nil
			},
			cleanup: cleanup,
		},
		{
			name:    "oRTB_bidder_found_but_invalid_json_present",
			dirPath: "test",
			want: want{
				mapper: nil,
				err:    "invalid character 'a' looking for beginning of value",
			},
			setup: func() error {
				err := os.MkdirAll("test", 0755)
				if err != nil {
					return err
				}
				err = os.WriteFile("test/owortb_test.json", []byte("any-invalid-json-data"), 0644)
				if err != nil {
					return err
				}
				return nil
			},
			cleanup: cleanup,
		},
		{
			name:    "oRTB_bidder_found_but_bidder-params_are_absent",
			dirPath: "test",
			want: want{
				mapper: &Mapper{bidderParamMapper: make(bidderParamMapper)},
				err:    "",
			},
			setup: func() error {
				err := os.MkdirAll("test", 0755)
				if err != nil {
					return err
				}
				err = os.WriteFile("test/owortb_test.json", []byte("{}"), 0644)
				if err != nil {
					return err
				}
				return nil
			},
			cleanup: cleanup,
		},
		{
			name:    "oRTB_bidder_found_but_updateBidderParamsMapper_returns_error",
			dirPath: "test",
			want: want{
				mapper: nil,
				err:    "error:[invalid_json_file_content_malformed_properties] bidderName:[owortb_test]",
			},
			setup: func() error {
				err := os.MkdirAll("test", 0755)
				if err != nil {
					return err
				}
				err = os.WriteFile("test/owortb_test.json", []byte(`{"properties":"invalid-properties"}`), 0644)
				if err != nil {
					return err
				}
				return nil
			},
			cleanup: cleanup,
		},
		{
			name:    "oRTB_bidder_found_and_valid_json_contents_present",
			dirPath: "test",
			want: want{
				mapper: &Mapper{bidderParamMapper: bidderParamMapper{
					"owortb_test": map[string]string{
						"adunitid": "req.app.adunit.id",
						"slotname": "req.ext.slotname",
					},
				}},
				err: "",
			},
			setup: func() error {
				err := os.MkdirAll("test", 0755)
				if err != nil {
					return err
				}
				err = os.WriteFile("test/owortb_test.json", []byte(`
				{
					"title":"ortb bidder",
					"properties": {
						"adunitid": {
							"type": "string",
							"location": "req.app.adunit.id"
						},
						"slotname": {
							"type": "string",
							"location": "req.ext.slotname"
						}
					}
				}
				`), 0644)
				if err != nil {
					return err
				}
				return nil
			},
			cleanup: cleanup,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				err := tt.cleanup()
				if err != nil {
					fmt.Printf("cleanup returned error for test:%s, err:%v , remove 'test' directory manually", tt.name, err)
				}
			}()
			err := tt.setup()
			assert.NoError(t, err, "setup returned unexpected error")
			got, err := prepareMapperFromFiles(tt.dirPath)
			assert.Equal(t, tt.want.mapper, got, "found incorrect mapper")
			assert.Equal(t, len(tt.want.err) == 0, err == nil, "mismatched error")
			if err != nil {
				assert.Equal(t, tt.want.err, err.Error(), "found incorrect error message")
			}
		})
	}
}

func Test_mapBidderParamsInRequest(t *testing.T) {
	type args struct {
		requestBody []byte
		mapper      map[string]string
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
				mapper: map[string]string{
					"adunit": "req.ext.adunit",
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
				mapper: map[string]string{
					"adunit": "req.ext.adunit",
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
				mapper: map[string]string{
					"adunit": "req.ext.adunit",
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
				mapper: map[string]string{
					"adunit": "req.ext.adunit",
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
				mapper: map[string]string{
					"adunit": "req.ext.adunit",
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
				mapper: map[string]string{
					"slot": "req.ext.slot",
				},
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
				mapper: map[string]string{
					"adunit": "req.ext.adunit",
				},
			},
			want: want{
				err:         "",
				requestBody: json.RawMessage(`{"ext":{"adunit":123},"imp":[{"ext":{"bidder":{}}}]}`),
			},
		},
		{
			name: "mapper_contains_bidder_param_invalid_location",
			args: args{
				requestBody: json.RawMessage(`{"imp":[{"ext":{"bidder":{"adunit":123}}}]}`),
				mapper: map[string]string{
					"adunit": "imp.ext.adunit",
				},
			},
			want: want{
				err:         "error:[invalid_bidder_param_location] param:[adunit] location:[imp.ext.adunit]",
				requestBody: nil,
			},
		},
		{
			name: "do_not_delete_bidder_param_if_failed_to_set_value",
			args: args{
				requestBody: json.RawMessage(`{"imp":[{"ext":{"bidder":{"adunit":123}}}]}`),
				mapper: map[string]string{
					"adunit": "req....",
				},
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
				mapper: map[string]string{
					"adunit":  "req.adunit.id",
					"slot":    "req.imp.tagid",
					"wrapper": "req.app.ext",
				},
			},
			want: want{
				err:         "",
				requestBody: json.RawMessage(`{"adunit":{"id":123},"app":{"ext":{"profile":1,"pubid":5890},"name":"sampleapp"},"imp":[{"ext":{"bidder":{"paramWithoutLocation":"value"}},"tagid":"test_slot"}]}`),
			},
		},
		{
			name: "multi_imps_bidder_params_mapping",
			args: args{
				requestBody: json.RawMessage(`{"app":{"name":"sampleapp"},"imp":[{"tagid":"tagid_1","ext":{"bidder":{"paramWithoutLocation":"value","adunit":111,"slot":"test_slot_1","wrapper":{"pubid":5890,"profile":1}}}},{"tagid":"tagid_2","ext":{"bidder":{"slot":"test_slot_2","adunit":222}}}]}`),
				mapper: map[string]string{
					"adunit":  "req.adunit.id",
					"slot":    "req.imp.tagid",
					"wrapper": "req.app.ext",
				},
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
				mapper: map[string]string{
					"adunit": "req.adunit.id",
				},
			},
			want: want{
				err:         "",
				requestBody: json.RawMessage(`{"adunit":{"id":222},"app":{"name":"sampleapp"},"imp":[{"ext":{"bidder":{"paramWithoutLocation":"value"}},"tagid":"tagid_1"},{"ext":{"bidder":{}},"tagid":"tagid_2"}]}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapBidderParamsInRequest(tt.args.requestBody, tt.args.mapper)
			assert.Equal(t, string(tt.want.requestBody), string(got), "mismatched request body")
			assert.Equal(t, len(tt.want.err) == 0, err == nil, "mismatched error")
			if err != nil {
				assert.Equal(t, err.Error(), tt.want.err, "mismatched error string")
			}
		})
	}
}

func TestInitMapper(t *testing.T) {
	tests := []struct {
		name    string
		dirPath string
		want    *Mapper
		wantErr bool
	}{
		{
			name:    "test_initMapper",
			dirPath: "../../static/bidder-params/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InitMapper(tt.dirPath)
			assert.NotNil(t, got, "mapper should be non-nil")
			assert.Nil(t, err, "error should be nil")
		})
	}
}
