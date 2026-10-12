package moduledeps

import (
	"net/http"

	"github.com/prebid/prebid-server/v4/config"
	"github.com/prebid/prebid-server/v4/currency"
	metricsCfg "github.com/prebid/prebid-server/v4/metrics/config"
)

// ModuleDeps provides dependencies that custom modules may need for hooks execution.
// Additional dependencies can be added here if modules need something more.
type ModuleDeps struct {
	HTTPClient      *http.Client
	RateConvertor   *currency.RateConverter
	Geoscope        map[string][]string
	MetricsCfg      *config.Metrics
	MetricsRegistry metricsCfg.MetricsRegistry
}
