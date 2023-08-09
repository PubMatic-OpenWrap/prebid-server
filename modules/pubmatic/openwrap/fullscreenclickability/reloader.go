package fullscreenclickability

import (
	"time"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
)

var initiateReloader = func(c cache.Cache, expiryTime int) {
	//  add log starting Reloader
	ticker := time.NewTicker(time.Duration(expiryTime) * time.Second)
	for {
		//Populating FscConfigMaps
		updateFscConfigMapsFromCache(c)
		select {
		case <-ticker.C:
			// add log for cache-refresh-reloader
		case <-fscConfigs.serviceStop:
			return
		}
	}
}
