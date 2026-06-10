package openwrap

import (
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/eds"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/sdk/signal"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/sdk/sdkutils"
)

// ResolveAtEntrypoint resolves EDS from SDK signal ext.eds for auction integrations.
func ResolveAtEntrypoint(endpoint string, originalBody []byte) models.ResolvedEds {
	if !sdkutils.IsSdkIntegration(endpoint) || len(originalBody) == 0 {
		return models.ResolvedEds{}
	}

	signalReq := signal.ParseForEndpoint(endpoint, originalBody)
	if signalReq == nil {
		return models.ResolvedEds{}
	}

	return eds.Resolve(eds.Sources{Signal: signalReq})
}
