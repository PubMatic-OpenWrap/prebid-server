package adbutler_onsite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
)

func (a *AdButlerOnsiteAdapter) MakeBids(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {

	if response.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if response.StatusCode == http.StatusBadRequest {
		err := &errortypes.BadInput{
			Message: "Unexpected status code: 400. Bad request from Adbutler.",
		}
		return nil, []error{err}
	}

	if response.StatusCode != http.StatusOK {
		err := &errortypes.BadServerResponse{
			Message: fmt.Sprintf("Unexpected status code: %d", response.StatusCode),
		}
		return nil, []error{err}
	}

	//Temporarily for Debugging
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, response.Body, "", "  ")
	if err != nil {
		fmt.Println("Failed to parse JSON:", err)
		return nil, []error{err}
	}
	fmt.Println(prettyJSON.String())

	return nil, nil
}

