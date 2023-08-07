package fullscreenclickability

import (
	"time"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
)

var initiateReloader = func(c cache.Cache, expiryTime int) {
	// logger.Info("FSC Reloader start")
	ticker := time.NewTicker(time.Duration(expiryTime) * time.Second)
	for {
		//Populating FscConfigMaps
		updateFscConfigMapsFromCache(c)
		select {
		case <-fscConfigs.serviceStop:
			return
		case <-ticker.C:
			// logger.Info("FSC Reloader loads cache @%v", t)
		}
	}
}
