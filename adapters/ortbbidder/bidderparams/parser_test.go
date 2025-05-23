package bidderparams

import (
	"fmt"
	"os"
	"testing"

	"github.com/prebid/prebid-server/v3/adapters/ortbbidder/util"
	"github.com/stretchr/testify/assert"
)

func TestPrepareRequestParams(t *testing.T) {
	type args struct {
		requestParamCfg map[string]any
		bidderName      string
	}
	type want struct {
		RequestParams map[string]BidderParamMapper
		err           error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "properties_missing_from_fileContents",
			args: args{
				requestParamCfg: map[string]any{
					"title": "test bidder parameters",
				},
				bidderName: "testbidder",
			},
			want: want{
				RequestParams: nil,
				err:           nil,
			},
		},
		{
			name: "properties_data_type_invalid",
			args: args{
				requestParamCfg: map[string]any{
					"title":      "test bidder parameters",
					"properties": "type invalid",
				},
				bidderName: "testbidder",
			},
			want: want{
				RequestParams: nil,
				err:           fmt.Errorf("error:[invalid_json_file_content_malformed_properties] bidderName:[testbidder]"),
			},
		},
		{
			name: "bidder-params_data_type_invalid",
			args: args{
				requestParamCfg: map[string]any{
					"title": "test bidder parameters",
					"properties": map[string]any{
						"adunitid": "invalid-type",
					},
				},
				bidderName: "testbidder",
			},
			want: want{
				RequestParams: nil,
				err:           fmt.Errorf("error:[invalid_json_file_content] bidder:[testbidder] bidderParam:[adunitid]"),
			},
		},
		{
			name: "bidder-params_properties_is_not_provided",
			args: args{
				requestParamCfg: map[string]any{
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
				RequestParams: map[string]BidderParamMapper{},
				err:           nil,
			},
		},
		{
			name: "bidder-params_location_is_not_in_string",
			args: args{
				requestParamCfg: map[string]any{
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
				RequestParams: nil,
				err:           fmt.Errorf("error:[incorrect_location_in_bidderparam] bidder:[testbidder] bidderParam:[adunitid]"),
			},
		},
		{
			name: "set_bidder-params_location_in_mapper",
			args: args{
				requestParamCfg: map[string]any{
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
				RequestParams: map[string]BidderParamMapper{
					"adunitid": {Location: "app.adunitid"},
				},
				err: nil,
			},
		},
		{
			name: "set_multiple_bidder-params_and_locations_in_mapper",
			args: args{
				requestParamCfg: map[string]any{
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
				RequestParams: map[string]BidderParamMapper{
					"adunitid": {Location: "app.adunitid"},
					"slotname": {Location: "ext.slot"},
				},
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RequestParams, err := prepareParams(tt.args.bidderName, tt.args.requestParamCfg)
			assert.Equalf(t, tt.want.err, err, "updateBidderParamsMapper returned unexpected error")
			assert.Equalf(t, tt.want.RequestParams, RequestParams, "updateBidderParamsMapper returned unexpected mapper")

		})
	}
}

func TestLoadBidderConfig(t *testing.T) {
	type want struct {
		biddersConfigMap *BidderConfig
		err              string
	}
	tests := []struct {
		name  string
		want  want
		setup func() (string, string, error)
	}{
		{
			name: "read_directory_fail",
			want: want{
				biddersConfigMap: nil,
				err:              "error handling request params: error:[open invalid-request-param-path: no such file or directory] dirPath:[invalid-request-param-path]",
			},
			setup: func() (string, string, error) {
				return "invalid-request-param-path", "invalid-response-param-path", nil
			},
		},
		{
			name: "found_file_without_.json_extension",
			want: want{
				biddersConfigMap: nil,
				err:              "error handling request params: error:[invalid_json_file_name] filename:[example.txt]",
			},
			setup: func() (string, string, error) {
				dirPath := t.TempDir()
				err := os.WriteFile(dirPath+"/example.txt", []byte("anything"), 0644)
				return dirPath, "", err
			},
		},
		{
			name: "response params - read_directory_fail",
			want: want{
				biddersConfigMap: nil,
				err:              "error handling response params: error:[open invalid-path: no such file or directory] dirPath:[invalid-path]",
			},
			setup: func() (string, string, error) {
				dirPath := t.TempDir()
				err := os.WriteFile(dirPath+"/example.json", []byte("anything"), 0644)
				return dirPath, "invalid-path", err
			},
		},
		{
			name: "response params - found_file_without_.json_extension",
			want: want{
				biddersConfigMap: nil,
				err:              "error handling response params: error:[invalid_json_file_name] filename:[example.txt]",
			},
			setup: func() (string, string, error) {
				requestDirPath := t.TempDir()
				err := os.WriteFile(requestDirPath+"/example.json", []byte("anything"), 0644)
				if err != nil {
					return "", "", err
				}
				responseDirPath := t.TempDir()
				err = os.WriteFile(responseDirPath+"/example.txt", []byte("anything"), 0644)
				return requestDirPath, responseDirPath, err
			},
		},
		{
			name: "oRTB_bidder_not_found",
			want: want{
				biddersConfigMap: &BidderConfig{BidderConfigMap: make(map[string]*Config)},
				err:              "",
			},
			setup: func() (string, string, error) {
				requestDirPath := t.TempDir()
				err := os.WriteFile(requestDirPath+"/example.json", []byte("anything"), 0644)
				if err != nil {
					return "", "", err
				}
				responseDirPath := t.TempDir()
				err = os.WriteFile(responseDirPath+"/example.json", []byte("anything"), 0644)
				return requestDirPath, responseDirPath, err
			},
		},
		{
			name: "oRTB_bidder_found_but_invalid_json_present",
			want: want{
				biddersConfigMap: nil,
				err:              "error handling request params: error:[fail_to_read_file]",
			},
			setup: func() (string, string, error) {
				requestDirPath := t.TempDir()
				err := os.WriteFile(requestDirPath+"/owortb_test.json", []byte("anything"), 0644)
				if err != nil {
					return "", "", err
				}
				responseDirPath := t.TempDir()
				err = os.WriteFile(responseDirPath+"/owortb_test.json", []byte("anything"), 0644)
				return requestDirPath, responseDirPath, err
			},
		},
		{
			name: "oRTB_bidder_found_but_bidder-params_are_absent",
			want: want{
				biddersConfigMap: &BidderConfig{BidderConfigMap: map[string]*Config{
					"owortb_test": {
						RequestParams: nil,
					},
				}},
				err: "",
			},
			setup: func() (string, string, error) {
				requestDirPath := t.TempDir()
				err := os.WriteFile(requestDirPath+"/owortb_test.json", []byte("{}"), 0644)
				if err != nil {
					return "", "", err
				}
				responseDirPath := t.TempDir()
				err = os.WriteFile(responseDirPath+"/owortb_test.json", []byte("{}"), 0644)
				return requestDirPath, responseDirPath, err
			},
		},
		{
			name: "oRTB_bidder_found_but_prepareBidderRequestProperties_returns_error",
			want: want{
				biddersConfigMap: nil,
				err:              "error:[invalid_json_file_content_malformed_properties] bidderName:[owortb_test]",
			},
			setup: func() (string, string, error) {
				requestDirPath := t.TempDir()
				err := os.WriteFile(requestDirPath+"/owortb_test.json", []byte(`{"properties":"invalid-properties"}`), 0644)
				if err != nil {
					return "", "", err
				}
				responseDirPath := t.TempDir()
				err = os.WriteFile(responseDirPath+"/owortb_test.json", []byte(`{"properties":"invalid-properties"}`), 0644)
				return requestDirPath, responseDirPath, err
			},
		},
		{
			name: "oRTB_bidder_found_and_valid_json_contents_present",
			want: want{
				biddersConfigMap: &BidderConfig{
					BidderConfigMap: map[string]*Config{
						"owortb_test": {
							RequestParams: map[string]BidderParamMapper{
								"adunitid": {Location: "app.adunit.id"},
								"slotname": {Location: "ext.slotname"},
							},
							ResponseParams: map[string]BidderParamMapper{
								"mtype": {Location: "seatbid.#.bid.#.ext.mtype"},
							},
						},
					}},
				err: "",
			},
			setup: func() (string, string, error) {
				requestDirPath := t.TempDir()
				err := os.WriteFile(requestDirPath+"/owortb_test.json", []byte(`
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
					return "", "", err
				}
				responseDirPath := t.TempDir()
				err = os.WriteFile(responseDirPath+"/owortb_test.json", []byte(`{
					"title":"ortb bidder",
					"properties": {
						"mtype": {
							"type": "string",
							"location": "seatbid.#.bid.#.ext.mtype"
						}
					}
				}`), 0644)
				return requestDirPath, responseDirPath, err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RequestParamsDirPath, ResponseParamsDirPath, err := tt.setup()
			assert.NoError(t, err, "setup returned unexpected error")
			got, err := LoadBidderConfig(RequestParamsDirPath, ResponseParamsDirPath, util.IsORTBBidder)
			assert.Equal(t, tt.want.biddersConfigMap, got, "found incorrect mapper")
			assert.Equal(t, len(tt.want.err) == 0, err == nil, "mismatched error")
			if err != nil {
				assert.ErrorContains(t, err, tt.want.err, "found incorrect error message")
			}
		})
	}
}

func TestReadFile(t *testing.T) {
	var setup = func() (string, error) {
		dir := t.TempDir()
		err := os.WriteFile(dir+"/owortb.json", []byte(`
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
		return dir, err
	}
	type args struct {
		file string
	}
	type want struct {
		err  bool
		node map[string]any
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "successful_readfile",
			args: args{
				file: "owortb.json",
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
		},
		{
			name: "fail_readfile",
			args: args{
				file: "invalid.json",
			},
			want: want{
				err:  true,
				node: nil,
			},
		},
	}
	path, err := setup()
	assert.Nil(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readFile(path, tt.args.file)
			assert.Equal(t, tt.want.err, err != nil, "mismatched error")
			assert.Equal(t, tt.want.node, got, "mismatched map[string]any")
		})
	}
}
