package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/endpoints/legacy/ctv"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func formJSONErrorResponse(request *http.Request, err error) []byte {
	type errResponse struct {
		Id  string                 `json:"id"`
		NBR *openrtb3.NoBidReason  `json:"nbr,omitempty"`
		Ext map[string]interface{} `json:"ext,omitempty"`
	}

	nbr := openrtb3.NoBidInvalidRequest.Ptr()
	if cerr, ok := err.(*ctv.ParseError); ok {
		nbr = cerr.NBR()
	}

	response := errResponse{
		Id:  request.URL.Query().Get(ctv.ORTBBidRequestID),
		NBR: nbr,
	}

	if request.URL.Query().Get(models.Debug) == "1" {
		response.Ext = map[string]interface{}{
			"prebid": map[string]interface{}{
				"errors": []string{err.Error()},
			},
		}
	}

	responseBytes, _ := json.Marshal(response)
	return responseBytes
}
