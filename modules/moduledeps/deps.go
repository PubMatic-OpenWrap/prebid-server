package moduledeps

import (
	"net/http"

	"github.com/prebid/prebid-server/v3/config"
	"github.com/prebid/prebid-server/v3/currency"
	metricsCfg "github.com/prebid/prebid-server/v3/metrics/config"
)

// ModuleDeps provides dependencies that custom modules may need for hooks execution.
// Additional dependencies can be added here if modules need something more.
type ModuleDeps struct {
	HTTPClient      *http.Client
	CacheHttpClient *http.Client
	Config          *config.Configuration
	RateConvertor   *currency.RateConverter
	MetricsRegistry metricsCfg.MetricsRegistry
}
