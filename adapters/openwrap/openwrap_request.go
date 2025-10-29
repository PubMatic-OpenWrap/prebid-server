package openwrap

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type ExtRequestORTB map[string]interface{}

func GetRequestExtORTB(prebidExt *openrtb_ext.ExtOWRequest) (*ExtRequestORTB, bool, error) {
	var requestExt *ExtRequestORTB
	var mapExt map[string]interface{}
	debug := prebidExt.Prebid.Debug

	if prebidExt.Prebid.BidderParams != nil {
		if err := json.Unmarshal(prebidExt.Prebid.BidderParams, &mapExt); err != nil {
			return nil, debug, &errortypes.BadInput{
				Message: "Impression extension not provided or can't be unmarshalled",
			}
		}

		if ext, ok := mapExt["requestExt"]; ok {
			extBytes, err := json.Marshal(ext)
			if err != nil {
				return nil, debug, &errortypes.BadInput{
					Message: "Error marshalling impression extension",
				}
			}

			if err := json.Unmarshal(extBytes, &requestExt); err != nil {
				return nil, debug, &errortypes.BadInput{
					Message: "Error unmarshalling impression extension to ExtRequestPrebidOnsite",
				}
			}
		} else {
			return nil, debug, &errortypes.BadInput{
				Message: "Impression extension not provided",
			}
		}
	}

	return requestExt, debug, nil
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
	*ExtRequestORTB, bool, []error) {
	var requestOWExt *openrtb_ext.ExtOWRequest
	var requestExtORTB *ExtRequestORTB
	var debug bool
	var err error
	var errors []error

	requestOWExt, err = GetOWRequestExt(request)
	if err != nil {
		errors = append(errors, err)
	}

	requestExtORTB, debug, err = GetRequestExtORTB(requestOWExt)
	if err != nil {
		errors = append(errors, err)
	} else {

	}

	if len(errors) > 0 {
		return nil, debug, errors
	}

	return requestExtORTB, debug, nil
}

func (a *OpenWrapAdapter) MakeRequests(request *openrtb2.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	requestExt, _, errors := GetRequestExt(request)
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

	// Backup Publisher ID
	pubID := ""
	if request.Site != nil && request.Site.Publisher != nil && request.Site.Publisher.ID != "" {
		pubID = request.Site.Publisher.ID
	} else if request.App != nil && request.App.Publisher != nil && request.App.Publisher.ID != "" {
		pubID = request.App.Publisher.ID
	}

	// Check if site.ext.sspreq is true and perform swapping
	var isSSPReq bool
	var queryParams string
	if request.Site != nil && request.Site.Ext != nil {
		var siteExt map[string]interface{}
		if err := json.Unmarshal(request.Site.Ext, &siteExt); err == nil {
			// Read queryparams and store in string
			if qp, exists := siteExt["queryParams"]; exists {
				if qpStr, ok := qp.(string); ok {
					queryParams = qpStr
				}
			}

			if sspreq, exists := siteExt["sspreq"]; exists {
				if sspreqBool, ok := sspreq.(bool); ok && sspreqBool {
					isSSPReq = true
					// Swap site.publisher.id with site.ext.cpid
					if cpid, exists := siteExt["cpid"]; exists {
						if cpidStr, ok := cpid.(string); ok && cpidStr != "" && request.Site.Publisher != nil && request.Site.Publisher.ID != "" {
							// Swap the values
							tempCpid := cpidStr
							siteExt["cpid"] = request.Site.Publisher.ID
							request.Site.Publisher.ID = tempCpid
						}
					}
					// Delete site.ext before sending to SSP
					request.Site.Ext = nil
				}
			}
		}
	}

	// Handle app extension similar to site
	if request.App != nil && request.App.Ext != nil {
		var appExt map[string]interface{}
		if err := json.Unmarshal(request.App.Ext, &appExt); err == nil {
			// Read queryparams and store in string
			if qp, exists := appExt["queryParams"]; exists {
				if qpStr, ok := qp.(string); ok {
					queryParams = qpStr
				}
			}

			if sspreq, exists := appExt["sspreq"]; exists {
				if sspreqBool, ok := sspreq.(bool); ok && sspreqBool {
					isSSPReq = true
					// Swap app.publisher.id with app.ext.cpid
					if cpid, exists := appExt["cpid"]; exists {
						if cpidStr, ok := cpid.(string); ok && cpidStr != "" && request.App.Publisher != nil && request.App.Publisher.ID != "" {
							// Swap the values
							tempCpid := cpidStr
							appExt["cpid"] = request.App.Publisher.ID
							request.App.Publisher.ID = tempCpid
						}
					}
					// Delete app.ext before sending to SSP
					request.App.Ext = nil
				}
			}
		}
	}

	for i := 0; i < len(request.Imp); i++ {
		var impExt map[string]interface{}
		if request.Imp[i].Ext != nil {
			var err1 error
			if err1 = json.Unmarshal(request.Imp[i].Ext, &impExt); err1 == nil {
				bidderExt := impExt["bidder"].(map[string]interface{})

				// Get the actual impression extension from bidderExt["impExt"]
				var actualImpExt map[string]interface{}
				if impExtData, exists := bidderExt["impExt"]; exists {
					if impExtMap, ok := impExtData.(map[string]interface{}); ok {
						actualImpExt = impExtMap

						// Check if cpinvid exists and swap with imp.tagid only if sspReq is true
						if isSSPReq {
							if cpinvid, exists := actualImpExt["cpinvid"]; exists {
								if cpinvidStr, ok := cpinvid.(string); ok && cpinvidStr != "" && request.Imp[i].TagID != "" {
									// Swap imp.tagid with imp.ext.cpinvid
									tempTagID := request.Imp[i].TagID
									request.Imp[i].TagID = cpinvidStr
									actualImpExt["cpinvid"] = tempTagID
								}
							}
							// Delete cpinvid from imp.ext before sending to SSP
							delete(actualImpExt, "cpinvid")
						}
					}
				}

				impExtJSON, err3 := json.Marshal(bidderExt["impExt"])
				if err3 != nil {
					return nil, []error{err}
				}
				request.Imp[i].Ext = impExtJSON

			} else {
				request.Imp[i].Ext = nil
			}
		} else {
			request.Imp[i].Ext = nil
		}

		// Create the Native object and fill with the desired JSON string.
		// This JSON string represents the native.request payload.
		// Retrieve width and height from Banner object if available.
		/*var width, height int64
		if request.Imp[i].Banner != nil {
			// Check if Banner has direct W and H values.
			if request.Imp[i].Banner.W != nil && request.Imp[i].Banner.H != nil {
				width = *request.Imp[i].Banner.W
				height = *request.Imp[i].Banner.H
			} else if len(request.Imp[i].Banner.Format) > 0 {
				// Fallback: use the first format's width and height, if present.
				if request.Imp[i].Banner.Format[0].W > 0 && request.Imp[i].Banner.Format[0].H > 0 {
					width = request.Imp[i].Banner.Format[0].W
					height = request.Imp[i].Banner.Format[0].H
				}
			}
		}
		// Build the native request JSON with dynamic width and height values.
		nativeReq := fmt.Sprintf(`{ "ver": "1.1", "assets": [ { "id": 1, "required": 1, "img": { "w": %d, "h": %d, "type": 3 } }, { "id": 2, "required": 1, "img": { "type": 1 } }, { "id": 12, "required": 1, "data": { "type": 2 } }], "eventtrackers": [ { "event": 1, "methods": [ 1, 2 ] } ] }`, width, height)
		request.Imp[i].Native = &openrtb2.Native{
			Request: nativeReq,
			Ver:     "1.1",
		}*/
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

	// Remove debug key from queryParams if present
	if queryParams != "" {
		// Split by & and filter out debug parameters
		params := strings.Split(queryParams, "&")
		var filteredParams []string
		for _, param := range params {
			if !strings.HasPrefix(param, "debug=") && !strings.HasPrefix(param, "debug&") {
				filteredParams = append(filteredParams, param)
			}
		}
		queryParams = strings.Join(filteredParams, "&")
	}

	// Determine which endpoint to use based on sspreq
	var endpoint string
	if isSSPReq {
		// Use SSP endpoint when sspreq is true
		if queryParams != "" {
			endpoint = a.sspEndpoint + "?" + queryParams
		} else {
			endpoint = a.sspEndpoint
		}
	} else {
		// Use regular endpoint when sspreq is false or not present
		endpoint = a.endpoint
	}
	// Print reqJSON if PubID is 167 (Kohls)
	if pubID == "167" || pubID == "164" {
		glog.Errorf("KOHLSEBAY_SSPREQUEST - PubID: %s, SSPReq: %v, RequestBody: %s", pubID, isSSPReq, string(reqJSON))
	}
	return []*adapters.RequestData{{
		Method:  "POST",
		Uri:     endpoint,
		Body:    reqJSON,
		Headers: headers,
	}}, nil
}
