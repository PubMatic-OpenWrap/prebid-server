package openwrap

import (
	"strconv"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

func (m OpenWrap) getProfileData(rCtx models.RequestCtx, bidRequest openrtb2.BidRequest) (map[int]map[string]string, error) {
	if bidRequest.Test == 2 { // skip db data for test=2
		//get platform from request, since test mode can be enabled for display and app platform only
		var platform string // TODO: should we've some default platform value
		if bidRequest.App != nil {
			platform = models.PLATFORM_APP
		}

		return getTestModePartnerConfigMap(platform, m.cfg.Timeout.HBTimeout, rCtx.DisplayID), nil
	}

	return m.cache.GetPartnerConfigMap(rCtx.PubID, rCtx.ProfileID, rCtx.DisplayID)
}

func getTestModePartnerConfigMap(platform string, timeout int64, displayVersion int) map[int]map[string]string {
	return map[int]map[string]string{
		1: {
			models.PARTNER_ID:          models.PUBMATIC_PARTNER_ID_STRING,
			models.PREBID_PARTNER_NAME: string(openrtb_ext.BidderPubmatic),
			models.BidderCode:          string(openrtb_ext.BidderPubmatic),
			models.SERVER_SIDE_FLAG:    models.PUBMATIC_SS_FLAG,
			models.KEY_GEN_PATTERN:     models.ADUNIT_SIZE_KGP,
			models.TIMEOUT:             strconv.Itoa(int(timeout)),
		},
		-1: {
			models.PLATFORM_KEY:     platform,
			models.DisplayVersionID: strconv.Itoa(displayVersion),
		},
	}
}
