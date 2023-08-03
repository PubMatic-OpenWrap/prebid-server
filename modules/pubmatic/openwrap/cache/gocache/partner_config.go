package gocache

import (
	"fmt"

	"github.com/pkg/errors"
)

// GetPartnerConfigMap returns partnerConfigMap using given parameters
func (c *cache) GetPartnerConfigMap(pubID, profileID, displayVersion int) (map[int]map[string]string, error) {
	dbAccessed := false
	var err error

	pubLockKey := key("%d", pubID)
	if mapNameHash, ok := c.cache.Get(key(PubSlotNameHash, pubID)); !ok || mapNameHash == nil {
		errPubSlotNameHash := c.LockAndLoad(pubLockKey, func() error {
			dbAccessed = true
			return c.populateCacheWithPubSlotNameHash(pubID)
		})
		if errPubSlotNameHash != nil {
			err = errors.Wrap(err, errPubSlotNameHash.Error())
		}
		//TODO: Add stat if error from the DB
	}

	if vastTags, ok := c.cache.Get(key(PubVASTTags, pubID)); !ok || vastTags == nil {
		errPublisherVASTTag := c.LockAndLoad(pubLockKey, func() error {
			dbAccessed = true
			return c.populatePublisherVASTTags(pubID)
		})
		if errPublisherVASTTag != nil {
			err = errors.Wrap(err, errPublisherVASTTag.Error())
		}
		//TODO: Add stat if error from the DB
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
		err = errors.Wrap(err, errGetPartnerConfig.Error())
	}

	var partnerConfigMap map[int]map[string]string
	if obj, ok := c.cache.Get(cacheKey); ok && obj != nil {
		partnerConfigMap = obj.(map[int]map[string]string)
	}

	if dbAccessed {
		//TODO: add stat to RecordGetProfileDataTime
	}
	return partnerConfigMap, err
}

func (c *cache) getActivePartnerConfigAndPopulateWrapperMappings(pubID, profileID, displayVersion int) (err error) {
	cacheKey := key(PUB_HB_PARTNER, pubID, profileID, displayVersion)
	partnerConfigMap, err := c.db.GetActivePartnerConfigurations(pubID, profileID, displayVersion)
	if err != nil {
		return
	}

	if len(partnerConfigMap) == 0 {
		return fmt.Errorf("there are no active partners for pubId:%d, profileId:%d, displayVersion:%d", pubID, profileID, displayVersion)
	}

	c.cache.Set(cacheKey, partnerConfigMap, getSeconds(c.cfg.CacheDefaultExpiry))
	if errWrapperSlotMapping := c.populateCacheWithWrapperSlotMappings(pubID, partnerConfigMap, profileID, displayVersion); errWrapperSlotMapping != nil {
		err = errors.Wrap(err, errWrapperSlotMapping.Error())
	}
	if errAdunitConfig := c.populateCacheWithAdunitConfig(pubID, profileID, displayVersion); errAdunitConfig != nil {
		err = errors.Wrap(err, errAdunitConfig.Error())
	}
	return
}
