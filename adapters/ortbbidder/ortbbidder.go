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
	responseMapper
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

var bidderParam openrtb_ext.BidderParamValidator

func SaveBidderParam(v openrtb_ext.BidderParamValidator) {
	bidderParam = v
}

func getAllFields(ext []byte, path string) (map[string][]byte, error) {
	fields := make(map[string][]byte)

	// Parse the ext JSON object
	err := jsonparser.ObjectEach(ext, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		fields[string(key)] = value
		return nil
	}, path)
	return fields, err
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

	fieldLocation := bidderParam.PropertyLocation(o.bidderName)
	fields, err := getAllFields(request.Imp[0].Ext, "bidder")
	if err != nil {
		fmt.Print("err-", err)
		return nil, err
	}
	for field, fieldVal := range fields {
		property := fieldLocation[field]
		if property.Location == "" {
			continue
		}
		keys := strings.Split(property.Location, ".")
		if property.Type == "string" {
			fieldVal = []byte(fmt.Sprintf(`"%s"`, fieldVal))
		}
		body, err = jsonparser.Set(body, fieldVal, keys...)
		if err != nil {
			fmt.Println("err-", err)
		}
	}

	return &adapters.RequestData{
		Method: http.MethodPost,
		Uri:    o.Endpoint,
		Body:   body,
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
		adapterInfo:    adapterInfo{config, extraAdapterInfo, bidderName},
		responseMapper: NewResponseMapper(),
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
			reqData, err := adapterInfo.prepareRequestData(&requestCopy)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			requestData = append(requestData, reqData)
		}
		return requestData, errs
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
