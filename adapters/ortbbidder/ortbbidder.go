package ortbbidder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// adapter implements adapters.Bidder interface
type adapter struct {
	adapterInfo
	mapper *Mapper
}

const (
	RequestModeSingle string = "single"
)

// adapterInfo contains oRTB bidder specific info required in MakeRequests/MakeBids functions
type adapterInfo struct {
	config.Adapter
	extraInfo  extraAdapterInfo
	bidderName openrtb_ext.BidderName
}
type extraAdapterInfo struct {
	RequestMode string `json:"requestMode"`
}

// prepareRequestData generates the RequestData by marshalling the request and returns it
// func (o adapterInfo) prepareRequestData(request *openrtb2.BidRequest, mapper map[string]paramDetails) (*adapters.RequestData, error) {
// 	if request == nil {
// 		return nil, fmt.Errorf("found nil request")
// 	}
// 	body, err := json.Marshal(request)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal request %s", err.Error())
// 	}
// 	// add bidder-params inside the reqData
// 	fmt.Println("body without -[%s]", string(body))
// 	// single impression in request
// 	bidderParamsExt, _, _, err := jsonparser.Get(request.Imp[0].Ext, "bidder")
// 	if err != nil {
// 		return nil, err
// 	}

// 	bidderParams := JSONNode{}
// 	err = json.Unmarshal(bidderParamsExt, &bidderParams)
// 	if err != nil {
// 		return nil, err
// 	}

// 	requestJsonNode := JSONNode{}
// 	err = json.Unmarshal(body, &requestJsonNode)
// 	if err != nil {
// 		return nil, err
// 	}

// 	for bidderParamName, bidderParamValue := range bidderParams {
// 		details, ok := mapper[bidderParamName]
// 		if !ok {
// 			continue
// 		}
// 		loc := details.location
// 		loc, _ = strings.CutPrefix(loc, "req.")
// 		if strings.HasPrefix(loc, "imp.") {
// 			loc, _ = strings.CutPrefix(loc, "imp.")
// 			imps, ok := requestJsonNode["imp"].([]interface{})
// 			if !ok {
// 				return nil, err
// 			}
// 			imp, ok := imps[0].(JSONNode)
// 			if !ok {
// 				return nil, err
// 			}
// 			SetValue(imp, loc, bidderParamValue)
// 		} else {
// 			SetValue(requestJsonNode, loc, bidderParamValue)
// 		}
// 	}

// 	body, err = json.Marshal(requestJsonNode)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal request %s", err.Error())
// 	}
// 	fmt.Println("body with -[%s]", string(body))

// 	return &adapters.RequestData{
// 		Method: http.MethodPost,
// 		Uri:    o.Endpoint,
// 		Body:   body,
// 	}, nil
// }

// multi-imp calls using map[string]interface{}
// func (o adapterInfo) prepareRequestData(request *openrtb2.BidRequest, mapper map[string]paramDetails) (*adapters.RequestData, error) {
// 	if request == nil {
// 		return nil, fmt.Errorf("found nil request")
// 	}
// 	body, err := json.Marshal(request)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal request %s", err.Error())
// 	}
// 	// add bidder-params inside the reqData
// 	fmt.Println("body without -[%s]", string(body))
// 	requestJsonNode := JSONNode{}
// 	err = json.Unmarshal(body, &requestJsonNode)
// 	if err != nil {
// 		return nil, err
// 	}
// 	impList, ok := requestJsonNode["imp"].([]interface{})
// 	if !ok {
// 		return nil, nil
// 	}
// 	for ind, eachImp := range impList {
// 		requestJsonNode["imp"] = eachImp
// 		imp, ok := eachImp.(map[string]interface{})
// 		if !ok {
// 			return nil, nil
// 		}
// 		ext, ok := imp["ext"].(map[string]interface{})
// 		if !ok {
// 			return nil, nil
// 		}
// 		bidderParams, ok := ext["bidder"].(map[string]interface{})
// 		if !ok {
// 			return nil, nil
// 		}
// 		for bidderParamName, bidderParamValue := range bidderParams {
// 			details, ok := mapper[bidderParamName]
// 			if !ok {
// 				continue
// 			}
// 			loc := details.location
// 			loc, _ = strings.CutPrefix(loc, "req.")
// 			if SetValue(requestJsonNode, loc, bidderParamValue) {
// 				delete(bidderParams, bidderParamName)
// 			}
// 		}
// 		impList[ind] = requestJsonNode["imp"]
// 	}
// 	requestJsonNode["imp"] = impList
// 	body, err = json.Marshal(requestJsonNode)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal request %s", err.Error())
// 	}
// 	fmt.Println("body with -[%s]", string(body))
// 	bidreq := &openrtb2.BidRequest{}
// 	err = json.Unmarshal(body, bidreq)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal request %s", err.Error())
// 	}
// 	return &adapters.RequestData{
// 		Method: http.MethodPost,
// 		Uri:    o.Endpoint,
// 		Body:   body,
// 	}, nil
// }

// multi-imp calls using jsonparser
func (o adapterInfo) prepareRequestData(request *openrtb2.BidRequest, mapper map[string]paramDetails) (*adapters.RequestData, error) {
	if request == nil {
		return nil, fmt.Errorf("found nil request")
	}
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request %s", err.Error())
	}
	// add bidder-params inside the reqData
	fmt.Println("body without -[%s]", string(requestBody))

	updatedImp := make([][]byte, 0)
	impIndex := 0
	_, err = jsonparser.ArrayEach(requestBody, func(imp []byte, dataType jsonparser.ValueType, offset int, _ error) {
		impIndex++

		impBody := imp
		bidderParamsBytes, _, _, err := jsonparser.Get(imp, "ext", "bidder")
		if err != nil {
			return
		}
		bidderParams := make(map[string]interface{})
		err = json.Unmarshal(bidderParamsBytes, &bidderParams)
		if err != nil {
			return
		}

		for bidderParamName, bidderParamValue := range bidderParams {
			paramDetails, ok := mapper[bidderParamName]
			if !ok {
				continue
			}
			loc := paramDetails.location
			loc, _ = strings.CutPrefix(loc, "req.")
			if strings.HasPrefix(loc, "imp.") {
				loc, _ = strings.CutPrefix(loc, "imp.")
				locs := strings.Split(loc, ".")
				// TODO - this will not handle the complex object (it will set it as '"wrapper": "map[profile:300 version:100]"')
				// need to handle the %d/%f for numeric values
				impBody, err = jsonparser.Set(impBody, []byte(fmt.Sprintf("\"%v\"", bidderParamValue)), locs...)
				if err != nil {
					return
				}

			} else {
				locs := strings.Split(loc, ".")
				// requestBody, err = jsonparser.Set(requestBody, []byte(fmt.Sprintf("\"%v\"", bidderParamValue)), locs...)
				requestBody, err = jsonparser.Set(requestBody, []byte(fmt.Sprintf("\"%s\"", bidderParamValue)), locs...)
				if err != nil {
					return
				}
			}
			// TODO - remove the bidder-params once successfully updated the bidRequest
		}
		updatedImp = append(updatedImp, impBody)
	}, "imp")

	finalImps := []byte{}
	for i, v := range updatedImp {
		finalImps = append(finalImps, v...)
		if i != len(updatedImp)-1 {
			finalImps = append(finalImps, []byte(",")...)
		}
	}
	requestBody, err = jsonparser.Set(requestBody, finalImps, "imp")
	if err != nil {
		return nil, fmt.Errorf("failed to set imps %s", err.Error())
	}

	fmt.Println("body with -[%s]", string(requestBody))

	bidreq := &openrtb2.BidRequest{}
	err = json.Unmarshal(requestBody, bidreq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request %s", err.Error())
	}

	return &adapters.RequestData{
		Method: http.MethodPost,
		Uri:    o.Endpoint,
		Body:   requestBody,
	}, nil
}

// Builder returns an instance of oRTB adapter
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter, server config.Server) (adapters.Bidder, error) {
	extraAdapterInfo := extraAdapterInfo{}
	if len(config.ExtraAdapterInfo) > 0 {
		err := json.Unmarshal([]byte(config.ExtraAdapterInfo), &extraAdapterInfo)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse extra_info for bidder:[%s] err:[%s]", bidderName, err.Error())
		}
	}
	mapper, err := NewMapper("./static/bidder-params")
	if err != nil {
		return nil, fmt.Errorf("Failed to prepare bidder-param mapper for bidder:[%s] err:[%s]", bidderName, err.Error())
	}
	return &adapter{
		adapterInfo: adapterInfo{config, extraAdapterInfo, bidderName},
		mapper:      mapper,
	}, nil
}

// MakeRequests prepares oRTB bidder-specific request information using which prebid server make call(s) to bidder.
func (o *adapter) MakeRequests(request *openrtb2.BidRequest, requestInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	if request == nil || requestInfo == nil {
		return nil, []error{fmt.Errorf("Found either nil request or nil requestInfo")}
	}
	var errs []error
	adapterInfo := o.adapterInfo
	// bidder request supports single impression in single HTTP call.
	if adapterInfo.extraInfo.RequestMode == RequestModeSingle {
		requestData := make([]*adapters.RequestData, 0, len(request.Imp))
		requestCopy := *request
		for _, imp := range request.Imp {
			requestCopy.Imp = []openrtb2.Imp{imp} // requestCopy contains single impression
			reqData, err := adapterInfo.prepareRequestData(&requestCopy, o.mapper.bidderParamMapper[o.adapterInfo.bidderName.String()])
			if err != nil {
				errs = append(errs, err)
				continue
			}
			requestData = append(requestData, reqData)
		}
		return requestData, errs
	}
	// bidder request supports multi impressions in single HTTP call.
	requestData, err := adapterInfo.prepareRequestData(request, o.mapper.bidderParamMapper[o.adapterInfo.bidderName.String()])
	if err != nil {
		return nil, []error{err}
	}
	return []*adapters.RequestData{requestData}, nil
}

// MakeBids prepares bidderResponse from the oRTB bidder server's http.Response
func (o *adapter) MakeBids(request *openrtb2.BidRequest, requestData *adapters.RequestData, responseData *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	if responseData == nil || adapters.IsResponseStatusCodeNoContent(responseData) {
		return nil, nil
	}

	if err := adapters.CheckResponseStatusCodeForErrors(responseData); err != nil {
		return nil, []error{err}
	}

	var response openrtb2.BidResponse
	if err := json.Unmarshal(responseData.Body, &response); err != nil {
		return nil, []error{err}
	}

	bidResponse := adapters.BidderResponse{
		Bids: make([]*adapters.TypedBid, 0),
	}
	for _, seatBid := range response.SeatBid {
		for bidInd, bid := range seatBid.Bid {
			bidResponse.Bids = append(bidResponse.Bids, &adapters.TypedBid{
				Bid:     &seatBid.Bid[bidInd],
				BidType: getMediaTypeForBid(bid),
			})
		}
	}
	return &bidResponse, nil
}

// getMediaTypeForBid returns the BidType as per the bid.MType field
// bid.MType has high priority over bidExt.Prebid.Type
func getMediaTypeForBid(bid openrtb2.Bid) openrtb_ext.BidType {
	var bidType openrtb_ext.BidType
	if bid.MType > 0 {
		bidType = getMediaTypeForBidFromMType(bid.MType)
	} else {
		if bid.Ext != nil {
			var bidExt openrtb_ext.ExtBid
			err := json.Unmarshal(bid.Ext, &bidExt)
			if err == nil && bidExt.Prebid != nil {
				bidType, _ = openrtb_ext.ParseBidType(string(bidExt.Prebid.Type))
			}
		}
	}
	if bidType == "" {
		// TODO : detect mediatype from bid.AdM and request.imp parameter
	}
	return bidType
}

// getMediaTypeForBidFromMType returns the bidType from the MarkupType field
func getMediaTypeForBidFromMType(mtype openrtb2.MarkupType) openrtb_ext.BidType {
	var bidType openrtb_ext.BidType
	switch mtype {
	case openrtb2.MarkupBanner:
		bidType = openrtb_ext.BidTypeBanner
	case openrtb2.MarkupVideo:
		bidType = openrtb_ext.BidTypeVideo
	case openrtb2.MarkupAudio:
		bidType = openrtb_ext.BidTypeAudio
	case openrtb2.MarkupNative:
		bidType = openrtb_ext.BidTypeNative
	}
	return bidType
}

// isORTBBidder returns true if the bidder is an oRTB bidder
func isORTBBidder(bidderName string) bool {
	return strings.HasPrefix(bidderName, "owortb_")
}
