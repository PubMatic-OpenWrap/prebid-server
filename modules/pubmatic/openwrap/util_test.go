package openwrap

import (
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	mock_metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/privacy"
	"github.com/prebid/prebid-server/usersync"
	"github.com/stretchr/testify/assert"
)

func TestRecordPublisherPartnerNoCookieStats(t *testing.T) {

	ctrl := gomock.NewController(t)
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
	defer ctrl.Finish()

	type args struct {
		rctx models.RequestCtx
	}

	tests := []struct {
		name  string
		args  args
		setup func(*mock_metrics.MockMetricsEngine)
	}{
		{
			name: "Empty cookies and empty partner config map",
			args: args{
				rctx: models.RequestCtx{},
			},
			setup: func(mme *mock_metrics.MockMetricsEngine) {},
		},
		{
			name: "Non-empty cookie and empty partner config map",
			args: args{
				rctx: models.RequestCtx{
					UidCookie: &http.Cookie{
						Name:  "uid",
						Value: "abc123",
					},
					PartnerConfigMap: map[int]map[string]string{},
				},
			},
			setup: func(mme *mock_metrics.MockMetricsEngine) {
				models.SyncerMap = make(map[string]usersync.Syncer)
			},
		},
		{
			name: "Empty cookie and non-empty partner config map",
			args: args{
				rctx: models.RequestCtx{
					UidCookie: nil,
					PartnerConfigMap: map[int]map[string]string{
						1: {
							models.SERVER_SIDE_FLAG:    "1",
							models.PREBID_PARTNER_NAME: "partner1",
							models.BidderCode:          "bidder1",
						},
					},
					PubIDStr: "5890",
				},
			},
			setup: func(mme *mock_metrics.MockMetricsEngine) {
				models.SyncerMap = make(map[string]usersync.Syncer)
				mme.EXPECT().RecordPublisherPartnerNoCookieStats("5890", "bidder1")
			},
		},
		{
			name: "Non-empty cookie and client side partner in config map",
			args: args{
				rctx: models.RequestCtx{
					UidCookie: &http.Cookie{
						Name:  "uid",
						Value: "abc123",
					},
					PartnerConfigMap: map[int]map[string]string{
						1: {
							models.SERVER_SIDE_FLAG:    "0",
							models.PREBID_PARTNER_NAME: "partner1",
							models.BidderCode:          "bidder1",
						},
					},
					PubIDStr: "5890",
				},
			},
			setup: func(mme *mock_metrics.MockMetricsEngine) {
				models.SyncerMap = make(map[string]usersync.Syncer)
			},
		},
		{
			name: "Non-empty cookie and client side partner in config map",
			args: args{
				rctx: models.RequestCtx{
					UidCookie: &http.Cookie{
						Name:  "uid",
						Value: "abc123",
					},
					PartnerConfigMap: map[int]map[string]string{
						1: {
							models.SERVER_SIDE_FLAG:    "0",
							models.PREBID_PARTNER_NAME: "partner1",
							models.BidderCode:          "bidder1",
						},
					},
					PubIDStr: "5890",
				},
			},
			setup: func(mme *mock_metrics.MockMetricsEngine) {
				models.SyncerMap = make(map[string]usersync.Syncer)
			},
		},
		{
			name: "GetUID returns empty uid",
			args: args{
				rctx: models.RequestCtx{
					UidCookie: &http.Cookie{
						Name:  "uid",
						Value: "ewoJInRlbXBVSURzIjogewoJCSJwdWJtYXRpYyI6IHsKCQkJInVpZCI6ICI3RDc1RDI1Ri1GQUM5LTQ0M0QtQjJEMS1CMTdGRUUxMUUwMjciLAoJCQkiZXhwaXJlcyI6ICIyMDIyLTEwLTMxVDA5OjE0OjI1LjczNzI1Njg5OVoiCgkJfQoJfSwKCSJiZGF5IjogIjIwMjItMDUtMTdUMDY6NDg6MzguMDE3OTg4MjA2WiIKfQ==",
					},
					PartnerConfigMap: map[int]map[string]string{
						1: {
							models.SERVER_SIDE_FLAG:    "1",
							models.PREBID_PARTNER_NAME: "pubmatic",
							models.BidderCode:          "pubmatic",
						},
					},
					PubIDStr: "5890",
				},
			},
			setup: func(mme *mock_metrics.MockMetricsEngine) {
				models.SyncerMap = map[string]usersync.Syncer{
					"pubmatic": fakeSyncer{
						key: "pubmatic",
					},
				}

			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(mockEngine)
			tc.args.rctx.MetricsEngine = mockEngine
			RecordPublisherPartnerNoCookieStats(tc.args.rctx)
		})
	}
}

// fakeSyncer implements syncer interface for unit test cases
type fakeSyncer struct {
	key string
}

func (s fakeSyncer) Key() string {
	return s.key
}

func (s fakeSyncer) DefaultSyncType() usersync.SyncType {
	return usersync.SyncType("")
}

func (s fakeSyncer) SupportsType(syncTypes []usersync.SyncType) bool {
	return false
}

func (fakeSyncer) GetSync(syncTypes []usersync.SyncType, privacyPolicies privacy.Policies) (usersync.Sync, error) {
	return usersync.Sync{}, nil
}

// TODO -Remove this function once we remove stats-server dependency from header-bidding repository.
func TestGetPubmaticPlatform(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{
			name: "empty string",
			arg:  "",
			want: "",
		},
		{
			name: "in-app",
			arg:  models.PLATFORM_APP,
			want: models.HB_PLATFORM_APP,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, getPubmaticPlatform(tc.arg))
		})
	}
}
