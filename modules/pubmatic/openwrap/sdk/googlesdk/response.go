package googlesdk

import (
	"encoding/json"
	"strings"

	"github.com/beevik/etree"
	"github.com/buger/jsonparser"
	"github.com/golang/glog"
	nativeResponse "github.com/prebid/openrtb/v20/native1/response"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
	"golang.org/x/net/html"
)

const videoClickThroughTagPath = "./VAST/Ad/InLine/Creatives/Creative/Linear/VideoClicks/ClickThrough"

func SetGoogleSDKResponseReject(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) bool {
	if rctx.Endpoint != models.EndpointGoogleSDK {
		return false
	}

	reject := false
	if bidResponse.NBR != nil {
		if !rctx.Debug {
			reject = true
		}
	} else if len(bidResponse.SeatBid) == 0 || len(bidResponse.SeatBid[0].Bid) == 0 {
		reject = true
	}
	return reject
}

func ApplyGoogleSDKResponse(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) *openrtb2.BidResponse {
	if rctx.Endpoint != models.EndpointGoogleSDK || rctx.GoogleSDK.Reject || bidResponse.NBR != nil {
		return bidResponse
	}

	bids, ok := customizeBid(rctx, bidResponse)
	if !ok {
		return bidResponse
	}

	*bidResponse = openrtb2.BidResponse{
		ID:    bidResponse.ID,
		BidID: utils.GetOriginalBidId(bidResponse.SeatBid[0].Bid[0].ID),
		Cur:   bidResponse.Cur,
		SeatBid: []openrtb2.SeatBid{
			{
				Bid: bids,
			},
		},
	}
	return bidResponse
}

func customizeBid(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) ([]openrtb2.Bid, bool) {
	resp, err := json.Marshal(bidResponse)
	if err != nil {
		*bidResponse = openrtb2.BidResponse{}
		return nil, false
	}

	if len(bidResponse.SeatBid) == 0 || len(bidResponse.SeatBid[0].Bid) == 0 {
		return nil, false
	}

	bid := bidResponse.SeatBid[0].Bid[0]
	declaredAd := getDeclaredAd(rctx, bid)

	//reject response if clickthroughURL is empty
	if len(declaredAd.ClickThroughURL) == 0 {
		if bidResponse.NBR == nil {
			bidResponse.NBR = new(openrtb3.NoBidReason)
		}
		*bidResponse.NBR = nbr.ResponseRejectedMissingParam
		return nil, false
	}

	bidExt := models.GoogleSDKBidExt{
		SDKRenderedAd: models.SDKRenderedAd{
			ID:            rctx.GoogleSDK.SDKRenderedAdID,
			RenderingData: string(resp),
			DeclaredAd:    declaredAd,
		},
		//EventNotificationToken: &models.EventNotificationToken{Payload: ""},
		BillingID: "",
	}
	bid.AdM = ""

	// update w and h with request sizes
	if len(rctx.GoogleSDK.FlexSlot) > 0 {
		if impCtx, ok := rctx.ImpBidCtx[bid.ImpID]; ok && impCtx.Banner != nil {
			if len(impCtx.Banner.Format) > 0 {
				bid.W = impCtx.Banner.Format[0].W
				bid.H = impCtx.Banner.Format[0].H
			} else if impCtx.Banner.W != nil && impCtx.Banner.H != nil {
				bid.W = *impCtx.Banner.W
				bid.H = *impCtx.Banner.H
			}
		}
	}

	bid.Ext, err = json.Marshal(bidExt)
	fmt.Println("bidExt", string(bid.Ext))
	if err != nil {
		glog.Errorf("[googlesdk] bidExt:[%s] error:[%s]", bidExt, err.Error())
	}
	return []openrtb2.Bid{bid}, true
}

func getDeclaredAd(rctx models.RequestCtx, bid openrtb2.Bid) models.DeclaredAd {
	var (
		declaredAd = models.DeclaredAd{}
		bidType    = rctx.Trackers[bid.ID].BidType
		nativeResp = nativeResponse.Response{}
	)

	if bidType == models.Banner {
		declaredAd.HTMLSnippet = bid.AdM
		declaredAd.ClickThroughURL = getBannerClickThroughURL(bid)
		return declaredAd
	}

	if bidType == models.Video || rctx.Platform == models.PLATFORM_VIDEO {
		declaredAd.VideoVastXML = bid.AdM
		declaredAd.ClickThroughURL = getVideoClickThroughURL(bid)
		return declaredAd
	}

	if bidType == models.Native {
		if err := json.Unmarshal([]byte(bid.AdM), &nativeResp); err != nil {
			glog.Errorf("[googlesdk] native:[%s] error:[%s]", bid.AdM, err.Error())
		}
		declaredAd.NativeResponse = &nativeResp
		declaredAd.ClickThroughURL = []string{nativeResp.Link.URL}
	}
	return declaredAd
}

func getVideoClickThroughURL(bid openrtb2.Bid) []string {
	videoCreative := strings.TrimSpace(bid.AdM)
	doc := etree.NewDocument()

	if err := doc.ReadFromString(videoCreative); err != nil {
		glog.Errorf("[googlesdk] video_creative:[%s] error:[%s]", videoCreative, err.Error())
		return bid.ADomain
	}

	clickThrough := doc.Element.FindElement(videoClickThroughTagPath)
	if clickThrough == nil || clickThrough.Text() == "" {
		return bid.ADomain
	}
	return []string{clickThrough.Text()}
}

func SetSDKRenderedAdID(app *openrtb2.App, endpoint string) string {
	if endpoint != models.EndpointGoogleSDK || app == nil || app.Ext == nil {
		return ""
	}

	if sdkRenderedAdID, err := jsonparser.GetString(app.Ext, "installed_sdk", "id"); err == nil {
		return sdkRenderedAdID
	}

	if sdkRenderedAdID, err := jsonparser.GetString(app.Ext, "installed_sdk", "[0]", "id"); err == nil {
		return sdkRenderedAdID
	}

	return ""
}

func getBannerClickThroughURL(bid openrtb2.Bid) []string {
	if strings.TrimSpace(bid.AdM) == "" {
		return []string{}
	}

	url := extractClickURLFromJSON(bid.AdM)
	if url != "" {
		return []string{url}
	}

	if url := extractClickURLFromHTML(bid.AdM); url != "" {
		return []string{url}
	}
	return bid.ADomain
}

// extractClickURLFromJSON Extracts click URL from JSON creative
func extractClickURLFromJSON(creative string) string {
	creative = strings.TrimSpace(creative)
	idx := strings.Index(creative, "click_urls")
	if idx == -1 {
		return ""
	}

	// move ahead to find colon
	colonIdx := strings.Index(creative[idx:], ":")
	if colonIdx == -1 {
		return ""
	}
	startIdx := idx + colonIdx + 1

	// Trim leading whitespace
	trimmed := strings.TrimLeft(creative[startIdx:], " \n\r\t")

	// Check if it's array or string
	if strings.HasPrefix(trimmed, "[") {
		// It's an array
		end := strings.Index(trimmed, "]")
		if end == -1 {
			return ""
		}
		arrayContent := trimmed[1:end]
		parts := strings.Split(arrayContent, ",")
		if len(parts) == 0 {
			return ""
		}
		return strings.Trim(parts[0], `" '`)

	} else if strings.HasPrefix(trimmed, "\"") || strings.HasPrefix(trimmed, "'") {
		quoteChar := trimmed[0] // either ' or "
		endIdx := strings.Index(trimmed[1:], string(quoteChar))
		if endIdx != -1 {
			return trimmed[1 : 1+endIdx]
		}
	}

	return ""
}

// extractClickURLFromHTML Parse HTML and find first anchor tag href
func extractClickURLFromHTML(htmlStr string) string {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil || doc == nil {
		return ""
	}
	return findFirstHref(doc)
}

// findFirstHref Recursively walks HTML tree and returns first <a href="">
func findFirstHref(n *html.Node) string {
	if n == nil {
		return ""
	}
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				return attr.Val
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findFirstHref(c); result != "" {
			return result
		}
	}
	return ""
}
