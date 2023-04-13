package openwrap

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/prebid-server/modules/moduledeps"
	ow_adapters "github.com/prebid/prebid-server/modules/pubmatic/openwrap/adapters"
	ow_cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	ow_gocache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache/gocache"
	ow_config "github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/database/mysql"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

const (
	CACHE_EXPIRY_ROUTINE_RUN_INTERVAL = 1 * time.Minute
)

type OpenWrap struct {
	cfg   ow_config.SSHB
	cache ow_cache.Cache
}

func initOpenWrap(rawCfg json.RawMessage, _ moduledeps.ModuleDeps) (OpenWrap, error) {
	cfg := ow_config.SSHB{}

	err := json.Unmarshal(rawCfg, &cfg)
	if err != nil {
		return OpenWrap{}, fmt.Errorf("invalid openwrap config: %v", err)
	}
	patchConfig(&cfg)

	mysqlDriver, err := open("mysql", cfg.OpenWrap.Database)
	if err != nil {
		return OpenWrap{}, fmt.Errorf("failed to open db connection: %v", err)
	}
	db := mysql.New(mysqlDriver, cfg.OpenWrap.Database)

	// NYC_TODO: replace this with freecache and use concrete structure
	cache := gocache.New(time.Duration(cfg.OpenWrap.Cache.CacheDefaultExpiry)*time.Second, CACHE_EXPIRY_ROUTINE_RUN_INTERVAL)
	if cache == nil {
		return OpenWrap{}, errors.New("error while initializing cache")
	}

	// NYC_TODO: remove this dependency
	if err := ow_adapters.InitBidders(cfg); err != nil {
		return OpenWrap{}, errors.New("error while initializing bidder params")
	}

	return OpenWrap{
		cfg:   cfg,
		cache: ow_gocache.New(cache, db, cfg.OpenWrap.Cache),
	}, nil
}

func open(driverName string, cfg ow_config.Database) (*sql.DB, error) {
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

func patchConfig(cfg *ow_config.SSHB) {
	cfg.OpenWrap.Server.HostName = getHostName()
	models.TrackerCallWrapOMActive = strings.Replace(models.TrackerCallWrapOMActive, "${OMScript}", cfg.OpenWrap.Pixelview.OMScript, 1)
	if cfg.OpenWrap.Stats.DefaultHostName == "" {
		cfg.OpenWrap.Stats.DefaultHostName = "N:P"
	}
}
