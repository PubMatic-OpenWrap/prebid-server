package entities

import "github.com/prebid/prebid-server/v3/openrtb_ext"

// GetNonBidParamsFromPbsOrtbBid function returns NonBidParams from PbsOrtbBid
func GetNonBidParamsFromPbsOrtbBid(bid *PbsOrtbBid, seat string) openrtb_ext.NonBidParams {
	adapterCode := seat
	if bid.AlternateBidderCode != "" {
		adapterCode = string(openrtb_ext.BidderName(bid.AlternateBidderCode))
	}
	if bid.BidMeta == nil {
		bid.BidMeta = &openrtb_ext.ExtBidPrebidMeta{}
	}
	bid.BidMeta.AdapterCode = adapterCode
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
