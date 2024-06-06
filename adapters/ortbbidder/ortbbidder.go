package ortbbidder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/buger/jsonparser"
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

// makeRequest constructs the endpoint URL and maps the bidder-parameters in request to create the RequestData objects.
// when supportSingleImpInRequest is true, it processes a request to generate 'N' RequestData objects, one for each of the 'N' impressions
// else it create single RequestData object for all impressions.
func (adapterInfo adapterInfo) makeRequest(rawRequest []byte, bidderParamMapper map[string]bidderparams.BidderParamMapper, supportSingleImpInRequest bool) ([]*adapters.RequestData, []error) {
	request, err := convertRequestToMap(rawRequest)
	if err != nil {
		return nil, []error{newBadInputError(err.Error())}
	}
	imps, ok := request[impKey].([]any)
	if !ok {
		return nil, []error{newBadInputError("imp object not found in request")}
	}
	// set "imp" object in request to empty to improve performance while creating deep copy of request
	if supportSingleImpInRequest {
		rawRequest, err = jsonparser.Set(rawRequest, []byte("[]"), impKey)
		if err != nil {
			return nil, []error{fmt.Errorf("failed to empty the imp key in request")}
		}
	}
	var (
		uri         string
		errs        []error
		requestData []*adapters.RequestData
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
		if supportSingleImpInRequest {
			// build endpoint-url from bidder-params, it must be done before calling setRequestParams, as it removes the imp.ext.bidder parameters.
			// for "single" requestMode, build endpoint-uri separately using each imp's bidder-params
			uri, err = buildEndpoint(adapterInfo.endpointTemplate, bidderParams)
			if err != nil {
				errs = append(errs, newBadInputError(err.Error()))
				continue
			}
			// override "imp" key in request to ensure request contains single imp
			request[impKey] = []any{imp}
			// update the request and imp object by mapping bidderParams at expected location.
			setRequestParams(request, bidderParams, bidderParamMapper, []int{0})
			requestData, err = appendRequestData(requestData, request, uri)
			if err != nil {
				errs = append(errs, newBadInputError(err.Error()))
			}
			// create a deep copy of the request to ensure common fields are not altered.
			// example - if imp2 modifies the original req.bcat field using its bidder-params, imp1 should still be able to use the original req.bcat value.
			if impIndex != 0 {
				request, err = convertRequestToMap(rawRequest)
				if err != nil {
					errs = append(errs, newBadInputError(fmt.Sprintf("failed to build request from rawRequest, err:%s", err.Error())))
					return requestData, errs
				}
			}
			continue
		}
		// processing for "multi" requestMode
		if impIndex == 0 {
			// build endpoint-url only once using first imp's bidder-params
			uri, err = buildEndpoint(adapterInfo.endpointTemplate, bidderParams)
			if err != nil {
				errs = append(errs, newBadInputError(err.Error()))
				return nil, errs
			}
		}
		// update the request and imp object by mapping bidderParams at expected location.
		setRequestParams(request, bidderParams, bidderParamMapper, []int{impIndex})
	}
	// for "multi" requestMode, combine all the prepared requests
	if !supportSingleImpInRequest {
		requestData, err = appendRequestData(requestData, request, uri)
		if err != nil {
			errs = append(errs, newBadInputError(err.Error()))
		}
	}

	return requestData, errs
}

// appendRequestData creates new RequestData using request and uri then appends it to requestData passed as argument
func appendRequestData(requestData []*adapters.RequestData, request map[string]any, uri string) ([]*adapters.RequestData, error) {
	rawRequest, err := jsonutil.Marshal(request)
	if err != nil {
		return requestData, fmt.Errorf("failed to marshal request after setting bidder-params, err:%s", err.Error())
	}
	requestData = append(requestData, &adapters.RequestData{
		Method: http.MethodPost,
		Uri:    uri,
		Body:   rawRequest,
		Headers: http.Header{
			"Content-Type": {"application/json;charset=utf-8"},
			"Accept":       {"application/json"},
		},
	})
	return requestData, nil
}

// buildEndpoint replaces macros present in the endpoint-url and returns the updated uri
func buildEndpoint(endpointTemplate *template.Template, bidderParams map[string]any) (uri string, err error) {
	uri, err = macros.ResolveMacros(endpointTemplate, bidderParams)
	if err != nil {
		return uri, fmt.Errorf("failed to replace macros in endpoint, err:%s", err.Error())
	}
	uri = strings.ReplaceAll(uri, "<no value>", "")
	return uri, err
}

// convertRequestToMap converts request from []byte to map[string]any, it creates deep copy of request using json-unmarshal
func convertRequestToMap(rawRequest []byte) (request map[string]any, err error) {
	err = jsonutil.Unmarshal(rawRequest, &request)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal request, err:%s", err.Error())
	}
	return request, err
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
	template, err := template.New("endpointTemplate").Option("missingkey=zero").Parse(config.Endpoint)
	if err != nil || template == nil {
		return nil, fmt.Errorf("failed to parse endpoint url template: %v", err)
	}
	return &adapter{
		adapterInfo:        adapterInfo{config, extraAdapterInfo, bidderName, template},
		bidderParamsConfig: g_bidderParamsConfig,
	}, nil
}

// MakeRequests prepares oRTB bidder-specific request information using which prebid server make call(s) to bidder.
func (o *adapter) MakeRequests(request *openrtb2.BidRequest, requestInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	if request == nil {
		return nil, []error{newBadInputError("found nil request")}
	}
	if o.bidderParamsConfig == nil {
		return nil, []error{newBadInputError("found nil bidderParamsConfig")}
	}
	rawRequest, err := jsonutil.Marshal(request)
	if err != nil {
		return nil, []error{newBadInputError(fmt.Sprintf("failed to marshal request, err:%s", err.Error()))}
	}
	requestParams := o.bidderParamsConfig.GetRequestParams(o.bidderName.String())
	return o.adapterInfo.makeRequest(
		rawRequest, requestParams, o.adapterInfo.extraInfo.RequestMode == requestModeSingle)
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
