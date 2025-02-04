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

func applyMutation(bidInfo []*rawBidderResponseHookResult, result *hookstage.HookResult[hookstage.RawBidderResponsePayload], payload hookstage.RawBidderResponsePayload) {
	var (
		newResultSet = []*adapters.TypedBid{}
		seatNonBid   = openrtb_ext.SeatNonBidBuilder{}
	)

	for _, bidResult := range bidInfo {
		if bidResult == nil || bidResult.bid == nil {
			continue
		}

		bidResult.bid.BidType = bidResult.bidtype
		bidResult.bid.Bid.Ext = bidResult.bidExt

		if rejectBid(bidResult.unwrapStatus) {
			seatNonBid.AddBid(openrtb_ext.NewNonBid(openrtb_ext.NonBidParams{
				Bid:            bidResult.bid.Bid,
				NonBidReason:   int(nbr.LossBidLostInVastUnwrap),
				DealPriority:   bidResult.bid.DealPriority,
				BidMeta:        bidResult.bid.BidMeta,
				BidType:        bidResult.bid.BidType,
				BidVideo:       bidResult.bid.BidVideo,
				OriginalBidCur: payload.BidderResponse.Currency,
			}), payload.Bidder)
			continue
		}

		newResultSet = append(newResultSet, bidResult.bid)
	}

	result.ChangeSet.RawBidderResponse().Bids().Update(newResultSet)
	result.SeatNonBid = seatNonBid
}

func (m OpenWrap) handleRawBidderResponseHook(
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawBidderResponsePayload,
) (result hookstage.HookResult[hookstage.RawBidderResponsePayload], err error) {
	var (
		rCtx, rCtxPresent    = miCtx.ModuleContext[models.RequestContext].(models.RequestCtx)
		isVastUnwrapEnabled  = rCtxPresent && rCtx.VastUnwrapEnabled
		isBidderCheckEnabled = isBidderInList(m.cfg.ResponseOverride.BidType, payload.Bidder)
	)

	if !(isBidderCheckEnabled || isVastUnwrapEnabled) {
		return result, nil
	}

	resultSet := []*rawBidderResponseHookResult{}
	for _, bid := range payload.BidderResponse.Bids {
		resultSet = append(resultSet, &rawBidderResponseHookResult{
			bid:     bid,
			bidtype: bid.BidType,
			bidExt:  bid.Bid.Ext,
		})
	}

	if isBidderCheckEnabled {
		m.updateBidderType(resultSet, payload.Bidder)
	}

	if isVastUnwrapEnabled {
		m.processVastUnwrap(resultSet, miCtx, payload.Bidder, rCtx)
	}

	applyMutation(resultSet, &result, payload)

	return result, nil
}

// updateBidderType updates the creative type if bidder check is enabled.
func (m OpenWrap) updateBidderType(resultSet []*rawBidderResponseHookResult, bidder string) {
	for _, bidResult := range resultSet {
		updateCreativeType(bidResult, m.cfg.ResponseOverride.BidType, bidder)
	}
}

// processVastUnwrap unwraps VAST creatives asynchronously if enabled.
func (m OpenWrap) processVastUnwrap(
	resultSet []*rawBidderResponseHookResult,
	miCtx hookstage.ModuleInvocationContext,
	bidder string,
	rCtx models.RequestCtx,
) {
	var wg sync.WaitGroup
	for _, bidResult := range resultSet {
		if isEligibleForUnwrap(*bidResult) {
			wg.Add(1)
			go func(iBid *rawBidderResponseHookResult) {
				defer wg.Done()
				iBid.unwrapStatus = m.unwrap.Unwrap(iBid.bid, miCtx.AccountID, bidder, rCtx.UA, rCtx.IP)
			}(bidResult)
		}
	}
	wg.Wait()
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
