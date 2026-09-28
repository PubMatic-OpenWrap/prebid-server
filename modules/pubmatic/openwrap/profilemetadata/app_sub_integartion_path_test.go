package profilemetadata

import (
	"sync"
	"testing"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func Test_profileMetaData_GetAppSubIntegrationPath(t *testing.T) {
	type fields struct {
		RWMutex               sync.RWMutex
		appSubIntegrationPath map[string]models.AppSubIntegrationPath
	}
	type args struct {
		appSubIntegrationPathStr string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   models.AppSubIntegrationPath
		want1  bool
	}{
		{
			name: "appIntegrationPath map has key",
			fields: fields{
				RWMutex: sync.RWMutex{},
				appSubIntegrationPath: map[string]models.AppSubIntegrationPath{
					"DFP":    {ID: 1},
					"CUSTOM": {ID: 2, MediationName: "AppLovin", MediationType: "waterfall"},
				},
			},
			args: args{
				appSubIntegrationPathStr: "CUSTOM",
			},
			want:  models.AppSubIntegrationPath{ID: 2, MediationName: "AppLovin", MediationType: "waterfall"},
			want1: true,
		},
		{
			name: "appIntegrationPath map does not have key",
			fields: fields{
				RWMutex: sync.RWMutex{},
				appSubIntegrationPath: map[string]models.AppSubIntegrationPath{
					"DFP":    {ID: 1},
					"CUSTOM": {ID: 2, MediationName: "AppLovin", MediationType: "waterfall"},
				},
			},
			args: args{
				appSubIntegrationPathStr: "test",
			},
			want:  models.AppSubIntegrationPath{},
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

func TestAppSubIntegrationPathImpMediation(t *testing.T) {
	got := models.AppSubIntegrationPath{ID: 16, MediationName: "AdMob", MediationType: "bidding"}.ImpMediation()
	assert.Equal(t, &models.ImpMediation{Name: "AdMob", WaterfallOrBidding: "bidding"}, got)

	assert.Nil(t, models.AppSubIntegrationPath{ID: 1}.ImpMediation())
}
