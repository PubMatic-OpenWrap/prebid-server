package ctvlegacy

import "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"

func GetWinningBidsIds(adpodBids []*AdPodBid, impCtxMap map[string]models.ImpCtx) (map[string][]string, error) {
	var winningBidIds map[string][]string
	if len(adpodBids) == 0 {
		return winningBidIds, nil
	}

	winningBidIds = make(map[string][]string)
	for _, eachAdpodBid := range adpodBids {
		impCtx := impCtxMap[eachAdpodBid.OriginalImpID]
		for _, bid := range eachAdpodBid.Bids {
			if len(bid.AdM) == 0 {
				continue
			}
			winningBidIds[eachAdpodBid.OriginalImpID] = append(winningBidIds[eachAdpodBid.OriginalImpID], bid.ID)
			impCtx.BidIDToAPRC[bid.ID] = models.StatusWinningBid
		}
		impCtxMap[eachAdpodBid.OriginalImpID] = impCtx
	}

	return winningBidIds, nil
}
