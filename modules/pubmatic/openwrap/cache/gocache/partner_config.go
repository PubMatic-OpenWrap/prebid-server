package gocache

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
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
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.metricEngine.RecordDBQueryFailure(models.LiveVersionInnerQuery, strconv.Itoa(pubID), strconv.Itoa(profileID))
			c.cache.Set(cacheKey, partnerConfigMap, getSeconds(c.cfg.CacheDefaultExpiry))
		} else {
			c.metricEngine.RecordDBQueryFailure(models.PartnerConfigQuery, strconv.Itoa(pubID), strconv.Itoa(profileID))
		}
		glog.Errorf(models.ErrDBQueryFailed, models.PartnerConfigQuery, pubID, profileID, err)
		return err
	}

	if len(partnerConfigMap) == 0 {
		c.cache.Set(cacheKey, partnerConfigMap, getSeconds(c.cfg.CacheDefaultExpiry))
		return fmt.Errorf(models.EmptyPartnerConfig, pubID, profileID, displayVersion)
	}

	err = c.populateCacheWithWrapperSlotMappings(pubID, partnerConfigMap, profileID, displayVersion)
	if err != nil {
		queryType := models.WrapperSlotMappingsQuery
		if displayVersion == 0 {
			queryType = models.WrapperLiveVersionSlotMappings
		}
		c.metricEngine.RecordDBQueryFailure(queryType, strconv.Itoa(pubID), strconv.Itoa(profileID))
		glog.Errorf(models.ErrDBQueryFailed, queryType, pubID, profileID, err)
		return err
	}

	err = c.populateCacheWithAdunitConfig(pubID, profileID, displayVersion)
	if err != nil {
		queryType := models.AdunitConfigQuery
		if displayVersion == 0 {
			queryType = models.AdunitConfigForLiveVersion
		}
		if errors.Is(err, adunitconfig.ErrAdUnitUnmarshal) {
			queryType = models.AdUnitFailUnmarshal
		}
		c.metricEngine.RecordDBQueryFailure(queryType, strconv.Itoa(pubID), strconv.Itoa(profileID))
		glog.Errorf(models.ErrDBQueryFailed, queryType, pubID, profileID, err)
		return err
	}

	c.updatePartnerConfigWithBidderFilters(partnerConfigMap, pubID, profileID, displayVersion)
	c.cache.Set(cacheKey, partnerConfigMap, getSeconds(c.cfg.CacheDefaultExpiry))
	return
}

func (c *cache) updatePartnerConfigWithBidderFilters(partnerConfigs map[int]map[string]string, pubID, profileID, displayVersion int) {

	cacheKey := key(PubAdunitConfig, pubID, profileID, displayVersion)
	obj, ok := c.cache.Get(cacheKey)
	if !ok {
		return
	}

	adUnitCfg, ok := obj.(*adunitconfig.AdUnitConfig)
	if !ok || adUnitCfg == nil {
		return
	}

	bidderfilter := map[string]string{}
	defaultAdUnitConfig := adUnitCfg.Config["default"]
	if defaultAdUnitConfig.BidderFilter != nil {
		for _, filter := range defaultAdUnitConfig.BidderFilter.Filters {
			for _, bidder := range filter.Bidders {
				bidderfilter[bidder] = string(filter.BiddingConditions)
			}
		}
	}

	if len(bidderfilter) == 0 {
		return
	}

	for id, cfg := range partnerConfigs {
		if biddingCodition, ok := bidderfilter[cfg[models.BidderCode]]; ok {
			partnerConfigs[id][models.BidderFilters] = biddingCodition
		}
	}
}
