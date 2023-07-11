package gocache

import (
	"strings"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

func (c *cache) populateCacheWithAdunitConfig(pubID int, profileID, displayVersion int) {
	adunitConfig, err := c.db.GetAdunitConfig(profileID, displayVersion)
	if err != nil {
		return
	}

	caseFoldConfigMap := make(map[string]*adunitconfig.AdConfig, len(adunitConfig.Config))
	for k, v := range adunitConfig.Config {
		caseFoldConfigMap[strings.ToLower(k)] = v
	}
	adunitConfig.Config = caseFoldConfigMap

	cacheKey := key(PubAdunitConfig, pubID, profileID, displayVersion)
	c.cache.Set(cacheKey, adunitConfig, getSeconds(c.cfg.CacheDefaultExpiry))
}

// GetAdunitConfigFromCache this function gets adunit config from cache for a given request
func (c *cache) GetAdunitConfigFromCache(request *openrtb2.BidRequest, pubID int, profileID, displayVersion int) *adunitconfig.AdUnitConfig {
	if request.Test == 2 {
		return nil
	}

	cacheKey := key(PubAdunitConfig, pubID, profileID, displayVersion)
	if obj, ok := c.cache.Get(cacheKey); ok {
		if v, ok := obj.(*adunitconfig.AdUnitConfig); ok {
			return v
		}
	}

	return nil
}
