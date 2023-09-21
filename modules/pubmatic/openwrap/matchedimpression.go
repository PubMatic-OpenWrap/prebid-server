package openwrap

import (
	"encoding/json"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/usersync"
)

func getMatchedImpression(rctx models.RequestCtx) json.RawMessage {
	var parsed *usersync.Cookie
	if rctx.UidCookie == nil {
		parsed = usersync.NewCookie()
	} else {
		parsed = usersync.ParseCookie(rctx.UidCookie)
	}

	cookieFlagMap := make(map[string]int)
	for _, partnerConfig := range rctx.PartnerConfigMap { // TODO: original code deos not handle throttled partners
		if partnerConfig[models.SERVER_SIDE_FLAG] != "1" {
			continue
		}

		partnerName := partnerConfig[models.PREBID_PARTNER_NAME]

		syncerCode := adapters.ResolveOWBidder(partnerName)

		status := 0
		if uid, _, _ := parsed.GetUID(syncerCode); uid != "" {
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
