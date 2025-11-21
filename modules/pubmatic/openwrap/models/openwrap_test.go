package models

import (
	"testing"
)

func TestRequestCtxGetVersionLevelKey(t *testing.T) {
	type fields struct {
		PartnerConfigMap map[int]map[string]string
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "get_version_level_platform_key",
			fields: fields{
				PartnerConfigMap: map[int]map[string]string{
					-1: {
						"platform": "in-app",
					},
				},
			},
			args: args{
				key: "platform",
			},
			want: "in-app",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RequestCtx{
				PartnerConfigMap: tt.fields.PartnerConfigMap,
			}
			if got := r.GetVersionLevelKey(tt.args.key); got != tt.want {
				t.Errorf("RequestCtx.GetVersionLevelKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
