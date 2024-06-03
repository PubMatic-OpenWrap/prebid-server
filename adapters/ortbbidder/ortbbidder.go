package ortbbidder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/macros"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/jsonutil"
)

// adapter implements adapters.Bidder interface
type adapter struct {
	adapterInfo
	bidderParamsConfig *bidderparams.BidderConfig
}

// adapterInfo contains oRTB bidder specific info required in MakeRequests/MakeBids functions
type adapterInfo struct {
	config.Adapter
	extraInfo        extraAdapterInfo
	bidderName       openrtb_ext.BidderName
	endpointTemplate *template.Template
}
type extraAdapterInfo struct {
	RequestMode string `json:"requestMode"`
}

// global instance to hold bidderParamsConfig
var g_bidderParamsConfig *bidderparams.BidderConfig

// InitBidderParamsConfig initializes a g_bidderParamsConfig instance from the files provided in dirPath.
func InitBidderParamsConfig(dirPath string) (err error) {
	g_bidderParamsConfig, err = bidderparams.LoadBidderConfig(dirPath, isORTBBidder)
	return err
}

// makeRequestForAllImps processes a request map to create single RequestData object for all impressions.
// It constructs the endpoint URL and maps the request-params in request to form the RequestData object.
func (adapterInfo adapterInfo) makeRequestForAllImps(request map[string]any, bidderParamMapper map[string]bidderparams.BidderParamMapper) ([]*adapters.RequestData, []error) {
	imps, ok := request[impKey].([]any)
	if !ok {
		return nil, []error{newBadInputError("invalid imp object found in request")}
	}
	var (
		err  error
		uri  string
		errs []error
	)
	// iterate through imps in reverse order to ensure setRequestParams prioritizes
	// the parameters from imp[0].ext.bidder over those from imp[1..N].ext.bidder.
	for impIndex := len(imps) - 1; impIndex >= 0; impIndex-- {
		imp, ok := imps[impIndex].(map[string]any)
		if !ok || imp == nil {
			errs = append(errs, newBadInputError(fmt.Sprintf("invalid imp object found at index:%d", impIndex)))
			continue
		}
		bidderParams := getImpExtBidderParams(imp)
		// build endpoint URL once, using the imp[0].ext.bidder parameters
		// this must be done before calling setRequestParams, as it removes the imp.ext.bidder parameters.
		if impIndex == 0 {
			uri, err = macros.ResolveMacros(adapterInfo.endpointTemplate, bidderParams)
			if err != nil {
				return nil, []error{newBadInputError(fmt.Sprintf("failed to form endpoint url, err:%s", err.Error()))}
			}
		}
		// update the request and imp object by mapping bidderParams at expected location.
		setRequestParams(request, imp, bidderParams, bidderParamMapper)
	}
	requestBody, err := jsonutil.Marshal(request)
	if err != nil {
		return nil, []error{newBadInputError(fmt.Sprintf("failed to marshal request after setting bidder-params, err:%s", err.Error()))}
	}
	return []*adapters.RequestData{
		{
			Method: http.MethodPost,
			Uri:    uri,
			Body:   requestBody,
			Headers: http.Header{
				"Content-Type": {"application/json;charset=utf-8"},
				"Accept":       {"application/json"},
			},
		},
	}, errs
}

// makeRequestPerImp processes a request map to generate 'N' RequestData objects, one for each of the 'N' impressions.
// It constructs the endpoint URL and maps the request parameters to create the RequestData objects.
func (adapterInfo adapterInfo) makeRequestPerImp(request map[string]any, requestParams map[string]bidderparams.BidderParamMapper) ([]*adapters.RequestData, []error) {
	imps, ok := request[impKey].([]any)
	if !ok {
		return nil, []error{newBadInputError("invalid imp object found in request")}
	}
	var (
		bidderParams map[string]any
		requestData  []*adapters.RequestData
		errs         []error
	)
	for impIndex, imp := range imps {
		imp, ok := imp.(map[string]any)
		if !ok || imp == nil {
			errs = append(errs, newBadInputError(fmt.Sprintf("invalid imp object found at index:%d", impIndex)))
			continue
		}
		bidderParams = getImpExtBidderParams(imp)
		// build endpoint url from imp.ext.bidder
		// this must be done before calling setRequestParams, as it removes the imp.ext.bidder parameters.
		uri, err := macros.ResolveMacros(adapterInfo.endpointTemplate, bidderParams)
		if err != nil {
			errs = append(errs, newBadInputError(fmt.Sprintf("failed to form endpoint url for imp at index:%d, err:%s", impIndex, err.Error())))
			continue
		}
		// update the request and imp object by mapping bidderParams at expected location.
		setRequestParams(request, imp, bidderParams, requestParams)
		// request should contain single impression so override the request["imp"] field
		request[impKey] = []any{imp}

		requestBody, err := jsonutil.Marshal(request)
		if err != nil {
			errs = append(errs, newBadInputError(fmt.Sprintf("failed to marshal request after seeting bidder-params for imp at index:%d, err:%s", impIndex, err.Error())))
			continue
		}
		requestData = append(requestData, &adapters.RequestData{
			Method: http.MethodPost,
			Uri:    uri,
			Body:   requestBody,
			Headers: http.Header{
				"Content-Type": {"application/json;charset=utf-8"},
				"Accept":       {"application/json"},
			},
		})
	}
	return requestData, errs
}

// Builder returns an instance of oRTB adapter
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter, server config.Server) (adapters.Bidder, error) {
	extraAdapterInfo := extraAdapterInfo{}
	if len(config.ExtraAdapterInfo) > 0 {
		err := jsonutil.Unmarshal([]byte(config.ExtraAdapterInfo), &extraAdapterInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to parse extra_info: %s", err.Error())
		}
	}
	template, err := template.New("endpointTemplate").Parse(config.Endpoint)
	if err != nil || template == nil {
		return nil, fmt.Errorf("failed to parse endpoint url template: %v", err)
	}
	return &adapter{
		adapterInfo:        adapterInfo{config, extraAdapterInfo, bidderName, template},
		bidderParamsConfig: g_bidderParamsConfig,
	}, nil
}

// MakeRequests prepares oRTB bidder-specific request information using which prebid server make call(s) to bidder.
func (o *adapter) MakeRequests(bidRequest *openrtb2.BidRequest, requestInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	if bidRequest == nil {
		return nil, []error{newBadInputError("found nil request")}
	}
	if o.bidderParamsConfig == nil {
		return nil, []error{newBadInputError("found nil bidderParamsConfig")}
	}
	rawRequest, err := jsonutil.Marshal(bidRequest)
	if err != nil {
		return nil, []error{newBadInputError(fmt.Sprintf("failed to marshal request, err:%s", err.Error()))}
	}
	var request map[string]any
	err = jsonutil.Unmarshal(rawRequest, &request)
	if err != nil {
		return nil, []error{newBadInputError(fmt.Sprintf("failed to unmarshal request, err:%s", err.Error()))}
	}
	var (
		requestData []*adapters.RequestData
		errs        []error
	)
	requestParams, _ := o.bidderParamsConfig.GetRequestParams(o.bidderName.String())
	switch o.adapterInfo.extraInfo.RequestMode {
	case requestModeSingle:
		requestData, errs = o.adapterInfo.makeRequestPerImp(request, requestParams)
	default:
		requestData, errs = o.adapterInfo.makeRequestForAllImps(request, requestParams)
	}
	return requestData, errs
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
	if err := jsonutil.Unmarshal(responseData.Body, &response); err != nil {
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
