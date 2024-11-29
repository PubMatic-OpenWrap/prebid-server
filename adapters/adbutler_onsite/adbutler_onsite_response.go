package adbutler_onsite

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type AdButlerBeacon struct {
	Type        string `json:"type,omitempty"`
	TrackingUrl string `json:"url,omitempty"`
}


const (
	MarkupInvalid openrtb2.MarkupType = 0
)


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
	AccupixelURL         string `json:"accupixel_url,omitempty"`
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

	//Temporarily for Debugging
	/*var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, response.Body, "", "  ")
	if err != nil {
		fmt.Println("Failed to parse JSON:", err)
		return nil, []error{err}
	}
	fmt.Println(prettyJSON.String())*/

	var adButlerResp AdButlerOnsiteResponse
	if err := json.Unmarshal(response.Body, &adButlerResp); err != nil {
		return nil, []error{&errortypes.BadServerResponse{
			Message: "Bad Server Response",
		}}
	}

	if len(adButlerResp) <= 0 {
		return nil, []error{&errortypes.NoValidBid{
			Message: "No Bid For the given Request",
		}}
	}

	noOfPlacements := 0

	for _, adSetObject := range adButlerResp {
		if adSetObject.Status == RESPONSE_SUCCESS {
			noOfPlacements += len(adSetObject.Placements)
		}
	}

	if noOfPlacements == 0 {
		return nil, []error{&errortypes.NoBidPrice{
			Message: "No Bid For the given Request",
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

// randomFloatInRange generates a random float64 in the range (1, 3].
func randomPriceInRange() float64 {
	// Seed the random number generator to ensure different results each time
	rand.Seed(time.Now().UnixNano())

	// Generate a random float64 in the range (0, 2] and then shift it to (1, 3]
	return 1 + rand.Float64()*2
}

func (a *AdButlerOnsiteAdapter) GetBidderResponse(request *openrtb2.BidRequest, adButlerResp *AdButlerOnsiteResponse, noOfBids int) *adapters.BidderResponse {

	impIDMap := getImpIDMap(request)

	bidResponse := adapters.NewBidderResponseWithBidsCapacity(noOfBids)

	for zoneID, adSetObject := range *adButlerResp {

		for _, adButlerBid := range adSetObject.Placements {

			var impID string
			_, ok := impIDMap[zoneID]
			if ok {
				impID = impIDMap[zoneID][0]
				impIDMap[zoneID] = impIDMap[zoneID][1:]
			} else {
				continue
			}

			bidID := adapters.GenerateUniqueBidIDComm()
			width, _ := strconv.Atoi(adButlerBid.Width)
			height, _ := strconv.Atoi(adButlerBid.Height)

			adm, adType := getADM(adButlerBid)

			adm = encodeRedirectURL(adm, Pattern_Click_URL, CLICK_KEY)

			if adType == Adtype_Invalid {
				continue
			}

			var nURL, viewURL, clickURL string

			if adButlerBid.EligibleURL != "" {
				nURL = IMP_KEY + adapters.EncodeURL(adButlerBid.EligibleURL)
			} else if adButlerBid.AccupixelURL != "" {
				nURL = IMP_KEY + adapters.EncodeURL(adButlerBid.AccupixelURL)
			}

			if adButlerBid.ViewableURL != "" {
				viewURL = VIEW_KEY + adapters.EncodeURL(adButlerBid.ViewableURL)
			}
			if adButlerBid.RedirectURL != "" {
				clickURL = CLICK_KEY + adapters.EncodeURL(adButlerBid.RedirectURL)
			}

			bidExt := &openrtb_ext.ExtBidCMOnsite{
				ViewUrl:  viewURL,
				ClickUrl: clickURL,
			}

			bid := &openrtb2.Bid{
				ID:    bidID,
				ImpID: impID,
				BURL:  nURL,
				W:     int64(width),
				H:     int64(height),
				AdM:   adm,
				MType:   adType,
				Price: randomPriceInRange(),  //Temporary calculation
				CrID: adButlerBid.BannerID,
			}

			adapters.AddDefaultFieldsComm(bid)

			bidExtJSON, err1 := json.Marshal(bidExt)
			if nil == err1 {
				bid.Ext = json.RawMessage(bidExtJSON)
			}

			typedbid := &adapters.TypedBid{
				Bid:  bid,
				Seat: openrtb_ext.BidderName(Seat_AdbutlerOnsite),
			}

			bidResponse.Bids = append(bidResponse.Bids, typedbid)
		}
	}
	return bidResponse
}

func getADM(adButlerBid *Placement) (string, openrtb2.MarkupType) {

	if adButlerBid.Body != "" {
		return adButlerBid.Body, openrtb2.MarkupNative
	}

	if adButlerBid.ImageURL != "" {
		return fmt.Sprintf(IMAGE_URL_TEMPLATE, adButlerBid.BannerID, adButlerBid.ImageURL, adButlerBid.Width, adButlerBid.Height), openrtb2.MarkupBanner
	}

	return "", MarkupInvalid
}

func getImpIDMap(request *openrtb2.BidRequest) map[string][]string {

	_, requestExt, errors := adapters.ValidateCMOnsiteRequest(request)

	if len(errors) > 0 {
		return nil
	}

	if requestExt == nil {
		return nil
	}

	inventoryDetails, _, _ := adapters.GetInventoryAndAccountDetailsCMOnsite(requestExt)

	impIDMap := make(map[string][]string)

	for _, imp := range request.Imp {
		inventory, ok := inventoryDetails[InventoryIDOnsite_Prefix+imp.TagID]
		if ok {
			zoneID := strconv.Itoa(inventory.AdbulterZoneID)
			impIDArray, ok := impIDMap[zoneID]
			var impID string
			if imp.Banner.Pos != nil {
				impID = strconv.Itoa(int(imp.Banner.Pos.Ptr().Val())) + imp.ID
			} else {
				impID = strconv.Itoa(0) + imp.ID
			}
			if ok {
				impIDArray = append(impIDArray, impID)
				impIDMap[zoneID] = impIDArray
			} else {
				impIDArray := make([]string, 0)
				impIDArray = append(impIDArray, impID)
				impIDMap[zoneID] = impIDArray
			}
		}
	}

	//Sorting according pos and then trimming the position
	for _, val := range impIDMap {
		sort.Strings(val)
		for i := 0; i < len(val); i++ {
			val[i] = val[i][1:]
		}
	}

	return impIDMap
}

func encodeRedirectURL(phrase, urlToSearch, preString string) string {
	regex := regexp.MustCompile(urlToSearch)
	matches := regex.FindAllStringSubmatch(phrase, -1)
	if len(matches) == 0 {
		return phrase
	}
	modifiedPhrase := phrase
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		encodedURL := preString + adapters.EncodeURL(match[1])
		modifiedPhrase = strings.Replace(modifiedPhrase, match[1], encodedURL, 1)
	}
	return modifiedPhrase
}



