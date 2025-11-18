package publisherfeature

import (
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

type Config struct {
	Cache                 cache.Cache
	DefaultExpiry         int
	AnalyticsThrottleList string
}

type feature struct {
	cache       cache.Cache
	serviceStop chan struct{}
	sync.RWMutex
	defaultExpiry           int
	publisherFeature        map[int]map[int]models.FeatureData
	fsc                     fsc
	tbf                     tbf
	ant                     analyticsThrottle
	ampMultiformat          ampMultiformat
	maxFloors               maxFloors
	bidRecovery             bidRecovery
	appLovinMultiFloors     appLovinMultiFloors
	appLovinSchainABTest    appLovinSchainABTest
	impCountingMethod       impCountingMethod
	gdprCountryCodes        gdprCountryCodes
	mbmf                    *mbmf
	dynamicFloor            dynamicFloor
	performanceDSPs         performanceDSPs
	inViewEnabledPublishers inViewEnabledPublishers
}

var fe *feature
var fOnce sync.Once

func New(config Config) *feature {
	fOnce.Do(func() {
		fe = &feature{
			cache:            config.Cache,
			serviceStop:      make(chan struct{}),
			defaultExpiry:    config.DefaultExpiry,
			publisherFeature: make(map[int]map[int]models.FeatureData),
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
			maxFloors: maxFloors{
				enabledPublishers: make(map[int]struct{}),
			},
			ant: analyticsThrottle{
				vault: newPubThrottling(config.AnalyticsThrottleList),
				db:    newPubThrottling(config.AnalyticsThrottleList),
			},
			appLovinMultiFloors: appLovinMultiFloors{
				enabledPublisherProfile: make(map[int]map[string]models.ApplovinAdUnitFloors),
			},
			appLovinSchainABTest: appLovinSchainABTest{
				schainABTestPercent: 0,
			},
			impCountingMethod:       newImpCountingMethod(),
			gdprCountryCodes:        newGDPRCountryCodes(),
			mbmf:                    newMBMF(),
			dynamicFloor:            newDynamicFloor(),
			performanceDSPs:         newPerformanceDSPs(),
			inViewEnabledPublishers: newInViewEnabledPublishers(),
		}
	})
	return fe
}

func (fe *feature) Start() {
	go initReloader(fe)
	glog.Info("Initialized feature reloader")
}

func (fe *feature) Stop() {
	//updating serviceStop flag to true
	close(fe.serviceStop)
}

// Initializing reloader with cache-refresh default-expiry + 30 mins (to avoid DB load post cache refresh)
var initReloader = func(fe *feature) {
	if fe.defaultExpiry <= 0 {
		return
	}
	glog.Info("Feature reloader start")
	ticker := time.NewTicker(time.Duration(fe.defaultExpiry+1800) * time.Second)
	for {
		//Populating feature config maps from cache
		fe.updateFeatureConfigMaps()
		//update gdprCountryCodes
		fe.updateGDPRCountryCodes()
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

	if publisherFeatureMap != nil {
		fe.publisherFeature = publisherFeatureMap
	}

	if errFscUpdate = fe.updateFscConfigMapsFromCache(); errFscUpdate != nil {
		err = models.ErrorWrap(err, errFscUpdate)
	}

	fe.updateTBFConfigMap()
	fe.updateAmpMutiformatEnabledPublishers()
	fe.updateMaxFloorsEnabledPublishers()
	fe.updateAnalyticsThrottling()
	fe.updateBidRecoveryEnabledPublishers()
	fe.updateApplovinMultiFloorsFeature()
	fe.updateApplovinSchainABTestFeature()
	fe.updateImpCountingMethodEnabledBidders()
	fe.updateMBMF()
	fe.updateDynamicFloorEnabledPublishers()
	fe.updatePerformanceDSPs()
	fe.updateInViewEnabledPublishers()

	if err != nil {
		glog.Error(err.Error())
	}
}
