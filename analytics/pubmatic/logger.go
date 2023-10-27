package pubmatic

import (
	"encoding/json"
	"fmt"
	"strings"

	"net/http"
	"strconv"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/hooks/hookexecution"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/utils"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func GetLogAuctionObjectAsURL(ao analytics.AuctionObject, rCtx *models.RequestCtx, logInfo, forRespExt bool) (string, http.Header) {

	wlog := WloggerRecord{
		record: record{
			PubID:             rCtx.PubID,
			ProfileID:         fmt.Sprintf("%d", rCtx.ProfileID),
			VersionID:         fmt.Sprintf("%d", rCtx.VersionID),
			Origin:            rCtx.Origin,
			PageURL:           rCtx.PageURL,
			IID:               rCtx.LoggerImpressionID,
			Timestamp:         rCtx.StartTime,
			ServerLogger:      1,
			TestConfigApplied: rCtx.ABTestConfigApplied,
			Timeout:           int(rCtx.TMax),
			PDC:               rCtx.DCName,
			CachePutMiss:      rCtx.CachePutMiss,
		},
	}

	requestExt := openrtb_ext.ExtRequest{}
	err := json.Unmarshal(ao.RequestWrapper.BidRequest.Ext, &requestExt)
	if err == nil {
		wlog.logFloorType(&requestExt.Prebid)
	}

	wlog.logIntegrationType(rCtx.Endpoint)
	wlog.logDeviceObject(rCtx, ao.RequestWrapper.BidRequest)

	if ao.RequestWrapper.User != nil {
		extUser := openrtb_ext.ExtUser{}
		_ = json.Unmarshal(ao.RequestWrapper.User.Ext, &extUser)
		wlog.ConsentString = extUser.Consent
	}

	if ao.RequestWrapper.Regs != nil {
		extReg := openrtb_ext.ExtRegs{}
		_ = json.Unmarshal(ao.RequestWrapper.Regs.Ext, &extReg)
		if extReg.GDPR != nil {
			wlog.GDPR = *extReg.GDPR
		}
	}

	if ao.RequestWrapper.Site != nil {
		wlog.logContentObject(ao.RequestWrapper.Site.Content)
	} else if ao.RequestWrapper.App != nil {
		wlog.logContentObject(ao.RequestWrapper.App.Content)
	}

	var ipr map[string][]PartnerRecord

	if logInfo {
		ipr = getDefaultPartnerRecordsByImp(rCtx)
	} else {
		ipr = getPartnerRecordsByImp(ao, rCtx)
	}

	// parent bidder could in one of the above and we need them by prebid's bidderCode and not seat(could be alias)
	slots := make([]SlotRecord, 0)
	for _, imp := range ao.RequestWrapper.Imp {
		reward := 0
		var incomingSlots []string
		if impCtx, ok := rCtx.ImpBidCtx[imp.ID]; ok {
			if impCtx.IsRewardInventory != nil {
				reward = int(*impCtx.IsRewardInventory)
			}
			incomingSlots = impCtx.IncomingSlots
		}

		// to keep existing response intact
		partnerData := make([]PartnerRecord, 0)
		if ipr[imp.ID] != nil {
			partnerData = ipr[imp.ID]
		}

		slots = append(slots, SlotRecord{
			SlotName:          getSlotName(imp.ID, imp.TagID),
			SlotSize:          incomingSlots,
			Adunit:            imp.TagID,
			PartnerData:       partnerData,
			RewardedInventory: int(reward),
			// AdPodSlot:         getAdPodSlot(imp, responseMap.AdPodBidsExt),
		})
	}

	wlog.Slots = slots

	headers := http.Header{
		models.USER_AGENT_HEADER: []string{rCtx.UA},
		models.IP_HEADER:         []string{rCtx.IP},
	}

	// TODO : confirm this header is not sent in HB ? do we need it here
	// if rCtx.KADUSERCookie != nil {
	// 	headers.Add(models.KADUSERCOOKIE, rCtx.KADUSERCookie.Value)
	// }

	var responseExt openrtb_ext.ExtBidResponse
	err = json.Unmarshal(ao.Response.Ext, &responseExt)
	if err == nil {
		if responseExt.Prebid != nil {
			wlog.SetFloorDetails(responseExt.Prebid.Floors)
		}
	}

	url := ow.cfg.Endpoint
	if logInfo {
		url = ow.cfg.PublicEndpoint
	}

	return PrepareLoggerURL(&wlog, url, GetGdprEnabledFlag(rCtx.PartnerConfigMap)), headers
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

type BidWrapper struct {
	*openrtb2.Bid
	Nbr *openrtb3.NonBidStatusCode
}

func convertNonBidToBidWrapper(nonBid *openrtb_ext.NonBid) (bid BidWrapper) {

	bid.Bid = &openrtb2.Bid{}
	bid.Price = nonBid.Ext.Prebid.Bid.Price
	bid.ADomain = nonBid.Ext.Prebid.Bid.ADomain
	bid.CatTax = nonBid.Ext.Prebid.Bid.CatTax
	bid.Cat = nonBid.Ext.Prebid.Bid.Cat
	bid.DealID = nonBid.Ext.Prebid.Bid.DealID
	bid.W = nonBid.Ext.Prebid.Bid.W
	bid.H = nonBid.Ext.Prebid.Bid.H
	bid.Dur = nonBid.Ext.Prebid.Bid.Dur
	bid.MType = nonBid.Ext.Prebid.Bid.MType
	bid.ID = nonBid.Ext.Prebid.Bid.ID
	bid.ImpID = nonBid.ImpId

	bidExt := models.BidExt{}
	bidExt.OriginalBidCPM = nonBid.Ext.Prebid.Bid.OriginalBidCPM
	bidExt.OriginalBidCPMUSD = nonBid.Ext.Prebid.Bid.OriginalBidCPMUSD
	bidExt.OriginalBidCur = nonBid.Ext.Prebid.Bid.OriginalBidCur
	bidExt.Prebid = new(openrtb_ext.ExtBidPrebid)
	bidExt.Prebid.DealPriority = nonBid.Ext.Prebid.Bid.DealPriority
	bidExt.Prebid.DealTierSatisfied = nonBid.Ext.Prebid.Bid.DealTierSatisfied
	bidExt.Prebid.Meta = nonBid.Ext.Prebid.Bid.Meta
	bidExt.Prebid.Targeting = nonBid.Ext.Prebid.Bid.Targeting
	bidExt.Prebid.Type = nonBid.Ext.Prebid.Bid.Type
	bidExt.Prebid.Video = nonBid.Ext.Prebid.Bid.Video
	bidExt.Prebid.BidId = nonBid.Ext.Prebid.Bid.BidId
	bidExt.Prebid.Floors = nonBid.Ext.Prebid.Bid.Floors
	// bidExt.Nbr = openwrap.GetNonBidStatusCodePtr(openrtb3.NonBidStatusCode(nonBid.StatusCode))

	bidExtBytes, err := json.Marshal(bidExt)
	if err == nil {
		bid.Ext = bidExtBytes
	}
	bid.Nbr = openwrap.GetNonBidStatusCodePtr(openrtb3.NonBidStatusCode(nonBid.StatusCode))
	return bid
}

func convertNonBidToBid(nonBid *openrtb_ext.NonBid) (bid openrtb2.Bid) {

	bid.Price = nonBid.Ext.Prebid.Bid.Price
	bid.ADomain = nonBid.Ext.Prebid.Bid.ADomain
	bid.CatTax = nonBid.Ext.Prebid.Bid.CatTax
	bid.Cat = nonBid.Ext.Prebid.Bid.Cat
	bid.DealID = nonBid.Ext.Prebid.Bid.DealID
	bid.W = nonBid.Ext.Prebid.Bid.W
	bid.H = nonBid.Ext.Prebid.Bid.H
	bid.Dur = nonBid.Ext.Prebid.Bid.Dur
	bid.MType = nonBid.Ext.Prebid.Bid.MType
	bid.ID = nonBid.Ext.Prebid.Bid.ID
	bid.ImpID = nonBid.ImpId

	bidExt := models.BidExt{}
	bidExt.OriginalBidCPM = nonBid.Ext.Prebid.Bid.OriginalBidCPM
	bidExt.OriginalBidCPMUSD = nonBid.Ext.Prebid.Bid.OriginalBidCPMUSD
	bidExt.OriginalBidCur = nonBid.Ext.Prebid.Bid.OriginalBidCur
	bidExt.Prebid = new(openrtb_ext.ExtBidPrebid)
	bidExt.Prebid.DealPriority = nonBid.Ext.Prebid.Bid.DealPriority
	bidExt.Prebid.DealTierSatisfied = nonBid.Ext.Prebid.Bid.DealTierSatisfied
	bidExt.Prebid.Meta = nonBid.Ext.Prebid.Bid.Meta
	bidExt.Prebid.Targeting = nonBid.Ext.Prebid.Bid.Targeting
	bidExt.Prebid.Type = nonBid.Ext.Prebid.Bid.Type
	bidExt.Prebid.Video = nonBid.Ext.Prebid.Bid.Video
	bidExt.Prebid.BidId = nonBid.Ext.Prebid.Bid.BidId
	bidExt.Prebid.Floors = nonBid.Ext.Prebid.Bid.Floors
	bidExt.Nbr = openwrap.GetNonBidStatusCodePtr(openrtb3.NonBidStatusCode(nonBid.StatusCode))

	bidExtBytes, err := json.Marshal(bidExt)
	if err == nil {
		bid.Ext = bidExtBytes
	}
	return bid
}

func getPartnerRecordsByImp(ao analytics.AuctionObject, rCtx *models.RequestCtx) map[string][]PartnerRecord {
	// impID-partnerRecords: partner records per impression
	ipr := make(map[string][]PartnerRecord)

	// Seat-impID (based on impID as default bids do not have ID). Shall we generate unique ID's for them?
	rejectedBids := map[string]map[string]struct{}{}
	loggerSeat := make(map[string][]BidWrapper)

	// currently, ao.SeatNonBid will contain the bids that got rejected in the prebid core
	// it does not contain the bids that got rejected inside pubmatic ow module
	// As of now, pubmatic ow module logs slot-not-map and partner-throttle related nonbids
	// for which we don't create partner-record in the owlogger
	for _, seatNonBid := range ao.SeatNonBid {
		if _, ok := rejectedBids[seatNonBid.Seat]; !ok {
			rejectedBids[seatNonBid.Seat] = map[string]struct{}{}
		}
		for _, nonBid := range seatNonBid.NonBid {
			rejectedBids[seatNonBid.Seat][nonBid.ImpId] = struct{}{}
			// loggerSeat[seatNonBid.Seat] = append(loggerSeat[seatNonBid.Seat], convertNonBidToBid(&nonBid))
			loggerSeat[seatNonBid.Seat] = append(loggerSeat[seatNonBid.Seat], convertNonBidToBidWrapper(&nonBid))
		}
	}

	// SeatBid contains valid-bids + default/proxy bids
	// loggerSeat should not contain duplicate entry for same imp-seat combination
	for seatIndex, seatBid := range ao.Response.SeatBid {
		for bidIndex, bid := range seatBid.Bid {
			// Check if this is a default as well as nonbid.
			// Ex. if only one bid is returned by pubmatic and it got rejected due to floors.
			// then OW-Module will add one default/proxy bid in seatbid.
			// and prebid core will add one nonbid in seat-non-bid.
			// So, we want to skip this default/proxy bid to avoid duplicate.
			// We don't have bid.ID here so using bid.ImpID
			// if bid.Price == 0 && bid.W == 0 && bid.H == 0 { //TODO ??
			if models.IsDefaultBid(&bid) {
				if _, ok := rejectedBids[seatBid.Seat]; ok {
					if _, ok := rejectedBids[seatBid.Seat][bid.ImpID]; ok {
						continue
					}
				}
			}
			loggerSeat[seatBid.Seat] = append(loggerSeat[seatBid.Seat], BidWrapper{Bid: &ao.Response.SeatBid[seatIndex].Bid[bidIndex]})
		}
	}

	// include bids that got dropped from ao.SeatBid by pubmatic ow module. Ex. sendAllBids=false
	for seat, Bids := range rCtx.DroppedBids {
		for bid := range Bids {
			loggerSeat[seat] = append(loggerSeat[seat], BidWrapper{Bid: &rCtx.DroppedBids[seat][bid]})
		}
	}

	// pubmatic's KGP details per impression
	// This is only required for groupm bids (groupm bids log pubmatic's data in ow logger)
	type pubmaticMarketplaceMeta struct {
		PubmaticKGP, PubmaticKGPV, PubmaticKGPSV string
	}
	pmMkt := make(map[string]pubmaticMarketplaceMeta)

	// loggerSeat contains valid-bids, default-bids, non-bids, dropped-bids
	for seat, bids := range loggerSeat {
		if seat == string(openrtb_ext.BidderOWPrebidCTV) {
			continue
		}

		// owlogger would not contain throttled bidders, do we need this
		if _, ok := rCtx.AdapterThrottleMap[seat]; ok {
			continue
		}

		for _, bid := range bids {
			impCtx, ok := rCtx.ImpBidCtx[bid.ImpID]
			if !ok {
				continue
			}

			// owlogger would not contain non-mapped bidders, do we need this
			if _, ok := impCtx.NonMapped[seat]; ok {
				break
			}

			revShare := 0.0
			partnerID := seat
			var isRegex bool
			var kgp, kgpv, kgpsv, adFormat string

			if bidderMeta, ok := impCtx.Bidders[seat]; ok {
				revShare, _ = strconv.ParseFloat(rCtx.PartnerConfigMap[bidderMeta.PartnerID][models.REVSHARE], 64)
				partnerID = bidderMeta.PrebidBidderCode
				kgp = bidderMeta.KGP           // _AU_@_W_x_H_
				kgpv = bidderMeta.KGPV         // ^/43743431/DMDemo[0-9]*@Div[12]@^728x90$
				kgpsv = bidderMeta.MatchedSlot // /43743431/DMDemo1@@728x90
				isRegex = bidderMeta.IsRegex
			}

			// impBidCtx contains bid.Id in format bid.id::uuid
			var bidExt models.BidExt
			json.Unmarshal(bid.Ext, &bidExt)

			bidIDForLookup := bid.ID
			if bidExt.Prebid != nil && !strings.Contains(bid.ID, models.BidIdSeparator) {
				// this block will not be executed for default-bids.
				bidIDForLookup = utils.SetUniqueBidID(bid.ID, bidExt.Prebid.BidId)
			}

			if bidCtx, ok := impCtx.BidCtx[bidIDForLookup]; ok {
				// override bidExt for seat-bids and default-bids since we have already prepared it under auction-response-hook
				// this block will not be executed for seat-non-bids because we dont create impBidCtx object
				// because bidder alias can return same bid-id for multiple bids and we can not differentiate them.
				bidExt = bidCtx.BidExt
			}

			var nbr *openrtb3.NonBidStatusCode
			nbr = bid.Nbr // only for seat-non-bids this will be non-nil
			if nbr == nil {
				nbr = bidExt.Nbr // valid-bids + default-bids
			}

			// adformat to be derived from prebid.type or bid.AdM
			adFormat = models.GetAdFormat(bid.Bid, &bidExt, &impCtx)

			if models.IsDefaultBid(bid.Bid) {
				//NOTE: kgpsv = bidderMeta.MatchedSlot above. Use the same
				if !isRegex && kgpv != "" { // unmapped pubmatic's slot
					kgpsv = kgpv // - KGP: _AU_@_DIV_@_W_x_H_
				} else if !isRegex {
					kgpv = kgpsv
				}
			} else if !isRegex {
				if kgpv != "" { // unmapped pubmatic's slot
					kgpsv = kgpv // /43743431/DMDemo1234@300x250 -->
				} else if adFormat == models.Video { // Check when adformat is video, bid.W and bid.H has to be zero with Price !=0. Ex: UOE-9222(0x0 default kgpv and kgpsv for video bid)
					// 2. valid video bid
					// kgpv has regex, do not generate slotName again
					// kgpsv could be unmapped or mapped slot, generate slotName with bid.W = bid.H = 0
					kgpsv = GenerateSlotName(0, 0, kgp, impCtx.TagID, impCtx.Div, rCtx.Source)
					kgpv = kgpsv // original /43743431/DMDemo1234@300x250 but new could be /43743431/DMDemo1234@0x0
				} else if bid.H != 0 && bid.W != 0 { // Check when bid.H and bid.W will be zero with Price !=0. Ex: MobileInApp-MultiFormat-OnlyBannerMapping_Criteo_Partner_Validaton
					// 3. valid bid
					// kgpv has regex, do not generate slotName again
					// kgpsv could be unmapped or mapped slot, generate slotName again based on bid.H and bid.W
					kgpsv = GenerateSlotName(bid.H, bid.W, kgp, impCtx.TagID, impCtx.Div, rCtx.Source)
					kgpv = kgpsv // original /43743431/DMDemo1234@300x250 but new could be /43743431/DMDemo1234@222x111
				}
			}

			if kgpv == "" {
				kgpv = kgpsv
			}

			price := bid.Price
			if ao.Response.Cur != "" && ao.Response.Cur != "USD" && bidExt.OriginalBidCPMUSD != 0 {
				price = bidExt.OriginalBidCPMUSD
			}

			// if bidExt.OriginalBidCPMUSD != bid.Price {
			// 	price = bidExt.OriginalBidCPMUSD
			// }

			if seat == "pubmatic" {
				pmMkt[bid.ImpID] = pubmaticMarketplaceMeta{
					PubmaticKGP:   kgp,
					PubmaticKGPV:  kgpv,
					PubmaticKGPSV: kgpsv,
				}
			}

			pr := PartnerRecord{
				PartnerID:         partnerID,                           // prebid biddercode
				BidderCode:        seat,                                // pubmatic biddercode: pubmatic2
				Latency1:          rCtx.BidderResponseTimeMillis[seat], // it is set inside auctionresponsehook for all bidders
				KGPV:              kgpv,
				KGPSV:             kgpsv,
				BidID:             bid.ID,
				OrigBidID:         bid.ID,
				DefaultBidStatus:  0, // this will be always 0 , decide whether to drop this field in future
				ServerSide:        1,
				MatchedImpression: rCtx.MatchedImpression[seat],
				NetECPM:           GetNetEcpm(price, revShare),
				GrossECPM:         GetGrossEcpm(price),
				OriginalCPM:       GetGrossEcpm(bidExt.OriginalBidCPM),
				OriginalCur:       bidExt.OriginalBidCur,
				PartnerSize:       getSizeForPlatform(bid.W, bid.H, rCtx.Platform),
				DealID:            bid.DealID,
				Nbr:               nbr,
				FloorRuleValue:    -1,
				Adformat:          adFormat,
			}

			if nbr != nil && *nbr == openrtb3.NoBidTimeoutError {
				pr.PostTimeoutBidStatus = 1
			}

			// TODO: WinningBids is set inside auctionresponsehook
			if b, ok := rCtx.WinningBids[bid.ImpID]; ok && b.ID == bidIDForLookup {
				pr.WinningBidStaus = 1
			}

			if len(pr.OriginalCur) == 0 {
				pr.OriginalCPM = float64(0)
				pr.OriginalCur = "USD"
			}

			if len(pr.DealID) != 0 {
				pr.DealChannel = models.DEFAULT_DEALCHANNEL
			} else {
				pr.DealID = "-1"
			}

			var floorCurrency string
			if bidExt.Prebid != nil {
				// don't want default banner for nobid in wl
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
					floorCurrency = bidExt.Prebid.Floors.FloorCurrency
					pr.FloorValue = roundToTwoDigit(bidExt.Prebid.Floors.FloorValue)
					pr.FloorRuleValue = pr.FloorValue
					if bidExt.Prebid.Floors.FloorRuleValue > 0 {
						pr.FloorRuleValue = roundToTwoDigit(bidExt.Prebid.Floors.FloorRuleValue)
					}
				}

				if len(bidExt.Prebid.BidId) > 0 {
					pr.BidID = bidExt.Prebid.BidId
				}
			}

			// if floor values are not set from bid.ext then fall back to imp.bidfloor
			if pr.FloorRuleValue == -1 && impCtx.BidFloor != 0 {
				pr.FloorValue = roundToTwoDigit(impCtx.BidFloor)
				pr.FloorRuleValue = pr.FloorValue
				floorCurrency = impCtx.BidFloorCur
			}

			if floorCurrency != "" && floorCurrency != models.USD {
				value, _ := rCtx.CurrencyConversion(floorCurrency, models.USD, pr.FloorValue)
				pr.FloorValue = roundToTwoDigit(value)
				value, _ = rCtx.CurrencyConversion(floorCurrency, models.USD, pr.FloorRuleValue)
				pr.FloorRuleValue = roundToTwoDigit(value)
			}

			if pr.FloorRuleValue == -1 {
				pr.FloorRuleValue = 0 //reset the value back to 0
			}

			if len(bid.ADomain) != 0 { // for non-bid , bid.ADomain  will be empty
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
			DealID:           "-1",
		}}
	}
	return ipr
}
