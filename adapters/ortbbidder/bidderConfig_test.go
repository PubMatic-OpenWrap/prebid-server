package ortbbidder

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareBidderRequestProperties(t *testing.T) {
	type args struct {
		propertiesMap map[string]any
		bidderName    string
	}
	type want struct {
		requestProperties map[string]bidderProperty
		err               error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "properties_missing_from_fileContents",
			args: args{
				propertiesMap: map[string]any{
					"title": "test bidder parameters",
				},
				bidderName: "testbidder",
			},
			want: want{
				requestProperties: nil,
				err:               nil,
			},
		},
		{
			name: "properties_data_type_invalid",
			args: args{
				propertiesMap: map[string]any{
					"title":      "test bidder parameters",
					"properties": "type invalid",
				},
				bidderName: "testbidder",
			},
			want: want{
				requestProperties: nil,
				err:               fmt.Errorf("error:[invalid_json_file_content_malformed_properties] bidderName:[testbidder]"),
			},
		},
		{
			name: "bidder-params_data_type_invalid",
			args: args{
				propertiesMap: map[string]any{
					"title": "test bidder parameters",
					"properties": map[string]any{
						"adunitid": "invalid-type",
					},
				},
				bidderName: "testbidder",
			},
			want: want{
				requestProperties: nil,
				err:               fmt.Errorf("error:[invalid_json_file_content] bidder:[testbidder] bidderParam:[adunitid]"),
			},
		},
		{
			name: "bidder-params_properties_is_not_provided",
			args: args{
				propertiesMap: map[string]any{
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
				requestProperties: map[string]bidderProperty{},
				err:               nil,
			},
		},
		{
			name: "bidder-params_location_is_not_in_string",
			args: args{
				propertiesMap: map[string]any{
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
				requestProperties: nil,
				err:               fmt.Errorf("error:[incorrect_location_in_bidderparam] bidder:[testbidder] bidderParam:[adunitid]"),
			},
		},
		{
			name: "set_bidder-params_location_in_mapper",
			args: args{
				propertiesMap: map[string]any{
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
				requestProperties: map[string]bidderProperty{
					"adunitid": {location: []string{"app", "adunitid"}},
				},
				err: nil,
			},
		},
		{
			name: "set_multiple_bidder-params_and_locations_in_mapper",
			args: args{
				propertiesMap: map[string]any{
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
				requestProperties: map[string]bidderProperty{
					"adunitid": {location: []string{"app", "adunitid"}},
					"slotname": {location: []string{"ext", "slot"}},
				},
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestProperties, err := prepareBidderRequestProperties(tt.args.bidderName, tt.args.propertiesMap)
			assert.Equalf(t, tt.want.err, err, "updateBidderParamsMapper returned unexpected error")
			assert.Equalf(t, tt.want.requestProperties, requestProperties, "updateBidderParamsMapper returned unexpected mapper")
		})
	}
}

func Test_prepareMapperFromFiles(t *testing.T) {
	var cleanup = func() error {
		err := os.RemoveAll("test")
		return err
	}
	type want struct {
		biddersConfigMap *biddersConfigMap
		err              string
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
				biddersConfigMap: nil,
				err:              "error:[open invalid-directory: no such file or directory] dirPath:[invalid-directory]",
			},
			setup:   func() error { return nil },
			cleanup: func() error { return nil },
		},
		{
			name:    "found_file_without_.json_extension",
			dirPath: "test",
			want: want{
				biddersConfigMap: nil,
				err:              "error:[invalid_json_file_name] filename:[example.txt]",
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
				biddersConfigMap: &biddersConfigMap{biddersConfig: make(map[string]*bidderConfig)},
				err:              "",
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
				biddersConfigMap: nil,
				err:              "error:[fail_to_read_file] dir:[test] filename:[owortb_test.json] err:[invalid character 'a' looking for beginning of value]",
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
				biddersConfigMap: &biddersConfigMap{biddersConfig: map[string]*bidderConfig{
					"owortb_test": {
						requestProperties: nil,
					},
				}},
				err: "",
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
				biddersConfigMap: nil,
				err:              "error:[invalid_json_file_content_malformed_properties] bidderName:[owortb_test]",
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
				biddersConfigMap: &biddersConfigMap{
					biddersConfig: map[string]*bidderConfig{
						"owortb_test": {
							requestProperties: map[string]bidderProperty{
								"adunitid": {location: []string{"app", "adunit", "id"}},
								"slotname": {location: []string{"ext", "slotname"}},
							},
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
			got, err := prepareBiddersConfigMap(tt.dirPath)
			assert.Equal(t, tt.want.biddersConfigMap, got, "found incorrect mapper")
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
		mapper      map[string]bidderProperty
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
				mapper: map[string]bidderProperty{
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
				mapper: map[string]bidderProperty{
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
				mapper: map[string]bidderProperty{
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
				mapper: map[string]bidderProperty{
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
				mapper: map[string]bidderProperty{
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
				mapper: map[string]bidderProperty{
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
				mapper: map[string]bidderProperty{
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
				mapper: map[string]bidderProperty{
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
				mapper: map[string]bidderProperty{
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
				mapper: map[string]bidderProperty{
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
				mapper: map[string]bidderProperty{
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
				mapper: map[string]bidderProperty{
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
				mapper: map[string]bidderProperty{
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

func TestInitBiddersConfigMap(t *testing.T) {
	tests := []struct {
		name    string
		dirPath string
		want    *biddersConfigMap
		wantErr bool
	}{
		{
			name:    "test_InitBiddersConfigMap",
			dirPath: "../../static/bidder-params/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitBiddersConfigMap(tt.dirPath)
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

func TestGetBidderRequestProperties(t *testing.T) {
	type fields struct {
		biddersConfig map[string]*bidderConfig
	}
	type args struct {
		bidderName string
	}
	type want struct {
		requestProperties map[string]bidderProperty
		found             bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "bidderName_absent_in_biddersConfigMap",
			fields: fields{
				biddersConfig: make(map[string]*bidderConfig),
			},
			args: args{
				bidderName: "test",
			},
			want: want{
				requestProperties: nil,
			},
		},
		{
			name: "bidderName_absent_in_biddersConfigMap",
			fields: fields{
				biddersConfig: nil,
			},
			args: args{
				bidderName: "test",
			},
			want: want{
				requestProperties: nil,
			},
		},
		{
			name: "bidderName_present_in_biddersConfigMap",
			fields: fields{
				biddersConfig: map[string]*bidderConfig{
					"test": {
						requestProperties: map[string]bidderProperty{
							"param-1": {
								location: []string{"value-1"},
							},
						},
					},
				},
			},
			args: args{
				bidderName: "test",
			},
			want: want{
				requestProperties: map[string]bidderProperty{
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
			bcm := &biddersConfigMap{
				biddersConfig: tt.fields.biddersConfig,
			}
			properties, found := bcm.getBidderRequestProperties(tt.args.bidderName)
			assert.Equal(t, tt.want.requestProperties, properties, "mismatched properties")
			assert.Equal(t, tt.want.found, found, "mismatched found value")
		})
	}
}

func Test_biddersConfigMap_setBidderRequestProperties(t *testing.T) {
	type fields struct {
		biddersConfig map[string]*bidderConfig
	}
	type args struct {
		bidderName string
		properties map[string]bidderProperty
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *biddersConfigMap
	}{
		{
			name: "bidderName_not_found",
			args: args{
				bidderName: "test",
				properties: map[string]bidderProperty{
					"param-1": {
						location: []string{"path"},
					},
				},
			},
			want: &biddersConfigMap{
				biddersConfig: map[string]*bidderConfig{
					"test": {
						requestProperties: map[string]bidderProperty{
							"param-1": {
								location: []string{"path"},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bcm := &biddersConfigMap{
				biddersConfig: tt.fields.biddersConfig,
			}
			bcm.setBidderRequestProperties(tt.args.bidderName, tt.args.properties)
			assert.Equal(t, tt.want, bcm)
		})
	}
}
