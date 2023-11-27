package vastbidder

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func Test_getJSONString(t *testing.T) {
	type args struct {
		kvmap any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty_map",
			args: args{kvmap: map[string]any{}},
			want: "{}",
		},
		{
			name: "map_without_nesting",
			args: args{kvmap: map[string]any{
				"k1": "v1",
				"k2": "v2",
			}},
			want: "{\"k1\":\"v1\",\"k2\":\"v2\"}",
		},
		{
			name: "map_with_nesting",
			args: args{kvmap: map[string]any{
				"k1": "v1",
				"metadata": map[string]any{
					"k2": "v2",
					"k3": "v3",
				},
			}},
			want: "{\"k1\":\"v1\",\"metadata\":{\"k2\":\"v2\",\"k3\":\"v3\"}}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getJSONString(tt.args.kvmap)
			assert.Equal(t, got, tt.want, tt.name)
		})
	}
}

func Test_getValueFromMap(t *testing.T) {
	type args struct {
		lookUpOrder []string
		m           map[string]any
	}
	tests := []struct {
		name string
		args args
		want any
	}{
		{
			name: "map_without_nesting",
			args: args{lookUpOrder: []string{"k1"},
				m: map[string]any{
					"k1": "v1",
					"k2": "v2",
				},
			},
			want: "v1",
		},
		{
			name: "map_with_nesting",
			args: args{lookUpOrder: []string{"country", "state"},
				m: map[string]any{
					"name": "test",
					"country": map[string]any{
						"state": "MH",
						"pin":   12345,
					},
				},
			},
			want: "MH",
		},
		{
			name: "key_not_exists",
			args: args{lookUpOrder: []string{"country", "name"},
				m: map[string]any{
					"name": "test",
					"country": map[string]any{
						"state": "MH",
						"pin":   12345,
					},
				},
			},
			want: "",
		},
		{
			name: "empty_map",
			args: args{lookUpOrder: []string{"country", "name"},
				m: map[string]any{},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getValueFromMap(tt.args.lookUpOrder, tt.args.m)
			assert.Equal(t, got, tt.want)
		})
	}
}

func Test_mapToQuery(t *testing.T) {
	type args struct {
		m map[string]any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "map_without_nesting",
			args: args{
				m: map[string]any{
					"k1": "v1",
					"k2": "v2",
				},
			},
			want: "k1=v1&k2=v2",
		},
		{
			name: "map_with_nesting",
			args: args{
				m: map[string]any{
					"name": "test",
					"country": map[string]any{
						"state": "MH",
						"pin":   12345,
					},
				},
			},
			want: "country=pin%3D12345%26state%3DMH&name=test",
		},
		{
			name: "empty_map",
			args: args{
				m: map[string]any{},
			},
			want: "",
		},
		{
			name: "nil_map",
			args: args{
				m: nil,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapToQuery(tt.args.m); got != tt.want {
				t.Errorf("mapToQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isMap(t *testing.T) {
	type args struct {
		data any
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "map_data_type",
			args: args{data: map[string]any{}},
			want: true,
		},
		{
			name: "string_data_type",
			args: args{data: "data type is string"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isMap(tt.args.data)
			assert.Equal(t, got, tt.want)
		})
	}
}
