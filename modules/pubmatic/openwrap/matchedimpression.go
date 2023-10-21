package openwrap

import (
	"encoding/json"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
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

// TODO : check if we can replace getMatchedImpression & RecordPublisherPartnerNoCookieStats with this one
func getMatchedImpressionWithParsedCookie(rctx models.RequestCtx) map[string]int {

	cookieFlagMap := make(map[string]int)
	for _, partnerConfig := range rctx.PartnerConfigMap {
		if partnerConfig[models.SERVER_SIDE_FLAG] != "1" {
			continue
		}

		syncerMap := models.SyncerMap
		partnerName := partnerConfig[models.PREBID_PARTNER_NAME]

		syncerCode := adapters.ResolveOWBidder(partnerName)

		matchedImpression := 0

		syncer := syncerMap[syncerCode]
		if syncer == nil {
			glog.V(3).Infof("Invalid bidder code passed to ParseRequestCookies: %s ", partnerName)
		} else {
			uid, _, _ := rctx.ParsedUidCookie.GetUID(syncer.Key())

			// Added flag in map for Cookie is present
			// we are not considering if the cookie is active
			if uid != "" {
				matchedImpression = 1
			}
		}
		cookieFlagMap[partnerConfig[models.BidderCode]] = matchedImpression
		if matchedImpression == 0 {
			rctx.MetricsEngine.RecordPublisherPartnerNoCookieStats(rctx.PubIDStr, partnerConfig[models.BidderCode])
		}
	}
	return cookieFlagMap
}
