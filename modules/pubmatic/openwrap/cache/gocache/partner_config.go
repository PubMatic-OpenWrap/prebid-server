package gocache

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

// GetPartnerConfigMap returns partnerConfigMap using given parameters
func (c *cache) GetPartnerConfigMap(pubID, profileID, displayVersion int, endpoint string) (map[int]map[string]string, error) {
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
			err = errorWrap(err, errPubSlotNameHash)
		}
	}

	if vastTags, ok := c.cache.Get(key(PubVASTTags, pubID)); !ok || vastTags == nil {
		errPublisherVASTTag := c.LockAndLoad(pubLockKey, func() error {
			dbAccessed = true
			return c.populatePublisherVASTTags(pubID)
		})
		if errPublisherVASTTag != nil {
			c.metricEngine.RecordDBQueryFailure(models.PublisherVASTTagsQuery, strconv.Itoa(pubID), strconv.Itoa(profileID))
			err = errorWrap(err, errPublisherVASTTag)
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
		err = errorWrap(err, errGetPartnerConfig)
	}

	var partnerConfigMap map[int]map[string]string
	if obj, ok := c.cache.Get(cacheKey); ok && obj != nil {
		partnerConfigMap = obj.(map[int]map[string]string)
	}

	if dbAccessed {
		c.metricEngine.RecordGetProfileDataTime(endpoint, strconv.Itoa(profileID), time.Since(startTime))
	}
	return partnerConfigMap, err
}

func (c *cache) getActivePartnerConfigAndPopulateWrapperMappings(pubID, profileID, displayVersion int) (err error) {
	cacheKey := key(PUB_HB_PARTNER, pubID, profileID, displayVersion)
	partnerConfigMap, err := c.db.GetActivePartnerConfigurations(pubID, profileID, displayVersion)
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.PartnerConfigQuery, strconv.Itoa(pubID), strconv.Itoa(profileID))
		return
	}

	if len(partnerConfigMap) == 0 {
		return fmt.Errorf("there are no active partners for pubId:%d, profileId:%d, displayVersion:%d", pubID, profileID, displayVersion)
	}

	c.cache.Set(cacheKey, partnerConfigMap, getSeconds(c.cfg.CacheDefaultExpiry))
	if errWrapperSlotMapping := c.populateCacheWithWrapperSlotMappings(pubID, partnerConfigMap, profileID, displayVersion); errWrapperSlotMapping != nil {
		err = errorWrap(err, errWrapperSlotMapping)
		queryType := models.WrapperSlotMappingsQuery
		if displayVersion == 0 {
			queryType = models.WrapperLiveVersionSlotMappings
		}
		c.metricEngine.RecordDBQueryFailure(queryType, strconv.Itoa(pubID), strconv.Itoa(profileID))
	}
	if errAdunitConfig := c.populateCacheWithAdunitConfig(pubID, profileID, displayVersion); errAdunitConfig != nil {
		queryType := models.AdunitConfigQuery
		if displayVersion == 0 {
			queryType = models.AdunitConfigForLiveVersion
		}
		if errors.Is(errAdunitConfig, adunitconfig.ErrAdUnitUnmarshal) {
			queryType = models.AdUnitFailUnmarshal
		}
		c.metricEngine.RecordDBQueryFailure(queryType, strconv.Itoa(pubID), strconv.Itoa(profileID))
		err = errorWrap(err, errAdunitConfig)
	}
	return
}
