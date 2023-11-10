package openwrap

import (
	"encoding/json"
	"strconv"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/adunitconfig"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
	uuid "github.com/satori/go.uuid"
)

func (m *OpenWrap) addDefaultBids(rctx *models.RequestCtx, bidResponse *openrtb2.BidResponse, bidResponseExt *openrtb_ext.ExtBidResponse) map[string]map[string][]openrtb2.Bid {
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
		for bidder := range impCtx.Bidders {
			if bidders, ok := seatBids[impID]; ok { // bid found for impID
				if _, ok := bidders[bidder]; ok { // bid found for seat
					continue
				}
			}

			if defaultBids[impID] == nil {
				defaultBids[impID] = make(map[string][]openrtb2.Bid)
			}

			// TODO: confirm this behaviour change
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
				},
			}

			// record error stats for each bidder
			m.recordErrorStats(*rctx, bidResponseExt, bidder)
		}
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

// getNonBRCodeFromBidRespExt maps the error-code present in prebid partner response with standard nonBR code
func getNonBRCodeFromBidRespExt(bidder string, bidResponseExt *openrtb_ext.ExtBidResponse) *openrtb3.NonBidStatusCode {
	errs := bidResponseExt.Errors[openrtb_ext.BidderName(bidder)]
	if len(errs) == 0 {
		return models.GetNonBidStatusCodePtr(openrtb3.NoBidGeneral)
	}

	switch errs[0].Code {
	case errortypes.TimeoutErrorCode:
		return models.GetNonBidStatusCodePtr(openrtb3.NoBidTimeoutError)
	case errortypes.UnknownErrorCode:
		return models.GetNonBidStatusCodePtr(openrtb3.NoBidGeneralError)
	default:
		return models.GetNonBidStatusCodePtr(openrtb3.NoBidGeneralError)
	}
}

func newDefaultBidExt(rctx models.RequestCtx, impID, bidder string, bidResponseExt *openrtb_ext.ExtBidResponse) *models.BidExt {

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
func (m *OpenWrap) recordErrorStats(rctx models.RequestCtx, bidResponseExt *openrtb_ext.ExtBidResponse, bidder string) {

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
