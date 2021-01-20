package smartyads

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/errortypes"
	"github.com/PubMatic-OpenWrap/prebid-server/macros"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

type SmartyAdsAdapter struct {
	endpoint template.Template
}

func NewSmartyAdsBidder(endpointTemplate string) *SmartyAdsAdapter {
	template, err := template.New("endpointTemplate").Parse(endpointTemplate)
	if err != nil {
		return nil
	}
	return &SmartyAdsAdapter{endpoint: *template}
}

func GetHeaders(request *openrtb.BidRequest) *http.Header {
	headers := http.Header{}
	headers.Add("Content-Type", "application/json;charset=utf-8")
	headers.Add("Accept", "application/json")
	headers.Add("X-Openrtb-Version", "2.5")

	if request.Device != nil {
		if len(request.Device.UA) > 0 {
			headers.Add("User-Agent", request.Device.UA)
		}

		if len(request.Device.IPv6) > 0 {
			headers.Add("X-Forwarded-For", request.Device.IPv6)
		}

		if len(request.Device.IP) > 0 {
			headers.Add("X-Forwarded-For", request.Device.IP)
		}

		if len(request.Device.Language) > 0 {
			headers.Add("Accept-Language", request.Device.Language)
		}

		if request.Device.DNT != nil {
			headers.Add("Dnt", strconv.Itoa(int(*request.Device.DNT)))
		}
	}

	return &headers
}

func (a *SmartyAdsAdapter) MakeRequests(
	openRTBRequest *openrtb.BidRequest,
	reqInfo *adapters.ExtraRequestInfo,
) (
	requestsToBidder []*adapters.RequestData,
	errs []error,
) {

	var errors []error
	var smartyadsExt *openrtb_ext.ExtSmartyAds
	var err error

	for i, imp := range openRTBRequest.Imp {
		smartyadsExt, err = a.getImpressionExt(&imp)
		if err != nil {
			errors = append(errors, err)
			break
		}
		openRTBRequest.Imp[i].Ext = nil
	}

	if len(errors) > 0 {
		return nil, errors
	}

	url, err := a.buildEndpointURL(smartyadsExt)
	if err != nil {
		return nil, []error{err}
	}

	reqJSON, err := json.Marshal(openRTBRequest)
	if err != nil {
		return nil, []error{err}
	}

	return []*adapters.RequestData{{
		Method:  http.MethodPost,
		Body:    reqJSON,
		Uri:     url,
		Headers: *GetHeaders(openRTBRequest),
	}}, nil
}

func (a *SmartyAdsAdapter) getImpressionExt(imp *openrtb.Imp) (*openrtb_ext.ExtSmartyAds, error) {
	var bidderExt adapters.ExtImpBidder
	if err := json.Unmarshal(imp.Ext, &bidderExt); err != nil {
		return nil, &errortypes.BadInput{
			Message: "ext.bidder not provided",
		}
	}
	var smartyadsExt openrtb_ext.ExtSmartyAds
	if err := json.Unmarshal(bidderExt.Bidder, &smartyadsExt); err != nil {
		return nil, &errortypes.BadInput{
			Message: "ext.bidder not provided",
		}
	}
	return &smartyadsExt, nil
}

func (a *SmartyAdsAdapter) buildEndpointURL(params *openrtb_ext.ExtSmartyAds) (string, error) {
	endpointParams := macros.EndpointTemplateParams{Host: params.Host, SourceId: params.SourceID, AccountID: params.AccountID}
	return macros.ResolveMacros(a.endpoint, endpointParams)
}

func (a *SmartyAdsAdapter) CheckResponseStatusCodes(response *adapters.ResponseData) error {
	if response.StatusCode == http.StatusNoContent {
		return &errortypes.BadInput{Message: "No bid response"}
	}

	if response.StatusCode == http.StatusBadRequest {
		return &errortypes.BadInput{
			Message: fmt.Sprintf("Unexpected status code: [ %d ]", response.StatusCode),
		}
	}

	if response.StatusCode == http.StatusServiceUnavailable {
		return &errortypes.BadInput{
			Message: fmt.Sprintf("Something went wrong, please contact your Account Manager. Status Code: [ %d ] ", response.StatusCode),
		}
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return &errortypes.BadInput{
			Message: fmt.Sprintf("Something went wrong, please contact your Account Manager. Status Code: [ %d ] ", response.StatusCode),
		}
	}

	return nil
}

func (a *SmartyAdsAdapter) MakeBids(
	openRTBRequest *openrtb.BidRequest,
	requestToBidder *adapters.RequestData,
	bidderRawResponse *adapters.ResponseData,
) (
	bidderResponse *adapters.BidderResponse,
	errs []error,
) {
	httpStatusError := a.CheckResponseStatusCodes(bidderRawResponse)
	if httpStatusError != nil {
		return nil, []error{httpStatusError}
	}

	responseBody := bidderRawResponse.Body
	var bidResp openrtb.BidResponse
	if err := json.Unmarshal(responseBody, &bidResp); err != nil {
		return nil, []error{&errortypes.BadServerResponse{
			Message: "Bad Server Response",
		}}
	}

	if len(bidResp.SeatBid) == 0 {
		return nil, []error{&errortypes.BadServerResponse{
			Message: "Empty SeatBid array",
		}}
	}

	bidResponse := adapters.NewBidderResponseWithBidsCapacity(len(bidResp.SeatBid[0].Bid))
	sb := bidResp.SeatBid[0]

	for _, bid := range sb.Bid {
		bidResponse.Bids = append(bidResponse.Bids, &adapters.TypedBid{
			Bid:     &bid,
			BidType: getMediaTypeForImp(bid.ImpID, openRTBRequest.Imp),
		})
	}
	return bidResponse, nil
}

func getMediaTypeForImp(impId string, imps []openrtb.Imp) openrtb_ext.BidType {
	mediaType := openrtb_ext.BidTypeBanner
	for _, imp := range imps {
		if imp.ID == impId {
			if imp.Video != nil {
				mediaType = openrtb_ext.BidTypeVideo
			} else if imp.Native != nil {
				mediaType = openrtb_ext.BidTypeNative
			}
			return mediaType
		}
	}
	return mediaType
}
