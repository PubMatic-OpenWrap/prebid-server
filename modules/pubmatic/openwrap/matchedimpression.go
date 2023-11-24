package openwrap

import (
	"encoding/json"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

func getMatchedImpression(rctx models.RequestCtx) json.RawMessage {

	cookieFlagMap := make(map[string]int)
	for _, partnerConfig := range rctx.PartnerConfigMap { // TODO: original code deos not handle throttled partners
		if partnerConfig[models.SERVER_SIDE_FLAG] != "1" {
			continue
		}

		partnerName := partnerConfig[models.PREBID_PARTNER_NAME]

		syncerCode := adapters.ResolveOWBidder(partnerName)

		status := 0
		if uid, _, _ := rctx.ParsedUidCookie.GetUID(syncerCode); uid != "" {
			status = 1
		}
		cookieFlagMap[partnerConfig[models.BidderCode]] = status
	}

	matchedImpression, err := json.Marshal(cookieFlagMap)
	if err != nil {
		return nil
	}

	return json.RawMessage(matchedImpression)
}
