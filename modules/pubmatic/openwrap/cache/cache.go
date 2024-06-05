package cache

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adpodconfig"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
)

type Cache interface {
	GetPartnerConfigMap(pubid, profileid, displayversion int) (map[int]map[string]string, error)
	GetAdunitConfigFromCache(request *openrtb2.BidRequest, pubID int, profileID, displayVersion int) *adunitconfig.AdUnitConfig
	GetMappingsFromCacheV25(rctx models.RequestCtx, partnerID int) map[string]models.SlotMapping
	GetSlotToHashValueMapFromCacheV25(rctx models.RequestCtx, partnerID int) models.SlotMappingInfo
	GetPublisherVASTTagsFromCache(pubID int) models.PublisherVASTTags
	GetAdpodConfig(pubID, profileID, displayVersion int) (*adpodconfig.AdpodConfig, error)

	GetFSCThresholdPerDSP() (map[int]int, error)
	GetPublisherFeatureMap() (map[int]map[int]models.FeatureData, error)

	Set(key string, value interface{})
	Get(key string) (interface{}, bool)
}
