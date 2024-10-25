package gocache

import (
	"strconv"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adpodconfig"
)

func (c *cache) populateCacheWithAdpodConfig(pubID, profileID, displayVersion int) (err error) {
	adpodConfig, err := c.db.GetAdpodConfig(pubID, profileID, displayVersion)
	if err != nil {
		return err
	}

	cacheKey := key(PubAdpodConfig, pubID, profileID, displayVersion)
	c.cache.Set(cacheKey, adpodConfig, getSeconds(c.cfg.CacheDefaultExpiry))
	return
}

// GetAdpodConfig this function gets adunit config from cache for a given request
func (c *cache) GetAdpodConfig(pubID, profileID, displayVersion int) (*adpodconfig.AdpodConfig, error) {
	var adpodConfig *adpodconfig.AdpodConfig

	cacheKey := key(PubAdpodConfig, pubID, profileID, displayVersion)
	if config, ok := c.cache.Get(cacheKey); ok {
		adpodConfig, _ = config.(*adpodconfig.AdpodConfig)
		return adpodConfig, nil
	}

	lockKey := cacheKey // Making cache key as lock key
	if err := c.LockAndLoad(lockKey, func() error {
		return c.populateCacheWithAdpodConfig(pubID, profileID, displayVersion)
	}); err != nil {
		c.metricEngine.RecordDBQueryFailure(models.GetAdpodConfig, strconv.Itoa(pubID), strconv.Itoa(profileID))
		return nil, err
	}

	if config, ok := c.cache.Get(cacheKey); ok && config != nil {
		if adpodConfig, ok = config.(*adpodconfig.AdpodConfig); ok {
			return adpodConfig, nil
		}
	}

	return adpodConfig, nil
}
