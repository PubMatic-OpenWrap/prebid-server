package main

import (
	"flag"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/currency"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/router"
	"github.com/prebid/prebid-server/server"
	"github.com/prebid/prebid-server/util/task"
	"github.com/pyroscope-io/client/pyroscope"

	"github.com/golang/glog"
	"github.com/spf13/viper"

	_ "net/http/pprof"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	flag.Parse() // required for glog flags and testing package flags

	hostname, _ := os.Hostname()

	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)
	pyroscope.Start(pyroscope.Config{
		ApplicationName: "nilesh.owpbsmodule.golang.app",

		// replace this with the address of pyroscope server
		ServerAddress: "https://ingest.pyroscope.cloud",

		// you can disable logging by setting this to nil
		// Logger: pyroscope.StandardLogger,
		Logger: nil,

		// optionally, if authentication is enabled, specify the API key:
		// AuthToken:    os.Getenv("PYROSCOPE_AUTH_TOKEN"),
		AuthToken: "psx-5NHb177FyMgtyzWqyDOkDjo-_kl0A7z7rTxr7LW9TZ_izUcqTPiMG_0",

		// you can provide static tags via a map:
		Tags: map[string]string{"hostname": hostname},

		ProfileTypes: []pyroscope.ProfileType{
			// these profile types are enabled by default:
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,

			// these profile types are optional:
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})

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
