package openwrap

import (
	"encoding/json"
	"slices"
	"sync"

	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v2/openrtb_ext"

	"github.com/buger/jsonparser"
	"github.com/prebid/prebid-server/v2/hooks/hookstage"
)

type rawBidderResponseHookResult struct {
	bid          *adapters.TypedBid
	unwrapStatus string
	bidtype      openrtb_ext.BidType
	bidExt       json.RawMessage
}

func applyMutation(bidInfo []rawBidderResponseHookResult, result *hookstage.HookResult[hookstage.RawBidderResponsePayload], payload hookstage.RawBidderResponsePayload) {
	result.ChangeSet.AddMutation(func(rp hookstage.RawBidderResponsePayload) (hookstage.RawBidderResponsePayload, error) {
		newResultSet := []*adapters.TypedBid{}
		unwrappedSuccessBidCnt := 0
		seatNonBid := openrtb_ext.SeatNonBidBuilder{}
		for _, bidResult := range bidInfo {
			bidResult.bid.BidType = bidResult.bidtype
			bidResult.bid.Bid.Ext = bidResult.bidExt
			if !rejectBid(bidResult.unwrapStatus) {
				unwrappedSuccessBidCnt++
				newResultSet = append(newResultSet, bidResult.bid)
			} else {
				seatNonBid.AddBid(openrtb_ext.NewNonBid(openrtb_ext.NonBidParams{
					Bid:            bidResult.bid.Bid,
					NonBidReason:   int(nbr.LossBidLostInVastUnwrap),
					DealPriority:   bidResult.bid.DealPriority,
					BidMeta:        bidResult.bid.BidMeta,
					BidType:        bidResult.bid.BidType,
					BidVideo:       bidResult.bid.BidVideo,
					OriginalBidCur: payload.BidderResponse.Currency,
				}), payload.Bidder,
				)
			}
		}
		rp.BidderResponse.Bids = newResultSet
		result.SeatNonBid = seatNonBid

		return rp, nil
	}, hookstage.MutationUpdate, "update-bidtype-for-multiformat-request")
}

func (m OpenWrap) handleRawBidderResponseHook(
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawBidderResponsePayload,
) (result hookstage.HookResult[hookstage.RawBidderResponsePayload], err error) {	
	//conditions
	var (
		isBidderCheckEnabled = isBidderInList(m.cfg.ResponseOverride.BidType, payload.Bidder)
		rCtx, rCtxPresent = miCtx.ModuleContext[models.RequestContext].(models.RequestCtx)
		isVastUnwrapEnabled  =rCtxPresent  && rCtx.VastUnwrapEnabled
	)

	if !(isBidderCheckEnabled || isVastUnwrapEnabled) {
		return result, nil
	}

	var resultSet []rawBidderResponseHookResult

	for _, bid := range payload.BidderResponse.Bids {
		resultSet = append(resultSet, rawBidderResponseHookResult{
			bid:     bid,
			bidtype: bid.BidType,
			bidExt:  bid.Bid.Ext,
		})
	}

	wg := sync.WaitGroup{}
	for i := range resultSet {
		bidResult := &resultSet[i]

		if isBidderCheckEnabled {
			updateCreativeType(bidResult, m.cfg.ResponseOverride.BidType, payload.Bidder)
		}

		if isVastUnwrapEnabled && isEligibleForUnwrap(*bidResult) {
			wg.Add(1)
			go func(iBid *rawBidderResponseHookResult) {
				defer wg.Done()
				iBid.unwrapStatus = m.unwrap.Unwrap(iBid.bid, miCtx.AccountID, payload.Bidder, rCtx.UA, rCtx.IP)
			}(bidResult)
		}
	}

	wg.Wait()

	applyMutation(resultSet, &result, payload)

	return result, nil
}

func isEligibleForUnwrap(bidResult rawBidderResponseHookResult) bool {
	return bidResult.bid != nil && bidResult.bidtype == openrtb_ext.BidTypeVideo && bidResult.bid.Bid != nil && bidResult.bid.Bid.AdM != ""
}

func rejectBid(bidUnwrapStatus string) bool {
	return bidUnwrapStatus == models.UnwrapEmptyVASTStatus || bidUnwrapStatus == models.UnwrapInvalidVASTStatus
}

func isBidderInList(bidderList []string, bidder string) bool {
	return slices.Contains(bidderList, bidder)
}

func updateCreativeType(adapterBid *rawBidderResponseHookResult, bidders []string, bidder string) {
	// Check if the bidder is in the bidders list
	if !isBidderInList(bidders, bidder) {
		return
	}

	bidType := openrtb_ext.GetCreativeTypeFromCreative(adapterBid.bid.Bid.AdM)
	if bidType == "" {
		return
	}

	newBidType := openrtb_ext.BidType(bidType)
	if adapterBid.bidtype != newBidType {
		adapterBid.bidtype = newBidType
	}

	// Update the "prebid.type" field in the bid extension
	updatedExt, err := jsonparser.Set(adapterBid.bidExt, []byte(`"`+bidType+`"`), "prebid", "type")
	if err != nil {
		return
	}

	// Assign the updated JSON only if `jsonparser.Set` succeeds
	adapterBid.bidExt = updatedExt
	return
}
