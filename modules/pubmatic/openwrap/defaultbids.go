package openwrap

import (
	"encoding/json"
	"strconv"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v2/errortypes"
	"github.com/prebid/prebid-server/v2/exchange"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/adunitconfig"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	uuid "github.com/satori/go.uuid"
)

func (m *OpenWrap) addDefaultBids(rctx *models.RequestCtx, bidResponse *openrtb2.BidResponse, bidResponseExt openrtb_ext.ExtBidResponse) map[string]map[string][]openrtb2.Bid {
	// responded bidders per impression
	seatBids := make(map[string]map[string]struct{}, len(bidResponse.SeatBid))
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			impId, _ := models.GetImpressionID(bid.ImpID)
			if seatBids[impId] == nil {
				seatBids[impId] = make(map[string]struct{})
			}
			seatBids[impId][seatBid.Seat] = struct{}{}
		}
	}

	// consider responded but dropped bids to avoid false nobid entries
	for seat, bids := range rctx.DroppedBids {
		for _, bid := range bids {
			impId, _ := models.GetImpressionID(bid.ImpID)
			if seatBids[impId] == nil {
				seatBids[impId] = make(map[string]struct{})
			}
			seatBids[impId][seat] = struct{}{}
		}
	}

	// bids per bidders per impression that did not respond
	defaultBids := make(map[string]map[string][]openrtb2.Bid, 0)
	for impID, impCtx := range rctx.ImpBidCtx {
		for bidder, meta := range impCtx.Bidders {
			if bidders, ok := seatBids[impID]; ok { // bid found for impID
				if _, ok := bidders[bidder]; ok { // bid found for seat
					continue
				}
			}

			if meta.PrebidBidderCode == models.BidderVASTBidder {
				continue
			}

			if defaultBids[impID] == nil {
				defaultBids[impID] = make(map[string][]openrtb2.Bid)
			}
			uuid, _ := m.uuidGenerator.Generate()
			bidExt := newDefaultBidExt(*rctx, impID, bidder, bidResponseExt)
			bidExtJson, _ := json.Marshal(bidExt)

			defaultBids[impID][bidder] = append(defaultBids[impID][bidder], openrtb2.Bid{
				ID:    uuid,
				ImpID: impID,
				Ext:   bidExtJson,
			})

			// create bidCtx because we need it for owlogger
			rctx.ImpBidCtx[impID].BidCtx[uuid] = models.BidCtx{
				BidExt: models.BidExt{
					Nbr: bidExt.Nbr,
				},
			}

			// record error stats for each bidder
			m.recordErrorStats(*rctx, bidResponseExt, bidder)
		}
	}

	// VastTags for a VastBidder that did not respond
	for impID, impCtx := range rctx.ImpBidCtx {
		for bidder, meta := range impCtx.Bidders {
			if meta.PrebidBidderCode != models.BidderVASTBidder {
				continue
			}

			var noBidVastTags []string
			for tag, status := range meta.VASTTagFlags {
				if !status {
					noBidVastTags = append(noBidVastTags, tag)
				}
			}

			if len(noBidVastTags) == 0 {
				continue
			}

			if defaultBids[impID] == nil {
				defaultBids[impID] = make(map[string][]openrtb2.Bid)
			}

			for i := range noBidVastTags {
				uuid := uuid.NewV4().String()
				bidExt := newDefaultBidExt(*rctx, impID, bidder, bidResponseExt)
				bidExtJson, _ := json.Marshal(bidExt)

				defaultBids[impID][bidder] = append(defaultBids[impID][bidder], openrtb2.Bid{
					ID:    uuid,
					ImpID: impID,
					Ext:   bidExtJson,
				})

				// create bidCtx because we need it for owlogger
				rctx.ImpBidCtx[impID].BidCtx[uuid] = models.BidCtx{
					BidExt: models.BidExt{
						Nbr: bidExt.Nbr,
						ExtBid: openrtb_ext.ExtBid{
							Prebid: &openrtb_ext.ExtBidPrebid{
								Video: &openrtb_ext.ExtBidPrebidVideo{
									VASTTagID: noBidVastTags[i],
								},
							},
						},
					},
				}
			}
		}
	}

	//Do not add nobids in default bids for throttled adapter and non-mapped bidders in case of web-s2s
	//as we are forming forming seatNonBids from defaultBids which is used for owlogger
	if rctx.Endpoint == models.EndpointWebS2S {
		return defaultBids
	}

	// add nobids for throttled adapter to all the impressions (how do we set profile with custom list of bidders at impression level?)
	for bidder := range rctx.AdapterThrottleMap {
		for impID := range rctx.ImpBidCtx { // ImpBidCtx is used only for list of impID, it does not have data of throttled adapters
			if defaultBids[impID] == nil {
				defaultBids[impID] = make(map[string][]openrtb2.Bid)
			}

			bidExt := newDefaultBidExt(*rctx, impID, bidder, bidResponseExt)
			bidExtJson, _ := json.Marshal(bidExt)
			// no need to create impBidCtx since we dont log partner-throttled bid in owlogger

			defaultBids[impID][bidder] = []openrtb2.Bid{
				{
					ID:    uuid.NewV4().String(),
					ImpID: impID,
					Ext:   bidExtJson,
				},
			}
		}
	}

	// add nobids for non-mapped bidders
	for impID, impCtx := range rctx.ImpBidCtx {
		for bidder := range impCtx.NonMapped {
			if defaultBids[impID] == nil {
				defaultBids[impID] = make(map[string][]openrtb2.Bid)
			}

			bidExt := newDefaultBidExt(*rctx, impID, bidder, bidResponseExt)
			bidExtJson, _ := json.Marshal(bidExt)
			// no need to create impBidCtx since we dont log slot-not-mapped bid in owlogger

			defaultBids[impID][bidder] = []openrtb2.Bid{
				{
					ID:    uuid.NewV4().String(),
					ImpID: impID,
					Ext:   bidExtJson,
				},
			}
		}
	}

	return defaultBids
}

func (m *OpenWrap) addDefaultBidsForMultiFloorsConfig(rctx *models.RequestCtx, bidResponse *openrtb2.BidResponse, bidResponseExt openrtb_ext.ExtBidResponse) map[string]map[string][]openrtb2.Bid {
	defaultBids := rctx.DefaultBids
	var excludeFloors []float64    //exclude floors which are already present in bidresponse
	var adunitFloors []float64     //floors for each adunit
	var defaultBidFloors []float64 //floors for default bids
	//MultiBidMultiFloor is enabled only for AppLovinMax
	if rctx.Endpoint != models.EndpointAppLovinMax || !rctx.AppLovinMax.MultiFloorsConfig.Enabled {
		return defaultBids
	}

	// count of pubmatic bids for each impression
	seatBidsMultiFloor := make(map[string]int, len(bidResponse.SeatBid))
	for _, seatBid := range bidResponse.SeatBid {
		if seatBid.Seat == models.BidderPubMatic || seatBid.Seat == models.BidderPubMaticSecondaryAlias {
			for _, bid := range seatBid.Bid {
				impId, _ := models.GetImpressionID(bid.ImpID)
				var bidExt models.BidExt
				seatBidsMultiFloor[impId]++
				if bid.Ext != nil {
					if err := json.Unmarshal(bid.Ext, &bidExt); err == nil && bidExt.MultiBidMultiFloorValue != 0 {
						excludeFloors = append(excludeFloors, bidExt.MultiBidMultiFloorValue)
					}
				}
			}
		}
	}

	for impID, impCtx := range rctx.ImpBidCtx {
		defaultBidsCount := len(rctx.AppLovinMax.MultiFloorsConfig.Config[impCtx.TagID]) - seatBidsMultiFloor[impID]

		if defaultBidsCount <= 0 {
			continue
		}

		if defaultBids[impID] == nil {
			defaultBids[impID] = make(map[string][]openrtb2.Bid)
		}

		//if defaultbid is already present for pubmatic, then reset it, as we are adding new defaultbids with MultiBidMultiFloor
		if _, ok := defaultBids[impID][models.BidderPubMatic]; ok && defaultBidsCount > 0 {
			defaultBids[impID][models.BidderPubMatic] = make([]openrtb2.Bid, 0, defaultBidsCount)
		}

		adunitFloors = rctx.AppLovinMax.MultiFloorsConfig.Config[rctx.ImpBidCtx[impID].TagID]

		//exclude floors which are already present in bidresponse for defaultbids
		defaultBidFloors = nil
		for _, floor := range adunitFloors {
			exclude := false
			for _, excludeFloor := range excludeFloors {
				if floor == excludeFloor {
					exclude = true
					break
				}
			}
			if !exclude {
				defaultBidFloors = append(defaultBidFloors, floor)
			}
		}

		for i := 0; i < defaultBidsCount; i++ {
			uuid, _ := m.uuidGenerator.Generate()
			bidExt := newDefaultBidExtMultiFloors(defaultBidFloors[i], models.BidderPubMatic, bidResponseExt)
			bidExtJson, _ := json.Marshal(bidExt)

			defaultBids[impID][models.BidderPubMatic] = append(defaultBids[impID][models.BidderPubMatic], openrtb2.Bid{
				ID:    uuid,
				ImpID: impID,
				Ext:   bidExtJson,
			})
		}
	}
	return defaultBids
}

// getNonBRCodeFromBidRespExt maps the error-code present in prebid partner response with standard nonBR code
func getNonBRCodeFromBidRespExt(bidder string, bidResponseExt openrtb_ext.ExtBidResponse) *openrtb3.NoBidReason {
	errs := bidResponseExt.Errors[openrtb_ext.BidderName(bidder)]
	if len(errs) == 0 {
		return openrtb3.NoBidUnknownError.Ptr()
	}

	switch errs[0].Code {
	case errortypes.TimeoutErrorCode:
		return exchange.ErrorTimeout.Ptr()
	case errortypes.UnknownErrorCode:
		return exchange.ErrorGeneral.Ptr()
	default:
		return exchange.ErrorGeneral.Ptr()
	}
}

func newDefaultBidExt(rctx models.RequestCtx, impID, bidder string, bidResponseExt openrtb_ext.ExtBidResponse) *models.BidExt {
	bidExt := models.BidExt{
		NetECPM: 0,
		Nbr:     getNonBRCodeFromBidRespExt(bidder, bidResponseExt),
	}
	if rctx.ClientConfigFlag == 1 {
		if cc := adunitconfig.GetClientConfigForMediaType(rctx, impID, "banner"); cc != nil {
			bidExt.Banner = &models.ExtBidBanner{
				ClientConfig: cc,
			}
		}

		if cc := adunitconfig.GetClientConfigForMediaType(rctx, impID, "video"); cc != nil {
			bidExt.Video = &models.ExtBidVideo{
				ClientConfig: cc,
			}
		}
	}

	if v, ok := rctx.PartnerConfigMap[models.VersionLevelConfigID]["refreshInterval"]; ok {
		n, err := strconv.Atoi(v)
		if err == nil {
			bidExt.RefreshInterval = n
		}
	}
	return &bidExt
}

func newDefaultBidExtMultiFloors(floor float64, bidder string, bidResponseExt openrtb_ext.ExtBidResponse) *models.BidExt {
	return &models.BidExt{
		NetECPM:                 0,
		Nbr:                     getNonBRCodeFromBidRespExt(bidder, bidResponseExt),
		MultiBidMultiFloorValue: floor,
	}
}

// TODO : Check if we need this?
// func newDefaultVastTagBidExt(rctx models.RequestCtx, impID, bidder, vastTag string, bidResponseExt openrtb_ext.ExtBidResponse) *models.BidExt {
// 	bidExt := models.BidExt{
// 		ExtBid: openrtb_ext.ExtBid{
// 			Prebid: &openrtb_ext.ExtBidPrebid{
// 				Video: &openrtb_ext.ExtBidPrebidVideo{
// 					VASTTagID: vastTag,
// 				},
// 			},
// 		},
// 		NetECPM: 0,
// 		Nbr:     getNonBRCodeFromBidRespExt(bidder, bidResponseExt),
// 	}

// 	if rctx.ClientConfigFlag == 1 {
// 		if cc := adunitconfig.GetClientConfigForMediaType(rctx, impID, "video"); cc != nil {
// 			bidExt.Video = &models.ExtBidVideo{
// 				ClientConfig: cc,
// 			}
// 		}
// 	}

// 	if v, ok := rctx.PartnerConfigMap[models.VersionLevelConfigID]["refreshInterval"]; ok {
// 		n, err := strconv.Atoi(v)
// 		if err == nil {
// 			bidExt.RefreshInterval = n
// 		}
// 	}

// 	return &bidExt
// }

func (m *OpenWrap) applyDefaultBids(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) (*openrtb2.BidResponse, error) {
	// update nobids in final response
	for i, seatBid := range bidResponse.SeatBid {
		for impID, noSeatBid := range rctx.DefaultBids {
			for seat, bids := range noSeatBid {
				if seatBid.Seat == seat {
					bidResponse.SeatBid[i].Bid = append(bidResponse.SeatBid[i].Bid, bids...)
					delete(noSeatBid, seat)
					rctx.DefaultBids[impID] = noSeatBid
				}
			}
		}
	}

	// no-seat case
	for _, noSeatBid := range rctx.DefaultBids {
		for seat, bids := range noSeatBid {
			bidResponse.SeatBid = append(bidResponse.SeatBid, openrtb2.SeatBid{
				Bid:  bids,
				Seat: seat,
			})
		}
	}

	return bidResponse, nil
}
func (m *OpenWrap) recordErrorStats(rctx models.RequestCtx, bidResponseExt openrtb_ext.ExtBidResponse, bidder string) {

	responseError := models.PartnerErrNoBid

	bidderErr, ok := bidResponseExt.Errors[openrtb_ext.BidderName(bidder)]
	if ok && len(bidderErr) > 0 {
		switch bidderErr[0].Code {
		case errortypes.TimeoutErrorCode:
			responseError = models.PartnerErrTimeout
		case errortypes.UnknownErrorCode:
			responseError = models.PartnerErrUnknownPrebidError
		}
	}
	m.metricEngine.RecordPartnerResponseErrors(rctx.PubIDStr, bidder, responseError)
}
