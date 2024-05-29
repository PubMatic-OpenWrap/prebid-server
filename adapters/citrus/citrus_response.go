package citrus

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
)


type CitrusResponse struct {
	Ads         []map[string]interface{} `json:"ads"`
	MemoryToken string `json:"memoryToken"`
}

func (a *CitrusAdapter) MakeBids(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	var errors []error

	commerceExt, err := adapters.GetImpressionExtComm(&(internalRequest.Imp[0]))
	if err != nil {
		errors := append(errors, err)
		return nil, errors
	}

	if commerceExt.ComParams.TestRequest {
		dummyResponse := a.GetMockResponse(internalRequest)
		return dummyResponse, nil
	}
	
	if response.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if response.StatusCode != http.StatusOK {
		return nil, []error{&errortypes.BadServerResponse{
			Message: fmt.Sprintf("Unexpected status code: %d", response.StatusCode),
		}}
	}

	citrusResponse, err := newCitrusResponseFromBytes(response.Body)
	if err != nil {
		return nil, []error{err}
	}

	if citrusResponse.Ads == nil || len(citrusResponse.Ads) <= 0 {
		return nil, []error{&errortypes.NoBidPrice{
			Message: "No Bid For the given Request",
		}}
	}
	
	var configValueMap = make(map[string]string)
        var configTypeMap = make(map[string]int)
	for _,obj := range commerceExt.Bidder.CustomConfig {
		configValueMap[obj.Key] = obj.Value
		configTypeMap[obj.Key] = obj.Type
	}

	impID := internalRequest.Imp[0].ID
	bidderResponse := a.getBidderResponse(internalRequest, &citrusResponse, impID, configValueMap)
	return bidderResponse, nil
}


func (a *CitrusAdapter) getBidderResponse(request *openrtb2.BidRequest, citrusResponse *CitrusResponse, requestImpID string, configValueMap map[string]string) *adapters.BidderResponse {

	noOfBids := countSponsoredProducts(citrusResponse)
	bidResponse := adapters.NewBidderResponseWithBidsCapacity(noOfBids)
	index := 1


	pubMaticTracking := false
	val, ok := configValueMap[adapters.AUCTIONDETAILS_PREFIX + adapters.PUBMATIC_TRACKING]
	if ok {
		if val == adapters.STRING_TRUE {
			pubMaticTracking = true
		}
	}
	bidderExtendedDetails := false
	val, ok = configValueMap[adapters.AUCTIONDETAILS_PREFIX + adapters.AD_BIDDER_EXTEN_DETAILS]
	if ok {
		if val == adapters.STRING_TRUE {
			bidderExtendedDetails = true
		}
	}

	for _, ad := range citrusResponse.Ads {
		bidID := adapters.GenerateUniqueBidIDComm()
		impID := requestImpID + "_" + strconv.Itoa(index)
		productID := ad[PRODUCT_ID].(string)
		bidTrackingID := ad[TRACKING_ID].(string)

		if productID == "" {
			continue
		}
			
		impressionURL := getTrackingURL(TRACKING_IMPURL, bidTrackingID, pubMaticTracking)
		clickURL := getTrackingURL(TRACKING_CLKURL, bidTrackingID, pubMaticTracking)

		index++

		// Add ProductDetails to bidExtension
		productDetails := make(map[string]interface{})
		if bidderExtendedDetails {
			for key, value := range ad {
				productDetails[key] = value
			}

			delete(productDetails, PRODUCT_ID)
			delete(productDetails, TRACKING_ID)
		}

		bidExt := &openrtb_ext.ExtBidCMSponsored{
			ProductId:      productID,
			ClickUrl:       clickURL,
			ProductDetails: productDetails,
		}

		bid := &openrtb2.Bid{
			ID:    bidID,
			ImpID: impID,
			IURL:  impressionURL,
		}

		adapters.AddDefaultFieldsComm(bid)
		bidExtJSON, err1 := json.Marshal(bidExt)
		if nil == err1 {
			bid.Ext = json.RawMessage(bidExtJSON)
		}

		seat := openrtb_ext.BidderName(SEAT_CITRUS)
		typedbid := &adapters.TypedBid{
			Bid:  bid,
			Seat: seat,
		}
		bidResponse.Bids = append(bidResponse.Bids, typedbid)
	}
			
	return bidResponse
}

func getTrackingURL(baseURL, bidTrackingID string, pubMaticTracking bool) string {
	if !pubMaticTracking {
		return adapters.IMP_KEY + adapters.EncodeURL(baseURL + bidTrackingID)
	}
	return baseURL + bidTrackingID
}

func newCitrusResponseFromBytes(bytes []byte) (CitrusResponse, error) {
	var err error
	var bidResponse CitrusResponse

	if err = json.Unmarshal(bytes, &bidResponse); err != nil {
		return bidResponse, err
	}

	return bidResponse, nil
}

func countSponsoredProducts(adResponse *CitrusResponse) int {
	count := 0

	for _, ad := range adResponse.Ads {
		// Check if "gtin" key exists in the ad map and is not an empty string
		if gtin, ok := ad[PRODUCT_ID].(string); ok && gtin != "" {
			count++
		}
	}
	return count
}


