package main_ow

import (
	"flag"
	"math/rand"
	"net/http"
	"path/filepath"
	"runtime"
	"time"

	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/currency"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/router"
	"github.com/prebid/prebid-server/v2/server"
	"github.com/prebid/prebid-server/v2/util/task"

	"github.com/golang/glog"
	"github.com/spf13/viper"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// TODO: revert this after PBS-OpenWrap module
func Main() {
	flag.Parse() // required for glog flags and testing package flags

	bidderInfoPath, err := filepath.Abs(infoDirectory)
	if err != nil {
		glog.Exitf("Unable to build configuration directory path: %v", err)
	}

	bidderInfos, err := config.LoadBidderInfoFromDisk(bidderInfoPath)
	if err != nil {
		glog.Exitf("Unable to load bidder configurations: %v", err)
	}
	cfg, err := loadConfig(bidderInfos)
	if err != nil {
		glog.Exitf("Configuration could not be loaded or did not pass validation: %v", err)
	}

	// Create a soft memory limit on the total amount of memory that PBS uses to tune the behavior
	// of the Go garbage collector. In summary, `cfg.GarbageCollectorThreshold` serves as a fixed cost
	// of memory that is going to be held garbage before a garbage collection cycle is triggered.
	// This amount of virtual memory won’t translate into physical memory allocation unless we attempt
	// to read or write to the slice below, which PBS will not do.
	garbageCollectionThreshold := make([]byte, cfg.GarbageCollectorThreshold)
	defer runtime.KeepAlive(garbageCollectionThreshold)

	err = serve(cfg)
	if err != nil {
		glog.Exitf("prebid-server failed: %v", err)
	}
}

const configFileName = "pbs.yaml"
const infoDirectory = "./static/bidder-info"

func loadConfig(bidderInfos config.BidderInfos) (*config.Configuration, error) {
	v := viper.New()
	config.SetupViper(v, configFileName, bidderInfos)
	return config.New(v, bidderInfos, openrtb_ext.NormalizeBidderName)
}

func serve(cfg *config.Configuration) error {
	fetchingInterval := time.Duration(cfg.CurrencyConverter.FetchIntervalSeconds) * time.Second
	staleRatesThreshold := time.Duration(cfg.CurrencyConverter.StaleRatesSeconds) * time.Second
	currencyConverter := currency.NewRateConverter(&http.Client{}, cfg.CurrencyConverter.FetchURL, staleRatesThreshold)

	currencyConverterTickerTask := task.NewTickerTask(fetchingInterval, currencyConverter)
	currencyConverterTickerTask.Start()

	r, err := router.New(cfg, currencyConverter)
	if err != nil {
		return err
	}

	corsRouter := router.SupportCORS(r)
	server.Listen(cfg, router.NoCache{Handler: corsRouter}, router.Admin(currencyConverter, fetchingInterval), r.MetricsEngine)

	r.Shutdown()
	return nil
}
