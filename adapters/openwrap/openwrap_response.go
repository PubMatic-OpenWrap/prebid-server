package openwrap

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/golang/glog"
	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

const (
	buyId                    = "buyid"
	clickScript              = "<script>async function handleAdClick_VALID_IMP_INDEX(redirectUrl,clickUrls,target){const clickPromises=clickUrls.map(url=>new Promise(resolve=>{if(navigator.sendBeacon){const success=navigator.sendBeacon(url);resolve(success)}else{const img=new Image();img.onload=()=>resolve(true);img.onerror=()=>resolve(false);img.src=url;setTimeout(()=>resolve(false),1000)}}));try{await Promise.all(clickPromises)}catch(e){}finally{if(target===\"_blank\"){window.open(redirectUrl,\"_blank\")}else{window.top.location.href=redirectUrl}}}document.addEventListener(\"DOMContentLoaded\",function(){var adLink=document.getElementById(\"ad-click-link-VALID_IMP_INDEX\");if(adLink){var redirecturl=\"CONVERT_LANDING_PAGE_DV\";var target=adLink.getAttribute(\"target\")||\"_top\";var clickurls=[ALL_CLICK_URLS];adLink.addEventListener(\"click\",function(e){e.preventDefault();handleAdClick_VALID_IMP_INDEX(redirecturl,clickurls,target)})}});</script>"
	admActivate              = "<div style='margin:0;padding:0;'><a href='CONVERT_LANDING_PAGE' target='_top'><img src='CONVERT_CREATIVE'></a></div>"
	admActivateNative        = "<div style='margin:0;padding:0;'> <a id=\"ad-click-link-VALID_IMP_INDEX\" href=\"#\"><img src='CONVERT_CREATIVE'></a><iframe width='0' scrolling='no' height='0' frameborder='0' src='DSP_IMP_URL' style='position:absolute;top:-15000px;left:-15000px' vspace='0' hspace='0' marginwidth='0' marginheight='0' allowtransparency='true' name='dspbeacon'></iframe> <iframe width='0' scrolling='no' height='0' frameborder='0' src='PUB_IMP_URL' style='position:absolute;top:-15000px;left:-15000px' vspace='0' hspace='0' marginwidth='0' marginheight='0' allowtransparency='true' name='pubmbeacon'></iframe></div>"
	landingUrl               = "https://ci-va2qa-mgmt.pubmatic.com/adservercommerce/convert/onsite/dv/redirect?redirectURL=CONVERT_LANDING_PAGE_DV&dvURL=DV_CLICK_URL&pubURL=PUB_CLICK_URL"
	redirectDVTestLandingUrl = "https://ci-va2qa-mgmt.pubmatic.com/v2/ui-demo-app/retailer1/coke"
	admActivateBanner        = "<div style='margin:0;padding:0;'> <a id=\"ad-click-link-VALID_IMP_INDEX\" href=\"#\"><img src='CONVERT_CREATIVE'></a></div>"
	thirdPartyTagCreative    = "https://go.trader.ca/wp-content/uploads/2022/02/250X250.png"
)

type pubmaticBidExt struct {
	BidType           *int                 `json:"BidType,omitempty"`
	VideoCreativeInfo *pubmaticBidExtVideo `json:"video,omitempty"`
	Marketplace       string               `json:"marketplace,omitempty"`
}

// Adm represents the top-level object for the adm.
type Adm struct {
	Ver    string  `json:"ver"`
	Assets []Asset `json:"assets"`
}

// Asset represents an asset within the adm.
type Asset struct {
	Id   int         `json:"id"`
	Data *AssetData  `json:"data,omitempty"`
	Img  *AssetImage `json:"img,omitempty"`
}

// AssetData represents the data asset (e.g. text).
type AssetData struct {
	Value string `json:"value"`
}

// AssetImage represents the image asset (e.g. url, w, h).
type AssetImage struct {
	Url string `json:"url"`
	W   int64  `json:"w"`
	H   int64  `json:"h"`
}

func replaceAdm(adm string, replace string) string {
	// Pattern 1: Match <script> block surrounded by optional whitespace and {}
	reWithBraces := regexp.MustCompile(`(?s)\{\s*<script[^>]*>.*?CONVERT_SSP_TAG.*?</script>\s*\}`)

	// Pattern 2: Match <script> block directly
	reWithoutBraces := regexp.MustCompile(`(?s)<script[^>]*>.*?CONVERT_SSP_TAG.*?</script>`)

	// First try replacing the script block wrapped in braces
	if reWithBraces.MatchString(adm) {
		return reWithBraces.ReplaceAllString(adm, replace)
	}

	// If not found, try replacing the standalone script block
	return reWithoutBraces.ReplaceAllString(adm, replace)
}

func getScriptContent(adm string) string {
	// Regex to match {<script>...</script>} with optional whitespace/newlines
	reWithBraces := regexp.MustCompile(`(?s)\{\s*<script[^>]*>.*?CONVERT_SSP_TAG.*?</script>\s*\}`)
	reWithoutBraces := regexp.MustCompile(`(?s)<script[^>]*>.*?CONVERT_SSP_TAG.*?</script>`)

	// Try match with curly braces first
	if match := reWithBraces.FindString(adm); match != "" {
		return match
	}

	// Fallback to no curly braces
	if match := reWithoutBraces.FindString(adm); match != "" {
		return match
	}

	return ""
}

// Function to extract creativeId and clickurl from script content
func parseScriptContent(script string) (string, string) {
	// Regex to match creativeId value (numeric or macro)
	creativeIdRegex := regexp.MustCompile(`(?i)creativeId\s*=\s*([^;]+)`)

	// Regex to match clickurl value
	clickurlRegex := regexp.MustCompile(`(?i)clickurl\s*=\s*([^\s;]+)`)

	creativeId := ""
	clickurl := ""

	// Extract creativeId and handle macro case
	if match := creativeIdRegex.FindStringSubmatch(script); len(match) > 1 {
		raw := strings.TrimSpace(match[1])
		if strings.Contains(strings.ToLower(raw), "creative_id") {
			creativeId = "" // macro found, skip
		} else {
			creativeId = raw
		}
	}

	// Extract clickurl
	if match := clickurlRegex.FindStringSubmatch(script); len(match) > 1 {
		clickurl = strings.TrimSpace(match[1])
	}

	return creativeId, clickurl
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
			activateCampaignId := extractWDSCampID(bid.AdM)
			if activateCampaignId != "" {
				bid.CID = activateCampaignId
			}
			if strings.Contains(bid.AdM, "CONVERT_SSP_TAG") {
				//updatedAdmActivate := strings.Replace(admActivateBanner, "CONVERT_CREATIVE", thirdPartyTagCreative, 1)
				//finalClickScript := strings.Replace(clickScript, "CONVERT_LANDING_PAGE_DV", redirectDVTestLandingUrl, 1)
				updatedAdmActivate := admActivateBanner
				finalClickScript := clickScript
				scriptContent := getScriptContent(bid.AdM)

				if scriptContent == "" {
					continue
				}
				creativeId, clickUrl := parseScriptContent(scriptContent)
				clickUrl = "\"" + clickUrl + "\""
				finalClickScript = strings.Replace(finalClickScript, "ALL_CLICK_URLS", clickUrl, 1)
				updatedAdmActivate = finalClickScript + updatedAdmActivate
				updatedAdmActivate = strings.Replace(updatedAdmActivate, "VALID_IMP_INDEX", strconv.Itoa(i), 4)
				bid.AdM = replaceAdm(bid.AdM, updatedAdmActivate)

				if bid.CrID == "" {
					bid.CrID = creativeId
				} else if creativeId != "" && bid.CrID != creativeId {
					bid.CrID = bid.CrID + "," + creativeId
				}

				// Log error with pubid and imp tagid
				pubID := ""
				if internalRequest.Site != nil && internalRequest.Site.Publisher != nil {
					pubID = internalRequest.Site.Publisher.ID
				}

				if internalRequest.App != nil && internalRequest.App.Publisher != nil {
					pubID = internalRequest.App.Publisher.ID
				}
				glog.Errorf("Openwrap creative processing - PubID: %s, CreativeId: %s, BidCrID: %s",
					pubID, creativeId, bid.CrID)

			} else if bid.MType == openrtb2.MarkupNative {
				// Define a structure to unmarshal the adm string.
				var admData struct {
					Link struct {
						URL           string   `json:"url"`
						Clicktrackers []string `json:"clicktrackers"`
					} `json:"link"`
					Imptrackers []string `json:"imptrackers"`
				}

				var adm Adm
				var width, height int64

				err := json.Unmarshal([]byte(bid.AdM), &adm)
				if err != nil {
					continue
				}
				// Iterate over assets to find asset with id==1.
				for _, asset := range adm.Assets {
					if asset.Id == 1 && asset.Img != nil {
						width = asset.Img.W
						height = asset.Img.H
					}
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

				updatedAdmActivate := strings.Replace(admActivateNative, "CONVERT_CREATIVE", bid.IURL, 1)
				updatedAdmActivate = strings.Replace(updatedAdmActivate, "DSP_IMP_URL", impTrackersStr, 1)
				if len(admData.Imptrackers) > 1 {
					updatedAdmActivate = strings.Replace(updatedAdmActivate, "PUB_IMP_URL", admData.Imptrackers[1], 1)
				}
				combinedClicks := "\"" + linkURL + "\",\"" + clickTrackersStr + "\""
				finalClickScript := strings.Replace(clickScript, "CONVERT_LANDING_PAGE_DV", redirectDVTestLandingUrl, 1)
				finalClickScript = strings.Replace(finalClickScript, "ALL_CLICK_URLS", combinedClicks, 1)
				updatedAdmActivateNative := finalClickScript + updatedAdmActivate
				updatedAdmActivateNative = strings.Replace(updatedAdmActivateNative, "VALID_IMP_INDEX", strconv.Itoa(i), 4)

				bid.AdM = updatedAdmActivateNative
				bid.MType = openrtb2.MarkupBanner
				bid.BURL = ""
				bidType = openrtb_ext.BidTypeBanner

				if width != 0 && height != 0 {
					bid.W = width
					bid.H = height
				}
			}

			bidResponse.Bids = append(bidResponse.Bids, &adapters.TypedBid{
				Bid:      &bid,
				BidType:  bidType,
				BidVideo: impVideo,
				Seat:     openrtb_ext.BidderName("openwrap"),
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

// getMapFromJSON converts JSON to map
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
