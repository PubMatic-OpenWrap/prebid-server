package openwrap

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/hooks/hookanalytics"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/adpod/auction"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/adunitconfig"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/tracker"
	"github.com/prebid/prebid-server/openrtb_ext"
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
		result.DebugMessages = append(result.DebugMessages, "error: module-ctx not found in handleBeforeValidationHook()")
		return result, nil
	}
	rctx, ok := moduleCtx.ModuleContext["rctx"].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleBeforeValidationHook()")
		return result, nil
	}
	defer func() {
		moduleCtx.ModuleContext["rctx"] = rctx
		m.metricEngine.RecordPublisherResponseTimeStats(rctx.PubIDStr, int(time.Since(time.Unix(rctx.StartTime, 0)).Milliseconds()))
	}()

	RecordPublisherPartnerNoCookieStats(rctx)

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

	// if payload.BidResponse.NBR != nil {
	// 	return result, nil
	// }

	var winningAdpodBidIds map[string][]string
	var errs []error
	if rctx.IsCTVRequest {
		winningAdpodBidIds, errs = auction.FormAdpodBidsAndPerformExclusion(payload.BidResponse, rctx.ImpBidCtx)
		if len(errs) > 0 {
			for i := range errs {
				result.Errors = append(result.Errors, errs[i].Error())
			}
		}
	}

	winningBids := models.WinningBids{}
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
			if bidderMeta, ok := impCtx.Bidders[seatBid.Seat]; ok {
				partnerID = bidderMeta.PartnerID
			}

			revShare := models.GetRevenueShare(rctx.PartnerConfigMap[partnerID])
			price := bid.Price

			bidId := bid.ID
			bidExt := &models.BidExt{}
			if len(bid.Ext) != 0 { //NYC_TODO: most of the fields should be filled even if unmarshal fails
				err := json.Unmarshal(bid.Ext, bidExt)
				if err != nil {
					result.Errors = append(result.Errors, "failed to unmarshal bid.ext for "+bidId)
					// continue
				}

				if bidExt.Prebid != nil && len(bidExt.Prebid.BidId) > 0 {
					bidId = bidExt.Prebid.BidId
				}

				// NYC_TODO: fix this in PBS-Core or ExecuteAllProcessedBidResponsesStage
				if bidExt.Prebid != nil && bidExt.Prebid.Video != nil && bidExt.Prebid.Video.Duration == 0 &&
					bidExt.Prebid.Video.PrimaryCategory == "" && bidExt.Prebid.Video.VASTTagID == "" {
					bidExt.Prebid.Video = nil
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
					bidExt.CreativeType = models.GetAdFormat(bid.AdM)
				}

				if payload.BidResponse.Cur != "USD" {
					price = bidExt.OriginalBidCPMUSD
				}

				bidExt.NetECPM = models.GetNetEcpm(price, revShare)

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
			}

			bidDealTierSatisfied := false
			if bidExt.Prebid != nil {
				bidDealTierSatisfied = bidExt.Prebid.DealTierSatisfied
			}

			owbid := models.OwBid{
				ID:                   bidId,
				NetEcpm:              bidExt.NetECPM,
				BidDealTierSatisfied: bidDealTierSatisfied,
			}

			if rctx.IsCTVRequest && impCtx.AdpodConfig != nil {
				if CheckWinningBidId(bidId, winningAdpodBidIds[impId]) {
					winningBids.AppendBid(impId, owbid)
				}
			} else {
				winningBids.AddBid(impId, owbid, rctx.SupportDeals)
			}

			if rctx.IsCTVRequest {
				bidExt.AdPod.Targeting = GetTargettingForAdpod(bid, rctx.PartnerConfigMap[models.VersionLevelConfigID], impCtx, bidExt, seatBid.Seat)
				if rctx.Debug {
					bidExt.AdPod.Debug.Targeting = GetTargettingForDebug(bid.ID, rctx.PubIDStr, rctx.ProfileIDStr, fmt.Sprint(rctx.DisplayID), impCtx.TagID, bidExt.NetECPM)
				}
			}

			// cache for bid details for logger and tracker
			if impCtx.BidCtx == nil {
				impCtx.BidCtx = make(map[string]models.BidCtx)
			}
			impCtx.BidCtx[bidId] = models.BidCtx{
				BidExt: *bidExt,
			}
			rctx.ImpBidCtx[impId] = impCtx
		}
	}

	rctx.WinningBids = winningBids
	if len(winningBids) == 0 {
		m.metricEngine.RecordNobidErrPrebidServerResponse(rctx.PubIDStr)
	}

	droppedBids, warnings := addPWTTargetingForBid(rctx, payload.BidResponse)
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

	rctx.DefaultBids = m.addDefaultBids(rctx, payload.BidResponse, &responseExt)

	rctx.Trackers = tracker.CreateTrackers(rctx, payload.BidResponse)

	for bidder, responseTimeMs := range responseExt.ResponseTimeMillis {
		rctx.BidderResponseTimeMillis[bidder.String()] = responseTimeMs
		m.metricEngine.RecordPartnerResponseTimeStats(rctx.PubIDStr, string(bidder), responseTimeMs)
	}

	// TODO: PBS-Core should pass the hostcookie for module to usersync.ParseCookieFromRequest()
	if matchedImpression := getMatchedImpression(rctx); matchedImpression != nil {
		responseExt.OwMatchedImpression = matchedImpression
	}

	if rctx.SendAllBids {
		responseExt.OwSendAllBids = 1
	}

	if rctx.LogInfoFlag == 1 {
		responseExt.OwLogInfo = &openrtb_ext.OwLogInfo{
			// Logger:  openwrap.GetLogAuctionObjectAsURL(ao, true, true), updated done later
			Tracker: tracker.GetTrackerInfo(rctx),
		}
	}

	if rctx.ReturnAllBidStatus {
		// prepare seat-non-bids and add them in the response-ext
		rctx.SeatNonBids = prepareSeatNonBids(rctx)
		addSeatNonBidsInResponseExt(rctx, &responseExt)
	}

	var err error
	rctx.ResponseExt, err = json.Marshal(responseExt)
	if err != nil {
		result.Errors = append(result.Errors, "failed to marshal response.ext err: "+err.Error())
	}

	if rctx.Debug {
		rCtxBytes, _ := json.Marshal(rctx)
		result.DebugMessages = append(result.DebugMessages, string(rCtxBytes))
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

		ap.BidResponse, err = m.applyDefaultBids(rctx, ap.BidResponse)

		ap.BidResponse.Ext = rctx.ResponseExt
		return ap, err
	}, hookstage.MutationUpdate, "response-body-with-sshb-format")

	// TODO: move debug here
	// result.ChangeSet.AddMutation(func(ap hookstage.AuctionResponsePayload) (hookstage.AuctionResponsePayload, error) {
	// }, hookstage.MutationUpdate, "response-body-with-sshb-format")

	return result, nil
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
				bid.ImpID, _ = models.GetImpressionID(bid.ImpID)
				bidId := bid.ID
				bidExt := &models.BidExt{}
				if len(bid.Ext) > 0 {
					_ = json.Unmarshal(bid.Ext, bidExt)

					if bidExt.Prebid != nil && len(bidExt.Prebid.BidId) > 0 {
						bidId = bidExt.Prebid.BidId
					}
				}
				if rctx.WinningBids.IsWinningBid(bid.ImpID, bidId) {
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

			bidId := bid.ID
			bidExt := &models.BidExt{}
			if len(bid.Ext) > 0 {
				_ = json.Unmarshal(bid.Ext, bidExt)

				if bidExt.Prebid != nil && len(bidExt.Prebid.BidId) > 0 {
					bidId = bidExt.Prebid.BidId
				}
			}

			bidCtx, ok := impCtx.BidCtx[bidId]
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

func getIntPtr(i int) *int {
	return &i
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

func GetTargettingForDebug(bidId, pubID, profileID, versionID, tagID string, ecpm float64) map[string]string {
	targeting := make(map[string]string)

	targeting[models.PwtBidID] = bidId
	targeting[models.PWT_CACHE_PATH] = models.AMP_CACHE_PATH
	targeting[models.PWT_ECPM] = fmt.Sprintf("%.2f", ecpm)
	targeting[models.PWT_PUBID] = pubID
	targeting[models.PWT_SLOTID] = tagID
	targeting[models.PWT_PROFILEID] = profileID

	if targeting[models.PWT_ECPM] == "" {
		targeting[models.PWT_ECPM] = "0"
	}

	if versionID != "0" {
		targeting[models.PWT_VERSIONID] = versionID
	}

	return targeting
}

func GetTargettingForAdpod(bid openrtb2.Bid, partnerConfig map[string]string, impCtx models.ImpCtx, bidExt *models.BidExt, seat string) map[string]string {
	targetingKeyValMap := make(map[string]string)
	targetingKeyValMap[models.PWT_PARTNERID] = seat

	if bidExt != nil {
		if bidExt.Prebid != nil {
			if bidExt.Prebid.Video != nil && bidExt.Prebid.Video.Duration > 0 {
				targetingKeyValMap[models.PWT_DURATION] = strconv.Itoa(bidExt.Prebid.Video.Duration)
			}

			prefix, _, _, err := jsonparser.Get(impCtx.NewExt, "prebid", "bidder", seat, "dealtier", "prefix")
			if bidExt.Prebid.DealTierSatisfied && partnerConfig[models.DealTierLineItemSetup] == "1" && err == nil && len(prefix) > 0 {
				targetingKeyValMap[models.PwtDealTier] = fmt.Sprintf("%s%d", string(prefix), bidExt.Prebid.DealPriority)
			} else if len(bid.DealID) > 0 && partnerConfig[models.DealIDLineItemSetup] == "1" {
				targetingKeyValMap[models.PWT_DEALID] = bid.DealID
			} else {
				priceBucket, ok := bidExt.Prebid.Targeting[string(openrtb_ext.HbpbConstantKey)]
				if ok {
					targetingKeyValMap[models.PwtPb] = priceBucket
				}
			}

			catDur, ok := bidExt.Prebid.Targeting[models.PwtPbCatDur]
			if ok {
				cat, dur := getCatAndDurFromPwtCatDur(catDur)
				if len(cat) > 0 {
					targetingKeyValMap[models.PwtCat] = cat
				}

				if len(dur) > 0 && targetingKeyValMap[models.PWT_DURATION] == "" {
					targetingKeyValMap[models.PWT_DURATION] = dur
				}
			}
		}
	}

	return targetingKeyValMap
}

func getCatAndDurFromPwtCatDur(pwtCatDur string) (string, string) {
	arr := strings.Split(pwtCatDur, "_")
	if len(arr) == 2 {
		return "", TrimRightByte(arr[1], 's')
	}
	if len(arr) == 3 {
		return arr[1], TrimRightByte(arr[2], 's')
	}
	return "", ""
}

func TrimRightByte(s string, b byte) string {
	if s[len(s)-1] == b {
		return s[:len(s)-1]
	}
	return s
}
