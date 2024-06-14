package profilemetadata

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_profileMetaData_GetProfileTypePlatform(t *testing.T) {
	type fields struct {
		RWMutex             sync.RWMutex
		profileTypePlatform map[string]int
	}
	type args struct {
		profileTypePlatformStr string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
		want1  bool
	}{
		{
			name: "profileTypePlatform map has key",
			fields: fields{
				RWMutex: sync.RWMutex{},
				profileTypePlatform: map[string]int{
					"openwrap": 1,
					"identity": 2,
				},
			},
			args: args{
				profileTypePlatformStr: "openwrap",
			},
			want:  1,
			want1: true,
		},
		{
			name: "profileTypePlatform map does not have key",
			fields: fields{
				RWMutex: sync.RWMutex{},
				profileTypePlatform: map[string]int{
					"openwrap": 1,
					"identity": 2,
				},
			},
			args: args{
				profileTypePlatformStr: "test",
			},
			want:  0,
			want1: false,
		},
	}
	for ind := range tests {
		tt := &tests[ind]
		t.Run(tt.name, func(t *testing.T) {
			pmd := &profileMetaData{
				profileTypePlatform: tt.fields.profileTypePlatform,
			}
			got, got1 := pmd.GetProfileTypePlatform(tt.args.profileTypePlatformStr)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
		})
	}
}
