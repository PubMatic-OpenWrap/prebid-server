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
	admActivate         = "<div style='margin:0;padding:0;'><a href='CONVERT_LANDING_PAGE' target='_blank'><img src='CONVERT_CREATIVE'></a></div>"
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

			if bidType == openrtb_ext.BidTypeNative {
				bid.AdM, err = getNativeAdm(bid.AdM)
				if err != nil {
					errs = append(errs, err)
				}
			}

			bUrl := extractBillingURL(bid.AdM)
			bid.BURL = bUrl
			activateCampaignId := extractWDSCampID(bid.AdM)
			if activateCampaignId != "" {
				bid.CID = activateCampaignId
			}

			updatedAdmActivate := strings.Replace(admActivate, "CONVERT_CREATIVE", bid.IURL, 1)
			bid.AdM = updatedAdmActivate

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





