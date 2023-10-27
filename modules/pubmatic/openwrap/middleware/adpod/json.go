package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	validator "github.com/asaskevich/govalidator"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

const (
	slotKeyFormat = "s%d_%s"
)

var (
	redirectTargetingKeys = []string{"pwtpb", "pwtdur", "pwtcid", "pwtpid", "pwtdealtier", "pwtdid"}
)

type jsonBid struct {
	*openrtb2.Bid
	Seat string
}

type adPodBid struct {
	ID        *string             `json:"id,omitempty"`
	Targeting []map[string]string `json:"targeting,omitempty"`
	Error     string              `json:"error,omitempty"`
	Ext       interface{}         `json:"ext,omitempty"`
}

type bidResponseAdpod struct {
	AdPodBids   []*adPodBid `json:"adpods,omitempty"`
	Ext         interface{} `json:"ext,omitempty"`
	RedirectURL *string     `json:"redirect_url,omitempty"`
}

func getAndValidateRedirectURL(r *http.Request) (string, string, CustomError) {
	params := r.URL.Query()
	debug := params.Get(models.Debug)

	format := strings.ToLower(strings.TrimSpace(params.Get(models.ResponseFormatKey)))
	if format != "" {
		if format != models.ResponseFormatJSON && format != models.ResponseFormatRedirect {
			return "", debug, NewError(634, "Invalid response format, must be 'json' or 'redirect'")
		}
	}

	owRedirectURL := params.Get(models.OWRedirectURLKey)
	if len(owRedirectURL) > 0 {
		owRedirectURL = strings.TrimSpace(owRedirectURL)
		if format == models.ResponseFormatRedirect && !isValidURL(owRedirectURL) {
			return "", debug, NewError(633, "Invalid redirect URL")
		}
	}

	return owRedirectURL, debug, nil
}

func isValidURL(urlVal string) bool {
	if !(strings.HasPrefix(urlVal, "http://") || strings.HasPrefix(urlVal, "https://")) {
		return false
	}
	return validator.IsRequestURL(urlVal) && validator.IsURL(urlVal)
}

func formJSONResponse(response []byte, redirectURL, debug string) []byte {
	var bidResponse *openrtb2.BidResponse

	err := json.Unmarshal(response, &bidResponse)
	if err != nil {
		return response
	}

	jsonResponse, err := getJsonResponse(bidResponse, redirectURL, debug)
	if err != nil {
		return response
	}

	return jsonResponse
}

func getJsonResponse(bidResponse *openrtb2.BidResponse, redirectURL, debug string) ([]byte, error) {
	if bidResponse == nil || bidResponse.SeatBid == nil {
		return nil, errors.New("recieved invalid bidResponse")
	}

	bidArrayMap := make(map[string][]jsonBid)
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			if bid.Price > 0 {
				impId, _ := models.GetImpressionID(bid.ImpID)
				bids, ok := bidArrayMap[impId]
				if !ok {
					bids = make([]jsonBid, 0)
				}

				bids = append(bids, jsonBid{Bid: &bid, Seat: seatBid.Seat})
				bidArrayMap[impId] = bids
			}
		}
	}

	adPodBids := formAdpodBids(bidArrayMap)

	var response []byte
	if len(redirectURL) > 0 && debug != "1" {
		response = getRedirectResponse(adPodBids, redirectURL)
	} else {
		var err error
		adpodResponse := bidResponseAdpod{AdPodBids: adPodBids, Ext: bidResponse.Ext}
		response, err = json.Marshal(adpodResponse)
		if err != nil {
			return nil, err
		}
	}

	return response, nil

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

func formAdpodBids(bidsMap map[string][]jsonBid) []*adPodBid {
	var adpodBids []*adPodBid
	for impId, bids := range bidsMap {
		adpodBid := adPodBid{
			ID: &impId,
		}
		sort.Slice(bids, func(i, j int) bool { return bids[i].Price > bids[j].Price })

		targetings := []map[string]string{}
		for i := 0; i < len(bids); i++ {
			slotNo := i + 1
			targeting := createTargetting(bids[i], slotNo)
			targetings = append(targetings, targeting)
		}
		adpodBid.Targeting = targetings
		adpodBids = append(adpodBids, &adpodBid)
	}

	return adpodBids
}

func prepareSlotLevelKey(slotNo int, key string) string {
	return fmt.Sprintf(slotKeyFormat, slotNo, key)
}

func createTargetting(bid jsonBid, slotNo int) map[string]string {
	targetingKeyValMap := make(map[string]string)

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
