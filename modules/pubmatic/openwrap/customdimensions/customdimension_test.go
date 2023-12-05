package customdimensions

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/prebid/prebid-server/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func TestGetCustomDimensions(t *testing.T) {
	type args struct {
		bidderParams json.RawMessage
	}
	tests := []struct {
		name  string
		args  args
		want  map[string]CustomDimension
		want1 bool
	}{
		{
			name: "bidderParams not present",
			args: args{
				bidderParams: json.RawMessage{},
			},
			want:  map[string]CustomDimension{},
			want1: false,
		},
		{
			name: "cds not present in bidderParams",
			args: args{
				bidderParams: json.RawMessage(`{}`),
			},
			want:  map[string]CustomDimension{},
			want1: false,
		},
		{
			name: "cds present",
			args: args{
				bidderParams: json.RawMessage(`{"pubmatic":{"cds":{"traffic":{"value":"email","sendtoGAM":true},"author":{"value":"henry","sendtoGAM":false},"age":{"value":"23"}}}}`),
			},
			want: map[string]CustomDimension{
				"traffic": {
					Value:     "email",
					SendToGAM: ptrutil.ToPtr(true),
				},
				"author": {
					Value:     "henry",
					SendToGAM: ptrutil.ToPtr(false),
				},
				"age": {
					Value: "23",
				},
			},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetCustomDimensions(tt.args.bidderParams)
			assert.Equal(t, tt.want, got, tt.name)
			assert.Equal(t, tt.want1, got1, tt.name)
		})
	}
}

func TestParseCustomDimensionsToString(t *testing.T) {
	type args struct {
		cdsMap map[string]CustomDimension
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want2 string
	}{
		{
			name: "valid custom dimensions map",
			args: args{
				cdsMap: map[string]CustomDimension{
					"k1": {Value: "v1"},
					"k2": {Value: "v2"},
				},
			},
			want: `k1=v1;k2=v2`,
		},
		{
			name: "empty custom dimensions map",
			args: args{
				cdsMap: map[string]CustomDimension{},
			},
			want: ``,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseCustomDimensionsToString(tt.args.cdsMap)
			expectedKeyVal := strings.Split(tt.want, ";")
			actualKeyVal := strings.Split(got, ";")
			assert.ElementsMatch(t, expectedKeyVal, actualKeyVal, tt.name)
		})
	}
}
