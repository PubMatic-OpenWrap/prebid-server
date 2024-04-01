package publisherfeature

import (
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

type feature struct {
	cache       cache.Cache
	serviceStop chan struct{}
	sync.RWMutex
	defaultExpiry    int
	publisherFeature map[int]int
	fsc              fsc
	tbf              tbf
	ampMultiformat   ampMultiformat
}

var fe *feature
var fOnce sync.Once

func New(c cache.Cache, defaultExpiry int) *feature {
	fOnce.Do(func() {
		fe = &feature{
			cache:            c,
			serviceStop:      make(chan struct{}),
			defaultExpiry:    defaultExpiry,
			publisherFeature: make(map[int]int),
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
	})
	return fe
}

func (fe *feature) Start() {
	go fe.init()
	glog.Info("Initialized feature reloader")
}

func (fe *feature) Stop() {
	//updating serviceStop flag to true
	close(fe.serviceStop)
}

// Initializing reloader with cache-refresh default-expiry + 30 mins (to avoid DB load post cache refresh)
func (fe *feature) init() {
	if fe.defaultExpiry <= 0 {
		return
	}
	glog.Info("Feature reloader start")
	ticker := time.NewTicker(time.Duration(fe.defaultExpiry+1800) * time.Second)
	for {
		//Populating feature config maps from cache
		fe.updateFeatureConfigMaps()
		select {
		case t := <-ticker.C:
			glog.Info("Feature Reloader loads cache @", t)
		case <-fe.serviceStop:
			return
		}
	}
}

func (fe *feature) updateFeatureConfigMaps() {
	var err error
	var errFscUpdate error
	publisherFeatureMap, errPubFeature := fe.cache.GetPublisherFeatureMap()
	if errPubFeature != nil {
		err = models.ErrorWrap(err, errPubFeature)
	}
	if err == nil {
		fe.Lock()
		fe.publisherFeature = publisherFeatureMap
		fe.Unlock()
		errFscUpdate = fe.updateFscConfigMapsFromCache()
	}
	if errFscUpdate != nil {
		err = models.ErrorWrap(err, errFscUpdate)
	}

	errTbfUpdate := fe.updateTBFConfigMapsFromCache()
	if errTbfUpdate != nil {
		err = models.ErrorWrap(err, errTbfUpdate)
	}
	glog.Error(err.Error())
	fe.updateAmpMutiformatConfigFromCache()
}