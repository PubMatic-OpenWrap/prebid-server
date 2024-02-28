package middleware

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/buger/jsonparser"
	"github.com/gofrs/uuid"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/utils"
)

type ortbResponse struct {
	debug              string
	WrapperLoggerDebug string
}

func (or *ortbResponse) formOperRTBResponse(adpodWriter *utils.HTTPResponseBufferWriter) ([]byte, map[string]string, int) {
	var statusCode = http.StatusOK
	var headers = map[string]string{
		ContentType:    ApplicationJSON,
		ContentOptions: NoSniff,
	}

	if adpodWriter.Code > 0 && adpodWriter.Code == http.StatusBadRequest {
		return adpodWriter.Response.Bytes(), headers, adpodWriter.Code
	}

	response, err := io.ReadAll(adpodWriter.Response)
	if err != nil {
		statusCode = http.StatusInternalServerError
		ext := addErrorInExtension(err.Error(), nil, or.debug)
		return formErrorBidResponse("", nbr.InternalError.Ptr(), ext), headers, statusCode
	}

	var bidResponse *openrtb2.BidResponse
	err = json.Unmarshal(response, &bidResponse)
	if err != nil {
		statusCode = http.StatusInternalServerError
		ext := addErrorInExtension(err.Error(), nil, or.debug)
		return formErrorBidResponse("", nbr.InternalError.Ptr(), ext), headers, statusCode
	}

	if bidResponse.NBR != nil {
		statusCode = http.StatusBadRequest
		return response, headers, statusCode
	}

	// TODO: Do not merge the response, respond with 2.6 response
	mergedBidResponse := mergeSeatBids(bidResponse)
	data, err := json.Marshal(mergedBidResponse)
	if err != nil {
		statusCode = 500
		var id string
		var bidExt json.RawMessage
		if bidResponse != nil {
			id = bidResponse.ID
			bidExt = bidResponse.Ext
		}
		bidExt = addErrorInExtension(err.Error(), bidExt, or.debug)
		return formErrorBidResponse(id, nbr.InternalError.Ptr(), bidExt), headers, statusCode
	}

	return data, headers, statusCode
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
		if len(bids) > 0 {
			bid.Cat = bids[0].Cat
			bid.ADomain = bids[0].ADomain
		}
		bid.ImpID = impId

		seatBid := openrtb2.SeatBid{}
		seatBid.Seat = models.BidderOWPrebidCTV
		seatBid.Bid = append(seatBid.Bid, bid)

		seatBids = append(seatBids, seatBid)
	}

	return seatBids
}

func formErrorBidResponse(id string, nbrCode *openrtb3.NoBidReason, ext json.RawMessage) []byte {
	response := openrtb2.BidResponse{
		ID:  id,
		NBR: nbrCode,
		Ext: ext,
	}
	data, _ := json.Marshal(response)
	return data
}
