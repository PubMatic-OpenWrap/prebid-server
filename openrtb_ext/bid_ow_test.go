package openrtb_ext

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetCreativeTypeFromCreative(t *testing.T) {
	type args struct {
		adm string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "video_creative",
			args: args{
				adm: "<VAST version=\"3.0\"></VAST>",
			},
			want: Video,
		},
		{
			name: "native_creative",
			args: args{
				adm: "{\"native\":{\"link\":{\"url\":\"http://example.com\"},\"assets\":[]}}",
			},
			want: Native,
		},
		{
			name: "banner_creative",
			args: args{
				adm: "<div>Banner Ad</div>",
			},
			want: Banner,
		},
		{
			name: "empty_AdM",
			args: args{
				adm: "",
			},
			want: "",
		},
		{
			name: "invalid_json_in_adm",
			args: args{
				adm: "{\"native\":{\"link\":{\"url\":\"http://example.com\"},\"assets\":[]",
			},
			want: Banner,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCreativeTypeFromCreative(tt.args.adm)
			assert.Equal(t, tt.want, got)
		})
	}
}
