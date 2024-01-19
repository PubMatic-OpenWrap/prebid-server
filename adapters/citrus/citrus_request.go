package citrus

import (
	"encoding/json"
	"net/http"

	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/adapters"
)

type CitrusRequest struct {
	SessionID       string     `json:"sessionId"`
	CatalogID       string     `json:"catalogId"`   
	Placement       string     `json:"placement"`
	SearchTerm      string     `json:"searchTerm"`
	ProductFilters  [][]string `json:"productFilters"`
	MaxNumberOfAds  int        `json:"maxNumberOfAds"`
	DynamicFields   map[string]interface{} `json:"-"`
}


func (a *CitrusAdapter) MakeRequests(request *openrtb2.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	commerceExt, siteExt, bidderParams,errors := adapters.ValidateCommRequest(request)
	if len(errors) > 0 {
		return nil, errors
	}

    var configValueMap = make(map[string]string)
    var configTypeMap = make(map[string]int)
	for _,obj := range commerceExt.Bidder.CustomConfig {
		configValueMap[obj.Key] = obj.Value
		configTypeMap[obj.Key] = obj.Type
	}

	var citrusReq CitrusRequest 
	//Assign Page Source if Present
	if siteExt != nil {
		citrusReq.Placement = siteExt.Page
	}

    //Retrieve AuthKey from Request and Build endpoint Url
	var authKey, catalogID string
	val, ok := configValueMap[adapters.AUCTIONDETAILS_PREFIX + AD_AUTH_KEY]
	if ok {
		authKey = val
	}
	val, ok = configValueMap[adapters.AUCTIONDETAILS_PREFIX + AD_CATALOG_ID]
	if ok {
		catalogID = val
	}
	
	citrusReq.CatalogID = catalogID
	citrusReq.SessionID = request.User.ID
	citrusReq.MaxNumberOfAds = commerceExt.ComParams.SlotsRequested
	citrusReq.SearchTerm = commerceExt.ComParams.SearchTerm
	
	//Add Category Params from AdRequest
	if commerceExt.ComParams.Filtering != nil && len(commerceExt.ComParams.Filtering) > 0 {
		productFilters := [][]string{}
	
		for _, filter := range commerceExt.ComParams.Filtering {
			filterValues := []string{}
			for _, value := range filter.Value {
				filterValues = append(filterValues, filter.Name+":"+value)
			}
			productFilters = append(productFilters, filterValues)
		}
	
		citrusReq.ProductFilters = productFilters
	}
	
	// Add other fields as needed
	citrusReq.DynamicFields = make(map[string]interface{})
	if bidderParams != nil {
		for key, value := range bidderParams {
			citrusReq.DynamicFields[key] = value
		}
	}
	
	reqJSON, err := json.Marshal(citrusReq)
	if err != nil {
		return nil, []error{err}
	}

	if len(citrusReq.DynamicFields) > 0 {
		dynamicFieldsJSON, err := json.Marshal(citrusReq.DynamicFields)
		if err != nil {
			return nil, []error{err}
		}
		reqJSON = append(reqJSON[:len(reqJSON)-1], byte(',')) // remove the closing brace
		reqJSON = append(reqJSON, dynamicFieldsJSON[1:]...)   // remove the opening brace of unknownFieldsJSON
	}
	
	headers := http.Header{}
	headers.Add("accept", "application/json")
	headers.Add("Content-Type", "application/json")
	headers.Add("Authorization", AUTH_PREFIX + authKey )

	return []*adapters.RequestData{{
		Method:  "POST",
		Uri:     a.endpoint,
		Body:    reqJSON,
		Headers: headers,
	}}, nil
}

