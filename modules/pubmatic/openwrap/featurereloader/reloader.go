package featurereloader

import (
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

type reloader struct {
	cache       cache.Cache
	serviceStop chan struct{}
	sync.RWMutex
	fsc            fsc
	tbf            tbf
	ampMultiformat ampMultiformat
}

var reloaderConfig reloader

// These reloaders will be called only to Forced-Write into Cache post timer based call.
var initiateReloader = func(c cache.Cache, expiryTime int) {
	if expiryTime <= 0 {
		return
	}
	glog.Info("Feature reloader start")
	ticker := time.NewTicker(time.Duration(expiryTime) * time.Second)
	for {
		//Populating feature config maps from cache
		updateFeatureConfigMapsFromCache(c)
		select {
		case t := <-ticker.C:
			glog.Info("Feature Reloader loads cache @", t)
		case <-reloaderConfig.serviceStop:
			return
		}
	}
}

func StopReloaderService() {
	//updating serviceStop flag to true
	close(reloaderConfig.serviceStop)
}

func ResetInitReloaderTest() {
	//setting empty to mock routine
	initiateReloader = func(c cache.Cache, expiryTime int) {}
}

// Initializing reloader with cache-refresh default-expiry + 30 mins (to avoid DB load post cache refresh)
func Init(c cache.Cache, defaultExpiry int) {
	//init fsc configs
	reloaderConfig = reloader{
		cache:       c,
		serviceStop: make(chan struct{}),
		fsc: fsc{
			disabledPublishers: make(map[int]struct{}),
			thresholdsPerDsp:   make(map[int]int),
		},
		tbf: tbf{
			pubProfileTraffic: make(map[int]map[int]int),
		},
		ampMultiformat: ampMultiformat{
			enabledPublishers: make(map[int]struct{}),
		},
	}

	go initiateReloader(c, defaultExpiry+1800)
	glog.Info("Initialized feature reloader")

}

func updateFeatureConfigMapsFromCache(c cache.Cache) {
	var err error
	publisherFeatureMap, errPubFeature := c.GetPublisherFeatureMap()
	if errPubFeature != nil {
		err = models.ErrorWrap(err, errPubFeature)
	}
	errFscUpdate := updateFscConfigMapsFromCache(c, publisherFeatureMap)
	if errFscUpdate != nil {
		err = models.ErrorWrap(err, errFscUpdate)
	}
	errTbfUpdate := updateTBFConfigMapsFromCache()
	if errTbfUpdate != nil {
		err = models.ErrorWrap(err, errTbfUpdate)
	}
	glog.Error(err.Error())
	updateAmpMutiformatConfigFromCache(publisherFeatureMap)
}
