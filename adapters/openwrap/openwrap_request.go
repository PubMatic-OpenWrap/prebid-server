package openwrap

import (
	"encoding/json"
	"net/http"

	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type ExtRequestORTB  map[string]interface{}     

func GetRequestExtORTB(prebidExt *openrtb_ext.ExtOWRequest) (*ExtRequestORTB, error) {
	var requestExt *ExtRequestORTB
	var mapExt map[string]interface{}

	if prebidExt.Prebid.BidderParams != nil {
		if err := json.Unmarshal(prebidExt.Prebid.BidderParams, &mapExt); err != nil {
			return nil, &errortypes.BadInput{
				Message: "Impression extension not provided or can't be unmarshalled",
			}
		}

		if ext, ok := mapExt["requestExt"]; ok {
			extBytes, err := json.Marshal(ext)
			if err != nil {
				return nil, &errortypes.BadInput{
					Message: "Error marshalling impression extension",
				}
			}

			if err := json.Unmarshal(extBytes, &requestExt); err != nil {
				return nil, &errortypes.BadInput{
					Message: "Error unmarshalling impression extension to ExtRequestPrebidOnsite",
				}
			}
		} else {
			return nil, &errortypes.BadInput{
				Message: "Impression extension not provided",
			}
		}
	}

	return requestExt, nil
}

func GetOWRequestExt(request *openrtb2.BidRequest) (*openrtb_ext.ExtOWRequest, error) {
	var requestExt openrtb_ext.ExtOWRequest

	if request.Ext != nil {
		if err := json.Unmarshal(request.Ext, &requestExt); err != nil {
			return nil, &errortypes.BadInput{
				Message: "Request extension not provided or can't be unmarshalled",
			}
		}
	}

	return &requestExt, nil
}


func GetRequestExt(request *openrtb2.BidRequest) (
	 *ExtRequestORTB, []error) {
	var requestOWExt *openrtb_ext.ExtOWRequest
	var requestExtORTB *ExtRequestORTB

	var err error
	var errors []error

	requestOWExt, err = GetOWRequestExt(request)
	if err != nil {
		errors = append(errors, err)
	}

	requestExtORTB, err = GetRequestExtORTB(requestOWExt)
	if err != nil {
		errors = append(errors, err)
	} else {

	}

	if len(errors) > 0 {
		return nil, errors
	}

	return requestExtORTB, nil
}

func (a *OpenWrapAdapter) MakeRequests(request *openrtb2.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	requestExt, errors := GetRequestExt(request)
	if len(errors) > 0 {
		return nil, errors
	}

	var headerValue interface{}

	// Check if the "header" key exists and extract its value
	if value, exists := (*requestExt)["user_headers"]; exists {
		headerValue = value
		// Remove the "header" key from the map
		delete(*requestExt, "user_headers")
	}
	// Convert requestExt to json.RawMessage
	extJSON, err := json.Marshal(requestExt)
	if err != nil {
		return nil, []error{err}
	}

	request.Ext = extJSON

	for i := 0; i < len(request.Imp); i++ {
		var impExt map[string]interface{}

		if request.Imp[i].Ext != nil {
			var err1 error
			if err1 = json.Unmarshal(request.Imp[i].Ext, &impExt); err1 == nil {
				bidderExt := impExt["bidder"].(map[string]interface{})
				impExtJSON, err3 := json.Marshal(bidderExt["impExt"])
					if err3 != nil {
						return nil, []error{err}
					}
					request.Imp[i].Ext = impExtJSON
				
				} else{
					request.Imp[i].Ext = nil
				}
		} else{
			request.Imp[i].Ext = nil
		}
	}

	reqJSON, err := json.Marshal(request)
	if err != nil {
		return nil, []error{err}
	}

	headers := http.Header{}
	// Assert headerValue to be map[string]interface{} and add to headers
	if headerMap, ok := headerValue.(map[string]interface{}); ok {
		for key, value := range headerMap {
			// Convert the value to a string if possible
			if strValue, ok := value.(string); ok {
				headers.Add(key, strValue)
			}
		}
	}
	// Check if "Content-Type" exists and delete it
	if _, ok := headers["Content-Type"]; ok {
		headers.Del("Content-Type")
	}

	// Add "Content-Type: application/json"
	headers.Add("Content-Type", "application/json")

	return []*adapters.RequestData{{
		Method:  "POST",
		Uri:     a.endpoint,
		Body:    reqJSON,
		Headers: headers,
	}}, nil
	}

