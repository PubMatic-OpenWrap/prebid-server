package middleware

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/buger/jsonparser"
	"github.com/gofrs/uuid"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
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

			impId := bid.ImpID
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
		bid := openrtb2.Bid{
			ImpID: impId,
		}

		bidID, err := uuid.NewV4()
		if err == nil {
			bid.ID = bidID.String()
		} else {
			bid.ID = bids[0].ID
		}

		// Get Categories and ad domain
		builder := GetAdPodBuilder()
		category := make(map[string]struct{})
		addomain := make(map[string]struct{})

		for _, eachBid := range bids {
			err := builder.Append(&eachBid)
			if err != nil {
				glog.Errorf("[CTV] type:[adpod_builder_append] parser:[%s] error:[%s] creative:[%s]", builder.Name(), err.Error(), base64.StdEncoding.EncodeToString([]byte(eachBid.AdM)))
				continue
			}

			bid.Price += eachBid.Price
			for _, cat := range eachBid.Cat {
				category[cat] = struct{}{}
			}

			for _, domain := range eachBid.ADomain {
				addomain[domain] = struct{}{}
			}
		}

		bid.AdM, err = builder.Build()
		if err != nil {
			glog.Errorf("[CTV] type:[adpod_builder_build] parser:[%s] error:[%s]", builder.Name(), err.Error())
			continue
		}

		bid.Cat = map2array(category)
		bid.ADomain = map2array(addomain)

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

func map2array(m map[string]struct{}) []string {
	var result []string
	for key := range m {
		result = append(result, key)
	}
	return result[:]
}
