package openwrap

import (
	"github.com/prebid/openrtb/v19/openrtb2"
	cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	vastmodels "github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/models"
)

const (
	VastUnwrapperEnableValue = "1"
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

// GetVastUnwrapEnabled function return vastunwrap flag from the database
func GetVastUnwrapEnabled(rctx vastmodels.RequestCtx) bool {
	rCtx := models.RequestCtx{
		Endpoint:  rctx.Endpoint,
		PubID:     rctx.PubID,
		ProfileID: rctx.ProfileID,
		DisplayID: rctx.DisplayID,
	}
	partnerConfigMap, err := ow.getProfileData(rCtx, openrtb2.BidRequest{})
	if err != nil || len(partnerConfigMap) == 0 {
		return false
	}
	rCtx.PartnerConfigMap = partnerConfigMap
	return models.GetVersionLevelPropertyFromPartnerConfig(rCtx.PartnerConfigMap, models.VastUnwrapperEnableKey) == VastUnwrapperEnableValue
}
