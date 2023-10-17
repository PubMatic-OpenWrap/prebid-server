package pubmatic

import (
	"encoding/json"
	"fmt"

	"net/http"
	"strconv"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/hooks/hookexecution"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// SetIntegrationType sets the integration type in WloggerRecord
func (wlog *WloggerRecord) setIntegrationType(endpoint string) {
	switch endpoint {
	case models.EndpointAMP:
		wlog.IntegrationType = models.TypeAmp
	case models.EndpointV25:
		wlog.IntegrationType = models.TypeSDK
	}
	// add CTV specific cases here
}

func GetLogAuctionObjectAsURL(ao analytics.AuctionObject, rCtx *models.RequestCtx, logInfo, forRespExt bool) (string, http.Header) {

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
			Timeout:           int(rCtx.TMax),
			PDC:               rCtx.DCName, //TODO: confirm this
			CachePutMiss:      rCtx.CachePutMiss,
		},
	}

	requestExt := openrtb_ext.ExtRequest{}
	err := json.Unmarshal(ao.RequestWrapper.BidRequest.Ext, &requestExt)
	if err == nil {
		wlog.logFloorType(&requestExt.Prebid)
	}

	wlog.setIntegrationType(rCtx.Endpoint)

	if ao.RequestWrapper.User != nil {
		extUser := openrtb_ext.ExtUser{}
		_ = json.Unmarshal(ao.RequestWrapper.User.Ext, &extUser)
		wlog.ConsentString = extUser.Consent
	}

	if ao.RequestWrapper.Device != nil {
		wlog.IP = ao.RequestWrapper.Device.IP
		wlog.UserAgent = ao.RequestWrapper.Device.UA
	}

	if ao.RequestWrapper.Regs != nil {
		extReg := openrtb_ext.ExtRegs{}
		_ = json.Unmarshal(ao.RequestWrapper.Regs.Ext, &extReg)
		if extReg.GDPR != nil {
			wlog.GDPR = *extReg.GDPR
		}
	}

	//log device object
	wlog.logDeviceObject(*rCtx, rCtx.UA, ao.RequestWrapper.BidRequest, rCtx.Platform)

	//log content object
	if nil != ao.RequestWrapper.Site {
		wlog.logContentObject(ao.RequestWrapper.Site.Content)
	} else if nil != ao.RequestWrapper.App {
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
	if rCtx.KADUSERCookie != nil {
		headers.Add(models.KADUSERCOOKIE, rCtx.KADUSERCookie.Value)
	}

	url := ow.cfg.Endpoint
	if logInfo || forRespExt {
		url = ow.cfg.PublicEndpoint
	}

	var responseExt openrtb_ext.ExtBidResponse
	err = json.Unmarshal(ao.Response.Ext, &responseExt)
	if err == nil {
		if responseExt.Prebid != nil && responseExt.Prebid.Floors != nil {
			wlog.SetFloorDetails(responseExt.Prebid.Floors)
		}
	}

	return PrepareLoggerURL(&wlog, url, GetGdprEnabledFlag(rCtx.PartnerConfigMap)), headers
}

func ConvertBoolToInt(val bool) int {
	if val {
		return 1
	}
	return 0
}

func (wlog *WloggerRecord) SetFloorDetails(floors *openrtb_ext.PriceFloorRules) {

	if floors == nil {
		return
	}

	if floors.Skipped != nil {
		skipped := ConvertBoolToInt(*floors.Skipped)
		for i := range wlog.Slots {
			wlog.Slots[i].FloorSkippedFlag = &skipped
		}
	}

	if floors.Data != nil && len(floors.Data.ModelGroups) > 0 {
		wlog.FloorModelVersion = floors.Data.ModelGroups[0].ModelVersion
	}

	if len(floors.PriceFloorLocation) > 0 {
		if source, ok := FloorSourceMap[floors.PriceFloorLocation]; ok {
			wlog.FloorSource = &source
		}
	}

	if status, ok := FetchStatusMap[floors.FetchStatus]; ok {
		wlog.FloorFetchStatus = &status
	}

	wlog.FloorProvider = floors.FloorProvider
	if floors.Data != nil && len(floors.Data.FloorProvider) > 0 {
		wlog.FloorProvider = floors.Data.FloorProvider
	}

	if floors.Enforcement != nil && floors.Enforcement.EnforcePBS != nil && *floors.Enforcement.EnforcePBS {
		wlog.record.FloorType = models.HardFloor
	}
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

func convertNonBidToBid(nonBid openrtb_ext.NonBid) (bid openrtb2.Bid) {

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

	bidExt.Nbr = func(nbr int) *openrtb3.NonBidStatusCode {
		temp := openrtb3.NonBidStatusCode(nbr)
		return &temp
	}(nonBid.StatusCode)

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
	loggerSeat := make(map[string][]openrtb2.Bid)

	// TODO : Uncomment and modify to add seatnonbids in logger
	for _, seatNonBid := range ao.SeatNonBid {
		if _, ok := rejectedBids[seatNonBid.Seat]; !ok {
			rejectedBids[seatNonBid.Seat] = map[string]struct{}{}
		}
		for _, nonBid := range seatNonBid.NonBid {
			rejectedBids[seatNonBid.Seat][nonBid.ImpId] = struct{}{}
			loggerSeat[seatNonBid.Seat] = append(loggerSeat[seatNonBid.Seat], convertNonBidToBid(nonBid))
		}
	}

	for _, seatBid := range ao.Response.SeatBid {
		for _, bid := range seatBid.Bid {
			// Check if this is a default and RejectedBids bid. Ex. only one bid by pubmatic it was rejected by floors.
			// Module would add a 0 bid. So, we want to skip this zero bid to avoid duplicate or incomplete data and log the correct one that was rejected.
			// We don't have bid.ID here so using bid.ImpID
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
		// include bids dropped by module. Ex. sendAllBids=false
		loggerSeat[seat] = append(loggerSeat[seat], Bids...)
	}

	// pubmatic's KGP details per impression
	// This is only required for groupm bids (groupm bids log pubmatic's data in ow logger)
	type pubmaticMarketplaceMeta struct {
		PubmaticKGP, PubmaticKGPV, PubmaticKGPSV string
	}
	pmMkt := make(map[string]pubmaticMarketplaceMeta)

	for seat, bids := range loggerSeat {
		if seat == string(openrtb_ext.BidderOWPrebidCTV) {
			continue
		}

		// Response would not contain non-mapped bids, do we need this
		if _, ok := rCtx.AdapterThrottleMap[seat]; ok {
			continue
		}

		for _, bid := range bids {
			impCtx, ok := rCtx.ImpBidCtx[bid.ImpID]
			if !ok {
				continue
			}

			// Response would not contain non-mapped bids, do we need this
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
				kgp = bidderMeta.KGP           // _AU_@_W_x_H_
				kgpv = bidderMeta.KGPV         // ^/43743431/DMDemo[0-9]*@Div[12]@^728x90$
				kgpsv = bidderMeta.MatchedSlot // /43743431/DMDemo1@@728x90
				isRegex = bidderMeta.IsRegex
			}

			// 1. nobid
			if bid.Price == 0 && bid.H == 0 && bid.W == 0 {
				//NOTE: kgpsv = bidderMeta.MatchedSlot above. Use the same
				if !isRegex && kgpv != "" { // unmapped pubmatic's slot
					kgpsv = kgpv // - KGP: _AU_@_DIV_@_W_x_H_
				} else if !isRegex {
					kgpv = kgpsv
				}
			} else if !isRegex {
				if kgpv != "" { // unmapped pubmatic's slot
					kgpsv = kgpv // /43743431/DMDemo1234@300x250 -->
				} else if bid.H != 0 && bid.W != 0 { // Check when bid.H and bid.W will be zero with Price !=0. Ex: MobileInApp-MultiFormat-OnlyBannerMapping_Criteo_Partner_Validaton
					// 2. valid bid
					// kgpv has regex, do not generate slotName again
					// kgpsv could be unmapped or mapped slot, generate slotName again based on bid.H and bid.W
					kgpsv = GenerateSlotName(bid.H, bid.W, kgp, impCtx.TagID, impCtx.Div, rCtx.Source)
					kgpv = kgpsv // original /43743431/DMDemo1234@300x250 but new could be /43743431/DMDemo1234@222x111
				}
			}

			if kgpv == "" {
				kgpv = kgpsv
			}

			var bidExt models.BidExt
			if bidCtx, ok := impCtx.BidCtx[bid.ID]; ok {
				bidExt = bidCtx.BidExt
			} else {
				// this is for nonbids
				json.Unmarshal(bid.Ext, &bidExt)
			}

			price := bid.Price
			// if ao.Response.Cur != "" && ao.Response.Cur != "USD" {
			// 	price = bidExt.OriginalBidCPMUSD
			// }

			if bidExt.OriginalBidCPMUSD != bid.Price {
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
				Nbr:         bidExt.Nbr,
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
			} else {
				pr.DealID = "-1"
			}

			var floorCurrency string
			if bidExt.Prebid != nil {
				// don't want default banner for nobid in wl
				if bidExt.Prebid.Type != "" {
					pr.Adformat = string(bidExt.Prebid.Type)
				}

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
					floorCurrency = bidExt.Prebid.Floors.FloorCurrency
					pr.FloorValue = roundToTwoDigit(bidExt.Prebid.Floors.FloorValue)
					pr.FloorRuleValue = roundToTwoDigit(bidExt.Prebid.Floors.FloorRuleValue)
				}

				if len(bidExt.Prebid.BidId) > 0 {
					pr.BidID = bidExt.Prebid.BidId
				}
			}

			if pr.FloorRuleValue == 0 && impCtx.BidFloor != nil {
				pr.FloorValue = roundToTwoDigit(*impCtx.BidFloor)
				pr.FloorRuleValue = pr.FloorValue
				if impCtx.BidFloorCur != nil {
					floorCurrency = *impCtx.BidFloorCur
				}
			}

			if floorCurrency != "" && floorCurrency != models.USD {
				value, _ := rCtx.GetPBSCurrencyConversion(floorCurrency, models.USD, pr.FloorValue)
				pr.FloorValue = roundToTwoDigit(value)
				value, _ = rCtx.GetPBSCurrencyConversion(floorCurrency, models.USD, pr.FloorRuleValue)
				pr.FloorRuleValue = roundToTwoDigit(value)
			}

			if pr.Adformat == "" && bid.AdM != "" {
				pr.Adformat = models.GetAdFormat(bid.AdM)
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
			DealID:           "-1",
		}}
	}
	return ipr
}
