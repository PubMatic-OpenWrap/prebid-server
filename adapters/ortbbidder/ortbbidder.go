package ortbbidder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// oRTBAdapter implements adapters.Bidder interface
// Note: Single instance of oRTBAdapter can run concurrently for multiple oRTB bidders;
// hence, do not store any bidder-specific data in this structure.
type oRTBAdapter struct {
	BidderInfo config.BidderInfos
}

var ortbAdapter *oRTBAdapter
var once sync.Once

// InitORTBAdapter initialises the instance of oRTBAdapter
func InitORTBAdapter(infos config.BidderInfos) {
	once.Do(func() {
		ortbAdapter = &oRTBAdapter{
			BidderInfo: infos,
		}
	})
}

// Builder returns an instance of oRTB adapter initialised by InitORTBAdapter,
// it returns an error if InitORTBAdapter is not called yet.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter, server config.Server) (adapters.Bidder, error) {
	if ortbAdapter == nil {
		return nil, fmt.Errorf("oRTB bidder is not initialised")
	}
	return ortbAdapter, nil
}

// MakeRequests prepares oRTB bidder-specific request information using which prebid server make call(s) to bidder.
func (o *oRTBAdapter) MakeRequests(request *openrtb2.BidRequest, requestInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	if request == nil || requestInfo == nil {
		return nil, []error{fmt.Errorf("Found either nil request or nil requestInfo")}
	}
	// fetch bidder specific info from bidder-info.yaml
	bidderInfo, found := o.BidderInfo[string(requestInfo.BidderCoreName)]
	if !found {
		return nil, []error{fmt.Errorf("bidder-info not found for bidder-[%s]", requestInfo.BidderCoreName)}
	}

	// bidder request supports single impression in single HTTP call.
	if bidderInfo.OpenWrap.RequestMode == config.RequestModeSingle {
		requestData := make([]*adapters.RequestData, 0, len(request.Imp))
		requestCopy := *request
		for _, imp := range request.Imp {
			requestCopy.Imp = []openrtb2.Imp{imp} // requestCopy contains single impression
			reqData, err := prepareRequestData(&requestCopy, bidderInfo.Endpoint)
			if err != nil {
				return nil, []error{err}
			}
			requestData = append(requestData, reqData)
		}
		return requestData, nil
	}

	// bidder request supports multi impressions in single HTTP call.
	requestData, err := prepareRequestData(request, bidderInfo.Endpoint)
	if err != nil {
		return nil, []error{err}
	}
	return []*adapters.RequestData{requestData}, nil
}

// prepareRequestData generates the RequestData by marshalling the request and returns it
func prepareRequestData(request *openrtb2.BidRequest, endpoint string) (*adapters.RequestData, error) {
	if request == nil {
		return nil, fmt.Errorf("found nil request")
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	return &adapters.RequestData{
		Method: http.MethodPost,
		Uri:    endpoint,
		Body:   body,
	}, nil
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

	// initialise bidResponse with zero bids
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
// bidExt.Prebid.Type has high priority over bid.MType
func getMediaTypeForBid(bid openrtb2.Bid) (openrtb_ext.BidType, error) {
	var bidType openrtb_ext.BidType
	if bid.Ext != nil {
		var bidExt openrtb_ext.ExtBid
		err := json.Unmarshal(bid.Ext, &bidExt)
		if err == nil && bidExt.Prebid != nil {
			return openrtb_ext.ParseBidType(string(bidExt.Prebid.Type))
		}
	}
	switch bid.MType {
	case openrtb2.MarkupBanner:
		bidType = openrtb_ext.BidTypeBanner
	case openrtb2.MarkupVideo:
		bidType = openrtb_ext.BidTypeVideo
	case openrtb2.MarkupAudio:
		bidType = openrtb_ext.BidTypeAudio
	case openrtb2.MarkupNative:
		bidType = openrtb_ext.BidTypeNative
	default:
		return bidType, fmt.Errorf("Failed to parse bid mType for bidID \"%s\"", bid.ID)
	}
	return bidType, nil
}
