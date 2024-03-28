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

const (
	RequestModeSingle = "single"
)

// adapter implements adapters.Bidder interface
type adapter struct {
	adapterInfoMap map[string]adapterInfo   // store the list of bidder and its respective info required in MakeRequest/MakeBids
	oRTBBidderList []openrtb_ext.BidderName // list of all instances of oRTB bidders
}

// adapterInstance is singleton instance of oRTB adapter
var adapterInstance *adapter

// adapterInfo contains oRTB bidder specific info required in MakeRequests/MakeBids functions
type adapterInfo struct {
	config.Adapter
	extraInfo extraAdapterInfo
}

// extraAdapterInfo holds the values for bidder-info.extra_info field
type extraAdapterInfo struct {
	RequestMode string `json:"requestMode"`
}

// prepareRequestData generates the RequestData by marshalling the request and returns it
func (o adapterInfo) prepareRequestData(request *openrtb2.BidRequest) (*adapters.RequestData, error) {
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
func Builder(bidderName openrtb_ext.BidderName, cfg config.Adapter, server config.Server) (adapters.Bidder, error) {
	if adapterInstance == nil {
		adapterInstance = &adapter{
			adapterInfoMap: make(map[string]adapterInfo),
		}
	}
	extraAdapterInfo := extraAdapterInfo{}
	if len(cfg.ExtraAdapterInfo) > 0 {
		err := json.Unmarshal([]byte(cfg.ExtraAdapterInfo), &extraAdapterInfo)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse extra_info for bidder:[%s] err:[%s]", bidderName, err.Error())
		}
	}
	adapterInstance.adapterInfoMap[string(bidderName)] = adapterInfo{cfg, extraAdapterInfo}
	return adapterInstance, nil
}

// MakeRequests prepares oRTB bidder-specific request information using which prebid server make call(s) to bidder.
func (a *adapter) MakeRequests(request *openrtb2.BidRequest, requestInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	if request == nil || requestInfo == nil {
		return nil, []error{fmt.Errorf("Found either nil request or nil requestInfo")}
	}
	adapterInfo, found := a.adapterInfoMap[requestInfo.BidderCoreName]
	if !found {
		return nil, []error{fmt.Errorf("adapter info not found")}
	}
	// bidder request supports single impression in single HTTP call.
	if adapterInfo.extraInfo.RequestMode == RequestModeSingle {
		requestData := make([]*adapters.RequestData, 0, len(request.Imp))
		requestCopy := *request
		for _, imp := range request.Imp {
			requestCopy.Imp = []openrtb2.Imp{imp} // requestCopy contains single impression
			reqData, err := adapterInfo.prepareRequestData(&requestCopy)
			if err != nil {
				return nil, []error{err} //TODO: check if we can send single imp
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
	var errs []error
	for _, seatBid := range response.SeatBid {
		for bidInd, bid := range seatBid.Bid {
			bidResponse.Bids = append(bidResponse.Bids, &adapters.TypedBid{
				Bid:     &seatBid.Bid[bidInd],
				BidType: getMediaTypeForBid(bid),
			})
		}
	}
	return &bidResponse, errs
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

// ReadORTBBidderList reads the "static/bidder-info" directory and stores the oRTB bidder names whose name starts with 'ortb_'
// func ReadORTBBidderList(dirPath string) (bidderList []openrtb_ext.BidderName, err error) {
// 	files, err := os.ReadDir(dirPath)
// 	if err != nil {
// 		return bidderList, err
// 	}
// 	for _, file := range files {
// 		bidderName := strings.TrimSuffix(file.Name(), ".yaml")
// 		if strings.HasPrefix(bidderName, oRTBBidderPrefix) {
// 			bidderList = append(bidderList, openrtb_ext.BidderName(bidderName))
// 		}
// 	}
// 	return bidderList, nil
// }

// AppendBuilders appends the builders received in argument by oRTB bidder specific builders
// func AppendBuilders(builders map[openrtb_ext.BidderName]adapters.Builder) map[openrtb_ext.BidderName]adapters.Builder {
// 	if adapterInstance == nil {
// 		return builders
// 	}
// 	for _, bidder := range adapterInstance.oRTBBidderList {
// 		builders[openrtb_ext.BidderName(bidder)] = Builder
// 	}
// 	return builders
// }

// var once sync.Once
// var err error

// // InitORTBAdapter makes sure that adapterInstance is initialised only once
// func InitORTBAdapter(bidderList []openrtb_ext.BidderName) *adapter {
// 	once.Do(func() {
// 		adapterInstance = &adapter{
// 			adapterInfoMap: make(map[string]adapterInfo),
// 			oRTBBidderList: bidderList,
// 		}
// 	})
// 	return adapterInstance
// }
