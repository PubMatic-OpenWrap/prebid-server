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
		node     map[string]any
		location []string
		value    any
	}
	type want struct {
		node   map[string]any
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
				node:     map[string]any{},
				location: []string{"key"},
				value:    nil,
			},
			want: want{
				status: false,
				node:   map[string]any{},
			},
		},
		{
			name: "set_value_in_empty_location",
			args: args{
				node:     map[string]any{},
				location: []string{},
				value:    123,
			},
			want: want{
				status: false,
				node:   map[string]any{},
			},
		},
		{
			name: "set_value_in_invalid_location_modifies_node",
			args: args{
				node:     map[string]any{},
				location: []string{"key", ""},
				value:    123,
			},
			want: want{
				status: false,
				node: map[string]any{
					"key": map[string]any{},
				},
			},
		},
		{
			name: "set_value_at_root_level_in_empty_node",
			args: args{
				node:     map[string]any{},
				location: []string{"key"},
				value:    123,
			},
			want: want{
				status: true,
				node:   map[string]any{"key": 123},
			},
		},
		{
			name: "set_value_at_root_level_in_non-empty_node",
			args: args{
				node:     map[string]any{"oldKey": "oldValue"},
				location: []string{"key"},
				value:    123,
			},
			want: want{
				status: true,
				node:   map[string]any{"oldKey": "oldValue", "key": 123},
			},
		},
		{
			name: "set_value_at_non-root_level_in_non-json_node",
			args: args{
				node:     map[string]any{"rootKey": "rootValue"},
				location: []string{"rootKey", "key"},
				value:    123,
			},
			want: want{
				status: false,
				node:   map[string]any{"rootKey": "rootValue"},
			},
		},
		{
			name: "set_value_at_non-root_level_in_json_node",
			args: args{
				node: map[string]any{"rootKey": map[string]any{
					"oldKey": "oldValue",
				}},
				location: []string{"rootKey", "newKey"},
				value:    123,
			},
			want: want{
				status: true,
				node: map[string]any{"rootKey": map[string]any{
					"oldKey": "oldValue",
					"newKey": 123,
				}},
			},
		},
		{
			name: "set_value_at_non-root_level_in_nested-json_node",
			args: args{
				node: map[string]any{"rootKey": map[string]any{
					"parentKey1": map[string]any{
						"innerKey": "innerValue",
					},
				}},
				location: []string{"rootKey", "parentKey2"},
				value:    "newKeyValue",
			},
			want: want{
				status: true,
				node: map[string]any{"rootKey": map[string]any{
					"parentKey1": map[string]any{
						"innerKey": "innerValue",
					},
					"parentKey2": "newKeyValue",
				}},
			},
		},
		{
			name: "override_existing_key's_value",
			args: args{
				node: map[string]any{"rootKey": map[string]any{
					"parentKey": map[string]any{
						"innerKey": "innerValue",
					},
				}},
				location: []string{"rootKey", "parentKey"},
				value:    "newKeyValue",
			},
			want: want{
				status: true,
				node: map[string]any{"rootKey": map[string]any{
					"parentKey": "newKeyValue",
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := setValue(tt.args.node, tt.args.location, tt.args.value)
			assert.Equalf(t, tt.want.node, tt.args.node, "SetValue failed to update node object")
			assert.Equalf(t, tt.want.status, got, "SetValue returned invalid status")
		})
	}
}

func Test_setBidderParamsDetails(t *testing.T) {
	type args struct {
		mapper       bidderParamMapper
		fileBytesMap map[string]any
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
				fileBytesMap: map[string]any{
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
				fileBytesMap: map[string]any{
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
				fileBytesMap: map[string]any{
					"title": "test bidder parameters",
					"properties": map[string]any{
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
				fileBytesMap: map[string]any{
					"title": "test bidder parameters",
					"properties": map[string]any{
						"adunitid": map[string]any{
							"type": "string",
						},
					},
				},
				bidderName: "testbidder",
			},
			want: want{
				mapper: bidderParamMapper{
					"testbidder": make(map[string]paramDetails),
				},
				err: nil,
			},
		},
		{
			name: "bidder-params_location_is_not_in_string",
			args: args{
				mapper: bidderParamMapper{},
				fileBytesMap: map[string]any{
					"title": "test bidder parameters",
					"properties": map[string]any{
						"adunitid": map[string]any{
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
			name: "set_bidder-params_location_in_mapper",
			args: args{
				mapper: bidderParamMapper{},
				fileBytesMap: map[string]any{
					"title": "test bidder parameters",
					"properties": map[string]any{
						"adunitid": map[string]any{
							"type":     "string",
							"location": "app.adunitid",
						},
					},
				},
				bidderName: "testbidder",
			},
			want: want{
				mapper: bidderParamMapper{
					"testbidder": map[string]paramDetails{
						"adunitid": {location: []string{"app", "adunitid"}},
					},
				},
				err: nil,
			},
		},
		{
			name: "set_multiple_bidder-params_and_locations_in_mapper",
			args: args{
				mapper: bidderParamMapper{},
				fileBytesMap: map[string]any{
					"title": "test bidder parameters",
					"properties": map[string]any{
						"adunitid": map[string]any{
							"type":     "string",
							"location": "app.adunitid",
						},
						"slotname": map[string]any{
							"type":     "string",
							"location": "ext.slot",
						},
					},
				},
				bidderName: "testbidder",
			},
			want: want{
				mapper: bidderParamMapper{
					"testbidder": map[string]paramDetails{
						"adunitid": {location: []string{"app", "adunitid"}},
						"slotname": {location: []string{"ext", "slot"}},
					},
				},
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.mapper.setBidderParamsDetails(tt.args.bidderName, tt.args.fileBytesMap)
			assert.Equalf(t, tt.want.err, err, "updateBidderParamsMapper returned unexpected error")
			assert.Equalf(t, tt.want.mapper, tt.args.mapper, "updateBidderParamsMapper returned unexpected mapper")
		})
	}
}

func Test_prepareMapperFromFiles(t *testing.T) {
	var cleanup = func() error {
		err := os.RemoveAll("test")
		return err
	}
	type want struct {
		mapper *mapper
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
				mapper: &mapper{bidderParamMapper: bidderParamMapper{}},
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
				err:    "error:[fail_to_read_file] dir:[test] filename:[owortb_test.json] err:[invalid character 'a' looking for beginning of value]",
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
				mapper: &mapper{bidderParamMapper: make(bidderParamMapper)},
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
				mapper: &mapper{bidderParamMapper: bidderParamMapper{
					"owortb_test": map[string]paramDetails{
						"adunitid": {location: []string{"app", "adunit", "id"}},
						"slotname": {location: []string{"ext", "slotname"}},
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
							"location": "app.adunit.id"
						},
						"slotname": {
							"type": "string",
							"location": "ext.slotname"
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
		mapper      map[string]paramDetails
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
				mapper: map[string]paramDetails{
					"adunit": {location: []string{"ext"}},
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
				mapper: map[string]paramDetails{
					"adunit": {location: []string{"ext"}},
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
				mapper: map[string]paramDetails{
					"adunit": {location: []string{"ext"}},
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
				mapper: map[string]paramDetails{
					"adunit": {location: []string{"ext"}},
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
				mapper: map[string]paramDetails{
					"adunit": {location: []string{"ext"}},
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
				mapper: map[string]paramDetails{
					"slot": {location: []string{"ext", "slot"}},
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
				mapper: map[string]paramDetails{
					"adunit": {location: []string{"ext", "adunit"}},
				},
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
				mapper: map[string]paramDetails{
					"adunit": {location: []string{"req", "", ""}},
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
				mapper: map[string]paramDetails{
					"adunit":  {location: []string{"adunit", "id"}},
					"slot":    {location: []string{"imp", "tagid"}},
					"wrapper": {location: []string{"app", "ext"}},
				},
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
				mapper: map[string]paramDetails{
					"wrapper": {location: []string{"appsite", "wrapper"}},
				},
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
				mapper: map[string]paramDetails{
					"wrapper": {location: []string{"appsite", "wrapper"}},
				},
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
				mapper: map[string]paramDetails{
					"adunit":  {location: []string{"adunit", "id"}},
					"slot":    {location: []string{"imp", "tagid"}},
					"wrapper": {location: []string{"app", "ext"}},
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
				mapper: map[string]paramDetails{
					"adunit": {location: []string{"adunit", "id"}},
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
		want    *mapper
		wantErr bool
	}{
		{
			name:    "test_initMapper",
			dirPath: "../../static/bidder-params/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitMapper(tt.dirPath)
			assert.Nil(t, err, "error should be nil")
		})
	}
}

func Test_readFile(t *testing.T) {
	var cleanup = func() error {
		err := os.RemoveAll("test")
		return err
	}
	type args struct {
		dirPath string
		file    string
	}
	type want struct {
		err  bool
		node map[string]any
	}
	tests := []struct {
		name    string
		args    args
		want    want
		setup   func() error
		cleanup func() error
	}{
		{
			name: "successful_readfile",
			args: args{
				dirPath: "test",
				file:    "owortb.json",
			},
			want: want{
				err: false,
				node: map[string]any{
					"title": "ortb bidder",
					"properties": map[string]any{
						"adunitid": map[string]any{
							"type":     "string",
							"location": "req.app.adunit.id",
						},
					},
				},
			},
			setup: func() error {
				err := os.MkdirAll("test", 0755)
				if err != nil {
					return err
				}
				err = os.WriteFile("test/owortb.json", []byte(`
				{
					"title":"ortb bidder",
					"properties": {
						"adunitid": {
							"type": "string",
							"location": "req.app.adunit.id"
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
		{
			name: "fail_readfile",
			args: args{
				dirPath: "test",
				file:    "owortb.json",
			},
			want: want{
				err:  true,
				node: nil,
			},
			setup: func() error {
				err := os.MkdirAll("test", 0755)
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
			got, err := readFile(tt.args.dirPath, tt.args.file)
			assert.Equal(t, tt.want.err, err != nil, "mismatched error")
			assert.Equal(t, tt.want.node, got, "mismatched map[string]any")
		})
	}
}

func Test_applyConditionalMapping(t *testing.T) {
	type args struct {
		requestBodyMap map[string]any
		details        paramDetails
	}
	tests := []struct {
		name string
		args args
		want paramDetails
	}{
		{
			name: "empty_location_for_bidder_param",
			args: args{
				requestBodyMap: map[string]any{
					"app": map[string]any{},
				},
				details: paramDetails{},
			},
			want: paramDetails{},
		},
		{
			name: "empty_request_body",
			args: args{
				requestBodyMap: map[string]any{},
				details: paramDetails{
					location: []string{"appsite", "publisher", "id"},
				},
			},
			want: paramDetails{
				location: []string{"appsite", "publisher", "id"},
			},
		},
		{
			name: "app_object_present_in_request_body",
			args: args{
				requestBodyMap: map[string]any{
					"app": map[string]any{
						"name": "sample_app",
					},
				},
				details: paramDetails{
					location: []string{"appsite", "publisher", "id"},
				},
			},
			want: paramDetails{
				location: []string{"app", "publisher", "id"},
			},
		},
		{
			name: "app_object_absent_in_request_body",
			args: args{
				requestBodyMap: map[string]any{
					"device": map[string]any{
						"name": "sample_app",
					},
				},
				details: paramDetails{
					location: []string{"appsite", "publisher", "id"},
				},
			},
			want: paramDetails{
				location: []string{"site", "publisher", "id"},
			},
		},
		{
			name: "site_object_present_in_request_body",
			args: args{
				requestBodyMap: map[string]any{
					"site": map[string]any{
						"name": "sample_app",
					},
				},
				details: paramDetails{
					location: []string{"appsite", "publisher", "id"},
				},
			},
			want: paramDetails{
				location: []string{"site", "publisher", "id"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyConditionalMapping(tt.args.requestBodyMap, tt.args.details)
			assert.Equal(t, tt.want, got, "mismatched paramDetails")
		})
	}
}
