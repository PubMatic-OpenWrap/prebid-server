package ortbbidder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

// adapter implements adapters.Bidder interface
type adapter struct {
	adapterInfo
	bidderParamsConfig *bidderparams.BidderConfig
	parser             Parser
	paramMapperFactory ParamMapperFactory
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

// global instance to hold bidderParamsConfig
var g_bidderParamsConfig *bidderparams.BidderConfig

// InitBidderParamsConfig initializes a g_bidderParamsConfig instance from the files provided in dirPath.
func InitBidderParamsConfig(requestParamsDirPath, responseParamsDirPath string) (err error) {
	g_bidderParamsConfig, err = bidderparams.LoadBidderConfig(requestParamsDirPath, responseParamsDirPath, isORTBBidder)
	return err
}

// makeRequest converts openrtb2.BidRequest to adapters.RequestData, sets requestParams in request if required
func (o adapterInfo) makeRequest(request *openrtb2.BidRequest, requestParams map[string]bidderparams.BidderParamMapper) (*adapters.RequestData, error) {
	if request == nil {
		return nil, fmt.Errorf("found nil request")
	}
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request %s", err.Error())
	}
	requestBody, err = setRequestParams(requestBody, requestParams)
	if err != nil {
		return nil, err
	}
	return &adapters.RequestData{
		Method: http.MethodPost,
		Uri:    o.Endpoint,
		Body:   requestBody,
		Headers: http.Header{
			"Content-Type": {"application/json;charset=utf-8"},
			"Accept":       {"application/json"},
		},
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
	return &adapter{
		adapterInfo:        adapterInfo{config, extraAdapterInfo, bidderName},
		bidderParamsConfig: g_bidderParamsConfig,
		parser:             &ParserImpl{},
		paramMapperFactory: ParamMapperFactoryImpl{},
	}, nil
}

// MakeRequests prepares oRTB bidder-specific request information using which prebid server make call(s) to bidder.
func (o *adapter) MakeRequests(request *openrtb2.BidRequest, requestInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	if request == nil || requestInfo == nil {
		return nil, []error{fmt.Errorf("Found either nil request or nil requestInfo")}
	}
	if o.bidderParamsConfig == nil {
		return nil, []error{fmt.Errorf("Found nil bidderParamsConfig")}
	}
	var errs []error
	adapterInfo := o.adapterInfo
	requestParams, _ := o.bidderParamsConfig.GetRequestParams(o.bidderName.String())

	// bidder request supports single impression in single HTTP call.
	if adapterInfo.extraInfo.RequestMode == RequestModeSingle {
		requestData := make([]*adapters.RequestData, 0, len(request.Imp))
		requestCopy := *request
		for _, imp := range request.Imp {
			requestCopy.Imp = []openrtb2.Imp{imp} // requestCopy contains single impression
			reqData, err := adapterInfo.makeRequest(&requestCopy, requestParams)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			requestData = append(requestData, reqData)
		}
		return requestData, errs
	}
	// bidder request supports multi impressions in single HTTP call.
	requestData, err := adapterInfo.makeRequest(request, requestParams)
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

	bidResponse := &adapters.BidderResponse{
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

	bidResponse, err := o.makeBids(response, bidResponse)
	if err != nil {
		return nil, []error{err}
	}
	return bidResponse, nil
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

// MakeBids prepares bidderResponse from the oRTB bidder server's http.Response
func (o *adapter) makeBids(response openrtb2.BidResponse, bidderResponse *adapters.BidderResponse) (*adapters.BidderResponse, error) {
	responseParmas, _ := o.bidderParamsConfig.GetResponseParams(o.bidderName.String())
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ortb response %s", err.Error())
	}

	adapterResonseBytes, err := json.Marshal(bidderResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ortb response %s", err.Error())
	}

	responseBytes, err = setResponseParams(responseBytes, adapterResonseBytes, responseParmas, o.parser)
	if err != nil {
		return bidderResponse, err
	}

	responseBytes, err = setResponseParams1(responseBytes, adapterResonseBytes, responseParmas, o.paramMapperFactory)
	if err != nil {
		return bidderResponse, err
	}

	var resp *adapters.BidderResponse
	err = json.Unmarshal(responseBytes, &resp)
	return resp, err
}

// implementation using Parser
func setResponseParams(responseBody, adapterResponseBody json.RawMessage, responseParams map[string]bidderparams.BidderParamMapper, parser Parser) ([]byte, error) {
	if len(responseBody) == 0 || len(adapterResponseBody) == 0 {
		return adapterResponseBody, nil
	}

	response := map[string]any{}
	err := json.Unmarshal(responseBody, &response)
	if err != nil {
		return nil, err
	}

	bidderResponse := map[string]any{}
	err = json.Unmarshal(adapterResponseBody, &bidderResponse)
	if err != nil {
		return nil, err
	}

	for paramName, callback := range getRequestParamParser() {
		paramMapper, ok := responseParams[paramName]
		if !ok {
			continue
		}
		callback(parser, response, bidderResponse, paramMapper.GetLocation())
	}

	seatBids, ok := response["seatbid"].([]any)
	if !ok {
		return nil, fmt.Errorf("error:[invalid_seatbid_found_in_responsebody], seatbid:[%v]", response["seatbid"])
	}
	index := 0
	for _, seatBid := range seatBids {
		seatBid, ok := seatBid.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("error:[invalid_seatbid_found_in_seatbids], seatbid:[%v]", seatBids)
		}

		bids, ok := seatBid["bid"].([]any)
		if !ok {
			return nil, fmt.Errorf("error:[invalid_bid_found_in_seatbid], bid:[%v]", seatBid["bid"])
		}

		typeBids := bidderResponse["Bids"].([]any)
		for _, bid := range bids {
			bid, ok := bid.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("error:[invalid_bid_found_in_bids], bid:[%v]", bid)
			}

			typeBid := typeBids[index]
			for paramName, callback := range getBidParamParser() {
				paramMapper, ok := responseParams[paramName]
				if !ok {
					continue
				}
				callback(parser, bid, typeBid.(map[string]any), paramMapper.GetLocation())
			}
			index++
		}
	}

	return json.Marshal(bidderResponse)
}

// implementation using ParamMapperFactory
func setResponseParams1(responseBody, adapterResponseBody json.RawMessage, responseParams map[string]bidderparams.BidderParamMapper, parserFactory ParamMapperFactory) ([]byte, error) {
	if len(responseBody) == 0 || len(adapterResponseBody) == 0 {
		return adapterResponseBody, nil
	}

	response := map[string]any{}
	err := json.Unmarshal(responseBody, &response)
	if err != nil {
		return nil, err
	}

	bidderResponse := map[string]any{}
	err = json.Unmarshal(adapterResponseBody, &bidderResponse)
	if err != nil {
		return nil, err
	}

	seatBids, ok := response["seatbid"].([]any)
	if !ok {
		return nil, fmt.Errorf("error:[invalid_seatbid_found_in_responsebody], seatbid:[%v]", response["seatbid"])
	}
	index := 0
	for _, seatBid := range seatBids {
		seatBid, ok := seatBid.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("error:[invalid_seatbid_found_in_seatbids], seatbid:[%v]", seatBids)
		}

		bids, ok := seatBid["bid"].([]any)
		if !ok {
			return nil, fmt.Errorf("error:[invalid_bid_found_in_seatbid], bid:[%v]", seatBid["bid"])
		}

		typeBids := bidderResponse["Bids"].([]any)
		for _, bid := range bids {
			bid, ok := bid.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("error:[invalid_bid_found_in_bids], bid:[%v]", bid)
			}

			typeBid := typeBids[index]
			for paramName, mapper := range parserFactory.NewBidParamMapper() {
				paramMapper, ok := responseParams[paramName]
				if !ok {
					continue
				}
				mapper.ProcessParam(bid, typeBid.(map[string]any), paramMapper.GetLocation())
			}
			index++
		}
	}

	return json.Marshal(adapterResponseBody)
}
