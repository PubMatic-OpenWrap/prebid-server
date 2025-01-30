package openwrap

import (
	"fmt"
	"slices"

	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v2/openrtb_ext"

	"github.com/buger/jsonparser"
	"github.com/prebid/prebid-server/v2/hooks/hookstage"
)

type BidUnwrapInfo struct {
	bid          *adapters.TypedBid
	unwrapStatus string
	bidtype      openrtb_ext.BidType
}

func applyMutation(bidInfo []BidUnwrapInfo, result hookstage.HookResult[hookstage.RawBidderResponsePayload]) {
	result.ChangeSet.AddMutation(func(rp hookstage.RawBidderResponsePayload) (hookstage.RawBidderResponsePayload, error) {
		var bids []*adapters.TypedBid
		for _, bidinfo := range bidInfo {
			bidinfo.bid.BidType = bidinfo.bidtype
			bids = append(bids, bidinfo.bid)
		}
		rp.BidderResponse.Bids = bids
		return rp, nil
	}, hookstage.MutationUpdate, "update-bidtype-for-multiformat-request")
}

func (m OpenWrap) handleRawBidderResponseHook(
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawBidderResponsePayload,
) (result hookstage.HookResult[hookstage.RawBidderResponsePayload], err error) {

	var bidInfo []BidUnwrapInfo

	for _, bid := range payload.BidderResponse.Bids {
		var bids BidUnwrapInfo
		bid, _ = updateCreativeType(bid, m.cfg.ResponseOverride.BidType, payload.Bidder)
		bids.bid = bid
		bids.bidtype = bid.BidType
		bidInfo = append(bidInfo, bids)
	}
	vastRequestContext, ok := miCtx.ModuleContext[models.RequestContext].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleRawBidderResponseHook()")
		applyMutation(bidInfo, result)
		return result, nil
	}

	if !vastRequestContext.VastUnwrapEnabled {
		applyMutation(bidInfo, result)
		return result, nil
	}
	seatNonBid := openrtb_ext.SeatNonBidBuilder{}
	unwrappedBids := make([]*adapters.TypedBid, 0, len(payload.BidderResponse.Bids))
	unwrappedBidsChan := make(chan BidUnwrapInfo, len(payload.BidderResponse.Bids))
	defer close(unwrappedBidsChan)

	unwrappedBidsCnt, unwrappedSuccessBidCnt := 0, 0
	totalBidCnt := len(payload.BidderResponse.Bids)
	for _, bid := range bidInfo {

		// send bids for unwrap
		if !isEligibleForUnwrap(bid.bid) {
			unwrappedBids = append(unwrappedBids, bid.bid)
			continue
		}
		unwrappedBidsCnt++
		go func(bid adapters.TypedBid, bidType openrtb_ext.BidType) {
			unwrapStatus := m.unwrap.Unwrap(&bid, miCtx.AccountID, payload.Bidder, vastRequestContext.UA, vastRequestContext.IP)
			unwrappedBidsChan <- BidUnwrapInfo{&bid, unwrapStatus, bidType}
		}(*bid.bid, bid.bidtype)

	}

	// collect bids after unwrap
	for i := 0; i < unwrappedBidsCnt; i++ {
		unwrappedBid := <-unwrappedBidsChan
		if !rejectBid(unwrappedBid.unwrapStatus) {
			unwrappedSuccessBidCnt++
			unwrappedBids = append(unwrappedBids, unwrappedBid.bid)
			continue
		}
		seatNonBid.AddBid(openrtb_ext.NewNonBid(openrtb_ext.NonBidParams{
			Bid:            unwrappedBid.bid.Bid,
			NonBidReason:   int(nbr.LossBidLostInVastUnwrap),
			DealPriority:   unwrappedBid.bid.DealPriority,
			BidMeta:        unwrappedBid.bid.BidMeta,
			BidType:        unwrappedBid.bid.BidType,
			BidVideo:       unwrappedBid.bid.BidVideo,
			OriginalBidCur: payload.BidderResponse.Currency,
		}), payload.Bidder,
		)
	}

	changeSet := hookstage.ChangeSet[hookstage.RawBidderResponsePayload]{}
	changeSet.RawBidderResponse().Bids().Update(unwrappedBids)
	result.ChangeSet = changeSet
	result.SeatNonBid = seatNonBid
	result.DebugMessages = append(result.DebugMessages,
		fmt.Sprintf("For pubid:[%d] VastUnwrapEnabled: [%v] Total Input Bids: [%d] Total Bids sent for unwrapping: [%d] Total Unwrap Success: [%d]", vastRequestContext.PubID, vastRequestContext.VastUnwrapEnabled, totalBidCnt, unwrappedBidsCnt, unwrappedSuccessBidCnt))
	return result, nil
}

func isEligibleForUnwrap(bid *adapters.TypedBid) bool {
	return bid != nil && bid.BidType == openrtb_ext.BidTypeVideo && bid.Bid != nil && bid.Bid.AdM != ""
}

func rejectBid(bidUnwrapStatus string) bool {
	return bidUnwrapStatus == models.UnwrapEmptyVASTStatus || bidUnwrapStatus == models.UnwrapInvalidVASTStatus
}

func isBidderInList(bidderList []string, bidder string) bool {
	return slices.Contains(bidderList, bidder)
}

func updateCreativeType(adapterBid *adapters.TypedBid, bidders []string, bidder string) (*adapters.TypedBid, error) {
	// Check if the bidder is in the bidders list
	if !isBidderInList(bidders, bidder) {
		return adapterBid, nil
	}

	bidType := GetCreativeTypeFromCreative(adapterBid.Bid)
	if bidType == "" {
		return adapterBid, nil
	}

	newBidType := openrtb_ext.BidType(bidType)
	if adapterBid.BidType != newBidType {
		adapterBid.BidType = newBidType
	}

	// Update the "prebid.type" field in the bid extension
	updatedExt, err := jsonparser.Set(adapterBid.Bid.Ext, []byte(fmt.Sprintf(`"%s"`, bidType)), "prebid", "type")
	if err != nil {
		return adapterBid, models.ErrorWrap(err, fmt.Errorf("error updating bid extension for bidder %s, bid ID %s: %v", bidder, adapterBid.Bid.ID, err))
	}

	// Assign the updated JSON only if `jsonparser.Set` succeeds
	adapterBid.Bid.Ext = updatedExt

	return adapterBid, nil
}
