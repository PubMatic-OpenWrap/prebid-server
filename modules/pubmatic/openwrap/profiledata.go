package openwrap

import (
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func (m OpenWrap) getProfileData(rCtx models.RequestCtx, bidRequest openrtb2.BidRequest) (map[int]map[string]string, error) {
	if rCtx.IsTestRequest == 2 { // skip db data for test=2
		//get platform from request, since test mode can be enabled for display and app platform only
		var platform string // TODO: should we've some default platform value
		if bidRequest.App != nil {
			platform = models.PLATFORM_APP
		}

		return getTestModePartnerConfigMap(platform, m.cfg.Timeout.HBTimeout, rCtx.DisplayID), nil
	}

	return m.cache.GetPartnerConfigMap(rCtx.PubID, rCtx.ProfileID, rCtx.DisplayID)
}
