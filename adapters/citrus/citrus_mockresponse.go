package citrus

import (
	"encoding/json"
	"math/rand"
	"strconv"

	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/openrtb_ext"
)
var MockProductDetails = map[string]interface{}{
    "discount": map[string]interface{}{
        "amount":          23.45,
        "minPrice":        21.00,
        "maxPerCustomer":  2.45,
    },
    "expiry":   "2021-05-12T04:17:50.400902957Z",
    "position": 1,
}

const (
	MAX_COUNT   = 9
	TRACKINGID  =  "display_QqHaKRrKlFm1Wxr9c_DXJN4HSE3NzMzNjM2"
)
func (a *CitrusAdapter) GetMockResponse(internalRequest *openrtb2.BidRequest) *adapters.BidderResponse {
	requestCount := GetRequestSlotCount(internalRequest)
	impiD := internalRequest.Imp[0].ID

	commerceExt, err := adapters.GetImpressionExtComm(&(internalRequest.Imp[0]))
	if err != nil {
		return nil
	}

	var configValueMap = make(map[string]string)
        var configTypeMap = make(map[string]int)
	for _,obj := range commerceExt.Bidder.CustomConfig {
		configValueMap[obj.Key] = obj.Value
		configTypeMap[obj.Key] = obj.Type
	}

	responseF := GetMockBids(requestCount, impiD, configValueMap )
	return responseF
}

func GetRequestSlotCount(internalRequest *openrtb2.BidRequest) int {
	impArray := internalRequest.Imp
	reqCount := 0
	for _, eachImp := range impArray {
		var commerceExt openrtb_ext.ExtImpCMSponsored
		json.Unmarshal(eachImp.Ext, &commerceExt)
		reqCount += commerceExt.ComParams.SlotsRequested
	}
	return reqCount
}

func GetRandomProductID() string {
	min := 100000
	max := 600000
	randomN := rand.Intn(max-min+1) + min
	t := strconv.Itoa(randomN)
	return t
}

func GetMockBids(requestCount int, ImpID string, configValueMap map[string]string ) *adapters.BidderResponse {
	var typedArray []*adapters.TypedBid

	if requestCount > MAX_COUNT {
		requestCount = MAX_COUNT
	}
	
	bidderExtendedDetails := false
	val, ok := configValueMap[adapters.AUCTIONDETAILS_PREFIX + adapters.AD_BIDDER_EXTEN_DETAILS]
	if ok {
		if val == adapters.STRING_TRUE {
			bidderExtendedDetails = true
		}
	}

	for i := 1; i <= requestCount; i++ {
		productid := GetRandomProductID()
		bidID := adapters.GenerateUniqueBidIDComm()

		impressionURL := TRACKING_IMPURL + TRACKINGID + "_" + strconv.Itoa(i)
		clickURL := TRACKING_CLKURL + TRACKINGID + "_" + strconv.Itoa(i)

		mockProductDetails := make(map[string]interface{})
		if bidderExtendedDetails {
			mockProductDetails = MockProductDetails
		}
		bidExt := &openrtb_ext.ExtBidCMSponsored{
			ProductId:  productid,
			ClickUrl: clickURL,
			ProductDetails: mockProductDetails,
		}

		bid := &openrtb2.Bid{
			ID:    bidID,
			NURL: impressionURL,
		}

		adapters.AddDefaultFieldsComm(bid)

		bidExtJSON, err1 := json.Marshal(bidExt)
		if nil == err1 {
			bid.Ext = json.RawMessage(bidExtJSON)
		}

		typedbid := &adapters.TypedBid{
			Bid:  bid,
			Seat: openrtb_ext.BidderName(SEAT_CITRUS),
		}
		typedArray = append(typedArray, typedbid)
	}

	responseF := &adapters.BidderResponse{
		Bids: typedArray,
	}
	return responseF
}



