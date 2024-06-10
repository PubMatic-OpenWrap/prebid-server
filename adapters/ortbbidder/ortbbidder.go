package ortbbidder

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/resolver"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/jsonutil"
)

// adapter implements adapters.Bidder interface
type adapter struct {
	adapterInfo
	bidderParamsConfig *bidderparams.BidderConfig
	processor          *resolver.ParamResolver
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
		processor:          &resolver.ParamResolver{},
	}, nil
}

// MakeRequests prepares oRTB bidder-specific request information using which prebid server make call(s) to bidder.
func (o *adapter) MakeRequests(request *openrtb2.BidRequest, requestInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	if request == nil || requestInfo == nil {
		return nil, []error{errors.New("found either nil request or nil requestInfo")}
	}
	if o.bidderParamsConfig == nil {
		return nil, []error{errors.New("found nil bidderParamsConfig")}
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

	responseParmas, _ := o.bidderParamsConfig.GetResponseParams(o.bidderName.String())
	if responseParmas == nil {
		responseParmas = make(map[string]bidderparams.BidderParamMapper)
	}
	bidResponse, err := o.makeBids(responseData.Body, responseParmas)
	if err != nil {
		return nil, []error{err}
	}

	return bidResponse, nil
}

// isORTBBidder returns true if the bidder is an oRTB bidder
func isORTBBidder(bidderName string) bool {
	return strings.HasPrefix(bidderName, "owortb_")
}

// MakeBids prepares bidderResponse from the oRTB bidder server's http.Response
func (o *adapter) makeBids(bidderResponseBytes json.RawMessage, responseParmas map[string]bidderparams.BidderParamMapper) (*adapters.BidderResponse, error) {

	adapterResponseBytes, err := o.setResponseParams(bidderResponseBytes, responseParmas, &resolver.ParamResolver{})
	if err != nil {
		return nil, err
	}

	var resp *adapters.BidderResponse
	err = jsonutil.UnmarshalValid(adapterResponseBytes, &resp)
	return resp, err
}

func (o *adapter) setResponseParams(bidderResponseBody json.RawMessage, responseParams map[string]bidderparams.BidderParamMapper, paramResolver *resolver.ParamResolver) ([]byte, error) {

	bidderResponse := map[string]any{}
	err := jsonutil.UnmarshalValid(bidderResponseBody, &bidderResponse)
	if err != nil {
		return nil, err
	}
	// setting bidder response in paramResolver one time
	paramResolver.BidderResponse = bidderResponse
	// build adapter response
	adapterResponse := map[string]any{}
	if curr, ok := bidderResponse["cur"].(string); ok {
		adapterResponse["Currency"] = curr
	}

	typeBids := []any{}
	// resolve response level params with bidder response and adapter response
	for _, paramName := range resolver.ResponseLevelParams {
		paramMapper, ok := responseParams[paramName]
		if !ok {
			continue
		}
		paramResolver.Resolve(bidderResponse, adapterResponse, paramMapper.GetPath(), paramName)
	}
	seatBids, ok := bidderResponse["seatbid"].([]any)
	if !ok {
		return nil, fmt.Errorf("error:[invalid_seatbid_found_in_responsebody], seatbid:[%v]", bidderResponse["seatbid"])
	}
	for seatIndex, seatBid := range seatBids {
		seatBid, ok := seatBid.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("error:[invalid_seatbid_found_in_seatbids_list], seatbid:[%v]", seatBids)
		}

		bids, ok := seatBid["bid"].([]any)
		if !ok {
			return nil, fmt.Errorf("error:[invalid_bid_found_in_seatbid], bid:[%v]", seatBid["bid"])
		}

		for bidIndex, bid := range bids {
			bid, ok := bid.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("error:[invalid_bid_found_in_bids_list], bid:[%v]", seatBid["bid"])
			}
			typeBid := map[string]any{
				"Bid": bid,
			}

			for _, paramName := range resolver.BidLevelParams {
				paramMapper, ok := responseParams[paramName]
				if !ok {
					continue
				}

				path := util.GetPath(paramMapper.GetPath(), []int{seatIndex, bidIndex})
				paramResolver.Resolve(bid, typeBid, path, paramName)
			}
			typeBids = append(typeBids, typeBid)
		}
	}

	adapterResponse["Bids"] = typeBids

	return jsonutil.Marshal(adapterResponse)
}
