package adbutler_onsite

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/PubMatic-OpenWrap/prebid-server/errortypes"
	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type AdButlerBeacon struct {
	Type        string `json:"type,omitempty"`
	TrackingUrl string `json:"url,omitempty"`
}

type Placement struct {
	BannerID             string `json:"banner_id,omitempty"`
	Width                string `json:"width,omitempty"`
	Height               string `json:"height,omitempty"`
	AltText              string `json:"alt_text,omitempty"`
	AccompaniedHTML      string `json:"accompanied_html,omitempty"`
	Target               string `json:"target,omitempty"`
	TrackingPixel        string `json:"tracking_pixel,omitempty"`
	Body                 string `json:"body,omitempty"`
	RedirectURL          string `json:"redirect_url,omitempty"`
	RefreshURL           string `json:"refresh_url,omitempty"`
	Rct                  string `json:"rct,omitempty"`
	Rcb                  string `json:"rcb,omitempty"`
	RefreshTime          string `json:"refresh_time,omitempty"`
	PlacementID          string `json:"placement_id,omitempty"`
	UserFrequencyViews   string `json:"user_frequency_views,omitempty"`
	UserFrequencyStart   string `json:"user_frequency_start,omitempty"`
	UserFrequencyExpiry  string `json:"user_frequency_expiry,omitempty"`
	ViewableURL          string `json:"viewable_url,omitempty"`
	EligibleURL          string `json:"eligible_url,omitempty"`
	ImageURL             string `json:"image_url,omitempty"`
	ImpressionsRemaining int    `json:"impressions_remaining,omitempty"`
	HasQuota             bool   `json:"has_quota,omitempty"`
}

type AdSet struct {
	Status     string       `json:"status,omitempty"`
	Placements []*Placement `json:"placements,omitempty"`
}

type AdButlerOnsiteResponse map[string]AdSet

func (a *AdButlerOnsiteAdapter) MakeBids(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {

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

	var adButlerResp AdButlerOnsiteResponse
	if err := json.Unmarshal(response.Body, &adButlerResp); err != nil {
		return nil, []error{&errortypes.BadServerResponse{
			Message: "Bad Server Response",
		}}
	}

	if len(adButlerResp) <= 0 {
		return nil, []error{&errortypes.NoBidPrice{
			Message: "No Bid For the given Request",
		}}
	}

	noOfSuccess := 0
	noOfPlacements := 0

	for _, adSetObject := range adButlerResp {
		if adSetObject.Status == RESPONSE_SUCCESS {
			noOfSuccess++
			noOfPlacements += len(adSetObject.Placements)
		}
	}

	if noOfSuccess == 0 {
		return nil, []error{&errortypes.NoValidBid{
			Message: "No Valid Bid For the given Request",
		}}
	}

	responseF := a.GetBidderResponse(internalRequest, &adButlerResp, noOfPlacements)
	if len(responseF.Bids) <= 0 {
		return nil, []error{&errortypes.NoValidBid{
			Message: "No Valid Bid For the given Request",
		}}
	}
	return responseF, errors

}

func (a *AdButlerOnsiteAdapter) GetBidderResponse(request *openrtb2.BidRequest, adButlerResp *AdButlerOnsiteResponse, noOfPlacement int) *adapters.BidderResponse {

	bidResponse := adapters.NewBidderResponseWithBidsCapacity(noOfPlacement)

	for _, adSetObject := range *adButlerResp {

		for index, adButlerBid := range adSetObject.Placements {

			requestImpID := strconv.Itoa(index)
			bidID := adapters.GenerateUniqueBidIDComm()
			impID := requestImpID + "_" + strconv.Itoa(index+1)
			width, _ := strconv.Atoi(adButlerBid.Width)
			height, _ := strconv.Atoi(adButlerBid.Height)

			adm := getADM(adButlerBid)

			bidExt := &openrtb_ext.ExtBidCMOnsite{
				ViewUrl:  adButlerBid.ViewableURL,
				ClickUrl: adButlerBid.RedirectURL,
			}

			bid := &openrtb2.Bid{
				ID:    bidID,
				ImpID: impID,
				NURL:  adButlerBid.EligibleURL,
				W:     int64(width),
				H:     int64(height),
				AdM:   adm,
			}

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
	}

	return bidResponse
}

func getADM(adButlerBid *Placement) string {

	if adButlerBid.Body != "" {
		return adButlerBid.Body
	}

	if adButlerBid.ImageURL != "" {
		return fmt.Sprintf(IMAGE_URL_TEMPLATE, adButlerBid.BannerID, adButlerBid.ImageURL, adButlerBid.Width, adButlerBid.Height)
	}

	return ""
}
