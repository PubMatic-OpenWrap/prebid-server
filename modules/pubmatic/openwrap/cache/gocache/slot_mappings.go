package gocache

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

// PopulateCacheWithPubSlotNameHash will put the slot names and hashes for a publisher in cache
func (c *cache) populateCacheWithPubSlotNameHash(pubID int) (err error) {
	cacheKey := key(PubSlotNameHash, pubID)

	publisherSlotNameHashMap, err := c.db.GetPublisherSlotNameHash(pubID)
	if err != nil {
		glog.Errorf(models.DBQueryFailure, "GetSlotNameHash", pubID, "", err)
		return err
	}
	//This call may set nil publisherSlotNameHashMap in cache
	c.cache.Set(cacheKey, publisherSlotNameHashMap, getSeconds(c.cfg.CacheDefaultExpiry))
	return
}

// PopulateCacheWithWrapperSlotMappings will get the SlotMappings from database and put them in cache.
func (c *cache) populateCacheWithWrapperSlotMappings(pubID int, partnerConfigMap map[int]map[string]string, profileID, displayVersion int) error {
	partnerSlotMappingMap, err := c.db.GetWrapperSlotMappings(partnerConfigMap, profileID, displayVersion)
	if err != nil {
		glog.Errorf(models.DBQueryFailure, "WrapperSlotMappingsQuery/WrapperLiveVersionSlotMappings", pubID, profileID, err)
		return err
	}

	//put a version level dummy entry in cache denoting mappings are present for this version
	cacheKey := key(PUB_SLOT_INFO, pubID, profileID, displayVersion, 0)
	c.cache.Set(cacheKey, make(map[string]models.SlotMapping, 0), getSeconds(c.cfg.CacheDefaultExpiry))

	if len(partnerSlotMappingMap) == 0 {
		for _, partnerConf := range partnerConfigMap {
			partnerID, _ := strconv.Atoi(partnerConf[models.PARTNER_ID])
			cacheKey = key(PUB_SLOT_INFO, pubID, profileID, displayVersion, partnerID)
			c.cache.Set(cacheKey, make(map[string]models.SlotMapping, 0), getSeconds(c.cfg.CacheDefaultExpiry))
		}
		return err
	}

	var nameHashMap map[string]string
	obj, ok := c.cache.Get(key(PubSlotNameHash, pubID))
	if ok && obj != nil {
		nameHashMap = obj.(map[string]string)
	}

	for partnerID, slotMappingList := range partnerSlotMappingMap {
		slotNameToMappingMap := make(map[string]models.SlotMapping, len(slotMappingList))
		slotNameToHashValueMap := make(map[string]string, len(slotMappingList))
		slotNameOrderedList := make([]string, 0)
		sort.Slice(slotMappingList, func(i, j int) bool {
			return slotMappingList[i].OrderID < slotMappingList[j].OrderID
		})
		for _, slotMapping := range slotMappingList {
			slotMapping.Hash = nameHashMap[slotMapping.SlotName]

			var mappingJSONObj map[string]interface{}
			if err := json.Unmarshal([]byte(slotMapping.MappingJson), &mappingJSONObj); err != nil {
				continue
			}

			cfgMap := partnerConfigMap[partnerID]
			bidderCode := cfgMap[models.BidderCode]
			if bidderCode == string(openrtb_ext.BidderPubmatic) {
				//Adding slotName from DB in fieldMap for PubMatic, as slotName from DB should be sent to PubMatic instead of slotName from request
				//This is required for case in-sensitive mapping
				mappingJSONObj[models.KEY_OW_SLOT_NAME] = slotMapping.SlotName
			}

			slotMapping.SlotMappings = mappingJSONObj
			slotNameToMappingMap[strings.ToLower(slotMapping.SlotName)] = slotMapping
			slotNameToHashValueMap[slotMapping.SlotName] = slotMapping.Hash
			slotNameOrderedList = append(slotNameOrderedList, slotMapping.SlotName)
		}
		cacheKey = key(PUB_SLOT_INFO, pubID, profileID, displayVersion, partnerID)
		c.cache.Set(cacheKey, slotNameToMappingMap, getSeconds(c.cfg.CacheDefaultExpiry))

		slotMappingInfoObj := models.SlotMappingInfo{
			OrderedSlotList: slotNameOrderedList,
			HashValueMap:    slotNameToHashValueMap,
		}
		cacheKey = key(PubSlotHashInfo, pubID, profileID, displayVersion, partnerID)
		c.cache.Set(cacheKey, slotMappingInfoObj, getSeconds(c.cfg.CacheDefaultExpiry))
	}

	return nil
}

// GetMappingsFromCacheV25 will return mapping of each partner in partnerConf map
func (c *cache) GetMappingsFromCacheV25(rctx models.RequestCtx, partnerID int) map[string]models.SlotMapping {
	cacheKey := key(PUB_SLOT_INFO, rctx.PubID, rctx.ProfileID, rctx.DisplayID, partnerID)
	if value, ok := c.cache.Get(cacheKey); ok {
		return value.(map[string]models.SlotMapping)
	}

	return nil
}

/*GetSlotToHashValueMapFromCacheV25 returns SlotMappingInfo object from cache that contains and ordered list of slot names by order_id field and a map of slot names to their hash values*/
func (c *cache) GetSlotToHashValueMapFromCacheV25(rctx models.RequestCtx, partnerID int) models.SlotMappingInfo {
	cacheKey := key(PubSlotHashInfo, rctx.PubID, rctx.ProfileID, rctx.DisplayID, partnerID)
	if value, ok := c.cache.Get(cacheKey); ok {
		return value.(models.SlotMappingInfo)
	}

	return models.SlotMappingInfo{}
}
