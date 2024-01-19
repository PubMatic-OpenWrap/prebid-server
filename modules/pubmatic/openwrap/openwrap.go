package openwrap

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"sync"

	"github.com/golang/glog"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/prebid-server/currency"
	"github.com/prebid/prebid-server/modules/moduledeps"
	ow_adapters "github.com/prebid/prebid-server/modules/pubmatic/openwrap/adapters"
	cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	ow_gocache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache/gocache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/database/mysql"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/fullscreenclickability"
	metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics"
	metrics_cfg "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/tbf"
)

const (
	CACHE_EXPIRY_ROUTINE_RUN_INTERVAL = 1 * time.Minute
)

type OpenWrap struct {
	cfg                config.Config
	cache              cache.Cache
	metricEngine       metrics.MetricsEngine
	currencyConversion currency.Conversions
}

var ow *OpenWrap

func initOpenWrap(rawCfg json.RawMessage, moduleDeps moduledeps.ModuleDeps) (OpenWrap, error) {
	var once sync.Once
	cfg := config.Config{}

	err := json.Unmarshal(rawCfg, &cfg)
	if err != nil {
		return OpenWrap{}, fmt.Errorf("invalid openwrap config: %v", err)
	}
	patchConfig(&cfg)

	glog.Info("Connecting to OpenWrap database...")
	mysqlDriver, err := open("mysql", cfg.Database)
	if err != nil {
		return OpenWrap{}, fmt.Errorf("failed to open db connection: %v", err)
	}
	db := mysql.New(mysqlDriver, cfg.Database)

	// NYC_TODO: replace this with freecache and use concrete structure
	cache := gocache.New(time.Duration(cfg.Cache.CacheDefaultExpiry)*time.Second, CACHE_EXPIRY_ROUTINE_RUN_INTERVAL)
	if cache == nil {
		return OpenWrap{}, errors.New("error while initializing cache")
	}

	// NYC_TODO: remove this dependency
	if err := ow_adapters.InitBidders("./static/bidder-params"); err != nil {
		return OpenWrap{}, errors.New("error while initializing bidder params")
	}

	metricEngine, err := metrics_cfg.NewMetricsEngine(&cfg, moduleDeps.MetricsCfg, moduleDeps.MetricsRegistry)
	if err != nil {
		return OpenWrap{}, fmt.Errorf("error while initializing metrics-engine: %v", err)
	}

	owCache := ow_gocache.New(cache, db, cfg.Cache, &metricEngine)

	// Init FSC and related services
	fullscreenclickability.Init(owCache, cfg.Cache.CacheDefaultExpiry)

	// Init TBF (tracking-beacon-first) feature related services
	tbf.Init(cfg.Cache.CacheDefaultExpiry, owCache)

	once.Do(func() {
		ow = &OpenWrap{
			cfg:                cfg,
			cache:              owCache,
			metricEngine:       &metricEngine,
			currencyConversion: moduleDeps.CurrencyConversion,
		}
	})
	return *ow, nil
}

func open(driverName string, cfg config.Database) (*sql.DB, error) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.Database)

	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(cfg.IdleConnection)
	db.SetMaxOpenConns(cfg.MaxConnection)
	db.SetConnMaxLifetime(time.Second * time.Duration(cfg.ConnMaxLifeTime))

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func patchConfig(cfg *config.Config) {
	cfg.Server.HostName = GetHostName()
	models.TrackerCallWrapOMActive = strings.Replace(models.TrackerCallWrapOMActive, "${OMScript}", cfg.PixelView.OMScript, 1)
}
