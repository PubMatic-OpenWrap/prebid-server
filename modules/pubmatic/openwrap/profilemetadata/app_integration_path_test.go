package profilemetadata

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_profileMetaData_GetAppIntegrationPath(t *testing.T) {
	type fields struct {
		RWMutex            sync.RWMutex
		appIntegrationPath map[string]int
	}
	type args struct {
		appIntegrationPathStr string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
		want1  bool
	}{
		{
			name: "appIntegrationPath map has key",
			fields: fields{
				RWMutex: sync.RWMutex{},
				appIntegrationPath: map[string]int{
					"iOS":     1,
					"Android": 2,
				},
			},
			args: args{
				appIntegrationPathStr: "Android",
			},
			want:  2,
			want1: true,
		},
		{
			name: "appIntegrationPath map does not have key",
			fields: fields{
				RWMutex: sync.RWMutex{},
				appIntegrationPath: map[string]int{
					"iOS":     1,
					"Android": 2,
				},
			},
			args: args{
				appIntegrationPathStr: "test",
			},
			want:  0,
			want1: false,
		},
	}
	for ind := range tests {
		tt := &tests[ind]
		t.Run(tt.name, func(t *testing.T) {
			pmd := &profileMetaData{
				appIntegrationPath: tt.fields.appIntegrationPath,
			}
			got, got1 := pmd.GetAppIntegrationPath(tt.args.appIntegrationPathStr)
			assert.Equal(t, got, tt.want)
			assert.Equal(t, got1, tt.want1)
		})
	}
}
