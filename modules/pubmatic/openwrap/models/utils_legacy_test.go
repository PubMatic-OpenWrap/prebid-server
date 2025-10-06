package models

import (
	"testing"

	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestGetRequestExt(t *testing.T) {
	type args struct {
		ext []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *RequestExt
		wantErr bool
	}{
		{
			name: "empty_request_ext",
			args: args{
				ext: []byte(`{}`),
			},
			want:    &RequestExt{},
			wantErr: false,
		},
		{
			name: "successfully_Unmarshaled_request_ext",
			args: args{
				ext: []byte(`{"prebid":{},"wrapper":{"ssauction":0,"profileid":2087,"sumry_disable":0,"clientconfig":1,"versionid":1}}`),
			},
			want: &RequestExt{
				ExtRequest: openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{},
				},
				Wrapper: &RequestExtWrapper{
					SSAuctionFlag:    0,
					ProfileId:        2087,
					SumryDisableFlag: 0,
					ClientConfigFlag: 1,
					VersionId:        1,
				},
			},
			wantErr: false,
		},
		{
			name: "failed_to_Unmarshaled_request_ext",
			args: args{
				ext: []byte(`{"prebid":{},"wrapper":{"ssauction":0,"profileid":"2087","sumry_disable":0,"clientconfig":1,"versionid":1}}`),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid_JSON_request_ext",
			args: args{
				ext: []byte(`Invalid json`),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unexpected_end_of_JSON_input",
			args: args{
				ext: []byte(`{`),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRequestExt(tt.args.ext)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRequestExt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
