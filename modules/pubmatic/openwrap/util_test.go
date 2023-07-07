package openwrap

import (
	"net/http"
	"testing"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/privacy"
	"github.com/prebid/prebid-server/usersync"
)

func TestParseRequestCookies(t *testing.T) {
	type args struct {
		httpReqUIDCookie *http.Cookie
		partnerConfigMap map[int]map[string]string
		syncerMap        map[string]usersync.Syncer
	}

	type want struct {
		partnerCookieMap map[string]int
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Empty cookies and empty partner config map",
			args: args{
				httpReqUIDCookie: nil,
				partnerConfigMap: map[int]map[string]string{},
				syncerMap:        make(map[string]usersync.Syncer),
			},
			want: want{
				partnerCookieMap: map[string]int{},
			},
		},
		{
			name: "Non-empty cookie and empty partner config map",
			args: args{
				httpReqUIDCookie: &http.Cookie{
					Name:  "uid",
					Value: "abc123",
				},
				partnerConfigMap: map[int]map[string]string{},
				syncerMap:        make(map[string]usersync.Syncer),
			},
			want: want{
				partnerCookieMap: map[string]int{},
			},
		},
		{
			name: "Empty cookie and non-empty partner config map",
			args: args{
				httpReqUIDCookie: nil,
				partnerConfigMap: map[int]map[string]string{
					1: {
						models.SERVER_SIDE_FLAG:    "1",
						models.PREBID_PARTNER_NAME: "partner1",
						models.BidderCode:          "bidder1",
					},
				},
				syncerMap: make(map[string]usersync.Syncer),
			},
			want: want{
				partnerCookieMap: map[string]int{
					"bidder1": 0,
				},
			},
		},
		{
			name: "Non-empty cookie and client side partner in config map",
			args: args{
				httpReqUIDCookie: &http.Cookie{
					Name:  "uid",
					Value: "abc123",
				},
				partnerConfigMap: map[int]map[string]string{
					1: {
						models.SERVER_SIDE_FLAG:    "0",
						models.PREBID_PARTNER_NAME: "partner1",
						models.BidderCode:          "bidder1",
					},
				},
				syncerMap: make(map[string]usersync.Syncer),
			},
			want: want{
				partnerCookieMap: map[string]int{},
			},
		},
		{
			name: "Non-empty cookie and client side partner in config map",
			args: args{
				httpReqUIDCookie: &http.Cookie{
					Name:  "uid",
					Value: "abc123",
				},
				partnerConfigMap: map[int]map[string]string{
					1: {
						models.SERVER_SIDE_FLAG:    "0",
						models.PREBID_PARTNER_NAME: "partner1",
						models.BidderCode:          "bidder1",
					},
				},
				syncerMap: make(map[string]usersync.Syncer),
			},
			want: want{
				partnerCookieMap: map[string]int{},
			},
		},
		{
			name: "GetUID returns empty uid",
			args: args{
				httpReqUIDCookie: &http.Cookie{
					Name:  "uid",
					Value: "ewoJInRlbXBVSURzIjogewoJCSJwdWJtYXRpYyI6IHsKCQkJInVpZCI6ICI3RDc1RDI1Ri1GQUM5LTQ0M0QtQjJEMS1CMTdGRUUxMUUwMjciLAoJCQkiZXhwaXJlcyI6ICIyMDIyLTEwLTMxVDA5OjE0OjI1LjczNzI1Njg5OVoiCgkJfQoJfSwKCSJiZGF5IjogIjIwMjItMDUtMTdUMDY6NDg6MzguMDE3OTg4MjA2WiIKfQ==",
				},
				partnerConfigMap: map[int]map[string]string{
					1: {
						models.SERVER_SIDE_FLAG:    "1",
						models.PREBID_PARTNER_NAME: "pubmatic",
						models.BidderCode:          "pubmatic",
					},
				},
				syncerMap: map[string]usersync.Syncer{
					"pubmatic": fakeSyncer{
						key: "pubmatic",
					},
				},
			},
			want: want{
				partnerCookieMap: map[string]int{
					"pubmatic": 1,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			models.SyncerMap = tc.args.syncerMap
			result := ParseRequestCookies(tc.args.httpReqUIDCookie, tc.args.partnerConfigMap)

			if len(result) != len(tc.want.partnerCookieMap) {
				t.Errorf("Unexpected length of cookie flag map. Expected: %d, Got: %d", len(tc.want.partnerCookieMap), len(result))
			}

			for bidder, expectedFlag := range tc.want.partnerCookieMap {
				if result[bidder] != expectedFlag {
					t.Errorf("Unexpected flag for bidder %s. Expected: %d, Got: %d", bidder, expectedFlag, result[bidder])
				}
			}
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
