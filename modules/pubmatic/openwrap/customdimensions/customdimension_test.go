package customdimensions

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/util/ptrutil"
)

func TestGetCustomDimensionsFromRequestExt(t *testing.T) {
	type args struct {
		bidderParams json.RawMessage
	}
	tests := []struct {
		name string
		args args
		want map[string]CustomDimension
	}{
		{
			name: "bidderParams not present",
			args: args{
				bidderParams: json.RawMessage{},
			},
			want: nil,
		},
		{
			name: "cds not present in bidderParams",
			args: args{
				bidderParams: json.RawMessage(`{}`),
			},
			want: nil,
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCustomDimensionsFromRequestExt(tt.args.bidderParams); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCustomDimensionsFromRequestExt() = %v, want %v", got, tt.want)
			}
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
			name: "when valid custom dimensions map, return a string ",
			args: args{
				cdsMap: map[string]CustomDimension{
					"k1": {Value: "v1"},
					"k2": {Value: "v2"},
				},
			},
			want:  `k1=v1;k2=v2`,
			want2: `k2=v2;k1=v1`,
		},
		{
			name: "when valid custom dimensions map is empty, return",
			args: args{
				cdsMap: map[string]CustomDimension{},
			},
			want:  ``,
			want2: ``,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseCustomDimensionsToString(tt.args.cdsMap); got != tt.want && got != tt.want2 {
				t.Errorf("ParseCustomDimensionsToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCustomDimensionsPresent(t *testing.T) {
	type args struct {
		ext interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  map[string]CustomDimension
		want1 bool
	}{
		{
			name: "if cds present return true and cds Map",
			args: args{
				ext: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						BidderParams: json.RawMessage(`{"pubmatic":{"cds":{"k1":{"sendtoGAM":false,"value":"v1"},"k3":{"value":"v3"},"k2":{"sendtoGAM":true,"value":"v2"}}}}`),
					},
				},
			},
			want: map[string]CustomDimension{
				"k1": {
					Value:     "v1",
					SendToGAM: ptrutil.ToPtr(false),
				},
				"k2": {
					Value:     "v2",
					SendToGAM: ptrutil.ToPtr(true),
				},
				"k3": {
					Value: "v3",
				},
			},
			want1: true,
		},
		{
			name: "if cds absent return false and empty cds Map",
			args: args{
				ext: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{},
				},
			},
			want:  map[string]CustomDimension{},
			want1: false,
		},
		{
			name: "ext data invalid",
			args: args{
				ext: "invalid_ext",
			},
			want:  map[string]CustomDimension{},
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := IsCustomDimensionsPresent(tt.args.ext)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IsCustomDimensionsPresent() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("IsCustomDimensionsPresent() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
