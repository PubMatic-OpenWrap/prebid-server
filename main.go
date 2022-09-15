package main

import (
	"context"
	"flag"
	"math/rand"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/account"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/currency"
	"github.com/prebid/prebid-server/router"
	"github.com/prebid/prebid-server/server"
	storedRequestsConf "github.com/prebid/prebid-server/stored_requests/config"
	"github.com/prebid/prebid-server/util/task"
	"github.com/spf13/viper"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var StartTime = time.Now()

func main() {
	flag.Parse() // required for glog flags and testing package flags
	flag.Lookup("log_dir").Value.Set("/tmp")
	cfg, err := loadConfig()
	if err != nil {
		glog.Exitf("Configuration could not be loaded or did not pass validation: %v", err)
	}

	// // Create a soft memory limit on the total amount of memory that PBS uses to tune the behavior
	// // of the Go garbage collector. In summary, `cfg.GarbageCollectorThreshold` serves as a fixed cost
	// // of memory that is going to be held garbage before a garbage collection cycle is triggered.
	// // This amount of virtual memory won’t translate into physical memory allocation unless we attempt
	// // to read or write to the slice below, which PBS will not do.
	// garbageCollectionThreshold := make([]byte, cfg.GarbageCollectorThreshold)
	// defer runtime.KeepAlive(garbageCollectionThreshold)

	// err = serve(cfg)
	// if err != nil {
	// 	glog.Exitf("prebid-server failed: %v", err)
	// }

	cfg.Accounts = config.StoredRequests{
		Files: config.FileFetcherConfig{
			Enabled: true,
			Path:    "/Users/test/prebid-server/stored_requests/backends/file_fetcher/test",
		},
	}

	cfg.Accounts.SetDataType(config.AccountDataType)
	cfg.Accounts.InMemoryCache = config.InMemoryCache{
		Type:             "lru",
		TTL:              10,
		Size:             1000000000,
		RequestCacheSize: 1000000000,
		ImpCacheSize:     10000000000,
		RespCacheSize:    10000000000,
	}

	fetcher, _ := storedRequestsConf.CreateStoredRequests(&cfg.Accounts, nil, nil, nil, nil)
	account.GetAccount(context.Background(), cfg, fetcher, "test")

}

const configFileName = "pbs"

func loadConfig() (*config.Configuration, error) {
	v := viper.New()
	config.SetupViper(v, configFileName)
	return config.New(v)
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
