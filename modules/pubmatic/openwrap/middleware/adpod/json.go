package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	pbc "github.com/prebid/prebid-server/prebid_cache_client"
)

const (
	slotKeyFormat = "s%d_%s"
)

var (
	redirectTargetingKeys = []string{"pwtpb", "pwtdur", "pwtcid", "pwtpid", "pwtdealtier", "pwtdid"}
)

type adPodBid struct {
	ID        string              `json:"id,omitempty"`
	Targeting []map[string]string `json:"targeting,omitempty"`
	Error     string              `json:"error,omitempty"`
	Ext       interface{}         `json:"ext,omitempty"`
}

type bidResponseAdpod struct {
	AdPodBids   []*adPodBid `json:"adpods,omitempty"`
	Ext         interface{} `json:"ext,omitempty"`
	RedirectURL *string     `json:"redirect_url,omitempty"`
}

func formAdpodBidErrorResponse(id string, err string, ext interface{}) []byte {
	errResponse := adPodBid{
		ID:    id,
		Error: err,
		Ext:   ext,
	}

	response, _ := json.Marshal(errResponse)
	return response
}

type jsonResponse struct {
	cacheClient *pbc.Client
	redirectURL string
	debug       string
}

func (jr *jsonResponse) formJSONResponse(response []byte) []byte {
	var bidResponse *openrtb2.BidResponse

	err := json.Unmarshal(response, &bidResponse)
	if err != nil {
		if len(jr.redirectURL) > 0 && jr.debug == "0" {
			return []byte(jr.redirectURL)
		}
		return formAdpodBidErrorResponse("", "error in unmarshaling the auction response", nil)
	}

	return jr.getJsonResponse(bidResponse)
}

func (jr *jsonResponse) getJsonResponse(bidResponse *openrtb2.BidResponse) []byte {
	if bidResponse == nil {
		if len(jr.redirectURL) > 0 && jr.debug == "0" {
			return []byte(jr.redirectURL)
		}
		return formAdpodBidErrorResponse("", "empty bid response recieved", nil)
	}

	if bidResponse.SeatBid == nil {
		if len(jr.redirectURL) > 0 && jr.debug == "0" {
			return []byte(jr.redirectURL)
		}
		return formAdpodBidErrorResponse("", "no seat bids in the response", bidResponse.Ext)
	}

	bidArrayMap := make(map[string][]openrtb2.Bid)
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			if bid.Price > 0 {
				impId, _ := models.GetImpressionID(bid.ImpID)
				bids := bidArrayMap[impId]
				bids = append(bids, bid)
				bidArrayMap[impId] = bids
			}
		}
	}
	adPodBids := formAdpodBids(bidArrayMap, jr.cacheClient)

	if len(jr.redirectURL) > 0 && jr.debug == "0" {
		return getRedirectResponse(adPodBids, jr.redirectURL)
	}

	adpodResponse := bidResponseAdpod{AdPodBids: adPodBids, Ext: bidResponse.Ext}
	response, _ := json.Marshal(adpodResponse)

	return response
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
			custParams.Set(targetingKey, target[targetingKey])
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
		adpodBid := adPodBid{
			ID: impId,
		}
		sort.Slice(bids, func(i, j int) bool { return bids[i].Price > bids[j].Price })

		cacheIds, err := cacheAllBids(cacheClient, bids)
		if err != nil {
			adpodBid.Error = err.Error()
			adpodBids = append(adpodBids, &adpodBid)
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
		adpodBids = append(adpodBids, &adpodBid)
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
			for k, v := range bidExt.Prebid.Targeting {
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
