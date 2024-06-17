package profilemetadata

import (
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

type Config struct {
	Cache                 cache.Cache
	ProfileMetaDataExpiry int
}

type profileMetaData struct {
	sync.RWMutex
	cache                 cache.Cache
	serviceStop           chan struct{}
	profileMetaDataExpiry int
	profileTypePlatform   map[string]int
	appIntegrationPath    map[string]int
	appSubIntegrationPath map[string]int
}

var pmd *profileMetaData
var pOnce sync.Once

func New(config Config) *profileMetaData {
	pOnce.Do(func() {
		pmd = &profileMetaData{
			cache:                 config.Cache,
			serviceStop:           make(chan struct{}),
			profileMetaDataExpiry: config.ProfileMetaDataExpiry,
			profileTypePlatform:   make(map[string]int),
			appIntegrationPath:    make(map[string]int),
			appSubIntegrationPath: make(map[string]int),
		}
	})
	return pmd
}

func (pmd *profileMetaData) Start() {
	go initReloader(pmd)
	glog.Info("Initialized profileMetaData reloader")
}

func (pmd *profileMetaData) Stop() {
	//updating serviceStop flag to true
	close(pmd.serviceStop)
}

// Initializing reloader with cache-refresh (to avoid DB load post cache refresh)
var initReloader = func(pmd *profileMetaData) {
	if pmd.profileMetaDataExpiry <= 0 {
		return
	}
	glog.Info("profileMetaData reloader start")
	ticker := time.NewTicker(time.Duration(pmd.profileMetaDataExpiry) * time.Second)
	for {
		//Populating pmdature config maps from cache
		pmd.updateProfileMetaDataMaps()
		select {
		case t := <-ticker.C:
			glog.Info("profileMetaData Reloader loads cache @", t)
		case <-pmd.serviceStop:
			return
		}
	}
}

func (pmd *profileMetaData) updateProfileMetaDataMaps() {
	var err error
	profileTypePlatfrom, errProfileTypePlatforms := pmd.cache.GetProfileTypePlatforms()
	if errProfileTypePlatforms != nil {
		err = models.ErrorWrap(err, errProfileTypePlatforms)
	} else {
		pmd.Lock()
		pmd.profileTypePlatform = profileTypePlatfrom
		pmd.Unlock()
	}

	appIntegrationPath, errAppIntegrationPath := pmd.cache.GetAppIntegrationPaths()
	if errAppIntegrationPath != nil {
		err = models.ErrorWrap(err, errAppIntegrationPath)
	} else {
		pmd.Lock()
		pmd.appIntegrationPath = appIntegrationPath
		pmd.Unlock()
	}

	appSubIntegrationPath, errAppSubIntegrationPath := pmd.cache.GetAppSubIntegrationPaths()
	if errAppSubIntegrationPath != nil {
		err = models.ErrorWrap(err, errAppSubIntegrationPath)
	} else {
		pmd.Lock()
		pmd.appSubIntegrationPath = appSubIntegrationPath
		pmd.Unlock()
	}

	if err != nil {
		glog.Error(err.Error())
	}
}
