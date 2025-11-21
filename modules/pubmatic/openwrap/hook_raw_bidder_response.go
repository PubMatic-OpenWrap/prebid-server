package openwrap

import (
	"encoding/json"
	"slices"
	"sync"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/adapters"
	endpointmanager "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/prebid/prebid-server/v3/privacy"
	"github.com/prebid/prebid-server/v3/util/iputil"

	"github.com/buger/jsonparser"
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
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
			}), payload.BidderResponse.BidderAlias.String())
			continue
		}

		newResultSet = append(newResultSet, bidResult.bid)
	}

	result.ChangeSet.RawBidderResponse().Bids().UpdateBids(newResultSet)
	result.SeatNonBid = seatNonBid
}

func (m OpenWrap) handleRawBidderResponseHook(
	miCtx hookstage.ModuleInvocationContext,
	payload hookstage.RawBidderResponsePayload,
) (result hookstage.HookResult[hookstage.RawBidderResponsePayload], err error) {
	rCtx, ok := utils.GetRequestContext(miCtx)
	if !ok {
		return result, nil
	}

	var (
		isVastUnwrapEnabled  = rCtx.VastUnWrap.Enabled
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
		m.updateBidderType(resultSet)
	}

	if isVastUnwrapEnabled {
		m.processVastUnwrap(resultSet, miCtx, payload.Bidder, rCtx)
	}

	applyMutation(resultSet, &result, payload)

	return result, nil
}

// updateBidderType updates the creative type if bidder check is enabled.
func (m OpenWrap) updateBidderType(resultSet []*rawBidderResponseHookResult) {
	for _, bidResult := range resultSet {
		updateCreativeType(bidResult)
	}
}

// processVastUnwrap unwraps VAST creatives asynchronously if enabled.
func (m OpenWrap) processVastUnwrap(
	resultSet []*rawBidderResponseHookResult,
	miCtx hookstage.ModuleInvocationContext,
	bidder string,
	rCtx models.RequestCtx,
) {
	ip := applyPrivacyMaskingToIP(rCtx.VastUnWrap, rCtx.DeviceCtx.IP)
	//TODO: remove this debug log after prod release once testing is done (Remove after 28th Aug 2025).
	glog.V(models.LogLevelDebug).Infof("processVastUnwrap: IP address is: %s", ip)

	var wg sync.WaitGroup
	for _, bidResult := range resultSet {
		if isEligibleForUnwrap(*bidResult) {
			wg.Add(1)
			go func(iBid *rawBidderResponseHookResult) {
				defer wg.Done()
				iBid.unwrapStatus = m.unwrap.Unwrap(iBid.bid, miCtx.AccountID, bidder, rCtx.DeviceCtx.UA, ip)
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

func updateCreativeType(adapterBid *rawBidderResponseHookResult) {

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

// applyPrivacyMaskingToIP returns the masked IP address if request is consented.
func applyPrivacyMaskingToIP(vastUnWrap models.VastUnWrap, ip string) string {
	if !vastUnWrap.IsPrivacyEnforced {
		return ip
	}

	_, ver := iputil.ParseIP(ip)
	switch ver {
	case iputil.IPv4:
		return privacy.ScrubIP(ip, iputil.IPv4DefaultMaskingBitSize, iputil.IPv4BitSize)
	case iputil.IPv6:
		return privacy.ScrubIP(ip, iputil.IPv6DefaultMaskingBitSize, iputil.IPv6BitSize)
	default:
		return ip
	}
}

func validateModuleContextRawBidderResponseHook(moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, endpointmanager.EndpointHookManager, hookstage.HookResult[hookstage.RawBidderResponsePayload], bool) {
	result := hookstage.HookResult[hookstage.RawBidderResponsePayload]{}

	if moduleCtx.ModuleContext == nil {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleBidderRequestHook()")
		return models.RequestCtx{}, nil, result, false
	}

	rCtxInterface, ok := moduleCtx.ModuleContext.Get("rctx")
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleBidderRequestHook()")
		return models.RequestCtx{}, nil, result, false
	}
	rCtx, ok := rCtxInterface.(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleBidderRequestHook()")
		return models.RequestCtx{}, nil, result, false
	}

	endpointHookManagerInterface, ok := moduleCtx.ModuleContext.Get("endpointhookmanager")
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: endpoint-hook-manager not found in handleBidderRequestHook()")
		return models.RequestCtx{}, nil, result, false
	}
	endpointHookManager, ok := endpointHookManagerInterface.(endpointmanager.EndpointHookManager)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: endpoint-hook-manager not found in handleBidderRequestHook()")
		return models.RequestCtx{}, nil, result, false
	}

	return rCtx, endpointHookManager, result, true
}
