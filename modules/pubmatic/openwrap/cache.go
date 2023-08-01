package openwrap

import (
	"strconv"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

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
