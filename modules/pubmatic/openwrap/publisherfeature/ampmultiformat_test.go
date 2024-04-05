package publisherfeature

import (
	"testing"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestFeature_updateAmpMutiformatEnabledPublishers(t *testing.T) {

	type fields struct {
		cache            cache.Cache
		publisherFeature map[int]map[int]models.FeatureData
		ampMultiformat   ampMultiformat
	}
	tests := []struct {
		name                               string
		fields                             fields
		wantAmpMultiformatEnabledPublisher map[int]struct{}
	}{
		{
			name: "publisherFeature map is nil",
			fields: fields{
				cache:            nil,
				publisherFeature: nil,
				ampMultiformat: ampMultiformat{
					enabledPublishers: make(map[int]struct{}),
				},
			},
			wantAmpMultiformatEnabledPublisher: map[int]struct{}{},
		},
		{
			name: "update amp feature enabled pub",
			fields: fields{
				cache: nil,
				publisherFeature: map[int]map[int]models.FeatureData{
					5890: {
						3: models.FeatureData{
							Enabled: 1,
						},
						1: models.FeatureData{
							Enabled: 1,
						},
					},
					5891: {
						3: models.FeatureData{
							Enabled: 1,
						},
						1: models.FeatureData{
							Enabled: 1,
						},
					},
				},
				ampMultiformat: ampMultiformat{
					enabledPublishers: make(map[int]struct{}),
				},
			},
			wantAmpMultiformatEnabledPublisher: map[int]struct{}{
				5890: {},
				5891: {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := feature{
				publisherFeature: tt.fields.publisherFeature,
				ampMultiformat:   tt.fields.ampMultiformat,
			}
			fe.updateAmpMutiformatEnabledPublishers()
			assert.Equal(t, tt.wantAmpMultiformatEnabledPublisher, fe.ampMultiformat.enabledPublishers)
		})
	}
}

func TestFeature_IsAmpMultiformatEnabled(t *testing.T) {
	type fields struct {
		ampMultiformat ampMultiformat
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
			name: "amp feature enabled pub",
			args: args{
				pubid: 5890,
			},
			fields: fields{
				ampMultiformat: ampMultiformat{
					enabledPublishers: map[int]struct{}{
						5890: {},
					},
				},
			},
			want: true,
		},
		{
			name: "amp feature disabled pub",
			args: args{
				pubid: 5891,
			},
			fields: fields{
				ampMultiformat: ampMultiformat{
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
				ampMultiformat: tt.fields.ampMultiformat,
			}
			got := fe.IsAmpMultiformatEnabled(tt.args.pubid)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
