package adpod

import (
	"encoding/json"

	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/util"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type StructuredAdpod struct {
	AdpodCtx
	ImpBidMap  map[string][]*types.Bid
	WinningBid map[string]types.Bid
}

func (da *StructuredAdpod) GetPodType() PodType {
	return da.Type
}

func (sa *StructuredAdpod) AddImpressions(imp openrtb2.Imp) {
	sa.Imps = append(sa.Imps, imp)
}

func (sa *StructuredAdpod) GenerateImpressions() {
	// We do not generate impressions in case of structured adpod
}

func (sa *StructuredAdpod) GetImpressions() []openrtb2.Imp {
	return sa.Imps
}

func (sa *StructuredAdpod) CollectBid(bid openrtb2.Bid, seat string) {
	ext := openrtb_ext.ExtBid{}
	if bid.Ext != nil {
		json.Unmarshal(bid.Ext, &ext)
	}

	adpodBid := types.Bid{
		Bid:               &bid,
		ExtBid:            ext,
		DealTierSatisfied: util.GetDealTierSatisfied(&ext),
		Seat:              seat,
	}
	bids := sa.ImpBidMap[bid.ImpID]

	bids = append(bids, &adpodBid)
	sa.ImpBidMap[bid.ImpID] = bids
}

func (sa *StructuredAdpod) PerformAuctionAndExclusion() {
	// Sort Bids
	for impId, bids := range sa.ImpBidMap {
		util.SortBids(bids)
		sa.ImpBidMap[impId] = bids
	}

	for impId, bids := range sa.ImpBidMap {
		sa.WinningBid[impId] = *bids[0]
	}

}

func (sa *StructuredAdpod) Validate() []error {
	return nil
}

func (sa *StructuredAdpod) GetAdpodSeatBids() []openrtb2.SeatBid {
	if len(sa.WinningBid) == 0 {
		return nil
	}

	seatBidMap := make(map[string][]openrtb2.Bid)
	for _, bid := range sa.WinningBid {
		seatBidMap[bid.Seat] = append(seatBidMap[bid.Seat], *bid.Bid)
	}

	var seatBids []openrtb2.SeatBid
	for seat, bids := range seatBidMap {
		adpodSeat := openrtb2.SeatBid{
			Bid:  bids,
			Seat: seat,
		}
		seatBids = append(seatBids, adpodSeat)
	}

	return seatBids
}

func (sa *StructuredAdpod) GetAdpodExtension(blockedVastTagID map[string]map[string][]string) *types.ImpData {
	return nil
}
