package openwrap

import (
	"fmt"

	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v2/openrtb_ext"

	"github.com/prebid/prebid-server/v2/hooks/hookstage"
)

type BidUnwrapInfo struct {
	bid          *adapters.TypedBid
	unwrapStatus string
}

func (m OpenWrap) handleRawBidderResponseHook(
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawBidderResponsePayload,
) (result hookstage.HookResult[hookstage.RawBidderResponsePayload], err error) {
	vastRequestContext, ok := miCtx.ModuleContext[models.RequestContext].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleRawBidderResponseHook()")
		return result, nil
	}

	if !vastRequestContext.VastUnwrapEnabled {
		return result, nil
	}

	seatNonBid := openrtb_ext.NonBidCollection{}
	unwrappedBids := make([]*adapters.TypedBid, 0, len(payload.BidderResponse.Bids))
	unwrappedBidsChan := make(chan BidUnwrapInfo, len(payload.BidderResponse.Bids))
	defer close(unwrappedBidsChan)

	unwrappedBidsCnt := 0
	// send bids for unwrap
	for _, bid := range payload.BidderResponse.Bids {
		if !isEligibleForUnwrap(bid) {
			continue
		}
		unwrappedBidsCnt++
		go func(bid adapters.TypedBid) {
			unwrapStatus := m.unwrap.Unwrap(&bid, miCtx.AccountID, payload.Bidder, vastRequestContext.UA, vastRequestContext.IP)
			unwrappedBidsChan <- BidUnwrapInfo{&bid, unwrapStatus}
		}(*bid)
	}

	// collect bids after unwrap
	for i := 0; i < unwrappedBidsCnt; i++ {
		unwrappedBid := <-unwrappedBidsChan
		if unwrappedBid.unwrapStatus == models.UnwrapEmptyVASTStatus || unwrappedBid.unwrapStatus == models.UnwrapInvalidVASTStatus {
			seatNonBid.AddBid(
				openrtb_ext.NewNonBid(openrtb_ext.NonBidParams{
					Bid:            unwrappedBid.bid.Bid,
					NonBidReason:   int(nbr.LossBidLostInVastUnwrap),
					DealPriority:   unwrappedBid.bid.DealPriority,
					BidMeta:        unwrappedBid.bid.BidMeta,
					BidType:        unwrappedBid.bid.BidType,
					BidVideo:       unwrappedBid.bid.BidVideo,
					OriginalBidCur: payload.BidderResponse.Currency,
				}), payload.Bidder,
			)
		} else {
			unwrappedBids = append(unwrappedBids, unwrappedBid.bid)
		}
	}

	changeSet := hookstage.ChangeSet[hookstage.RawBidderResponsePayload]{}
	changeSet.RawBidderResponse().Bids().Update(unwrappedBids)
	result.ChangeSet = changeSet
	result.SeatNonBid = seatNonBid
	result.DebugMessages = append(result.DebugMessages,
		fmt.Sprintf("For pubid:[%d] VastUnwrapEnabled: [%v]", vastRequestContext.PubID, vastRequestContext.VastUnwrapEnabled))

	return result, nil
}

func isEligibleForUnwrap(bid *adapters.TypedBid) bool {
	return bid != nil && bid.Bid != nil && bid.Bid.AdM != "" && bid.BidType == openrtb_ext.BidTypeVideo
}
