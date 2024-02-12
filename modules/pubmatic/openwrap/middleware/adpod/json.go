package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/utils"
	"github.com/prebid/prebid-server/openrtb_ext"
	pbc "github.com/prebid/prebid-server/prebid_cache_client"
)

const (
	slotKeyFormat = "s%d_%s"
)

var (
	redirectTargetingKeys = []string{"pwtpb", "pwtdur", "pwtcid", "pwtpid", "pwtdealtier", "pwtdid"}
)

type adPodBid struct {
	ID        string                `json:"id,omitempty"`
	NBR       *openrtb3.NoBidReason `json:"nbr,omitempty"`
	Targeting []map[string]string   `json:"targeting,omitempty"`
	Error     string                `json:"error,omitempty"`
	Ext       interface{}           `json:"ext,omitempty"`
}

type bidResponseAdpod struct {
	AdPodBids []*adPodBid `json:"adpods,omitempty"`
	Ext       interface{} `json:"ext,omitempty"`
}

type jsonResponse struct {
	cacheClient *pbc.Client
	debug       string
}

func (jr *jsonResponse) formJSONResponse(adpodWriter *utils.HTTPResponseBufferWriter) ([]byte, map[string]string, int) {
	var statusCode = http.StatusOK
	var headers = map[string]string{
		ContentType:    ApplicationJSON,
		ContentOptions: NoSniff,
	}

	if adpodWriter.Code > 0 && adpodWriter.Code == http.StatusBadRequest {
		return formJSONErrorResponse("", adpodWriter.Response.String(), GetNoBidReasonCode(nbr.InvalidVideoRequest), nil, jr.debug), headers, adpodWriter.Code
	}

	response, err := io.ReadAll(adpodWriter.Response)
	if err != nil {
		statusCode = http.StatusInternalServerError
		return formJSONErrorResponse("", "error in reading response, reason: "+err.Error(), GetNoBidReasonCode(nbr.InternalError), nil, jr.debug), headers, statusCode
	}

	var bidResponse *openrtb2.BidResponse
	err = json.Unmarshal(response, &bidResponse)
	if err != nil {
		statusCode = http.StatusInternalServerError
		return formJSONErrorResponse("", "error in unmarshaling the auction response, reason: "+err.Error(), GetNoBidReasonCode(nbr.InternalError), nil, jr.debug), headers, statusCode
	}

	if bidResponse.NBR != nil {
		statusCode = http.StatusBadRequest
		return formJSONErrorResponse(bidResponse.ID, "", bidResponse.NBR, bidResponse.Ext, jr.debug), headers, statusCode
	}

	var finalResponse []byte
	finalResponse, statusCode = jr.getJsonResponse(bidResponse)

	return finalResponse, headers, statusCode
}

func (jr *jsonResponse) getJsonResponse(bidResponse *openrtb2.BidResponse) ([]byte, int) {
	if bidResponse == nil {
		return formJSONErrorResponse("", "empty bid response recieved", GetNoBidReasonCode(int(openrtb3.NoBidGeneralError)), nil, jr.debug), http.StatusOK
	}

	var reqExt openrtb_ext.ExtBidResponse
	err := json.Unmarshal(bidResponse.Ext, &reqExt)
	if err != nil {
		return formJSONErrorResponse("", "error in unmarshaling request extension, reason: "+err.Error(), GetNoBidReasonCode(nbr.InternalError), nil, jr.debug), http.StatusInternalServerError
	}

	var responseFormat, redirectURL string
	if reqExt.Wrapper != nil {
		responseFormat = reqExt.Wrapper.ResponseFormat
		redirectURL = reqExt.Wrapper.RedirectURL
		reqExt.Wrapper = nil
	}
	bidResponse.Ext, _ = json.Marshal(reqExt)

	if bidResponse.SeatBid == nil {
		if len(redirectURL) > 0 && responseFormat == models.ResponseFormatRedirect && jr.debug != "1" {
			return []byte(redirectURL), http.StatusFound
		}
		return formJSONErrorResponse("", "No Bid", GetNoBidReasonCode(int(openrtb3.NoBidGeneralError)), bidResponse.Ext, jr.debug), http.StatusOK
	}

	bidArrayMap := make(map[string][]openrtb2.Bid)
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			impId, _ := models.GetImpressionID(bid.ImpID)
			bids, ok := bidArrayMap[impId]
			if !ok {
				bidArrayMap[impId] = make([]openrtb2.Bid, 0)
			}
			if bid.Price > 0 {
				bids = append(bids, bid)
				bidArrayMap[impId] = bids
			}
		}
	}
	adPodBids := formAdpodBids(bidArrayMap, jr.cacheClient)

	if len(redirectURL) > 0 && responseFormat == models.ResponseFormatRedirect && jr.debug != "1" {
		return getRedirectResponse(adPodBids, redirectURL), http.StatusFound
	}

	adpodResponse := bidResponseAdpod{AdPodBids: adPodBids, Ext: bidResponse.Ext}
	response, _ := json.Marshal(adpodResponse)

	return response, http.StatusOK
}

func getRedirectResponse(adpodBids []*adPodBid, redirectURL string) []byte {
	if len(adpodBids) == 0 {
		return []byte(redirectURL)
	}

	if len(adpodBids[0].Targeting) == 0 {
		return []byte(redirectURL)
	}

	parsedURL, err := url.ParseRequestURI(redirectURL)
	if err != nil {
		return []byte(redirectURL)
	}

	redirectQuery := parsedURL.Query()
	custParams, err := url.ParseQuery(strings.TrimSpace(redirectQuery.Get(models.CustParams)))
	if err != nil {
		return []byte(redirectURL)
	}

	for i, target := range adpodBids[0].Targeting {
		sNo := i + 1
		for _, tk := range redirectTargetingKeys {
			targetingKey := prepareSlotLevelKey(sNo, tk)
			if value, ok := target[targetingKey]; ok {
				custParams.Set(targetingKey, value)
			}
		}
	}

	redirectQuery.Set(models.CustParams, custParams.Encode())
	parsedURL.RawQuery = redirectQuery.Encode()

	rURL := parsedURL.String()

	return []byte(rURL)
}

func formAdpodBids(bidsMap map[string][]openrtb2.Bid, cacheClient *pbc.Client) []*adPodBid {
	var adpodBids []*adPodBid
	for impId, bids := range bidsMap {
		adpodBid := &adPodBid{
			ID: impId,
		}
		if len(bids) == 0 {
			adpodBid.NBR = GetNoBidReasonCode(int(openrtb3.NoBidGeneralError))
			adpodBids = append(adpodBids, adpodBid)
			continue
		}
		sort.Slice(bids, func(i, j int) bool { return bids[i].Price > bids[j].Price })

		cacheIds, err := cacheAllBids(cacheClient, bids)
		if err != nil {
			adpodBid.Error = err.Error()
			adpodBids = append(adpodBids, adpodBid)
			continue
		}

		targetings := []map[string]string{}
		for i := 0; i < len(bids); i++ {
			slotNo := i + 1
			targeting := createTargetting(bids[i], slotNo, cacheIds[i])
			if len(targeting) > 0 {
				targetings = append(targetings, targeting)
			}
		}

		if len(targetings) > 0 {
			adpodBid.Targeting = targetings
		}
		adpodBids = append(adpodBids, adpodBid)
	}

	return adpodBids
}

func prepareSlotLevelKey(slotNo int, key string) string {
	return fmt.Sprintf(slotKeyFormat, slotNo, key)
}

func createTargetting(bid openrtb2.Bid, slotNo int, cacheId string) map[string]string {
	targetingKeyValMap := make(map[string]string)
	targetingKeyValMap[prepareSlotLevelKey(slotNo, models.PWT_CACHEID)] = cacheId

	if len(bid.Ext) > 0 {
		bidExt := models.BidExt{}
		err := json.Unmarshal(bid.Ext, &bidExt)
		if err != nil {
			return targetingKeyValMap
		}

		for k, v := range bidExt.AdPod.Targeting {
			targetingKeyValMap[prepareSlotLevelKey(slotNo, k)] = v
		}

		if bidExt.AdPod.Debug.Targeting != nil {
			for k, v := range bidExt.AdPod.Debug.Targeting {
				targetingKeyValMap[k] = v
			}
		}

	}

	return targetingKeyValMap

}

func writeErrorResponse(w http.ResponseWriter, code int, err CustomError) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	errResponse := GetErrorResponse(err)
	fmt.Fprintln(w, errResponse)
}

func GetErrorResponse(err CustomError) []byte {
	if err == nil {
		return nil
	}

	response, _ := json.Marshal(map[string]interface{}{
		"ErrorCode": err.Code(),
		"Error":     err.Error(),
	})
	return response
}

func formJSONErrorResponse(id string, errMessage string, nbr *openrtb3.NoBidReason, ext json.RawMessage, debug string) []byte {
	type errResponse struct {
		Id  string                `json:"id"`
		NBR *openrtb3.NoBidReason `json:"nbr,omitempty"`
		Ext json.RawMessage       `json:"ext,omitempty"`
	}

	if len(errMessage) > 0 {
		ext = addErrorInExtension(errMessage, ext, debug)
	}

	response := errResponse{
		Id:  id,
		NBR: nbr,
		Ext: ext,
	}

	responseBytes, _ := json.Marshal(response)
	return responseBytes
}