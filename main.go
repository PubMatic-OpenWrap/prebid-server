package main

import (
	"flag"
	"math/rand"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/currency"
	pbc "github.com/prebid/prebid-server/prebid_cache_client"
	"github.com/prebid/prebid-server/router"
	"github.com/prebid/prebid-server/server"
	"github.com/prebid/prebid-server/util/task"

	"github.com/golang/glog"
	"github.com/spf13/viper"
)

// Rev holds binary revision string
// Set manually at build time using:
//    go build -ldflags "-X main.Rev=`git rev-parse --short HEAD`"
// Populated automatically at build / releases
//   `gox -os="linux" -arch="386" -output="{{.Dir}}_{{.OS}}_{{.Arch}}" -ldflags "-X main.Rev=`git rev-parse --short HEAD`" -verbose ./...;`
// See issue #559
var Rev string

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	flag.Parse() // required for glog flags and testing package flags

	cfg, err := loadConfig()
	if err != nil {
		glog.Exitf("Configuration could not be loaded or did not pass validation: %v", err)
	}

	err = serve(Rev, cfg)
	if err != nil {
		glog.Exitf("prebid-server failed: %v", err)
	}
}

func InitPrebidServer(configFile string) {
	//init contents
	rand.Seed(time.Now().UnixNano())

	//main contents
	cfg, err := loadConfig(configFile)
	if err != nil {
		glog.Fatalf("Configuration could not be loaded or did not pass validation: %v", err)
	}

	err = serve(Rev, cfg)
	if err != nil {
		glog.Errorf("prebid-server failed: %v", err)
	}
}

//const configFileName = "pbs"

func loadConfig(configFileName string) (*config.Configuration, error) {
	v := viper.New()
	config.SetupViper(v, configFileName)
	v.SetConfigFile(configFileName)
	v.ReadInConfig()
	return config.New(v)
}

func serve(revision string, cfg *config.Configuration) error {
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
	server.Listen(cfg, router.NoCache{Handler: corsRouter}, router.Admin(revision, currencyConverter, fetchingInterval), r.MetricsEngine)

	r.Shutdown()
	return nil
}

func OrtbAuction(w http.ResponseWriter, r *http.Request) error {
	return router.OrtbAuctionEndpointWrapper(w, r)
}

func VideoAuction(w http.ResponseWriter, r *http.Request) error {
	return router.VideoAuctionEndpointWrapper(w, r)
}

func Auction(w http.ResponseWriter, r *http.Request) {
	router.AuctionWrapper(w, r)
}

func GetUIDS(w http.ResponseWriter, r *http.Request) {
	router.GetUIDSWrapper(w, r)
}

func SetUIDS(w http.ResponseWriter, r *http.Request) {
	router.SetUIDSWrapper(w, r)
}

func CookieSync(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	router.CookieSync(w, r)
}

func SyncerMap() map[openrtb_ext.BidderName]usersync.Usersyncer {
	return router.SyncerMap()
}
