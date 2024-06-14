package util

import (
	"testing"

	"github.com/prebid/prebid-server/v2/util/jsonutil"
	"github.com/stretchr/testify/assert"
)

func TestSetValue(t *testing.T) {
	type args struct {
		requestNode map[string]any
		location    []string
		value       any
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
				requestNode: map[string]any{},
				location:    []string{"key"},
				value:       nil,
			},
			want: want{
				status: false,
				node:   map[string]any{},
			},
		},
		{
			name: "set_value_in_empty_location",
			args: args{
				requestNode: map[string]any{},
				location:    []string{},
				value:       123,
			},
			want: want{
				status: false,
				node:   map[string]any{},
			},
		},
		{
			name: "set_value_in_invalid_location_modifies_node",
			args: args{
				requestNode: map[string]any{},
				location:    []string{"key", ""},
				value:       123,
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
				requestNode: map[string]any{},
				location:    []string{"key"},
				value:       123,
			},
			want: want{
				status: true,
				node:   map[string]any{"key": 123},
			},
		},
		{
			name: "set_value_at_root_level_in_non-empty_node",
			args: args{
				requestNode: map[string]any{"oldKey": "oldValue"},
				location:    []string{"key"},
				value:       123,
			},
			want: want{
				status: true,
				node:   map[string]any{"oldKey": "oldValue", "key": 123},
			},
		},
		{
			name: "set_value_at_non-root_level_in_non-json_node",
			args: args{
				requestNode: map[string]any{"rootKey": "rootValue"},
				location:    []string{"rootKey", "key"},
				value:       123,
			},
			want: want{
				status: false,
				node:   map[string]any{"rootKey": "rootValue"},
			},
		},
		{
			name: "set_value_at_non-root_level_in_json_node",
			args: args{
				requestNode: map[string]any{"rootKey": map[string]any{
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
				requestNode: map[string]any{"rootKey": map[string]any{
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
				requestNode: map[string]any{"rootKey": map[string]any{
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
		{
			name: "appsite_key_app_object_present",
			args: args{
				requestNode: map[string]any{"app": map[string]any{
					"parentKey": "oldValue",
				}},
				location: []string{"appsite", "parentKey"},
				value:    "newKeyValue",
			},
			want: want{
				status: true,
				node: map[string]any{"app": map[string]any{
					"parentKey": "newKeyValue",
				}},
			},
		},
		{
			name: "appsite_key_site_object_present",
			args: args{
				requestNode: map[string]any{"site": map[string]any{
					"parentKey": "oldValue",
				}},
				location: []string{"appsite", "parentKey"},
				value:    "newKeyValue",
			},
			want: want{
				status: true,
				node: map[string]any{"site": map[string]any{
					"parentKey": "newKeyValue",
				}},
			},
		},
		{
			name: "request_has_list_of_interface",
			args: args{
				requestNode: map[string]any{
					"id": "req_1",
					"imp": []any{
						map[string]any{
							"id": "imp_1",
						},
					},
				},
				location: []string{"imp", "0", "ext"},
				value:    "value",
			},
			want: want{
				status: true,
				node: map[string]any{
					"id": "req_1",
					"imp": []any{
						map[string]any{
							"id":  "imp_1",
							"ext": "value",
						},
					},
				},
			},
		},
		{
			name: "request_has_list_of_interface_with_multi_items",
			args: args{
				requestNode: map[string]any{
					"id": "req_1",
					"imp": []any{
						map[string]any{
							"id": "imp_1",
						},
						map[string]any{
							"id": "imp_2",
						},
					},
				},
				location: []string{"imp", "1", "ext"},
				value:    "value",
			},
			want: want{
				status: true,
				node: map[string]any{
					"id": "req_1",
					"imp": []any{
						map[string]any{
							"id": "imp_1",
						},
						map[string]any{
							"id":  "imp_2",
							"ext": "value",
						},
					},
				},
			},
		},
		{
			name: "request_has_list_of_interface_with_multi_items_but_invalid_index_to_update",
			args: args{
				requestNode: map[string]any{
					"id": "req_1",
					"imp": []any{
						map[string]any{
							"id": "imp_1",
						},
					},
				},
				location: []string{"imp", "3", "ext"},
				value:    "value",
			},
			want: want{
				status: false,
				node: map[string]any{
					"id": "req_1",
					"imp": []any{
						map[string]any{
							"id": "imp_1",
						},
					},
				},
			},
		},
		{
			name: "request_has_list_of_interface_with_multi_items_but_valid_index_to_update",
			args: args{
				requestNode: map[string]any{
					"id": "req_1",
					"imp": []any{
						map[string]any{
							"id": "imp_1",
						},
					},
				},
				location: []string{"imp", "0"},
				value: map[string]any{
					"id": "updated_id",
				},
			},
			want: want{
				status: true,
				node: map[string]any{
					"id": "req_1",
					"imp": []any{
						map[string]any{
							"id": "updated_id",
						},
					},
				},
			},
		},
		{
			name: "request_has_list_of_interface_where_new_node_need_to_be_created",
			args: args{
				requestNode: map[string]any{
					"id": "req_1",
					"imp": []any{
						nil, nil,
					},
				},
				location: []string{"imp", "0", "ext"},
				value: map[string]any{
					"id": "updated_id",
				},
			},
			want: want{
				status: true,
				node: map[string]any{
					"id": "req_1",
					"imp": []any{
						map[string]any{
							"ext": map[string]any{
								"id": "updated_id",
							},
						},
						nil,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SetValue(tt.args.requestNode, tt.args.location, tt.args.value)
			assert.Equalf(t, tt.want.node, tt.args.requestNode, "SetValue failed to update node object")
			assert.Equalf(t, tt.want.status, got, "SetValue returned invalid status")
		})
	}
}

func TestGetNode(t *testing.T) {
	type args struct {
		requestNode map[string]any
		key         string
	}
	tests := []struct {
		name string
		args args
		want any
	}{
		{
			name: "appsite_key_present_when_app_object_present",
			args: args{
				requestNode: map[string]any{"app": map[string]any{
					"parentKey": "oldValue",
				}},
				key: "appsite",
			},
			want: map[string]any{"parentKey": "oldValue"},
		},
		{
			name: "appsite_key_present_when_site_object_present",
			args: args{
				requestNode: map[string]any{"site": map[string]any{
					"siteKey": "siteValue",
				}},
				key: "appsite",
			},
			want: map[string]any{"siteKey": "siteValue"},
		},
		{
			name: "appsite_key_absent",
			args: args{
				requestNode: map[string]any{"device": map[string]any{
					"deviceKey": "deviceVal",
				}},
				key: "appsite",
			},
			want: nil,
		},
		{
			name: "imp_key_present",
			args: args{
				requestNode: map[string]any{
					"device": map[string]any{
						"deviceKey": "deviceVal",
					},
					"imp": []any{
						map[string]any{
							"id": "imp_1",
						},
					},
				},
				key: "imp",
			},
			want: []any{
				map[string]any{
					"id": "imp_1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := getNode(tt.args.requestNode, tt.args.key)
			assert.Equal(t, tt.want, node)
		})
	}
}

func TestReplaceLocationMacro(t *testing.T) {
	tests := []struct {
		name              string
		path              string
		array             []int
		expectedValuePath string
	}{
		{
			name:              "Empty path",
			path:              "",
			array:             []int{0, 1},
			expectedValuePath: "",
		},
		{
			name:              "Replace # in path with array",
			path:              "seatbid.#.bid.#.ext.mtype",
			array:             []int{0, 1},
			expectedValuePath: "seatbid.0.bid.1.ext.mtype",
		},
		{
			name:              "Array length less than # count in path",
			path:              "seatbid.#.bid.#.ext.mtype",
			array:             []int{0},
			expectedValuePath: "seatbid.0.bid.#.ext.mtype",
		},
		{
			name:              "No # in path",
			path:              "seatbid.bid.ext.mtype",
			array:             []int{0, 1},
			expectedValuePath: "seatbid.bid.ext.mtype",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceLocationMacro(tt.path, tt.array)
			assert.Equal(t, tt.expectedValuePath, result)
		})
	}
}

func TestGetValueFromLocation(t *testing.T) {

	jsonToMap := func(jsonStr string) (result map[string]any) {
		jsonutil.Unmarshal([]byte(jsonStr), &result)
		return
	}

	node := jsonToMap(`{"seatbid":[{"bid":[{"ext":{"mtype":"video"}}]}]}`)

	tests := []struct {
		name          string
		node          interface{}
		path          string
		expectedValue interface{}
		ok            bool
	}{
		{
			name:          "Node is empty",
			node:          nil,
			path:          "seatbid.0.bid.0.ext.mtype",
			expectedValue: nil,
			ok:            false,
		},
		{
			name:          "Path is empty",
			node:          node,
			path:          "",
			expectedValue: nil,
			ok:            false,
		},
		{
			name:          "Value is present in node",
			node:          node,
			path:          "seatbid.0.bid.0.ext.mtype",
			expectedValue: "video",
			ok:            true,
		},
		{
			name:          "Value is not present in node",
			node:          node,
			path:          "seatbid.0.bid.0.ext.mtype1",
			expectedValue: nil,
			ok:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := GetValueFromLocation(tt.node, tt.path)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.expectedValue, result)
		})
	}
}

func TestIsORTBBidder(t *testing.T) {
	tests := []struct {
		name     string
		bidder   string
		expected bool
	}{
		{
			name:     "ORTB bidder",
			bidder:   "owortb_test",
			expected: true,
		},
		{
			name:     "Non-ORTB bidder",
			bidder:   "test",
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsORTBBidder(tt.bidder)
			assert.Equal(t, tt.expected, result)
		})
	}
}
