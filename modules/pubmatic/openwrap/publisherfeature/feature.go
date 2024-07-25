package publisherfeature

import "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"

type Feature interface {
	IsFscApplicable(pubId int, seat string, dspId int) bool
	IsAmpMultiformatEnabled(pubId int) bool
	IsMaxFloorsEnabled(pubId int) bool
	IsTBFFeatureEnabled(pubid int, profid int) bool
	IsAnalyticsTrackingThrottled(pubID, profileID int) (bool, bool)
	IsBidRecoveryEnabled(pubID int, profileID int) bool
	IsApplovinABTestEnabled(pubID int, profileID string) bool
	GetApplovinABTestFloors(pubID int, profileID string) models.ApplovinAdUnitFloors
}
