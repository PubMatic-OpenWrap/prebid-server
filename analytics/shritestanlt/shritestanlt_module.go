package shritestanlt

import (
	"fmt"

	"github.com/prebid/prebid-server/analytics"
)

type ShriTestAnalyticsModule struct {
}

func NewShriTestAnalyticsModule() analytics.PBSAnalyticsModule {
	return &ShriTestAnalyticsModule{}
}

func (s *ShriTestAnalyticsModule) LogAuctionObject(auc *analytics.AuctionObject) {
	fmt.Printf("Received auction obj: %v\n", auc)
}
func (s *ShriTestAnalyticsModule) LogVideoObject(vid *analytics.VideoObject) {
	fmt.Printf("Received video obj: %v\n", *vid)
}
func (s *ShriTestAnalyticsModule) LogCookieSyncObject(cke *analytics.CookieSyncObject) {
	fmt.Printf("Received cookie obj: %v\n", cke)
}
func (s *ShriTestAnalyticsModule) LogSetUIDObject(suid *analytics.SetUIDObject) {

}
func (s *ShriTestAnalyticsModule) LogAmpObject(amp *analytics.AmpObject) {
	fmt.Printf("Received amp obj: %v\n", amp)
}
func (s *ShriTestAnalyticsModule) LogNotificationEventObject(ntf *analytics.NotificationEvent) {
	fmt.Printf("Received notification obj: %v\n", ntf)
}
