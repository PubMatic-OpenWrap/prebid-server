package publisherfeature

import (
	"testing"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestFeature_updateMaxFloorsEnabledPublishers(t *testing.T) {

	type fields struct {
		cache            cache.Cache
		publisherFeature map[int]map[int]models.FeatureData
		maxFloors        maxFloors
	}
	tests := []struct {
		name                          string
		fields                        fields
		wantMaxFloorsEnabledPublisher map[int]struct{}
	}{
		{
			name: "publisherFeature map is nil",
			fields: fields{
				cache:            nil,
				publisherFeature: nil,
				maxFloors: maxFloors{
					enabledPublishers: make(map[int]struct{}),
				},
			},
			wantMaxFloorsEnabledPublisher: map[int]struct{}{},
		},
		{
			name: "update max floors feature enabled pub",
			fields: fields{
				cache: nil,
				publisherFeature: map[int]map[int]models.FeatureData{
					5890: {
						5: models.FeatureData{
							Enabled: 1,
						},
					},
					5891: {
						5: models.FeatureData{
							Enabled: 1,
						},
					},
				},
				maxFloors: maxFloors{
					enabledPublishers: make(map[int]struct{}),
				},
			},
			wantMaxFloorsEnabledPublisher: map[int]struct{}{
				5890: {},
				5891: {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := feature{
				publisherFeature: tt.fields.publisherFeature,
				maxFloors:        tt.fields.maxFloors,
			}
			fe.updateMaxFloorsEnabledPublishers()
			assert.Equal(t, tt.wantMaxFloorsEnabledPublisher, fe.maxFloors.enabledPublishers)
		})
	}
}

func TestFeature_IsMaxFloorsEnabled(t *testing.T) {
	type fields struct {
		maxFloors maxFloors
	}
	type args struct {
		pubid int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "max floors feature enabled pub",
			args: args{
				pubid: 5890,
			},
			fields: fields{
				maxFloors: maxFloors{
					enabledPublishers: map[int]struct{}{
						5890: {},
					},
				},
			},
			want: true,
		},
		{
			name: "max floors feature disabled pub",
			args: args{
				pubid: 5891,
			},
			fields: fields{
				maxFloors: maxFloors{
					enabledPublishers: map[int]struct{}{
						5890: {},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				maxFloors: tt.fields.maxFloors,
			}
			got := fe.IsMaxFloorsEnabled(tt.args.pubid)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
