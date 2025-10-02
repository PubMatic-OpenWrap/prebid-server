package openwrap

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"sync"

	vastunwrap "git.pubmatic.com/vastunwrap"
	"github.com/golang/glog"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/prebid-server/v3/currency"
	"github.com/prebid/prebid-server/v3/modules/moduledeps"
	ow_adapters "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adapters"
	cache "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache"
	ow_gocache "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache/gocache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/creativecache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/database/mysql"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/feature"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/geodb"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/geodb/netacuity"
	metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics"
	metrics_cfg "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics/config"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/profilemetadata"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/publisherfeature"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/unwrap"
	"github.com/prebid/prebid-server/v3/util/uuidutil"
)

const (
	CACHE_EXPIRY_ROUTINE_RUN_INTERVAL = 1 * time.Minute
)

type OpenWrap struct {
	cfg             config.Config
	cache           cache.Cache
	metricEngine    metrics.MetricsEngine
	rateConvertor   *currency.RateConverter
	geoInfoFetcher  geodb.Geography
	pubFeatures     publisherfeature.Feature
	unwrap          unwrap.Unwrap
	profileMetaData profilemetadata.ProfileMetaData
	uuidGenerator   uuidutil.UUIDGenerator
	features        feature.Features
	creativeCache   creativecache.Client
}

var ow *OpenWrap

func initOpenWrap(rawCfg json.RawMessage, moduleDeps moduledeps.ModuleDeps) (OpenWrap, error) {
	var once sync.Once
	cfg := config.Config{}

	err := json.Unmarshal(rawCfg, &cfg)
	if err != nil {
		return OpenWrap{}, fmt.Errorf("invalid openwrap config: %v", err)
	}
	err = patchConfig(&cfg)
	if err != nil {
		return OpenWrap{}, err
	}
	glog.Info("Connecting to OpenWrap database...")
	mysqlDriver, err := open("mysql", cfg.Database)
	if err != nil {
		return OpenWrap{}, fmt.Errorf("failed to open db connection: %v", err)
	}
	db := mysql.New(mysqlDriver, cfg.Database, cfg.Cache)

	// NYC_TODO: replace this with freecache and use concrete structure
	cache := gocache.New(time.Duration(cfg.Cache.CacheDefaultExpiry)*time.Second, CACHE_EXPIRY_ROUTINE_RUN_INTERVAL)
	if cache == nil {
		return OpenWrap{}, errors.New("error while initializing cache")
	}

	// NYC_TODO: remove this dependency
	if err := ow_adapters.InitBidders("./static/bidder-params"); err != nil {
		return OpenWrap{}, errors.New("error while initializing bidder params")
	}

	metricEngine, err := metrics_cfg.NewMetricsEngine(&cfg, &moduleDeps.Config.Metrics, moduleDeps.MetricsRegistry)
	if err != nil {
		return OpenWrap{}, fmt.Errorf("error while initializing metrics-engine: %v", err)
	}

	owCache := ow_gocache.New(cache, db, cfg.Cache, &metricEngine)

	// Init Feature reloader service
	pubFeatures := publisherfeature.New(publisherfeature.Config{
		Cache:                 owCache,
		DefaultExpiry:         cfg.Cache.CacheDefaultExpiry,
		AnalyticsThrottleList: cfg.Features.AnalyticsThrottlingPercentage,
	})
	pubFeatures.Start()

	// Init ProfileMetaData reloader service
	profileMetaData := profilemetadata.New(profilemetadata.Config{
		Cache:                 owCache,
		ProfileMetaDataExpiry: cfg.Cache.ProfileMetaDataCacheExpiry,
	})
	if err = profileMetaData.Start(); err != nil {
		glog.Error("Failed to load profileMetaData from DB")
		return OpenWrap{}, fmt.Errorf("error while initializing profile-metadata: %v", err)
	}
	glog.Info("Initialized profileMetaData reloader")

	// Features
	featureLoader := feature.NewFeatureLoader(mysqlDriver, cfg.Database)
	features := featureLoader.LoadFeatures()

	// Init VAST Unwrap
	vastunwrap.InitUnWrapperConfig(cfg.VastUnwrapCfg)
	uw := unwrap.NewUnwrap(fmt.Sprintf("http://%s:%d/unwrap", cfg.VastUnwrapCfg.APPConfig.Host, cfg.VastUnwrapCfg.APPConfig.Port),
		cfg.VastUnwrapCfg.APPConfig.UnwrapDefaultTimeout, nil, &metricEngine)

	// Creative Cache
	creativeCache := creativecache.NewClient(moduleDeps.CacheHttpClient, &moduleDeps.Config.CacheURL, &moduleDeps.Config.ExtCacheURL, metricEngine)

	initOpenWrapServer(&cfg)

	// init geoDBClient
	netacuityClient, err := netacuity.NewNetacuity(cfg.GeoDB.Location)
	if err != nil {
		return OpenWrap{}, fmt.Errorf("error initializing geoDB client host:[%s] err:[%v]", GetHostName(), err)
	}
	geodb.SetGeography(netacuityClient)

	once.Do(func() {
		ow = &OpenWrap{
			cfg:             cfg,
			cache:           owCache,
			metricEngine:    &metricEngine,
			rateConvertor:   moduleDeps.RateConvertor,
			geoInfoFetcher:  geodb.GetGeography(),
			pubFeatures:     pubFeatures,
			unwrap:          uw,
			profileMetaData: profileMetaData,
			uuidGenerator:   uuidutil.UUIDRandomGenerator{},
			features:        features,
			creativeCache:   creativeCache,
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

func patchConfig(cfg *config.Config) error {
	cfg.Server.HostName = GetHostName()

	models.TrackerCallWrapOMActive = strings.Replace(models.TrackerCallWrapOMActive, "${OMScript}", cfg.PixelView.OMScript, 1)
	sort.Strings(cfg.ResponseOverride.BidType)

	if len(cfg.VastUnwrapCfg.HTTPConfig.SSLCertificates) > 0 {
		decodedCert, err := base64.StdEncoding.DecodeString(cfg.VastUnwrapCfg.HTTPConfig.SSLCertificates)
		if err != nil {
			return fmt.Errorf("error decoding base64 SSL certificates: %v", err)
		}
		cfg.VastUnwrapCfg.HTTPConfig.SSLCertificates = string(decodedCert)
	}
	if len(cfg.VastUnwrapCfg.HTTPConfig.SSLKey) > 0 {
		decodedKey, err := base64.StdEncoding.DecodeString(cfg.VastUnwrapCfg.HTTPConfig.SSLKey)
		if err != nil {
			return fmt.Errorf("error decoding base64 SSL Key: %v", err)
		}
		cfg.VastUnwrapCfg.HTTPConfig.SSLKey = string(decodedKey)
	}
	return nil
}
