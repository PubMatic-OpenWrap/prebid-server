package profilemetadata

import (
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache"
)

type Config struct {
	Cache         cache.Cache
	DefaultExpiry int
}

type profileMetaData struct {
	cache       cache.Cache
	serviceStop chan struct{}
	sync.RWMutex
	defaultExpiry         int
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
			defaultExpiry:         config.DefaultExpiry,
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

// Initializing reloader with cache-refresh default-expiry + 23 Hrs (to avoid DB load post cache refresh)
var initReloader = func(pmd *profileMetaData) {
	if pmd.defaultExpiry <= 0 {
		return
	}
	glog.Info("profileMetaData reloader start")
	ticker := time.NewTicker(time.Duration(pmd.defaultExpiry+(23*60*60)) * time.Second)
	for {
		//Populating pmdature config maps from cache
		pmd.updateProfileMetadaMaps()
		select {
		case t := <-ticker.C:
			glog.Info("profileMetaData Reloader loads cache @", t)
		case <-pmd.serviceStop:
			return
		}
	}
}

func (pmd *profileMetaData) updateProfileMetadaMaps() {
	if profileTypePlatfrom, err := pmd.cache.GetProfileTypePlatform(); err == nil {
		pmd.Lock()
		pmd.profileTypePlatform = profileTypePlatfrom
		pmd.Unlock()
	}

	if appIntegrationPath, err := pmd.cache.GetAppIntegrationPath(); err == nil {
		pmd.Lock()
		pmd.appIntegrationPath = appIntegrationPath
		pmd.Unlock()
	}

	if AppSubIntegrationPath, err := pmd.cache.GetAppSubIntegrationPath(); err == nil {
		pmd.Lock()
		pmd.appSubIntegrationPath = AppSubIntegrationPath
		pmd.Unlock()
	}
}
