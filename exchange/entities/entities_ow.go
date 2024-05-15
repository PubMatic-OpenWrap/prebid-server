package entities

import "github.com/prebid/prebid-server/v2/openrtb_ext"

func GetNonBidParamsFromPbsOrtbBid(bid *PbsOrtbBid) openrtb_ext.NonBidParams {
	return openrtb_ext.NonBidParams{
		Bid:               bid.Bid,
		OriginalBidCPM:    bid.OriginalBidCPM,
		OriginalBidCur:    bid.OriginalBidCur,
		DealPriority:      bid.DealPriority,
		DealTierSatisfied: bid.DealTierSatisfied,
		GeneratedBidID:    bid.GeneratedBidID,
		TargetBidderCode:  bid.TargetBidderCode,
		OriginalBidCPMUSD: bid.OriginalBidCPMUSD,
		BidMeta:           bid.BidMeta,
		BidType:           bid.BidType,
		BidTargets:        bid.BidTargets,
		BidVideo:          bid.BidVideo,
		BidEvents:         bid.BidEvents,
		BidFloors:         bid.BidFloors,
	}
}
