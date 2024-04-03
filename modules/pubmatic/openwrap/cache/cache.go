package cache

import (
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

type Cache interface {
	GetPartnerConfigMap(pubid, profileid, displayversion int, endpoint string) (map[int]map[string]string, error)
	GetAdunitConfigFromCache(request *openrtb2.BidRequest, pubID int, profileID, displayVersion int) *adunitconfig.AdUnitConfig
	GetMappingsFromCacheV25(rctx models.RequestCtx, partnerID int) map[string]models.SlotMapping
	GetSlotToHashValueMapFromCacheV25(rctx models.RequestCtx, partnerID int) models.SlotMappingInfo
	GetPublisherVASTTagsFromCache(pubID int) models.PublisherVASTTags

	GetFSCThresholdPerDSP() (map[int]int, error)
	GetPublisherFeatureMap() (map[int]map[int]models.FeatureData, error)

	Set(key string, value interface{})
	Get(key string) (interface{}, bool)
}
