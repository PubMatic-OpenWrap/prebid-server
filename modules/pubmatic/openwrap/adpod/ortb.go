package adpod

import (
	"encoding/json"
	"net/http"

	"github.com/buger/jsonparser"
	"github.com/gofrs/uuid"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
)

type ortbResp struct {
	rctx models.RequestCtx
}

func newOrtbResponder(rctx models.RequestCtx) Responder {
	return &ortbResp{rctx: rctx}
}

func (ortb *ortbResp) FormResponse(bidResponse *openrtb2.BidResponse, headers http.Header) (interface{}, http.Header, error) {
	mergedBidResponse := mergeSeatBids(bidResponse)
	data, err := json.Marshal(mergedBidResponse)
	if err != nil {
		var (
			id     string
			bidExt json.RawMessage
		)
		if bidResponse != nil {
			id = bidResponse.ID
			bidExt = bidResponse.Ext
		}
		bidExt = addErrorInExtension(err.Error(), bidExt, ortb.rctx.Debug)
		return formErrorBidResponse(id, nbr.InternalError.Ptr(), bidExt), headers, nil
	}

	return data, headers, nil
}

func mergeSeatBids(bidResponse *openrtb2.BidResponse) *openrtb2.BidResponse {
	if bidResponse == nil || bidResponse.SeatBid == nil {
		return bidResponse
	}

	var seatBids []openrtb2.SeatBid
	bidArrayMap := make(map[string][]openrtb2.Bid)
	for _, seatBid := range bidResponse.SeatBid {
		//Copy seatBid and reset its bids
		videoSeatBid := seatBid
		videoSeatBid.Bid = nil
		for _, bid := range seatBid.Bid {
			if bid.Price == 0 {
				continue
			}

			adpodBid, _ := jsonparser.GetBoolean(bid.Ext, "adpod", "isAdpodBid")
			bid.Ext = jsonparser.Delete(bid.Ext, "adpod")
			if !adpodBid {
				videoSeatBid.Bid = append(videoSeatBid.Bid, bid)
				continue
			}

			impId, _ := models.GetImpressionID(bid.ImpID)
			bidArrayMap[impId] = append(bidArrayMap[impId], bid)
		}

		if len(videoSeatBid.Bid) > 0 {
			seatBids = append(seatBids, videoSeatBid)
		}
	}

	// Get Merged prebid_ctv bid
	ctvSeatBid := getPrebidCTVSeatBid(bidArrayMap)

	seatBids = append(seatBids, ctvSeatBid...)
	bidResponse.SeatBid = seatBids

	return bidResponse
}

func getPrebidCTVSeatBid(bidsMap map[string][]openrtb2.Bid) []openrtb2.SeatBid {
	seatBids := []openrtb2.SeatBid{}

	for impId, bids := range bidsMap {
		bid := openrtb2.Bid{}
		bidID, err := uuid.NewV4()
		if err == nil {
			bid.ID = bidID.String()
		} else {
			bid.ID = bids[0].ID
		}
		creative, price := getAdPodBidCreativeAndPrice(bids)
		bid.AdM = creative
		bid.Price = price
		bid.ImpID = impId
		// Get Categories and ad domain
		category := make(map[string]bool)
		addomain := make(map[string]bool)
		for _, eachBid := range bids {
			for _, cat := range eachBid.Cat {
				if _, ok := category[cat]; !ok {
					category[cat] = true
					bid.Cat = append(bid.Cat, cat)
				}
			}
			for _, domain := range eachBid.ADomain {
				if _, ok := addomain[domain]; !ok {
					addomain[domain] = true
					bid.ADomain = append(bid.ADomain, domain)
				}
			}
		}

		seatBid := openrtb2.SeatBid{}
		seatBid.Seat = models.BidderOWPrebidCTV
		seatBid.Bid = append(seatBid.Bid, bid)

		seatBids = append(seatBids, seatBid)
	}

	return seatBids
}

func formErrorBidResponse(id string, nbrCode *openrtb3.NoBidReason, ext json.RawMessage) openrtb2.BidResponse {
	return openrtb2.BidResponse{
		ID:  id,
		NBR: nbrCode,
		Ext: ext,
	}
}
