package gocache

import (
	"fmt"
	"strconv"
	"time"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

// GetPartnerConfigMap returns partnerConfigMap using given parameters
func (c *cache) GetPartnerConfigMap(pubID, profileID, displayVersion int) (map[int]map[string]string, error) {
	dbAccessed := false
	var err error
	startTime := time.Now()

	pubLockKey := key("%d", pubID)
	if mapNameHash, ok := c.cache.Get(key(PubSlotNameHash, pubID)); !ok || mapNameHash == nil {
		errPubSlotNameHash := c.LockAndLoad(pubLockKey, func() error {
			dbAccessed = true
			return c.populateCacheWithPubSlotNameHash(pubID)
		})
		if errPubSlotNameHash != nil {
			c.metricEngine.RecordDBQueryFailure(models.SlotNameHash, strconv.Itoa(pubID), strconv.Itoa(profileID))
			err = models.ErrorWrap(err, errPubSlotNameHash)
		}
	}

	if vastTags, ok := c.cache.Get(key(PubVASTTags, pubID)); !ok || vastTags == nil {
		errPublisherVASTTag := c.LockAndLoad(pubLockKey, func() error {
			dbAccessed = true
			return c.populatePublisherVASTTags(pubID)
		})
		if errPublisherVASTTag != nil {
			c.metricEngine.RecordDBQueryFailure(models.PublisherVASTTagsQuery, strconv.Itoa(pubID), strconv.Itoa(profileID))
			err = models.ErrorWrap(err, errPublisherVASTTag)
		}
	}

	cacheKey := key(PUB_HB_PARTNER, pubID, profileID, displayVersion)
	if obj, ok := c.cache.Get(cacheKey); ok && obj != nil {
		return obj.(map[int]map[string]string), err
	}

	lockKey := key("%d%d%d", pubID, profileID, displayVersion)
	if errGetPartnerConfig := c.LockAndLoad(lockKey, func() error {
		dbAccessed = true
		return c.getActivePartnerConfigAndPopulateWrapperMappings(pubID, profileID, displayVersion)
	}); errGetPartnerConfig != nil {
		err = models.ErrorWrap(err, errGetPartnerConfig)
	}

	var partnerConfigMap map[int]map[string]string
	if obj, ok := c.cache.Get(cacheKey); ok && obj != nil {
		partnerConfigMap = obj.(map[int]map[string]string)
	}

	if dbAccessed {
		c.metricEngine.RecordGetProfileDataTime(time.Since(startTime))
	}
	return partnerConfigMap, err
}

func (c *cache) getActivePartnerConfigAndPopulateWrapperMappings(pubID, profileID, displayVersion int) (err error) {
	cacheKey := key(PUB_HB_PARTNER, pubID, profileID, displayVersion)
	partnerConfigMap, err := c.db.GetActivePartnerConfigurations(pubID, profileID, displayVersion)
	if models.GetErrorCode(err) == models.DBErrorCode {
		c.metricEngine.RecordDBQueryFailure(models.PartnerConfigQuery, strconv.Itoa(pubID), strconv.Itoa(profileID))
		return
	}

	if len(partnerConfigMap) == 0 {
		c.cache.Set(cacheKey, partnerConfigMap, getSeconds(c.cfg.CacheDefaultExpiry)) // Setting empty partner Config map
		return fmt.Errorf("there are no active partners for pubId:%d, profileId:%d, displayVersion:%d", pubID, profileID, displayVersion)
	}

	err = c.populateCacheWithWrapperSlotMappings(pubID, partnerConfigMap, profileID, displayVersion)
	if models.GetErrorCode(err) == models.DBErrorCode {
		queryType := models.WrapperSlotMappingsQuery
		if displayVersion == 0 {
			queryType = models.WrapperLiveVersionSlotMappings
		}
		c.metricEngine.RecordDBQueryFailure(queryType, strconv.Itoa(pubID), strconv.Itoa(profileID))
		return err
	}

	err = c.populateCacheWithAdunitConfig(pubID, profileID, displayVersion)
	if err != nil {
		queryType := models.AdunitConfigQuery
		if displayVersion == 0 {
			queryType = models.AdunitConfigForLiveVersion
		}
		if models.GetErrorCode(err) == models.AdUnitUnmarshalErrorCode {
			queryType = models.AdUnitFailUnmarshal
		}
		c.metricEngine.RecordDBQueryFailure(queryType, strconv.Itoa(pubID), strconv.Itoa(profileID))
		// In case of Error in AdUnit Unmarshal, push PartnerConfig and process request
		if models.GetErrorCode(err) == models.DBErrorCode {
			return err
		}
	}

	c.cache.Set(cacheKey, partnerConfigMap, getSeconds(c.cfg.CacheDefaultExpiry))
	return nil
}
