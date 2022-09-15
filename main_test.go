package main

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/account"
	"github.com/prebid/prebid-server/config"
	metricsConf "github.com/prebid/prebid-server/metrics/config"
	"github.com/prebid/prebid-server/openrtb_ext"
	storedRequestsConf "github.com/prebid/prebid-server/stored_requests/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func compareStrings(t *testing.T, message string, expect string, actual string) {
	if expect != actual {
		t.Errorf(message, expect, actual)
	}
}

// forceEnv sets an environment variable to a certain value, and return a deferable function to reset it to the original value.
func forceEnv(t *testing.T, key string, val string) func() {
	orig, set := os.LookupEnv(key)
	err := os.Setenv(key, val)
	if err != nil {
		t.Fatalf("Error setting environment %s", key)
	}
	if set {
		return func() {
			if os.Setenv(key, orig) != nil {
				t.Fatalf("Error unsetting environment %s", key)
			}
		}
	}
	return func() {
		if os.Unsetenv(key) != nil {
			t.Fatalf("Error unsetting environment %s", key)
		}
	}
}

// Test the viper setup
func TestViperInit(t *testing.T) {
	v := viper.New()
	config.SetupViper(v, "")
	compareStrings(t, "Viper error: external_url expected to be %s, found %s", "http://localhost:8000", v.Get("external_url").(string))
	compareStrings(t, "Viper error: adapters.pulsepoint.endpoint expected to be %s, found %s", "http://bid.contextweb.com/header/s/ortb/prebid-s2s", v.Get("adapters.pulsepoint.endpoint").(string))
}

func TestViperEnv(t *testing.T) {
	v := viper.New()
	config.SetupViper(v, "")
	port := forceEnv(t, "PBS_PORT", "7777")
	defer port()

	endpt := forceEnv(t, "PBS_ADAPTERS_PUBMATIC_ENDPOINT", "not_an_endpoint")
	defer endpt()

	ttl := forceEnv(t, "PBS_HOST_COOKIE_TTL_DAYS", "60")
	defer ttl()

	ipv4Networks := forceEnv(t, "PBS_REQUEST_VALIDATION_IPV4_PRIVATE_NETWORKS", "1.1.1.1/24 2.2.2.2/24")
	defer ipv4Networks()

	assert.Equal(t, 7777, v.Get("port"), "Basic Config")
	assert.Equal(t, "not_an_endpoint", v.Get("adapters.pubmatic.endpoint"), "Nested Config")
	assert.Equal(t, 60, v.Get("host_cookie.ttl_days"), "Config With Underscores")
	assert.ElementsMatch(t, []string{"1.1.1.1/24", "2.2.2.2/24"}, v.Get("request_validation.ipv4_private_networks"), "Arrays")
}

func BenchmarkXxx(b *testing.B) {
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

	fetcher, _ := storedRequestsConf.CreateStoredRequests(&cfg.Accounts, metricsConf.NewMetricsEngine(cfg, openrtb_ext.CoreBidderNames(), []string{}), nil, nil, nil)

	b.Run("New Changes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			account.GetAccount(context.Background(), cfg, fetcher, "test")
		}
	})

}
