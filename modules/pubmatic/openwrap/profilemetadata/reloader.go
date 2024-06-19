package profilemetadata

import (
	"fmt"
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
	failToLoadDBData      chan bool
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
			failToLoadDBData:      make(chan bool),
			profileMetaDataExpiry: config.ProfileMetaDataExpiry,
			profileTypePlatform:   make(map[string]int),
			appIntegrationPath:    make(map[string]int),
			appSubIntegrationPath: make(map[string]int),
		}
	})
	return pmd
}

func (pmd *profileMetaData) Start() error {
	go initReloader(pmd)
	//Waiting for the  profileMetaData to load from DB
	if <-pmd.failToLoadDBData {
		glog.Error("Failed to load profileMetaData")
		return fmt.Errorf("failed to load profileMetaData")
	}
	glog.Info("Initialized profileMetaData reloader")
	return nil
}

func (pmd *profileMetaData) Stop() {
	//updating serviceStop flag to true
	close(pmd.serviceStop)
}

// Initializing reloader with cache-refresh (to avoid DB load post cache refresh)
var initReloader = func(pmd *profileMetaData) {
	firstdbLoad := true
	if pmd.profileMetaDataExpiry <= 0 {
		return
	}
	glog.Info("profileMetaData reloader start")
	ticker := time.NewTicker(time.Duration(pmd.profileMetaDataExpiry) * time.Second)
	for {
		//Populating pmdata config maps from cache (if data is not loaded from DB for first instance then do not start the service)
		if err := pmd.updateProfileMetaDataMaps(); err != nil && firstdbLoad {
			pmd.failToLoadDBData <- true
			return
		} else {
			firstdbLoad = false
			pmd.failToLoadDBData <- false
		}
		select {
		case t := <-ticker.C:
			glog.Info("profileMetaData Reloader loads cache @", t)
		case <-pmd.serviceStop:
			return
		}
	}
}

func (pmd *profileMetaData) updateProfileMetaDataMaps() error {
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
	return err
}
