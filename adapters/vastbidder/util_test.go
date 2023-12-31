package vastbidder

import (
	"errors"
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

// TestGetDurationInSeconds ...
// hh:mm:ss.mmm => 3:40:43.5 => 3 hours, 40 minutes, 43 seconds and 5 milliseconds
// => 3*60*60 + 40*60 + 43 + 5*0.001 => 10800 + 2400 + 43 + 0.005 => 13243.005
func Test_parseDuration(t *testing.T) {
	type want struct {
		duration int // seconds  (will converted from string with format as  HH:MM:SS.mmm)
		err      error
	}
	tests := []struct {
		name string
		args string
		want want
	}{
		// duration validation tests
		{name: "duration 00:00:25 (= 25 seconds)", want: want{duration: 25}, args: "00:00:25"},
		{name: "duration 00:00:-25 (= -25 seconds)", want: want{err: errors.New("Invalid Duration")}, args: "00:00:-25"},
		{name: "duration 00:00:30.999 (= 30.990 seconds (int -> 30 seconds))", want: want{duration: 30}, args: "00:00:30.999"},
		{name: "duration 00:01:08 (1 min 8 seconds = 68 seconds)", want: want{duration: 68}, args: "00:01:08"},
		{name: "duration 02:13:12 (2 hrs 13 min  12 seconds) = 7992 seconds)", want: want{duration: 7992}, args: "02:13:12"},
		{name: "duration 3:40:43.5 (3 hrs 40 min  43 seconds 5 ms) = 6043.005 seconds (int -> 6043 seconds))", want: want{duration: 13243}, args: "3:40:43.5"},
		{name: "duration 00:00:25.0005458 (0 hrs 0 min  25 seconds 0005458 ms) - invalid max ms is 999", want: want{err: errors.New("Invalid Duration")}, args: "00:00:25.0005458"},
		{name: "invalid duration 3:13:900 (3 hrs 13 min  900 seconds) = Invalid seconds )", want: want{err: errors.New("Invalid Duration")}, args: "3:13:900"},
		{name: "invalid duration 3:13:34:44 (3 hrs 13 min 34 seconds :44=invalid) = ?? )", want: want{err: errors.New("Invalid Duration")}, args: "3:13:34:44"},
		{name: "duration = 0:0:45.038 , with milliseconds duration (0 hrs 0 min 45 seconds and 038 millseconds) = 45.038 seconds (int -> 45 seconds) )", want: want{duration: 45}, args: "0:0:45.038"},
		{name: "duration = 0:0:48.50  = 48.050 seconds (int -> 48 seconds))", want: want{duration: 48}, args: "0:0:48.50"},
		{name: "duration = 0:0:28.59  = 28.059 seconds  (int -> 28 seconds))", want: want{duration: 28}, args: "0:0:28.59 "},
		{name: "duration = 56 (ambiguity w.r.t. HH:MM:SS.mmm format)", want: want{err: errors.New("Invalid Duration")}, args: "56"},
		{name: "duration = :56 (ambiguity w.r.t. HH:MM:SS.mmm format)", want: want{err: errors.New("Invalid Duration")}, args: ":56"},
		{name: "duration = :56: (ambiguity w.r.t. HH:MM:SS.mmm format)", want: want{err: errors.New("Invalid Duration")}, args: ":56:"},
		{name: "duration = ::56 (ambiguity w.r.t. HH:MM:SS.mmm format)", want: want{err: errors.New("Invalid Duration")}, args: "::56"},
		{name: "duration = 56.445 (ambiguity w.r.t. HH:MM:SS.mmm format)", want: want{err: errors.New("Invalid Duration")}, args: "56.445"},
		{name: "duration = a:b:c.d (no numbers)", want: want{err: errors.New("Invalid Duration")}, args: "a:b:c.d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dur, err := parseDuration(tt.args)
			assert.Equal(t, tt.want.duration, dur)
			assert.Equal(t, tt.want.err, err)
			// if error expects 0 value for duration
			if nil != err {
				assert.Equal(t, 0, dur)
			}
		})
	}
}

func Test_parseVASTVersion(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		wantVersion float64
		wantErr     error
	}{
		{name: `empty_version`, version: "", wantVersion: 2.0, wantErr: nil},
		{name: `negative_version`, version: "-4.2", wantVersion: 0.0, wantErr: errInvalidVASTVersion},
		{name: `string_version`, version: "abc", wantVersion: 0.0, wantErr: errInvalidVASTVersion},
		{name: `half_floating_point`, version: "2.", wantVersion: 2.0, wantErr: nil},
		{name: `version_2.0`, version: "2.0", wantVersion: 2.0, wantErr: nil},
		{name: `version_3.0`, version: "3.0", wantVersion: 3.0, wantErr: nil},
		{name: `version_4.2`, version: "4.2", wantVersion: 4.2, wantErr: nil},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := parseVASTVersion(tt.version)
			assert.Equal(t, version, tt.wantVersion)
			assert.Equal(t, err, tt.wantErr)
		})
	}
}
