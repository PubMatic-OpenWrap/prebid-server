package openwrap

import (
	"github.com/prebid/openrtb/v19/openrtb2"
	cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	vastmodels "github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/models"
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

// GetVastUnwrapEnable
func GetVastUnwrapEnable(rctx vastmodels.RequestCtx) bool {
	rCtx := models.RequestCtx{}
	rCtx.Endpoint = rctx.Endpoint
	rCtx.PubID = rctx.PubID
	rCtx.ProfileID = rctx.ProfileID
	rCtx.DisplayID = rctx.DisplayID
	// rCtx.VersionID = rctx.VersionID
	partnerConfigMap, err := ow.getProfileData(rCtx, openrtb2.BidRequest{})
	if err != nil || len(partnerConfigMap) == 0 {
		return false
	}
	rCtx.PartnerConfigMap = partnerConfigMap
	if models.GetVersionLevelPropertyFromPartnerConfig(rCtx.PartnerConfigMap, models.VastUnwrapperEnableKey) == "1" {
		return true
	}
	return false
}
