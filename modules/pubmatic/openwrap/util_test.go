package openwrap

import (
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/macros"
	mock_metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/usersync"
)

func TestRecordPublisherPartnerNoCookieStats(t *testing.T) {

	ctrl := gomock.NewController(t)
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
	defer ctrl.Finish()

	type args struct {
		rctx models.RequestCtx
	}

	tests := []struct {
		name           string
		args           args
		getHttpRequest func() *http.Request
		setup          func(*mock_metrics.MockMetricsEngine)
	}{
		{
			name: "Empty cookies and empty partner config map",
			args: args{
				rctx: models.RequestCtx{},
			},
			setup: func(mme *mock_metrics.MockMetricsEngine) {},
			getHttpRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://anyurl.com", nil)
				return req
			},
		},
		{
			name: "Empty cookie and non-empty partner config map",
			args: args{
				rctx: models.RequestCtx{
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
			getHttpRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://anyurl.com", nil)
				return req
			},
		},
		{
			name: "only client side partner in config map",
			args: args{
				rctx: models.RequestCtx{
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
			getHttpRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://anyurl.com", nil)
				return req
			},
		},
		{
			name: "GetUID returns empty uid",
			args: args{
				rctx: models.RequestCtx{
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
				mme.EXPECT().RecordPublisherPartnerNoCookieStats("5890", "pubmatic")
			},
			getHttpRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://anyurl.com", nil)
				return req
			},
		},
		{
			name: "GetUID returns non empty uid",
			args: args{
				rctx: models.RequestCtx{
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
			getHttpRequest: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://anyurl.com", nil)

				cookie := &http.Cookie{
					Name:  "uids",
					Value: "ewoJInRlbXBVSURzIjogewoJCSJwdWJtYXRpYyI6IHsKCQkJInVpZCI6ICI3RDc1RDI1Ri1GQUM5LTQ0M0QtQjJEMS1CMTdGRUUxMUUwMjciLAoJCQkiZXhwaXJlcyI6ICIyMDIyLTEwLTMxVDA5OjE0OjI1LjczNzI1Njg5OVoiCgkJfQoJfSwKCSJiZGF5IjogIjIwMjItMDUtMTdUMDY6NDg6MzguMDE3OTg4MjA2WiIKfQ==",
				}
				req.AddCookie(cookie)
				return req
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(mockEngine)
			tc.args.rctx.MetricsEngine = mockEngine
			tc.args.rctx.ParsedUidCookie = usersync.ReadCookie(tc.getHttpRequest(), usersync.Base64Decoder{}, &config.HostCookie{})
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

func (fakeSyncer) GetSync([]usersync.SyncType, macros.UserSyncPrivacy) (usersync.Sync, error) {
	return usersync.Sync{}, nil
}
