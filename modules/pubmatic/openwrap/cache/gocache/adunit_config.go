package gocache

import (
	"strings"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
)

func (c *cache) populateCacheWithAdunitConfig(pubID int, profileID, displayVersion int) (err error) {
	adunitConfig, err := c.db.GetAdunitConfig(profileID, displayVersion)
	if err != nil {
		glog.Errorf(models.DBQueryFailure, "adunitConfigQuery", pubID, profileID, err)
		return err
	}

	if adunitConfig != nil {
		caseFoldConfigMap := make(map[string]*adunitconfig.AdConfig, len(adunitConfig.Config))
		for k, v := range adunitConfig.Config {
			v.UniversalPixel = validUPixels(v.UniversalPixel)
			caseFoldConfigMap[strings.ToLower(k)] = v
		}
		adunitConfig.Config = caseFoldConfigMap
	}

	cacheKey := key(PubAdunitConfig, pubID, profileID, displayVersion)
	c.cache.Set(cacheKey, adunitConfig, getSeconds(c.cfg.CacheDefaultExpiry))
	return
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
