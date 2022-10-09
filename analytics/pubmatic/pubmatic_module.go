package pubmatic

import (
	"github.com/prebid/prebid-server/analytics"
)

type PubMaticModule struct {
}

/*
LogAuctionObject provides rejected banner bid context to header-bidding shared req-logger map object
	This is used in case of banner format
*/
func (m *PubMaticModule) LogAuctionObject(a *analytics.AuctionObject) {
}

/*
LogVideoObject provides rejected video bid context to header-bidding shared req-logger map object
	This is used in case of banner format
*/
func (m *PubMaticModule) LogVideoObject(v *analytics.VideoObject)                   {}
func (m *PubMaticModule) LogCookieSyncObject(c *analytics.CookieSyncObject)         {}
func (m *PubMaticModule) LogSetUIDObject(u *analytics.SetUIDObject)                 {}
func (m *PubMaticModule) LogAmpObject(a *analytics.AmpObject)                       {}
func (m *PubMaticModule) LogNotificationEventObject(n *analytics.NotificationEvent) {}
