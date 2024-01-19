package citrus

import (
	"encoding/json"
	"math/rand"
	"strconv"

	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/openrtb_ext"
)
var mockProductDetails = map[string]interface{}{
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
	IMP_URL  = TRACKING_IMPURL + "id=display_QqHaKRrKlFm1Wxr9c_DXJN4HSE3NzMzNjM2"
	CLICK_URL  = TRACKING_CLKURL + "id=display_QqHaKRrKlFm1Wxr9c_DXJN4HSE3NzMzNjM2"
)
func (a *CitrusAdapter) GetMockResponse(internalRequest *openrtb2.BidRequest) *adapters.BidderResponse {
	requestCount := GetRequestSlotCount(internalRequest)
	impiD := internalRequest.Imp[0].ID

	responseF := GetMockBids(requestCount, impiD)
	return responseF
}

func GetRequestSlotCount(internalRequest *openrtb2.BidRequest) int {
	impArray := internalRequest.Imp
	reqCount := 0
	for _, eachImp := range impArray {
		var commerceExt openrtb_ext.ExtImpCommerce
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

func GetMockBids(requestCount int, ImpID string) *adapters.BidderResponse {
	var typedArray []*adapters.TypedBid

	if requestCount > MAX_COUNT {
		requestCount = MAX_COUNT
	}
	
	for i := 1; i <= requestCount; i++ {
		productid := GetRandomProductID()
		bidID := adapters.GenerateUniqueBidIDComm()
		impID := ImpID + "_" + strconv.Itoa(i)

		bidExt := &openrtb_ext.ExtBidCommerce{
			ProductId:  productid,
			ClickUrl: CLICK_URL,
			ProductDetails: mockProductDetails,
		}

		bid := &openrtb2.Bid{
			ID:    bidID,
			ImpID: impID,
			IURL: IMP_URL,
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

