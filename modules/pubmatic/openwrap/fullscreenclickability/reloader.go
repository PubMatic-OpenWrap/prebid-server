package fullscreenclickability

import (
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
)

// These reloaders will be called only to Forced-Write into Cache post timer based call.
var initiateReloader = func(c cache.Cache, expiryTime int) {
	if expiryTime <= 0 {
		return
	}
	glog.Info("FSC Reloader start")
	ticker := time.NewTicker(time.Duration(expiryTime) * time.Second)
	for {
		//Populating FscConfigMaps
		updateFscConfigMapsFromCache(c)
		select {
		case t := <-ticker.C:
			glog.Info("FSC Reloader loads cache @", t)
		case <-fscConfigs.serviceStop:
			return
		}
	}
}

func StopReloaderService() {
	//updating serviceStop flag to true
	close(fscConfigs.serviceStop)
}

func ResetInitFscReloaderTest() {
	//setting empty to mock routine
	initiateReloader = func(c cache.Cache, expiryTime int) {}
}
