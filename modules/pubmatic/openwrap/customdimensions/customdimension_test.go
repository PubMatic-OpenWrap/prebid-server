package customdimensions

import (
	"encoding/json"
	"testing"

	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/util/ptrutil"
	"github.com/stretchr/testify/assert"
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
			got := GetCustomDimensionsFromRequestExt(tt.args.bidderParams)
			assert.Equal(t, tt.want, got, tt.name)
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
			assert.Equal(t, tt.want, got, tt.name)
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
			name: "valid cds present",
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
			name: "cds present with repeated key",
			args: args{
				ext: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						BidderParams: json.RawMessage(`{"pubmatic":{"cds":{"k10":{"sendtoGAM":false,"value":"v1"},"k10":{"sendtoGAM":true,"value":"v101"},"k3":{"value":"v3"},"k2":{"sendtoGAM":true,"value":"v2"}}}}`),
					},
				},
			},
			want: map[string]CustomDimension{
				"k10": {
					Value:     "v101",
					SendToGAM: ptrutil.ToPtr(true),
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
			name: "cds not present",
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
		{
			name: "ext data is nil",
			args: args{
				ext: nil,
			},
			want:  map[string]CustomDimension{},
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := IsCustomDimensionsPresent(tt.args.ext)
			assert.Equal(t, tt.want, got, tt.name)
			assert.Equal(t, tt.want1, got1, tt.name)
		})
	}
}
