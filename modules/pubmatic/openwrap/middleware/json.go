package middleware

import (
	"encoding/json"

	"github.com/prebid/openrtb/v19/openrtb2"
	pbc "github.com/prebid/prebid-server/prebid_cache_client"
)

func FormJSONResponse(client *pbc.Client, response []byte) []byte {
	bidResponse := openrtb2.BidResponse{}

	err := json.Unmarshal(response, &bidResponse)
	if err != nil {
		return response
	}

	return response
}
