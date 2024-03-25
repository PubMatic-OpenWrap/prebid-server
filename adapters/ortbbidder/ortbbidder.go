package ortbbidder

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// oRTBAdapter implements adapters.Bidder interface
type oRTBAdapter struct {
	oRTBAdapterInfo
}

const (
	RequestModeSingle string = "single"
)

// oRTBAdapterInfo contains oRTB bidder specific info required in MakeRequests/MakeBids functions
type oRTBAdapterInfo struct {
	config.Adapter
	requestMode string
}
type ExtraAdapterInfo struct {
	RequestMode string `json:"requestMode"`
}

// prepareRequestData generates the RequestData by marshalling the request and returns it
func (o oRTBAdapterInfo) prepareRequestData(request *openrtb2.BidRequest) (*adapters.RequestData, error) {
	if request == nil {
		return nil, fmt.Errorf("found nil request")
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request %s", err.Error())
	}
	return &adapters.RequestData{
		Method: http.MethodPost,
		Uri:    o.Endpoint,
		Body:   body,
	}, nil
}

// Builder returns an instance of oRTB adapter
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter, server config.Server) (adapters.Bidder, error) {
	extraAdapterInfo := ExtraAdapterInfo{}
	if len(config.ExtraAdapterInfo) > 0 {
		err := json.Unmarshal([]byte(config.ExtraAdapterInfo), &extraAdapterInfo)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse extra_info for bidder:[%s] err:[%s]", bidderName, err.Error())
		}
	}
	return &oRTBAdapter{
		oRTBAdapterInfo: oRTBAdapterInfo{config, extraAdapterInfo.RequestMode},
	}, nil
}

// MakeRequests prepares oRTB bidder-specific request information using which prebid server make call(s) to bidder.
func (o *oRTBAdapter) MakeRequests(request *openrtb2.BidRequest, requestInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	if request == nil || requestInfo == nil {
		return nil, []error{fmt.Errorf("Found either nil request or nil requestInfo")}
	}
	adapterInfo := o.oRTBAdapterInfo
	// bidder request supports single impression in single HTTP call.
	if adapterInfo.requestMode == RequestModeSingle {
		requestData := make([]*adapters.RequestData, 0, len(request.Imp))
		requestCopy := *request
		for _, imp := range request.Imp {
			requestCopy.Imp = []openrtb2.Imp{imp} // requestCopy contains single impression
			reqData, err := adapterInfo.prepareRequestData(&requestCopy)
			if err != nil {
				return nil, []error{err}
			}
			requestData = append(requestData, reqData)
		}
		return requestData, nil
	}
	// bidder request supports multi impressions in single HTTP call.
	requestData, err := adapterInfo.prepareRequestData(request)
	if err != nil {
		return nil, []error{err}
	}
	return []*adapters.RequestData{requestData}, nil
}

// MakeBids prepares bidderResponse from the oRTB bidder server's http.Response
func (o *oRTBAdapter) MakeBids(request *openrtb2.BidRequest, requestData *adapters.RequestData, responseData *adapters.ResponseData) (*adapters.BidderResponse, []error) {
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
	var errs []error
	for _, seatBid := range response.SeatBid {
		for bidInd, bid := range seatBid.Bid {
			bidType, err := getMediaTypeForBid(bid)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			bidResponse.Bids = append(bidResponse.Bids, &adapters.TypedBid{
				Bid:     &seatBid.Bid[bidInd],
				BidType: bidType,
			})
		}
	}
	return &bidResponse, errs
}

// getMediaTypeForBid returns the BidType as per the bid.MType field
// bid.MType has high priority over bidExt.Prebid.Type
func getMediaTypeForBid(bid openrtb2.Bid) (openrtb_ext.BidType, error) {
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
	// TODO : detect mediatype from bid.AdM and request.imp parameter
	if bidType == "" {
		return bidType, fmt.Errorf("Failed to parse bid mType for bidID \"%s\"", bid.ID)
	}
	return bidType, nil
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
