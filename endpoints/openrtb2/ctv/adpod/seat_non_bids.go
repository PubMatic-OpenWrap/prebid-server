package adpod

import (
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/v2/exchange/entities"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

// GetNonBidParamsFromTypesBid function returns NonBidParams from types Bid
func GetNonBidParamsFromTypesBid(bid *types.Bid, seat string) openrtb_ext.NonBidParams {
	if bid.ExtBid.Prebid == nil {
		bid.ExtBid.Prebid = &openrtb_ext.ExtBidPrebid{}
	}
	if bid.ExtBid.Prebid.Video != nil && bid.ExtBid.Prebid.Video.Duration == 0 && bid.Duration > 0 {
		bid.ExtBid.Prebid.Video.Duration = bid.Duration
	}
	pbsOrtbBid := entities.PbsOrtbBid{
		Bid:               bid.Bid,
		BidMeta:           bid.ExtBid.Prebid.Meta,
		BidType:           bid.ExtBid.Prebid.Type,
		BidTargets:        bid.ExtBid.Prebid.Targeting,
		BidVideo:          bid.ExtBid.Prebid.Video,
		BidEvents:         bid.ExtBid.Prebid.Events,
		BidFloors:         bid.ExtBid.Prebid.Floors,
		DealPriority:      bid.ExtBid.Prebid.DealPriority,
		DealTierSatisfied: bid.DealTierSatisfied,
		GeneratedBidID:    bid.ExtBid.Prebid.BidId,
		OriginalBidCPM:    bid.OriginalBidCPM,
		OriginalBidCur:    bid.OriginalBidCur,
		TargetBidderCode:  bid.ExtBid.Prebid.TargetBidderCode,
		OriginalBidCPMUSD: bid.OriginalBidCPMUSD,
	}
	return entities.GetNonBidParamsFromPbsOrtbBid(&pbsOrtbBid, seat)
}

func addSeatNonBids(bids []*types.Bid) openrtb_ext.NonBidCollection {
	var snb openrtb_ext.NonBidCollection
	for _, bid := range bids {
		if bid.Nbr != nil {
			nonBidParams := GetNonBidParamsFromTypesBid(bid, bid.Seat)
			nonBidParams.NonBidReason = int(*bid.Nbr)
			snb.AddBid(openrtb_ext.NewNonBid(nonBidParams), bid.Seat)
		}
	}
	return snb
}
