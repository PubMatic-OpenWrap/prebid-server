package openwrap

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

const (
	buyId               = "buyid"
	admActivate         = "<div style='margin:0;padding:0;'><a href='CONVERT_LANDING_PAGE' target='_top'><img src='CONVERT_CREATIVE'></a></div>"
	landingUrl 			= "https://cmpbid.pubmatic.com/convert/onsite/dv/redirect?redirectURL='CONVERT_LANDING_PAGE_DV'&dvURL='DV_CLICK_URL'&pubURL='PUB_CLICK_URL'"
	redirectDVTestLandingUrl = "https://ci-va2qa-mgmt.pubmatic.com/v2/ui-demo-app/retailer1/coke"
)

type pubmaticBidExt struct {
	BidType           *int                 `json:"BidType,omitempty"`
	VideoCreativeInfo *pubmaticBidExtVideo `json:"video,omitempty"`
	Marketplace       string               `json:"marketplace,omitempty"`
}

func extractBillingURL(adm string) string {
	// Define the regular expression pattern to match the URL
	// that contains "/AdServer/AdDisplayTrackerServlet"
	pattern := `https?://[^\s"]+/AdServer/AdDisplayTrackerServlet[^\s"]*`

	// Compile the regular expression
	re := regexp.MustCompile(pattern)

	// Find the first match for the pattern in the adm string
	match := re.FindString(adm)

	return match
}


// Function to extract the value of wDspCampId from the input string
func extractWDSCampID(input string) string {
	// Define the prefix and suffix to search for
	prefix := "&wDspCampId="
	suffix := "&"

	// Find the starting position of the prefix
	start := strings.Index(input, prefix)
	if start == -1 {
		return "" // Return empty string if prefix is not found
	}

	// Move the starting position past the prefix
	start += len(prefix)

	// Find the ending position of the suffix
	end := strings.Index(input[start:], suffix)
	if end == -1 {
		return "" // Return empty string if suffix is not found
	}

	// Extract and return the value between the prefix and suffix
	return input[start : start+end]
}

type pubmaticBidExtVideo struct {
	Duration *int `json:"duration,omitempty"`
}
func (a *OpenWrapAdapter) MakeBids(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	if response.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if response.StatusCode == http.StatusBadRequest {
		return nil, []error{&errortypes.BadInput{
			Message: fmt.Sprintf("Unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode),
		}}
	}

	if response.StatusCode != http.StatusOK {
		return nil, []error{fmt.Errorf("Unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode)}
	}

	var bidResp openrtb2.BidResponse
	if err := json.Unmarshal(response.Body, &bidResp); err != nil {
		return nil, []error{err}
	}

	bidResponse := adapters.NewBidderResponseWithBidsCapacity(5)

	var errs []error
	for _, sb := range bidResp.SeatBid {
		for i := 0; i < len(sb.Bid); i++ {
			bid := sb.Bid[i]
		
			impVideo := &openrtb_ext.ExtBidPrebidVideo{}

			if len(bid.Cat) > 1 {
				bid.Cat = bid.Cat[0:1]
			}

			var bidExt *pubmaticBidExt
			bidType := openrtb_ext.BidTypeBanner
			err := json.Unmarshal(bid.Ext, &bidExt)
			if err != nil {
				errs = append(errs, err)
			} else if bidExt != nil {
				if bidExt.VideoCreativeInfo != nil && bidExt.VideoCreativeInfo.Duration != nil {
					impVideo.Duration = *bidExt.VideoCreativeInfo.Duration
				}
				bidType = getBidType(bidExt)
			}



			bUrl := extractBillingURL(bid.AdM)
			bid.BURL = bUrl
			activateCampaignId := extractWDSCampID(bid.AdM)
			if activateCampaignId != "" {
				bid.CID = activateCampaignId
			}

			updatedAdmActivate := strings.Replace(admActivate, "CONVERT_CREATIVE", bid.IURL, 1)
			if bid.MType ==  openrtb2.MarkupBanner{
				bid.AdM = updatedAdmActivate
			} else if bid.MType ==  openrtb2.MarkupNative{
				// Define a structure to unmarshal the adm string.
				var admData struct {
					Link struct {
						URL string `json:"url"`
						Clicktrackers []string `json:"clicktrackers"`
					} `json:"link"`
					Imptrackers []string `json:"imptrackers"`
				}
				// Unmarshal the adm JSON string.
				// Unmarshal the adm JSON string and check for errors.
				if err := json.Unmarshal([]byte(bid.AdM), &admData); err != nil {
					continue // or handle the error as appropriate
				}
				// Check if imptrackers and clicktrackers slices contain at least one element.
				if len(admData.Imptrackers) == 0 {
					continue // or handle the situation as needed
				}
				if len(admData.Link.Clicktrackers) == 0 {
					continue // or handle the situation as needed
				}
				
				// Extract the link URL.
				linkURL := admData.Link.URL
				impTrackersStr := admData.Imptrackers[0]
				clickTrackersStr := admData.Link.Clicktrackers[0]
	
				updatedFinalLandingUrl := strings.Replace(landingUrl, "CONVERT_LANDING_PAGE_DV", redirectDVTestLandingUrl, 1)
				updatedFinalLandingUrl = strings.Replace(updatedFinalLandingUrl, "DV_CLICK_URL", adapters.EncodeURL(linkURL), 1)
				updatedFinalLandingUrl = strings.Replace(updatedFinalLandingUrl, "PUB_CLICK_URL", adapters.EncodeURL(clickTrackersStr), 1)
				updatedAdmActivateNative := strings.Replace(updatedAdmActivate, "CONVERT_LANDING_PAGE", updatedFinalLandingUrl, 1)
				bid.AdM = updatedAdmActivateNative
				bid.MType = openrtb2.MarkupBanner
				bid.BURL = impTrackersStr
				bidType = openrtb_ext.BidTypeBanner
			}

			bidResponse.Bids = append(bidResponse.Bids, &adapters.TypedBid{
				Bid:        &bid,
				BidType:    bidType,
				BidVideo:   impVideo,
				Seat:       openrtb_ext.BidderName(sb.Seat),
			})

		}
	}
	if bidResp.Cur != "" {
		bidResponse.Currency = bidResp.Cur
	}
	return bidResponse, errs
}

func getBidType(bidExt *pubmaticBidExt) openrtb_ext.BidType {
	// setting "banner" as the default bid type
	bidType := openrtb_ext.BidTypeBanner
	if bidExt != nil && bidExt.BidType != nil {
		switch *bidExt.BidType {
		case 0:
			bidType = openrtb_ext.BidTypeBanner
		case 1:
			bidType = openrtb_ext.BidTypeVideo
		case 2:
			bidType = openrtb_ext.BidTypeNative
		default:
			// default value is banner
			bidType = openrtb_ext.BidTypeBanner
		}
	}
	return bidType
}

func getNativeAdm(adm string) (string, error) {
	var err error
	nativeAdm := make(map[string]interface{})
	err = json.Unmarshal([]byte(adm), &nativeAdm)
	if err != nil {
		return adm, errors.New("unable to unmarshal native adm")
	}

	// move bid.adm.native to bid.adm
	if _, ok := nativeAdm["native"]; ok {
		//using jsonparser to avoid marshaling, encode escape, etc.
		value, _, _, err := jsonparser.Get([]byte(adm), string(openrtb_ext.BidTypeNative))
		if err != nil {
			return adm, errors.New("unable to get native adm")
		}
		adm = string(value)
	}

	return adm, nil
}

//getMapFromJSON converts JSON to map
func getMapFromJSON(source json.RawMessage) map[string]interface{} {
	if source != nil {
		dataMap := make(map[string]interface{})
		err := json.Unmarshal(source, &dataMap)
		if err == nil {
			return dataMap
		}
	}
	return nil
}










