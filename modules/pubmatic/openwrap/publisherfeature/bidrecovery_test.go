package publisherfeature

import (
	"testing"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestUpdateBidRecoveryEnabledPublishers(t *testing.T) {
	type fields struct {
		cache            cache.Cache
		publisherFeature map[int]map[int]models.FeatureData
		bidRecovery      bidRecovery
	}
	tests := []struct {
		name                             string
		fields                           fields
		wantBidrecoveryEnabledPublishers map[int]map[int]struct{}
	}{
		{
			name: "publisherFeature map is nil",
			fields: fields{
				cache:            nil,
				publisherFeature: nil,
			},
		},
		{
			name: "update bid recovery feature enabled pub",
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
						6: models.FeatureData{
							Enabled: 1,
							Value:   "[222]",
						},
					},
					5891: {
						3: models.FeatureData{
							Enabled: 1,
						},
						1: models.FeatureData{
							Enabled: 1,
						},
						6: models.FeatureData{
							Enabled: 1,
							Value:   "[123,456,789]",
						},
					},
				},
				bidRecovery: bidRecovery{
					enabledPublisherProfile: make(map[int]map[int]struct{}),
				},
			},
			wantBidrecoveryEnabledPublishers: map[int]map[int]struct{}{
				5890: {
					222: {},
				},
				5891: {
					123: {},
					456: {},
					789: {},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				cache:            tt.fields.cache,
				publisherFeature: tt.fields.publisherFeature,
				bidRecovery:      tt.fields.bidRecovery,
			}
			fe.updateBidRecoveryEnabledPublishers()
			assert.Equal(t, tt.wantBidrecoveryEnabledPublishers, fe.bidRecovery.enabledPublisherProfile)
		})
	}
}

func TestIsBidRecoveryEnabled(t *testing.T) {
	type fields struct {
		bidRecovery bidRecovery
	}
	type args struct {
		pubID     int
		profileID int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "bid recovery enabled for pub and profile",
			args: args{
				pubID:     5890,
				profileID: 3,
			},
			fields: fields{
				bidRecovery: bidRecovery{
					enabledPublisherProfile: map[int]map[int]struct{}{
						5890: {
							3: {},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "bid recovery not enabled for pub",
			args: args{
				pubID:     5891,
				profileID: 3,
			},
			fields: fields{
				bidRecovery: bidRecovery{
					enabledPublisherProfile: map[int]map[int]struct{}{
						5890: {
							3: {},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "bid recovery not enabled for profile",
			args: args{
				pubID:     5890,
				profileID: 1,
			},
			fields: fields{
				bidRecovery: bidRecovery{
					enabledPublisherProfile: map[int]map[int]struct{}{
						5890: {
							3: {},
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				bidRecovery: tt.fields.bidRecovery,
			}
			got := fe.IsBidRecoveryEnabled(tt.args.pubID, tt.args.profileID)
			assert.Equal(t, tt.want, got)
		})
	}
}
