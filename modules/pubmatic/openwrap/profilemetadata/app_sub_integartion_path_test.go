package profilemetadata

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_profileMetaData_GetAppSubIntegrationPath(t *testing.T) {
	type fields struct {
		RWMutex               sync.RWMutex
		appSubIntegrationPath map[string]int
	}
	type args struct {
		appSubIntegrationPathStr string
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
				appSubIntegrationPath: map[string]int{
					"DFP":    1,
					"CUSTOM": 2,
				},
			},
			args: args{
				appSubIntegrationPathStr: "CUSTOM",
			},
			want:  2,
			want1: true,
		},
		{
			name: "appIntegrationPath map does not have key",
			fields: fields{
				RWMutex: sync.RWMutex{},
				appSubIntegrationPath: map[string]int{
					"DFP":    1,
					"CUSTOM": 2,
				},
			},
			args: args{
				appSubIntegrationPathStr: "test",
			},
			want:  0,
			want1: false,
		},
	}
	for ind := range tests {
		tt := &tests[ind]
		t.Run(tt.name, func(t *testing.T) {
			pmd := &profileMetaData{
				appSubIntegrationPath: tt.fields.appSubIntegrationPath,
			}
			got, got1 := pmd.GetAppSubIntegrationPath(tt.args.appSubIntegrationPathStr)
			assert.Equal(t, got, tt.want)
			assert.Equal(t, got1, tt.want1)
		})
	}
}
