package openwrap

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"time"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/hooks/hookanalytics"
	"github.com/prebid/prebid-server/v2/hooks/hookstage"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/adpod/auction"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/adunitconfig"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/tracker"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/utils"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

func (m OpenWrap) handleAuctionResponseHook(
	ctx context.Context,
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.AuctionResponsePayload,
) (hookstage.HookResult[hookstage.AuctionResponsePayload], error) {
	result := hookstage.HookResult[hookstage.AuctionResponsePayload]{}
	result.ChangeSet = hookstage.ChangeSet[hookstage.AuctionResponsePayload]{}

	// absence of rctx at this hook means the first hook failed!. Do nothing
	if len(moduleCtx.ModuleContext) == 0 {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleAuctionResponseHook()")
		return result, nil
	}
	rctx, ok := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleAuctionResponseHook()")
		return result, nil
	}

	//SSHB request should not execute module
	if rctx.Sshb == "1" || rctx.Endpoint == models.EndpointHybrid {
		return result, nil
	}

	defer func() {
		moduleCtx.ModuleContext["rctx"] = rctx
		m.metricEngine.RecordPublisherResponseTimeStats(rctx.PubIDStr, int(time.Since(time.Unix(rctx.StartTime, 0)).Milliseconds()))
	}()

	// cache rctx for analytics
	result.AnalyticsTags = hookanalytics.Analytics{
		Activities: []hookanalytics.Activity{
			{
				Name: "openwrap_request_ctx",
				Results: []hookanalytics.Result{
					{
						Values: map[string]interface{}{
							"request-ctx": &rctx,
						},
					},
				},
			},
		},
	}

	if rctx.IsCTVRequest && payload.BidResponse.NBR != nil {
		return result, nil
	}

	var winningAdpodBidIds map[string][]string
	var errs []error
	if rctx.IsCTVRequest {
		winningAdpodBidIds, errs = auction.FormAdpodBidsAndPerformExclusion(payload.BidResponse, rctx)
		if len(errs) > 0 {
			for i := range errs {
				result.Errors = append(result.Errors, errs[i].Error())
			}
			result.NbrCode = int(nbr.InternalError)
		}
	}

	anyDealTierSatisfyingBid := false
	winningBids := make(models.WinningBids)

	payload.BidResponse = updateMultiFloorResponse(payload.BidResponse)

	for _, seatBid := range payload.BidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			m.metricEngine.RecordPlatformPublisherPartnerResponseStats(rctx.Platform, rctx.PubIDStr, seatBid.Seat)

			impId := bid.ImpID
			if rctx.IsCTVRequest {
				impId, _ = models.GetImpressionID(bid.ImpID)
			}

			impCtx, ok := rctx.ImpBidCtx[impId]
			if !ok {
				result.Errors = append(result.Errors, "invalid impCtx.ID for bid"+impId)
				continue
			}

			partnerID := 0
			bidderMeta, ok := impCtx.Bidders[seatBid.Seat]
			if ok {
				partnerID = bidderMeta.PartnerID
			}

			var eg, en float64
			bidExt := &models.BidExt{}

			if len(bid.Ext) != 0 {
				err := json.Unmarshal(bid.Ext, bidExt)
				if err != nil {
					result.Errors = append(result.Errors, "failed to unmarshal bid.ext for "+utils.GetOriginalBidId(bid.ID))
					// continue
				}
			}

			if rctx.IsCTVRequest {
				if dur, ok := impCtx.BidIDToDur[bid.ID]; ok {
					bidExt.Prebid.Video.Duration = int(dur)
				}
			}

			if impCtx.Video != nil && bidExt.Prebid != nil && bidExt.Prebid.Type == openrtb_ext.BidTypeVideo && bidExt.Prebid.Video != nil && bidExt.Prebid.Video.Duration == 0 {
				bidExt.Prebid.Video.Duration = int(impCtx.Video.MaxDuration)
			}

			// NYC_TODO: fix this in PBS-Core or ExecuteAllProcessedBidResponsesStage
			if bidExt.Prebid != nil && bidExt.Prebid.Video != nil && bidExt.Prebid.Video.Duration == 0 &&
				bidExt.Prebid.Video.PrimaryCategory == "" && bidExt.Prebid.Video.VASTTagID == "" {
				bidExt.Prebid.Video = nil
			}

			// Update VastTagFlags for the bids
			if len(bidderMeta.VASTTagFlags) > 0 {
				if bidExt.Prebid != nil && bidExt.Prebid.Video != nil && len(bidExt.Prebid.Video.VASTTagID) > 0 {
					bidderMeta.VASTTagFlags[bidExt.Prebid.Video.VASTTagID] = true
					impCtx.Bidders[seatBid.Seat] = bidderMeta
				}
			}

			if v, ok := rctx.PartnerConfigMap[models.VersionLevelConfigID]["refreshInterval"]; ok {
				n, err := strconv.Atoi(v)
				if err == nil {
					bidExt.RefreshInterval = n
				}
			}

			if bidExt.Prebid != nil {
				bidExt.CreativeType = string(bidExt.Prebid.Type)
			}
			if bidExt.CreativeType == "" {
				bidExt.CreativeType = models.GetCreativeType(&bid, bidExt, &impCtx)
			}

			// set response netecpm and logger/tracker en
			revShare := models.GetRevenueShare(rctx.PartnerConfigMap[partnerID])
			bidExt.NetECPM = models.ToFixed(bid.Price, models.BID_PRECISION)
			eg = models.GetGrossEcpmFromNetEcpm(bid.Price, revShare)
			en = bidExt.NetECPM
			if payload.BidResponse.Cur != "USD" {
				eg = models.GetGrossEcpmFromNetEcpm(bidExt.OriginalBidCPMUSD, revShare)
				en = bidExt.OriginalBidCPMUSD
				bidExt.OriginalBidCPMUSD = 0
			}

			if impCtx.Video != nil && impCtx.Type == "video" && bidExt.CreativeType == "video" {
				if bidExt.Video == nil {
					bidExt.Video = &models.ExtBidVideo{}
				}
				if impCtx.Video.MaxDuration != 0 {
					bidExt.Video.MaxDuration = impCtx.Video.MaxDuration
				}
				if impCtx.Video.MinDuration != 0 {
					bidExt.Video.MinDuration = impCtx.Video.MinDuration
				}
				if impCtx.Video.Skip != nil {
					bidExt.Video.Skip = impCtx.Video.Skip
				}
				if impCtx.Video.SkipAfter != 0 {
					bidExt.Video.SkipAfter = impCtx.Video.SkipAfter
				}
				if impCtx.Video.SkipMin != 0 {
					bidExt.Video.SkipMin = impCtx.Video.SkipMin
				}
				bidExt.Video.BAttr = impCtx.Video.BAttr
				bidExt.Video.PlaybackMethod = impCtx.Video.PlaybackMethod
				if rctx.ClientConfigFlag == 1 {
					bidExt.Video.ClientConfig = adunitconfig.GetClientConfigForMediaType(rctx, impId, "video")
				}
			} else if impCtx.Banner && bidExt.CreativeType == "banner" && rctx.ClientConfigFlag == 1 {
				cc := adunitconfig.GetClientConfigForMediaType(rctx, impId, "banner")
				if len(cc) != 0 {
					if bidExt.Banner == nil {
						bidExt.Banner = &models.ExtBidBanner{}
					}
					bidExt.Banner.ClientConfig = cc
				}
			}

			bidDealTierSatisfied := false
			if bidExt.Prebid != nil {
				bidDealTierSatisfied = bidExt.Prebid.DealTierSatisfied
				if bidDealTierSatisfied {
					anyDealTierSatisfyingBid = true // found at least one bid which satisfies dealTier
				}
			}

			owbid := models.OwBid{
				ID:                   bid.ID,
				NetEcpm:              bidExt.NetECPM,
				BidDealTierSatisfied: bidDealTierSatisfied,
			}

			var wbid models.OwBid
			var wbids []*models.OwBid
			var oldWinBidFound bool
			if rctx.IsCTVRequest && impCtx.AdpodConfig != nil {
				if CheckWinningBidId(bid.ID, winningAdpodBidIds[impId]) {
					winningBids.AppendBid(impId, &owbid)
				}
			} else {
				wbids, oldWinBidFound = winningBids[bid.ImpID]
				if len(wbids) > 0 {
					wbid = *wbids[0]
				}
				if !oldWinBidFound {
					winningBids[bid.ImpID] = make([]*models.OwBid, 1)
					winningBids[bid.ImpID][0] = &owbid
				} else if models.IsNewWinningBid(&owbid, &wbid, rctx.SupportDeals) {
					winningBids[bid.ImpID][0] = &owbid
				}
			}

			// update NonBr codes for current bid
			if owbid.Nbr != nil {
				bidExt.Nbr = owbid.Nbr
			}

			if rctx.IsCTVRequest && impCtx.AdpodConfig != nil {
				bidExt.Nbr = auction.ConvertAPRCToNBRC(impCtx.BidIDToAPRC[bid.ID])
			} else {
				// if current bid is winner then update NonBr code for earlier winning bid
				if winningBids.IsWinningBid(impId, owbid.ID) && oldWinBidFound {
					winBidCtx := rctx.ImpBidCtx[impId].BidCtx[wbid.ID]
					winBidCtx.BidExt.Nbr = wbid.Nbr
					rctx.ImpBidCtx[impId].BidCtx[wbid.ID] = winBidCtx
				}
			}

			// cache for bid details for logger and tracker
			if impCtx.BidCtx == nil {
				impCtx.BidCtx = make(map[string]models.BidCtx)
			}
			impCtx.BidCtx[bid.ID] = models.BidCtx{
				BidExt: *bidExt,
				EG:     eg,
				EN:     en,
			}
			rctx.ImpBidCtx[impId] = impCtx
		}
	}

	rctx.WinningBids = winningBids
	if len(winningBids) == 0 {
		m.metricEngine.RecordNobidErrPrebidServerResponse(rctx.PubIDStr)
	}

	/*
		At this point of time,
		1. For price-based auction (request with supportDeals = false),
				all rejected bids will have NonBR code as LossLostToHigherBid which is expected.
		2. For request with supportDeals = true :
			2.1) If all bids are non-deal-bids (bidExt.Prebid.DealTierSatisfied = false)
					then NonBR code for them will be LossLostToHigherBid which is expected.
			2.2) If one of the bid is deal-bid (bidExt.Prebid.DealTierSatisfied = true)
				expectation:
					all rejected non-deal bids should have NonBR code as LossLostToDealBid
					all rejected deal-bids should have NonBR code as LossLostToHigherBid
				addLostToDealBidNonBRCode function will make sure that above expectation are met.
	*/
	if anyDealTierSatisfyingBid {
		addLostToDealBidNonBRCode(&rctx)
	}

	droppedBids, warnings := m.addPWTTargetingForBid(rctx, payload.BidResponse)
	if len(droppedBids) != 0 {
		rctx.DroppedBids = droppedBids
	}
	if len(warnings) != 0 {
		result.Warnings = append(result.Warnings, warnings...)
	}

	responseExt := openrtb_ext.ExtBidResponse{}
	// TODO use concrete structure
	if len(payload.BidResponse.Ext) != 0 {
		if err := json.Unmarshal(payload.BidResponse.Ext, &responseExt); err != nil {
			result.Errors = append(result.Errors, "failed to unmarshal response.ext err: "+err.Error())
		}
	}

	if rctx.IsCTVRequest && rctx.Endpoint == models.EndpointJson {
		if len(rctx.RedirectURL) > 0 {
			responseExt.Wrapper = &openrtb_ext.ExtWrapper{
				ResponseFormat: rctx.ResponseFormat,
				RedirectURL:    rctx.RedirectURL,
			}
		}

		impToAdserverURL := make(map[string]string)
		for _, impCtx := range rctx.ImpBidCtx {
			if impCtx.AdserverURL != "" {
				impToAdserverURL[impCtx.ImpID] = impCtx.AdserverURL
			}
		}

		if len(impToAdserverURL) > 0 {
			if responseExt.Wrapper == nil {
				responseExt.Wrapper = &openrtb_ext.ExtWrapper{}
			}
			responseExt.Wrapper.ImpToAdServerURL = impToAdserverURL
		}
	}

	rctx.ResponseExt = responseExt
	rctx.DefaultBids = m.addDefaultBids(&rctx, payload.BidResponse, responseExt)

	rctx.Trackers = tracker.CreateTrackers(rctx, payload.BidResponse)

	for bidder, responseTimeMs := range responseExt.ResponseTimeMillis {
		rctx.BidderResponseTimeMillis[bidder.String()] = responseTimeMs
		m.metricEngine.RecordPartnerResponseTimeStats(rctx.PubIDStr, string(bidder), responseTimeMs)
	}

	// TODO: PBS-Core should pass the hostcookie for module to usersync.ParseCookieFromRequest()
	rctx.MatchedImpression = getMatchedImpression(rctx)
	matchedImpression, err := json.Marshal(rctx.MatchedImpression)
	if err == nil {
		responseExt.OwMatchedImpression = matchedImpression
	}

	if rctx.SendAllBids {
		responseExt.OwSendAllBids = 1
	}

	if rctx.LogInfoFlag == 1 {
		responseExt.OwLogInfo = &openrtb_ext.OwLogInfo{
			// Logger:  openwrap.GetLogAuctionObjectAsURL(ao, true, true), updated done later
			Tracker: tracker.GetTrackerInfo(rctx, responseExt),
		}
	}

	result.SeatNonBid = prepareSeatNonBids(rctx)

	if rctx.Debug {
		rCtxBytes, _ := json.Marshal(rctx)
		result.DebugMessages = append(result.DebugMessages, string(rCtxBytes))
	}

	rctx.AppLovinMax = updateAppLovinMaxResponse(rctx, payload.BidResponse)

	if rctx.Endpoint == models.EndpointWebS2S {
		result.ChangeSet.AddMutation(func(ap hookstage.AuctionResponsePayload) (hookstage.AuctionResponsePayload, error) {
			rctx := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
			var err error
			ap.BidResponse, err = tracker.InjectTrackers(rctx, ap.BidResponse)
			if err == nil {
				resetBidIdtoOriginal(ap.BidResponse)
			}
			return ap, err
		}, hookstage.MutationUpdate, "response-body-with-webs2s-format")
		return result, nil
	}

	result.ChangeSet.AddMutation(func(ap hookstage.AuctionResponsePayload) (hookstage.AuctionResponsePayload, error) {
		rctx := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
		var err error
		ap.BidResponse, err = m.updateORTBV25Response(rctx, ap.BidResponse)
		if err != nil {
			return ap, err
		}

		ap.BidResponse, err = tracker.InjectTrackers(rctx, ap.BidResponse)
		if err != nil {
			return ap, err
		}

		var responseExtjson json.RawMessage
		responseExtjson, err = json.Marshal(responseExt)
		if err != nil {
			result.Errors = append(result.Errors, "failed to marshal response.ext err: "+err.Error())
		}
		ap.BidResponse, err = m.applyDefaultBids(rctx, ap.BidResponse)
		ap.BidResponse.Ext = responseExtjson

		resetBidIdtoOriginal(ap.BidResponse)

		if rctx.Endpoint == models.EndpointAppLovinMax {
			ap.BidResponse = applyAppLovinMaxResponse(rctx, ap.BidResponse)
		}
		return ap, err
	}, hookstage.MutationUpdate, "response-body-with-sshb-format")

	// TODO: move debug here
	// result.ChangeSet.AddMutation(func(ap hookstage.AuctionResponsePayload) (hookstage.AuctionResponsePayload, error) {
	// }, hookstage.MutationUpdate, "response-body-with-sshb-format")
	return result, nil
}

func updateMultiFloorResponse(bidResponse *openrtb2.BidResponse) *openrtb2.BidResponse {
	pubMaticResponseMap := map[string][]openrtb2.Bid{}
	for _, seatBid := range bidResponse.SeatBid {
		if seatBid.Seat == "pubmatic" {
			for _, bid := range seatBid.Bid {
				if bid.Price == 0 {
					continue
				}
				bid.ImpID = trimSuffixWithPattern(bid.ImpID)
				pubMaticResponseMap[bid.ImpID] = append(pubMaticResponseMap[bid.ImpID], bid)
			}
		}
	}
	winningBids := []openrtb2.Bid{}
	for _, bid := range pubMaticResponseMap {
		if len(bid) == 0 {
			continue
		}
		winningBid := bid[0]
		for i := 1; i < len(bid); i++ {
			if bid[i].Price > winningBid.Price {
				winningBid = bid[i]
			}
		}
		winningBids = append(winningBids, winningBid)
	}

	for i, seatBid := range bidResponse.SeatBid {
		if seatBid.Seat == "pubmatic" {
			bidResponse.SeatBid[i].Bid = winningBids
		}
	}
	return bidResponse
}

func trimSuffixWithPattern(input string) string {
	re := regexp.MustCompile(models.AppLovinMaxImpressionPattern)
	return re.ReplaceAllString(input, "")
}

func (m *OpenWrap) updateORTBV25Response(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) (*openrtb2.BidResponse, error) {
	if len(bidResponse.SeatBid) == 0 {
		return bidResponse, nil
	}

	// remove non-winning bids if sendallbids=1
	if !rctx.SendAllBids {
		for i := range bidResponse.SeatBid {
			filteredBid := make([]openrtb2.Bid, 0, len(bidResponse.SeatBid[i].Bid))
			for _, bid := range bidResponse.SeatBid[i].Bid {
				impId := bid.ImpID
				if rctx.IsCTVRequest {
					impId, _ = models.GetImpressionID(bid.ImpID)
				}
				if rctx.WinningBids.IsWinningBid(impId, bid.ID) {
					filteredBid = append(filteredBid, bid)
				}
			}
			bidResponse.SeatBid[i].Bid = filteredBid
		}
	}

	// remove seats with empty bids (will add nobids later)
	filteredSeatBid := make([]openrtb2.SeatBid, 0, len(bidResponse.SeatBid))
	for _, seatBid := range bidResponse.SeatBid {
		if len(seatBid.Bid) > 0 {
			filteredSeatBid = append(filteredSeatBid, seatBid)
		}
	}
	bidResponse.SeatBid = filteredSeatBid

	// keep pubmatic 1st to handle automation failure.
	if len(bidResponse.SeatBid) != 0 {
		if bidResponse.SeatBid[0].Seat != "pubmatic" {
			for i := 0; i < len(bidResponse.SeatBid); i++ {
				if bidResponse.SeatBid[i].Seat == "pubmatic" {
					temp := bidResponse.SeatBid[0]
					bidResponse.SeatBid[0] = bidResponse.SeatBid[i]
					bidResponse.SeatBid[i] = temp
				}
			}
		}
	}

	// update bid ext and other details
	for i, seatBid := range bidResponse.SeatBid {
		for j, bid := range seatBid.Bid {
			impId := bid.ImpID
			if rctx.IsCTVRequest {
				impId, _ = models.GetImpressionID(bid.ImpID)
			}
			impCtx, ok := rctx.ImpBidCtx[impId]
			if !ok {
				continue
			}

			bidCtx, ok := impCtx.BidCtx[bid.ID]
			if !ok {
				continue
			}

			bidResponse.SeatBid[i].Bid[j].Ext, _ = json.Marshal(bidCtx.BidExt)
		}
	}

	return bidResponse, nil
}

func getPlatformName(platform string) string {
	if platform == models.PLATFORM_APP {
		return models.PlatformAppTargetingKey
	}
	return platform
}

func resetBidIdtoOriginal(bidResponse *openrtb2.BidResponse) {
	for i, seatBid := range bidResponse.SeatBid {
		for j, bid := range seatBid.Bid {
			bidResponse.SeatBid[i].Bid[j].ID = utils.GetOriginalBidId(bid.ID)
		}
	}
}

func CheckWinningBidId(bidId string, wbidIds []string) bool {
	if len(wbidIds) == 0 {
		return false
	}

	for i := range wbidIds {
		if bidId == wbidIds[i] {
			return true
		}
	}

	return false
}
