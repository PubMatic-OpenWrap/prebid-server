package auction

func GetWinningBidsIds(adpodBids []*AdPodBid) map[string][]string {
	var winningBidIds map[string][]string
	if len(adpodBids) == 0 {
		return winningBidIds
	}

	winningBidIds = make(map[string][]string)
	for _, eachAdpodBid := range adpodBids {
		for _, bid := range eachAdpodBid.Bids {
			bidId := bid.ID
			if bid.ExtBid.Prebid != nil && bid.ExtBid.Prebid.BidId != "" {
				bidId = bid.ExtBid.Prebid.BidId
			}
			winningBidIds[eachAdpodBid.OriginalImpID] = append(winningBidIds[eachAdpodBid.OriginalImpID], bidId)
		}
	}

	return winningBidIds
}
