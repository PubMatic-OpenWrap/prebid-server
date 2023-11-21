package auction

func GetWinningBidsIds(adpodBids []*AdPodBid) (map[string][]string, error) {
	var winningBidIds map[string][]string
	if len(adpodBids) == 0 {
		return winningBidIds, nil
	}

	winningBidIds = make(map[string][]string)
	for _, eachAdpodBid := range adpodBids {
		for _, bid := range eachAdpodBid.Bids {
			if len(bid.AdM) == 0 {
				continue
			}
			winningBidIds[eachAdpodBid.OriginalImpID] = append(winningBidIds[eachAdpodBid.OriginalImpID], bid.ID)
		}
	}

	return winningBidIds, nil
}
