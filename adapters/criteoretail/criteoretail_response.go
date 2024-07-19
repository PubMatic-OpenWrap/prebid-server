package criteoretail

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

type Placement struct {
	Format       string                   `json:"format"`
	Products     []map[string]interface{} `json:"products"`
	OnLoadBeacon string                   `json:"OnLoadBeacon,omitempty"`
	OnViewBeacon string                   `json:"OnViewBeacon,omitempty"`
}

type CriteoResponse struct {
	Status               string                   `json:"status"`
	OnAvailabilityUpdate interface{}              `json:"OnAvailabilityUpdate"`
	Placements           []map[string][]Placement `json:"placements"`
}

func (a *CriteoRetailAdapter) MakeBids(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {

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

	criteoResponse, err := newcriteoretailResponseFromBytes(response.Body)
	if err != nil {
		return nil, []error{err}
	}

	if criteoResponse.Status != RESPONSE_OK {
		return nil, []error{&errortypes.BidderFailedSchemaValidation{
			Message: "Error Occured at Criteo for the given request ",
		}}
	}

	if criteoResponse.Placements == nil || len(criteoResponse.Placements) <= 0 {
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
	bidderResponse := a.getBidderResponse(internalRequest, &criteoResponse, impID, configValueMap)
	return bidderResponse, nil
}

func (a *CriteoRetailAdapter) getBidderResponse(request *openrtb2.BidRequest, criteoResponse *CriteoResponse, requestImpID string, configValueMap map[string]string) *adapters.BidderResponse {

	noOfBids := a.countSponsoredProducts(criteoResponse)
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
	for _, placementMap := range criteoResponse.Placements {
		for _, placements := range placementMap {
			for _, placement := range placements {
				if placement.Format == FORMAT_SPONSORED {
					for _, productMap := range placement.Products {
						bidID := adapters.GenerateUniqueBidIDComm()
						impID := requestImpID + "_" + strconv.Itoa(index)
						productID := productMap[PRODUCT_ID].(string)

						var impressionURL,clickURL string
						if pubMaticTracking {
							impressionURL = adapters.IMP_KEY + adapters.EncodeURL(productMap[VIEW_BEACON].(string))
							clickURL = adapters.CLICK_KEY + adapters.EncodeURL(productMap[CLICK_BEACON].(string))
						} else {
							impressionURL = productMap[VIEW_BEACON].(string)
							clickURL = productMap[CLICK_BEACON].(string)
						}
						index++

						// Add ProductDetails to bidExtension
						productDetails := make(map[string]interface{})
						if bidderExtendedDetails {
				
							for key, value := range productMap {
								productDetails[key] = value
							}

							delete(productDetails, PRODUCT_ID)
							delete(productDetails, VIEW_BEACON)
							delete(productDetails, CLICK_BEACON)
						}

						bidExt := &openrtb_ext.ExtBidCMSponsored{
							ProductId:      productID,
							ClickUrl:       clickURL,
							ProductDetails: productDetails,
						}

						bid := &openrtb2.Bid{
							ID:    bidID,
							ImpID: impID,
							NURL:  impressionURL,
						}

						adapters.AddDefaultFieldsComm(bid)
						bidExtJSON, err1 := json.Marshal(bidExt)
						if nil == err1 {
							bid.Ext = json.RawMessage(bidExtJSON)
						}

						seat := openrtb_ext.BidderName(SEAT_CRITEORETAIL)

						typedbid := &adapters.TypedBid{
							Bid:  bid,
							Seat: seat,
						}
						bidResponse.Bids = append(bidResponse.Bids, typedbid)
					}
				}
			}
		}
	}
	return bidResponse
}

func newcriteoretailResponseFromBytes(bytes []byte) (CriteoResponse, error) {
	var err error
	var bidResponse CriteoResponse

	if err = json.Unmarshal(bytes, &bidResponse); err != nil {
		return bidResponse, err
	}

	return bidResponse, nil
}

func (a *CriteoRetailAdapter) countSponsoredProducts(adResponse *CriteoResponse) int {
	count := 0

	// Iterate through placements
	for _, placementMap := range adResponse.Placements {
		for _, placements := range placementMap {
			for _, placement := range placements {
				if placement.Format == FORMAT_SPONSORED {
					count += len(placement.Products)
				}
			}
		}
	}

	return count
}



