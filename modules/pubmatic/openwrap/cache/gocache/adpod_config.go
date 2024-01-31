package gocache

import (
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adpodconfig"
)

func (c *cache) populateCacheWithAdpodConfig(pubID int, profileID, displayVersion int) (err error) {
	adpodConfig, err := c.db.GetAdpodConfigs(profileID, displayVersion)
	if err != nil {
		return err
	}

	cacheKey := key(PubAdpodConfig, pubID, profileID, displayVersion)
	c.cache.Set(cacheKey, adpodConfig, getSeconds(c.cfg.CacheDefaultExpiry))
	return
}

// GetAdpodConfig this function gets adunit config from cache for a given request
func (c *cache) GetAdpodConfigs(request *openrtb2.BidRequest, pubID int, profileID, displayVersion int) (*adpodconfig.AdpodConfig, error) {
	if request.Test == 2 {
		return nil, nil
	}

	cacheKey := key(PubAdpodConfig, pubID, profileID, displayVersion)
	if adpodConfig, ok := c.cache.Get(cacheKey); ok {
		return adpodConfig.(*adpodconfig.AdpodConfig), nil
	}

	lockKey := key("%d", pubID)
	if err := c.LockAndLoad(lockKey, func() error {
		return c.populateCacheWithAdpodConfig(pubID, profileID, displayVersion)
	}); err != nil {
		return nil, err
	}

	var adpodConfig *adpodconfig.AdpodConfig
	if config, ok := c.cache.Get(cacheKey); ok && config != nil {
		adpodConfig = config.(*adpodconfig.AdpodConfig)
	}

	return adpodConfig, nil
}
