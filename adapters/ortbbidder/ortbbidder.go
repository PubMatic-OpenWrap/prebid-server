package ortbbidder

import (
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
	"github.com/prebid/prebid-server/v2/config"
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
func InitBidderParamsConfig(requestParamsDirPath, responseParamsDirPath string) (err error) {
	g_bidderParamsConfig, err = bidderparams.LoadBidderConfig(requestParamsDirPath, responseParamsDirPath, util.IsORTBBidder)
	return err
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
	template, err := template.New(endpointTemplate).Option(templateOption).Parse(config.Endpoint)
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
	if o.bidderParamsConfig == nil {
		return nil, []error{newBadInputError(errNilBidderParamCfg.Error())}
	}

	requestBuilder := newRequestBuilder(
		o.adapterInfo.extraInfo.RequestMode,
		o.Endpoint,
		o.endpointTemplate,
		o.bidderParamsConfig.GetRequestParams(o.bidderName.String()))

	if err := requestBuilder.parseRequest(request); err != nil {
		return nil, []error{newBadInputError(err.Error())}
	}

	return requestBuilder.makeRequest()
}

// MakeBids prepares bidderResponse from the oRTB bidder server's http.Response
func (o *adapter) MakeBids(request *openrtb2.BidRequest, requestData *adapters.RequestData, responseData *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	if responseData == nil || adapters.IsResponseStatusCodeNoContent(responseData) {
		return nil, nil
	}

	if err := adapters.CheckResponseStatusCodeForErrors(responseData); err != nil {
		return nil, []error{err}
	}

	response, err := o.makeBids(responseData.Body)
	if err != nil {
		return nil, []error{err}
	}

	return response, nil
}

// MakeBids prepares bidderResponse from the oRTB bidder server's http.Response
func (o *adapter) makeBids(bidderResponseBytes json.RawMessage) (*adapters.BidderResponse, error) {
	responseParmas := o.bidderParamsConfig.GetResponseParams(o.bidderName.String())
	rb := newResponseBuilder(responseParmas)

	err := rb.parseResponse(bidderResponseBytes)
	if err != nil {
		return nil, err
	}

	err = rb.buildResponse()
	if err != nil {
		return nil, err
	}

	return rb.convertToAdapterResponse()
}
