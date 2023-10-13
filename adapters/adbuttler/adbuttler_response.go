package adbuttler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type AdButlerBeacon struct {
	Type        string `json:"type,omitempty"`
	TrackingUrl string `json:"url,omitempty"`
}

type AdButlerBid struct {
	CPCBid      float64           `json:"cpc_bid,omitempty"`
	CPCSpend    float64           `json:"cpc_spend,omitempty"`
	CampaignID  int64             `json:"campaign_id,omitempty"`
	ProductData map[string]string `json:"item,omitempty"`
	Beacons     []*AdButlerBeacon `json:"beacons,omitempty"`
}

type AdButlerResponse struct {
	Status string         `json:"status,omitempty"`
	Code   int32          `json:"code,omitempty"`
	Bids   []*AdButlerBid `json:"items,omitempty"`
}

func AddDefaultFields(bid *openrtb2.Bid) {
	if bid != nil {
		bid.CrID = "DefaultCRID"
	}
}

func GenerateUniqueBidID() string {
	id := uuid.New()
	return id.String()
}

func (a *AdButtlerAdapter) MakeBids(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	var errors []error

	if response.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if response.StatusCode == http.StatusBadRequest {
		err := &errortypes.BadInput{
			Message: "Unexpected status code: 400. Bad request from Adbutler.",
		}
		return nil, []error{err}
	}

	if response.StatusCode != http.StatusOK {
		err := &errortypes.BadServerResponse{
			Message: fmt.Sprintf("Unexpected status code: %d", response.StatusCode),
		}
		return nil, []error{err}
	}

	var adButlerResp AdButlerResponse
	if err := json.Unmarshal(response.Body, &adButlerResp); err != nil {
		return nil, []error{&errortypes.BadServerResponse{
			Message: "Bad Server Response",
		}}
	}

	//Temporarily for Debugging
	//u, _ := json.Marshal(adButlerResp)
	//fmt.Println(string(u))

	if adButlerResp.Status == RESPONSE_NOADS {
		return nil, []error{&errortypes.BidderFailedSchemaValidation{
			Message: fmt.Sprintf("Error Occured at Adbutler for the given request with ErrorCode %d", adButlerResp.Code),
		}}
	}

	if adButlerResp.Status == RESPONSE_SUCCESS && (adButlerResp.Bids == nil ||
		len(adButlerResp.Bids) <= 0) {
		return nil, []error{&errortypes.NoBidPrice{
			Message: "No Bid For the given Request",
		}}
	}

	if adButlerResp.Status == RESPONSE_SUCCESS && (adButlerResp.Bids != nil &&
		len(adButlerResp.Bids) > 0) {
		impID := internalRequest.Imp[0].ID
		responseF := a.GetBidderResponse(internalRequest, &adButlerResp, impID)
		if len(responseF.Bids) <= 0 {
			return nil, []error{&errortypes.NoValidBid{
				Message: "No Valid Bid For the given Request",
			}}
		}
		return responseF, errors
	}

	err := fmt.Errorf("unknown error occcured for the given request from adbutler")
	errors = append(errors, err)

	return nil, errors

}

func EncodeURl(url string) string {
	str := base64.StdEncoding.EncodeToString([]byte(url))
	return str
}

func (a *AdButtlerAdapter) GetBidderResponse(request *openrtb2.BidRequest, adButlerResp *AdButlerResponse, requestImpID string) *adapters.BidderResponse {

	bidResponse := adapters.NewBidderResponseWithBidsCapacity(len(adButlerResp.Bids))
	var commerceExt *openrtb_ext.ExtImpCommerce
	var adbutlerID, zoneID, adbUID string
	var configValueMap = make(map[string]string)

	if len(request.Imp) > 0 {
		commerceExt, _ = a.getImpressionExt(&(request.Imp[0]))
		for _, obj := range commerceExt.Bidder.CustomConfig {
			configValueMap[obj.Key] = obj.Value
		}

		val, ok := configValueMap[BIDDERDETAILS_PREFIX+BD_ACCOUNT_ID]
		if ok {
			adbutlerID = val
		}

		val, ok = configValueMap[BIDDERDETAILS_PREFIX+BD_ZONE_ID]
		if ok {
			zoneID = val
		}
		adbUID = request.User.ID

	}

	for index, adButlerBid := range adButlerResp.Bids {

		bidID := GenerateUniqueBidID()
		impID := requestImpID + "_" + strconv.Itoa(index+1)
		bidPrice := adButlerBid.CPCBid
		campaignID := strconv.FormatInt(adButlerBid.CampaignID, 10)
		clickPrice := adButlerBid.CPCSpend

		var productid string
		//Retailer Specific ProductID is present from Product Feed Template
		val, ok := configValueMap[PRODUCTTEMPLATE_PREFIX+PD_TEMPLATE_PRODUCTID]
		if ok {
			productid = adButlerBid.ProductData[val]
		}
		if productid == "" {
			productid = adButlerBid.ProductData[DEFAULT_PRODUCTID]
		}

		var impressionUrl, clickUrl, conversionUrl string
		for _, beacon := range adButlerBid.Beacons {
			switch beacon.Type {
			case BEACONTYPE_IMP:
				if beacon.TrackingUrl != "" {
					impressionUrl = IMP_KEY + EncodeURl(beacon.TrackingUrl)
				}
			case BEACONTYPE_CLICK:
				if beacon.TrackingUrl != "" {
					clickUrl = CLICK_KEY + EncodeURl(beacon.TrackingUrl)
				}
			}
		}

		conversionUrl = GenerateConversionUrl(adbutlerID, zoneID, adbUID, productid)

		bidExt := &openrtb_ext.ExtBidCommerce{
			ProductId:     productid,
			ClickUrl:      clickUrl,
			ClickPrice:    clickPrice,
			ConversionUrl: conversionUrl,
		}

		bid := &openrtb2.Bid{
			ID:    bidID,
			ImpID: impID,
			Price: bidPrice,
			CID:   campaignID,
			IURL:  impressionUrl,
		}

		if !areMandatoryFieldsPresent(bidExt, bid) {
			continue
		}

		AddDefaultFields(bid)

		bidExtJSON, err1 := json.Marshal(bidExt)
		if nil == err1 {
			bid.Ext = json.RawMessage(bidExtJSON)
		}

		typedbid := &adapters.TypedBid{
			Bid:  bid,
			Seat: openrtb_ext.BidderName(SEAT_ADBUTLER),
		}

		bidResponse.Bids = append(bidResponse.Bids, typedbid)
	}
	return bidResponse
}

func areMandatoryFieldsPresent(bidExt *openrtb_ext.ExtBidCommerce, bid *openrtb2.Bid) bool {

	if bid.Price == 0 || bid.IURL == "" {
		return false
	}
	if bidExt.ProductId == "" || bidExt.ClickUrl == "" {
		return false
	}

	return true
}

func GenerateConversionUrl(adbutlerID, zoneID, adbUID, productID string) string {
	/*var hostname string
		url, err := url.Parse(clickurl)
	    if err == nil {
			hostname = url.Hostname()
		 }

		conversionUrl := strings.Replace(CONVERSION_URL, CONV_HOSTNAME, hostname, 1)*/
	conversionUrl := strings.Replace(CONVERSION_URL, CONV_ADBUTLERID, adbutlerID, 1)
	conversionUrl = strings.Replace(conversionUrl, CONV_ZONEID, zoneID, 1)
	conversionUrl = strings.Replace(conversionUrl, CONV_ADBUID, adbUID, 1)
	conversionUrl = strings.Replace(conversionUrl, CONV_IDENTIFIER, productID, 1)

	return conversionUrl

}
