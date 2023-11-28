package openwrap

import (
	cache "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/config"
	metrics "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics"
)

// GetConfig Temporary function to expose config to SSHB
func (ow OpenWrap) GetConfig() config.Config {
	return ow.cfg

}

// GetCache Temporary function to expose cache to SSHB
func (ow OpenWrap) GetCache() cache.Cache {
	return ow.cache
}

// GetMetricEngine Temporary function to expose mertics to SSHB
func (ow OpenWrap) GetMetricEngine() metrics.MetricsEngine {
	return ow.metricEngine
}

// SetConfig Temporary function to expose config to SSHB
func (ow *OpenWrap) SetConfig(c config.Config) {
	ow.cfg = c
}

// GetCache Temporary function to expose cache to SSHB
func (ow *OpenWrap) SetCache(c cache.Cache) {
	ow.cache = c
}

// GetMetricEngine Temporary function to expose mertics to SSHB
func (ow *OpenWrap) SetMetricEngine(m metrics.MetricsEngine) {
	ow.metricEngine = m
}
