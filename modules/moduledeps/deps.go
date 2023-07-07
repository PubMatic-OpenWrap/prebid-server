package moduledeps

import (
	"net/http"

	"github.com/prebid/prebid-server/config"
	"github.com/prometheus/client_golang/prometheus"
)

// ModuleDeps provides dependencies that custom modules may need for hooks execution.
// Additional dependencies can be added here if modules need something more.
type ModuleDeps struct {
	HTTPClient        *http.Client
	PrometheusMetrics config.PrometheusMetrics
	Registry          *prometheus.Registry
}
