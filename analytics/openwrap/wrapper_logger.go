package openwrap

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strconv"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/hooks/hookexecution"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func GetLogAuctionObjectAsURL(ao analytics.AuctionObject, logInfo bool) string {
	defer func() {
		if r := recover(); r != nil {
			glog.Error(string(debug.Stack()))
		}
	}()

	rCtx := GetRequestCtx(ao.HookExecutionOutcome)
	if rCtx == nil {
		return ""
	}

	wlog := WloggerRecord{
		record: record{
			PubID:             rCtx.PubID,
			ProfileID:         fmt.Sprintf("%d", rCtx.ProfileID),
			VersionID:         fmt.Sprintf("%d", rCtx.DisplayID),
			Origin:            rCtx.Origin,
			PageURL:           rCtx.PageURL,
			IID:               rCtx.LoggerImpressionID,
			Timestamp:         rCtx.StartTime,
			ServerLogger:      1,
			TestConfigApplied: rCtx.ABTestConfigApplied,
			Timeout:           int(ao.Request.TMax),
		},
	}

	extWrapper := models.RequestExtWrapper{}
	err := json.Unmarshal(ao.Request.Ext, &extWrapper)
	if err != nil {
		return ""
	}

	if ao.Request.User != nil {
		extUser := openrtb_ext.ExtUser{}
		_ = json.Unmarshal(ao.Request.User.Ext, &extUser)
		wlog.ConsentString = extUser.Consent
	}

	if ao.Request.Device != nil {
		wlog.IP = ao.Request.Device.IP
		wlog.UserAgent = ao.Request.Device.UA
	}

	if ao.Request.Regs != nil {
		extReg := openrtb_ext.ExtRegs{}
		_ = json.Unmarshal(ao.Request.Regs.Ext, &extReg)
		if extReg.GDPR != nil {
			wlog.GDPR = *extReg.GDPR
		}
	}

	//log device object
	wlog.logDeviceObject(*rCtx, rCtx.UA, ao.Request, rCtx.Platform)

	//log content object
	if nil != ao.Request.Site {
		wlog.logContentObject(ao.Request.Site.Content)
	} else if nil != ao.Request.App {
		wlog.logContentObject(ao.Request.App.Content)
	}

	var ipr map[string][]PartnerRecord

	if logInfo {
		ipr = getDefaultPartnerRecordsByImp(rCtx)
	} else {
		ipr = getPartnerRecordsByImp(ao, rCtx)
	}

	// parent bidder could in one of the above and we need them by prebid's bidderCode and not seat(could be alias)
	slots := make([]SlotRecord, 0)
	for _, imp := range ao.Request.Imp {
		reward := 0
		var incomingSlots []string
		if impCtx, ok := rCtx.ImpBidCtx[imp.ID]; ok {
			if impCtx.IsRewardInventory != nil {
				reward = int(*impCtx.IsRewardInventory)
			}
			incomingSlots = impCtx.IncomingSlots
		}

		slots = append(slots, SlotRecord{
			SlotName:          getSlotName(imp.ID, imp.TagID),
			SlotSize:          incomingSlots,
			Adunit:            imp.TagID,
			PartnerData:       ipr[imp.ID],
			RewardedInventory: int(reward),
			// AdPodSlot:         getAdPodSlot(imp, responseMap.AdPodBidsExt),
		})
	}

	wlog.Slots = slots

	return PrepareLoggerURL(&wlog, rCtx.URL, GetGdprEnabledFlag(rCtx.PartnerConfigMap))
}

// TODO filter by name. (*stageOutcomes[8].Groups[0].InvocationResults[0].AnalyticsTags.Activities[0].Results[0].Values["request-ctx"].(data))
func GetRequestCtx(hookExecutionOutcome []hookexecution.StageOutcome) *models.RequestCtx {
	for _, stageOutcome := range hookExecutionOutcome {
		for _, groups := range stageOutcome.Groups {
			for _, invocationResult := range groups.InvocationResults {
				for _, activity := range invocationResult.AnalyticsTags.Activities {
					for _, result := range activity.Results {
						if result.Values != nil {
							if irctx, ok := result.Values["request-ctx"]; ok {
								rctx, ok := irctx.(*models.RequestCtx)
								if !ok {
									return nil
								}
								return rctx
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func getPartnerRecordsByImp(ao analytics.AuctionObject, rCtx *models.RequestCtx) map[string][]PartnerRecord {
	// impID-partnerRecords: partner records per impression
	ipr := make(map[string][]PartnerRecord)

	// Seat-impID
	rejectedBids := map[string]map[string]struct{}{}
	loggerSeat := make(map[string][]openrtb2.Bid)
	for _, seatBids := range ao.RejectedBids {
		if _, ok := rejectedBids[seatBids.Seat]; !ok {
			rejectedBids[seatBids.Seat] = map[string]struct{}{}
		}

		if seatBids.Bid != nil && seatBids.Bid.Bid != nil {
			rejectedBids[seatBids.Seat][seatBids.Bid.Bid.ImpID] = struct{}{}

			bidExt := models.BidExt{}
			_ = json.Unmarshal(seatBids.Bid.Bid.Ext, &bidExt)
			bidExt.OriginalBidCPM = seatBids.Bid.OriginalBidCPM
			bidExt.OriginalBidCPMUSD = seatBids.Bid.OriginalBidCPMUSD
			bidExt.OriginalBidCur = seatBids.Bid.OriginalBidCur
			if bidExt.Prebid == nil {
				bidExt.Prebid = &openrtb_ext.ExtBidPrebid{}
			}
			bidExt.Prebid.Floors = seatBids.Bid.BidFloors
			bidExt.Prebid.DealPriority = seatBids.Bid.DealPriority
			bidExt.Prebid.Meta = seatBids.Bid.BidMeta
			bidExt.Prebid.Video = seatBids.Bid.BidVideo

			loggerSeat[seatBids.Seat] = append(loggerSeat[seatBids.Seat], *seatBids.Bid.Bid)
		}
	}
	for _, seatBid := range ao.Response.SeatBid {
		for _, bid := range seatBid.Bid {
			// Check if this is a default bid of the RejectedBids
			if bid.Price == 0 && bid.W == 0 && bid.H == 0 {
				if _, ok := rejectedBids[seatBid.Seat]; ok {
					if _, ok := rejectedBids[seatBid.Seat][bid.ImpID]; ok {
						continue
					}
				}
			}
			loggerSeat[seatBid.Seat] = append(loggerSeat[seatBid.Seat], bid)
		}
	}
	for seat, Bids := range rCtx.DroppedBids {
		loggerSeat[seat] = append(loggerSeat[seat], Bids...)
	}

	// pubmatic's KGP details per impression
	type pubmaticMarketplaceMeta struct {
		PubmaticKGP, PubmaticKGPV, PubmaticKGPSV string
	}
	pmMkt := make(map[string]pubmaticMarketplaceMeta)

	for seat, bids := range loggerSeat {
		if seat == string(openrtb_ext.BidderOWPrebidCTV) {
			continue
		}

		if _, ok := rCtx.AdapterThrottleMap[seat]; ok {
			continue
		}

		for _, bid := range bids {
			impCtx, ok := rCtx.ImpBidCtx[bid.ImpID]
			if !ok {
				continue
			}

			if _, ok := impCtx.NonMapped[seat]; ok {
				break
			}

			revShare := 0.0
			partnerID := seat
			var isRegex bool
			var kgp, kgpv, kgpsv string

			if bidderMeta, ok := impCtx.Bidders[seat]; ok {
				revShare, _ = strconv.ParseFloat(rCtx.PartnerConfigMap[bidderMeta.PartnerID][models.REVSHARE], 64)
				partnerID = bidderMeta.PrebidBidderCode
				kgp = bidderMeta.KGP
				kgpv = bidderMeta.KGPV
				kgpsv = bidderMeta.MatchedSlot
				isRegex = bidderMeta.IsRegex
			}

			// 1. nobid
			if bid.Price == 0 && bid.H == 0 && bid.W == 0 {
				//NOTE: kgpsv = bidderMeta.MatchedSlot above. Use the same
				if !isRegex && kgpv != "" { // unmapped pubmatic's slot
					kgpsv = kgpv
				} else if !isRegex {
					kgpv = kgpsv
				}
			} else if !isRegex {
				if kgpv != "" { // unmapped pubmatic's slot
					kgpsv = kgpv
				} else if bid.H != 0 && bid.W != 0 { // Check when bid.H and bid.W will be zero with Price !=0. Ex: MobileInApp-MultiFormat-OnlyBannerMapping_Criteo_Partner_Validaton
					// 2. valid bid
					// kgpv has regex, do not generate slotName again
					// kgpsv could be unmapped or mapped slot, generate slotName again based on bid.H and bid.W
					kgpsv = GenerateSlotName(bid.H, bid.W, kgp, impCtx.TagID, impCtx.Div, rCtx.Source)
					kgpv = kgpsv
				}
			}

			if kgpv == "" {
				kgpv = kgpsv
			}

			bidExt := models.BidExt{}
			_ = json.Unmarshal(bid.Ext, &bidExt)

			price := bid.Price
			if ao.Response.Cur != "USD" {
				price = bidExt.OriginalBidCPMUSD
			}

			if seat == "pubmatic" {
				pmMkt[bid.ImpID] = pubmaticMarketplaceMeta{
					PubmaticKGP:   kgp,
					PubmaticKGPV:  kgpv,
					PubmaticKGPSV: kgpsv,
				}
			}

			pr := PartnerRecord{
				PartnerID:  partnerID, // prebid biddercode
				BidderCode: seat,      // pubmatic biddercode: pubmatic2
				// AdapterCode: adapterCode, // prebid adapter that brought the bid
				Latency1:         rCtx.BidderResponseTimeMillis[seat],
				KGPV:             kgpv,
				KGPSV:            kgpsv,
				BidID:            bid.ID,
				OrigBidID:        bid.ID,
				DefaultBidStatus: 0,
				ServerSide:       1,
				// MatchedImpression: matchedImpression,
				NetECPM: func() float64 {
					if revShare != 0.0 {
						return GetNetEcpm(price, revShare)
					}
					return price
				}(),
				GrossECPM:   GetGrossEcpm(price),
				OriginalCPM: GetGrossEcpm(bidExt.OriginalBidCPM),
				OriginalCur: bidExt.OriginalBidCur,
				PartnerSize: getSizeForPlatform(bid.W, bid.H, rCtx.Platform),
				DealID:      bid.DealID,
			}

			// don't want default banner for nobid in wl
			if len(bid.AdM) != 0 {
				pr.Adformat = models.GetAdFormat(bid.AdM)
			}

			if b, ok := rCtx.WinningBids[bid.ImpID]; ok && b.ID == bid.ID {
				pr.WinningBidStaus = 1
			}

			if len(pr.OriginalCur) == 0 {
				pr.OriginalCPM = float64(0)
				pr.OriginalCur = "USD"
			}

			if len(pr.DealID) != 0 {
				pr.DealChannel = models.DEFAULT_DEALCHANNEL
			}

			if bidExt.Prebid != nil {
				if bidExt.Prebid.DealTierSatisfied && bidExt.Prebid.DealPriority > 0 {
					pr.DealPriority = bidExt.Prebid.DealPriority
				}

				if bidExt.Prebid.Video != nil && bidExt.Prebid.Video.Duration > 0 {
					pr.AdDuration = &bidExt.Prebid.Video.Duration
				}

				if bidExt.Prebid.Meta != nil {
					pr.setMetaDataObject(bidExt.Prebid.Meta)
				}

				if bidExt.Prebid.Floors != nil {
					pr.FloorRule = bidExt.Prebid.Floors.FloorRule
					pr.FloorRuleValue = roundToTwoDigit(bidExt.Prebid.Floors.FloorRuleValue)
					if bidExt.Prebid.Floors.FloorCurrency == "USD" {
						pr.FloorValue = roundToTwoDigit(bidExt.Prebid.Floors.FloorValue)
					} else {
						pr.FloorValue = roundToTwoDigit(bidExt.Prebid.Floors.FloorValueUSD)
					}
				}
			}

			if len(bid.ADomain) != 0 {
				if domain, err := ExtractDomain(bid.ADomain[0]); err == nil {
					pr.ADomain = domain
				}
			}

			ipr[bid.ImpID] = append(ipr[bid.ImpID], pr)
		}
	}

	// overwrite marketplace bid details with that of partner adatper
	if rCtx.MarketPlaceBidders != nil {
		for impID, partnerRecords := range ipr {
			for i := 0; i < len(partnerRecords); i++ {
				if _, ok := rCtx.MarketPlaceBidders[partnerRecords[i].BidderCode]; ok {
					partnerRecords[i].PartnerID = "pubmatic"
					partnerRecords[i].KGPV = pmMkt[impID].PubmaticKGPV
					partnerRecords[i].KGPSV = pmMkt[impID].PubmaticKGPSV
				}
			}
			ipr[impID] = partnerRecords
		}
	}

	return ipr
}

func getDefaultPartnerRecordsByImp(rCtx *models.RequestCtx) map[string][]PartnerRecord {
	ipr := make(map[string][]PartnerRecord)
	for impID := range rCtx.ImpBidCtx {
		ipr[impID] = []PartnerRecord{{
			ServerSide:       1,
			DefaultBidStatus: 1,
			PartnerSize:      "0x0",
		}}
	}
	return ipr
}
