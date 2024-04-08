package openwrap

import (
	"strconv"

	"github.com/prebid/openrtb/v20/openrtb2"
	cache "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/config"
	metrics "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/publisherfeature"
	vastmodels "github.com/prebid/prebid-server/v2/modules/pubmatic/vastunwrap/models"
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

// GetFeature Temporary function to expose feature to SSHB
func (ow *OpenWrap) GetFeature() publisherfeature.Feature {
	return ow.featureConfig
}

// GetVastUnwrapEnabled return whether to enable vastunwrap or not
func GetVastUnwrapEnabled(rctx vastmodels.RequestCtx, VASTUnwrapTraffic int) bool {
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
	trafficPercentage := VASTUnwrapTraffic
	unwrapEnabled := models.GetVersionLevelPropertyFromPartnerConfig(rCtx.PartnerConfigMap, models.VastUnwrapperEnableKey) == VastUnwrapperEnableValue
	if unwrapEnabled {
		if value := models.GetVersionLevelPropertyFromPartnerConfig(rCtx.PartnerConfigMap, models.VastUnwrapTrafficPercentKey); len(value) > 0 {
			if trafficPercentDB, err := strconv.Atoi(value); err == nil {
				trafficPercentage = trafficPercentDB
			}
		}
	}
	return unwrapEnabled && GetRandomNumberIn1To100() <= trafficPercentage
}
