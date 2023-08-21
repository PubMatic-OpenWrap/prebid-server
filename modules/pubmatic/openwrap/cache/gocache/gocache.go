package gocache

import (
	"fmt"
	"sync"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/database"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics"
)

const (
	PUB_SLOT_INFO  = "pslot_%d_%d_%d_%d" // publisher slot mapping at publisher, profile, display version and adapter level
	PUB_HB_PARTNER = "hbplist_%d_%d_%d"  // header bidding partner list at publishr,profile, display version level
	//HB_PARTNER_CFG = "hbpcfg_%d"         // header bidding partner configuration at partner level
	//PubAadunitConfig - this key for storing adunit config at pub, profile and version level
	PubAdunitConfig = "aucfg_%d_%d_%d"
	PubSlotHashInfo = "pshash_%d_%d_%d_%d"     // slot and its hash info at publisher, profile, display version and adapter level
	PubSlotRegex    = "psregex_%d_%d_%d_%d_%s" // slot and its matching regex info at publisher, profile, display version and adapter level
	PubSlotNameHash = "pslotnamehash_%d"       //publisher slotname hash mapping cache key
	PubVASTTags     = "pvasttags_%d"           //publisher level vasttags
)

func key(format string, v ...interface{}) string {
	return fmt.Sprintf(format, v...)
}

// NYC_TODO: refactor this to inject any kind of cache,replace cache with freecache library
// any db or cache should be injectable
type cache struct {
	sync.Map
	cache        *gocache.Cache
	cfg          config.Cache
	db           database.Database
	metricEngine metrics.MetricsEngine
}

var c *cache
var cOnce sync.Once

func New(goCache *gocache.Cache, database database.Database, cfg config.Cache, metricEngine metrics.MetricsEngine) *cache {
	cOnce.Do(
		func() {
			c = &cache{
				cache:        goCache,
				db:           database,
				cfg:          cfg,
				metricEngine: metricEngine,
			}
		})
	return c
}

func getSeconds(duration int64) time.Duration {
	return time.Duration(duration) * time.Second
}

func (c *cache) Set(key string, value interface{}) {
	c.cache.Set(key, value, getSeconds(c.cfg.CacheDefaultExpiry))
}

func (c *cache) Get(key string) (interface{}, bool) {
	return c.cache.Get(key)
}
