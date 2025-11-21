package openwrap

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/hooks/hookanalytics"
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adunitconfig"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/auction"
	endpointmanager "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/enpdointmanager"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/feature"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/parser"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/sdk/googlesdk"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/sdk/sdkutils"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/sdk/unitylevelplay"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/tracker"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/prebid/prebid-server/v3/util/jsonutil"
)

func (m OpenWrap) handleAuctionResponseHook(
	_ context.Context,
	moduleCtx hookstage.ModuleInvocationContext,
	payload hookstage.AuctionResponsePayload,
) (hookstage.HookResult[hookstage.AuctionResponsePayload], error) {
	rctx, endpointHookManager, result, ok := validateModuleContextAuctionResponseHook(moduleCtx)
	if !ok {
		return result, nil
	}

	//SSHB request should not execute module
	if rctx.Sshb == "1" || rctx.Endpoint == models.EndpointHybrid {
		return result, nil
	}

	defer func() {
		moduleCtx.ModuleContext.Set("rctx", rctx)
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

	if payload.BidResponse.NBR != nil && rctx.IsCTVRequest {
		return result, nil
	}

	//Impression counting method enabled bidders
	if rctx.Endpoint == models.EndpointV25 || sdkutils.IsSdkIntegration(rctx.Endpoint) {
		rctx.ImpCountingMethodEnabledBidders = m.pubFeatures.GetImpCountingMethodEnabledBidders()
	}

	// Populate Bid extension
	result, ok = populateBidExt(&rctx, result, payload.BidResponse)
	if !ok {
		return result, nil
	}

	// Initialize winning bids
	rctx.WinningBids = make(models.WinningBids)

	// Handle Auction Response Hook (Perform endpoint specific auction)
	rctx, result, err := endpointHookManager.HandleAuctionResponseHook(payload, rctx, result, moduleCtx)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, nil
	}

	// Perform auction
	auction.Auction(rctx, payload.BidResponse)

	if len(rctx.WinningBids) == 0 {
		m.metricEngine.RecordNobidErrPrebidServerResponse(rctx.PubIDStr)
	}

	droppedBids, warnings := m.addPWTTargetingForBid(rctx, payload.BidResponse)
	if len(droppedBids) != 0 {
		rctx.DroppedBids = droppedBids
	}
	if len(warnings) != 0 {
		result.Warnings = append(result.Warnings, warnings...)
	}

	responseExt := openrtb_ext.ExtBidResponse{}
	if len(payload.BidResponse.Ext) != 0 {
		if err := json.Unmarshal(payload.BidResponse.Ext, &responseExt); err != nil {
			result.Errors = append(result.Errors, "failed to unmarshal response.ext err: "+err.Error())
		}
	}

	rctx.ResponseExt = responseExt
	rctx.DefaultBids = m.addDefaultBids(&rctx, payload.BidResponse, responseExt)
	rctx.DefaultBids = m.addDefaultBidsForMultiFloorsConfig(&rctx, payload.BidResponse, responseExt)

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

	result.SeatNonBid = prepareSeatNonBids(rctx)

	if rctx.Debug {
		rCtxBytes, _ := json.Marshal(rctx)
		result.DebugMessages = append(result.DebugMessages, string(rCtxBytes))
	}

	rctx.AppLovinMax = updateAppLovinMaxResponse(rctx, payload.BidResponse)
	rctx.GoogleSDK.Reject = googlesdk.SetGoogleSDKResponseReject(rctx, payload.BidResponse)
	rctx.UnityLevelPlay.Reject = unitylevelplay.SetUnityLevelPlayResponseReject(rctx, payload.BidResponse)

	if rctx.Endpoint == models.EndpointWebS2S {
		result.ChangeSet.AddMutation(func(ap hookstage.AuctionResponsePayload) (hookstage.AuctionResponsePayload, error) {
			rctx, ok := utils.GetRequestContext(moduleCtx)
			if !ok {
				result.Errors = append(result.Errors, "error: request-ctx not found in handleAuctionResponseHook mutation")
				return ap, nil
			}

			var err error
			ap.BidResponse, err = tracker.InjectTrackers(rctx, ap.BidResponse)

			if rctx.NewReqExt != nil && rctx.NewReqExt.Prebid.GoogleSSUFeatureEnabled && rctx.Endpoint == models.EndpointVAST {
				feature.EnrichVASTForSSUFeature(ap.BidResponse, parser.GetTrackerInjector())
			}
			return ap, err
		}, hookstage.MutationUpdate, "response-body-with-webs2s-format")
		return result, nil
	}

	result.ChangeSet.AddMutation(func(ap hookstage.AuctionResponsePayload) (hookstage.AuctionResponsePayload, error) {
		rctx, ok := utils.GetRequestContext(moduleCtx)
		if !ok {
			result.Errors = append(result.Errors, "error: request-ctx not found in handleAuctionResponseHook mutation")
			return ap, nil
		}

		var err error
		ap.BidResponse, err = m.updateORTBV25Response(rctx, ap.BidResponse)
		if err != nil {
			return ap, err
		}

		ap.BidResponse, err = tracker.InjectTrackers(rctx, ap.BidResponse)
		if err != nil {
			return ap, err
		}
		if rctx.NewReqExt != nil && rctx.NewReqExt.Prebid.GoogleSSUFeatureEnabled && rctx.Endpoint == models.EndpointVAST {
			feature.EnrichVASTForSSUFeature(ap.BidResponse, parser.GetTrackerInjector())
		}

		var responseExtjson json.RawMessage
		responseExtjson, err = json.Marshal(responseExt)
		if err != nil {
			result.Errors = append(result.Errors, "failed to marshal response.ext err: "+err.Error())
		}
		ap.BidResponse, err = m.applyDefaultBids(rctx, ap.BidResponse)
		ap.BidResponse.Ext = responseExtjson

		ap.BidResponse = googlesdk.ApplyGoogleSDKResponse(rctx, ap.BidResponse)

		ap.BidResponse = unitylevelplay.ApplyUnityLevelPlayResponse(rctx, ap.BidResponse)
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

// validateModuleContext validates that required context is available
func validateModuleContextAuctionResponseHook(moduleCtx hookstage.ModuleInvocationContext) (models.RequestCtx, endpointmanager.EndpointHookManager, hookstage.HookResult[hookstage.AuctionResponsePayload], bool) {
	result := hookstage.HookResult[hookstage.AuctionResponsePayload]{}
	result.ChangeSet = hookstage.ChangeSet[hookstage.AuctionResponsePayload]{}

	if moduleCtx.ModuleContext == nil {
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in auctionresponsehook()")
		return models.RequestCtx{}, nil, result, false
	}

	rContext, ok := moduleCtx.ModuleContext.Get("rctx")
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in auctionresponsehook()")
		return models.RequestCtx{}, nil, result, false
	}
	rCtx, ok := rContext.(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in auctionresponsehook()")
		return models.RequestCtx{}, nil, result, false
	}

	hookManager, ok := moduleCtx.ModuleContext.Get("endpointhookmanager")
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: endpoint-hook-manager not found in auctionresponsehook()")
		return rCtx, nil, result, false
	}
	endpointHookManager, ok := hookManager.(endpointmanager.EndpointHookManager)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: endpoint-hook-manager not found in auctionresponsehook()")
		return rCtx, nil, result, false
	}

	return rCtx, endpointHookManager, result, true
}

func populateBidExt(rctx *models.RequestCtx, result hookstage.HookResult[hookstage.AuctionResponsePayload], bidResponse *openrtb2.BidResponse) (hookstage.HookResult[hookstage.AuctionResponsePayload], bool) {
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			rctx.MetricsEngine.RecordPlatformPublisherPartnerResponseStats(rctx.Platform, rctx.PubIDStr, seatBid.Seat)

			impId := bid.ImpID
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

			bidExt := &models.BidExt{}
			if len(bid.Ext) != 0 {
				err := jsonutil.Unmarshal(bid.Ext, bidExt)
				if err != nil {
					result.Errors = append(result.Errors, "failed to unmarshal bid.ext for "+utils.GetOriginalBidId(bid.ID))
					// continue
				}
			}

			// Explicitly set the bid.ext.mbmfv value if it is present in the bid.ext since we need it for logging but do not want it in the response
			mbmfv, err := jsonparser.GetFloat(bid.Ext, models.MultiBidMultiFloorValue)
			if err == nil && mbmfv > 0 {
				bidExt.MultiBidMultiFloorValue = mbmfv
			}

			if bidExt.InBannerVideo {
				rctx.MetricsEngine.RecordIBVRequest(rctx.PubIDStr, rctx.ProfileIDStr)
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

			if bidExt.CreativeType != string(openrtb_ext.BidTypeBanner) {
				bidExt.ClickTrackers = nil
			}

			// set response netecpm and logger/tracker en
			var eg, en float64
			revShare := models.GetRevenueShare(rctx.PartnerConfigMap[partnerID])
			bidExt.NetECPM = models.ToFixed(bid.Price, models.BID_PRECISION)
			eg = models.GetGrossEcpmFromNetEcpm(bid.Price, revShare)
			en = bidExt.NetECPM
			if bidResponse.Cur != "USD" {
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
					bidExt.Video.ClientConfig = adunitconfig.GetClientConfigForMediaType(*rctx, impId, "video")
				}
			} else if impCtx.IsBanner && bidExt.CreativeType == "banner" && rctx.ClientConfigFlag == 1 {
				cc := adunitconfig.GetClientConfigForMediaType(*rctx, impId, "banner")
				if len(cc) != 0 {
					if bidExt.Banner == nil {
						bidExt.Banner = &models.ExtBidBanner{}
					}
					bidExt.Banner.ClientConfig = cc
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

	return result, true
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
				if rctx.WinningBids.IsWinningBid(bid.ImpID, bid.ID) {
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
			impCtx, ok := rctx.ImpBidCtx[bid.ImpID]
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
