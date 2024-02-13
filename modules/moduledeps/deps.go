package moduledeps

import (
	"net/http"

	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/currency"
	metricsCfg "github.com/prebid/prebid-server/v2/metrics/config"
)

// ModuleDeps provides dependencies that custom modules may need for hooks execution.
// Additional dependencies can be added here if modules need something more.
type ModuleDeps struct {
	HTTPClient      *http.Client
	RateConvertor   *currency.RateConverter
	MetricsCfg      *config.Metrics
	MetricsRegistry metricsCfg.MetricsRegistry
}
