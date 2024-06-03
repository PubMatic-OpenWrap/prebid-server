package ortbbidder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetValue(t *testing.T) {
	type args struct {
		requestNode map[string]any
		impNode     map[string]any
		location    []string
		value       any
	}
	type want struct {
		node   map[string]any
		status bool
		imp    map[string]any
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
			name: "imp_key_present",
			args: args{
				requestNode: map[string]any{"id": map[string]any{
					"id": "req_1",
					"imp": map[string]any{
						"id": "imp_1",
					},
				}},
				impNode: map[string]any{
					"id": "imp_1",
				},
				location: []string{"imp", "ext"},
				value:    "value",
			},
			want: want{
				status: true,
				imp: map[string]any{
					"ext": "value",
					"id":  "imp_1",
				},
				node: map[string]any{"id": map[string]any{
					"id": "req_1",
					"imp": map[string]any{
						"id": "imp_1",
					},
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := setValue(tt.args.requestNode, tt.args.impNode, tt.args.location, tt.args.value)
			assert.Equalf(t, tt.want.node, tt.args.requestNode, "SetValue failed to update node object")
			assert.Equalf(t, tt.want.imp, tt.args.impNode, "SetValue failed to update imp object")
			assert.Equalf(t, tt.want.status, got, "SetValue returned invalid status")
		})
	}
}

func TestGetNode(t *testing.T) {
	type args struct {
		requestNode map[string]any
		impNode     map[string]any
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
					"imp": []string{},
				},
				impNode: map[string]any{
					"id": "imp_1",
				},
				key: "imp",
			},
			want: map[string]any{
				"id": "imp_1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := getNode(tt.args.requestNode, tt.args.impNode, tt.args.key)
			assert.Equal(t, tt.want, node)
		})
	}
}
