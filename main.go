package prebidServer

import (
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/usersync"

	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/currency"
	pbc "github.com/prebid/prebid-server/prebid_cache_client"
	"github.com/prebid/prebid-server/router"
	"github.com/prebid/prebid-server/util/task"

	"github.com/spf13/viper"
)

func InitPrebidServer(configFile string) {
	cfg, err := loadConfig(configFile)
	if err != nil {
		glog.Exitf("Configuration could not be loaded or did not pass validation: %v", err)
	}

	err = serve(cfg)
	if err != nil {
		glog.Exitf("prebid-server failed: %v", err)
	}
}

func loadConfig(configFileName string) (*config.Configuration, error) {
	v := viper.New()
	config.SetupViper(v, configFileName)
	v.SetConfigFile(configFileName)
	v.ReadInConfig()
	return config.New(v)
}

func serve(cfg *config.Configuration) error {
	fetchingInterval := time.Duration(cfg.CurrencyConverter.FetchIntervalSeconds) * time.Second
	staleRatesThreshold := time.Duration(cfg.CurrencyConverter.StaleRatesSeconds) * time.Second
	currencyConverter := currency.NewRateConverter(&http.Client{}, cfg.CurrencyConverter.FetchURL, staleRatesThreshold)

	currencyConverterTickerTask := task.NewTickerTask(fetchingInterval, currencyConverter)
	currencyConverterTickerTask.Start()

	_, err := router.New(cfg, currencyConverter)
	if err != nil {
		return err
	}

	pbc.InitPrebidCache(cfg.CacheURL.GetBaseURL())

	corsRouter := router.SupportCORS(r)
	server.Listen(cfg, router.NoCache{Handler: corsRouter}, router.Admin(currencyConverter, fetchingInterval), r.MetricsEngine)

	r.Shutdown()
	return nil
}

func OrtbAuction(w http.ResponseWriter, r *http.Request) error {
	return router.OrtbAuctionEndpointWrapper(w, r)
}

var VideoAuction = func(w http.ResponseWriter, r *http.Request) error {
	return router.VideoAuctionEndpointWrapper(w, r)
}

func GetUIDS(w http.ResponseWriter, r *http.Request) {
	router.GetUIDSWrapper(w, r)
}

func SetUIDS(w http.ResponseWriter, r *http.Request) {
	router.SetUIDSWrapper(w, r)
}

func CookieSync(w http.ResponseWriter, r *http.Request) {
	router.CookieSync(w, r)
}

func SyncerMap() map[string]usersync.Syncer {
	return router.SyncerMap()
}
