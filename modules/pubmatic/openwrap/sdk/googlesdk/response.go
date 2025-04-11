package googlesdk

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/buger/jsonparser"
	"github.com/golang/glog"
	nativeResponse "github.com/prebid/openrtb/v20/native1/response"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

const videoClickThroughTagPath = "./VAST/Ad/InLine/Creatives/Creative/Linear/VideoClicks/ClickThrough"

func SetGoogleSDKResponseReject(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) bool {
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
	if rctx.Endpoint != models.EndpointGoogleSDK {
		return bidResponse
	}

	if rctx.Debug {
		if len(bidResponse.SeatBid) == 0 || len(bidResponse.SeatBid[0].Bid) == 0 {
			return bidResponse
		}
		if bidResponse.NBR != nil {
			return bidResponse
		}
	}

	if rctx.GoogleSDK.Reject {
		processingTimeValue := time.Since(time.Unix(rctx.StartTime, 0)).Milliseconds()
		ext := json.RawMessage([]byte(fmt.Sprintf(`{"%s":%d}`, models.ProcessingTime, processingTimeValue)))
		*bidResponse = openrtb2.BidResponse{
			ID:  bidResponse.ID,
			NBR: bidResponse.NBR,
			Ext: ext,
		}
		return bidResponse
	}

	bids, ok := customizeBid(rctx, bidResponse)
	if !ok {
		return bidResponse
	}

	*bidResponse = openrtb2.BidResponse{
		ID:    bidResponse.ID,
		BidID: bidResponse.SeatBid[0].Bid[0].ID,
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

	bid := bidResponse.SeatBid[0].Bid[0]
	bidExt := models.GoogleSDKBidExt{
		SDKRenderedAd: models.SDKRenderedAd{
			ID:            rctx.GoogleSDK.SDKRenderedAdID,
			RenderingData: string(resp),
			DeclaredAd:    getDeclaredAd(rctx, bid),
		},
		//EventNotificationToken: &models.EventNotificationToken{Payload: ""},
		BillingID: "",
	}
	bid.AdM = ""

	bid.Ext, err = json.Marshal(bidExt)
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

	if err := json.Unmarshal([]byte(bid.AdM), &nativeResp); err != nil {
		glog.Errorf("[googlesdk] native:[%s] error:[%s]", bid.AdM, err.Error())
	}
	declaredAd.NativeResponse = &nativeResp
	declaredAd.ClickThroughURL = nativeResp.Link.URL
	return declaredAd
}

func getBannerClickThroughURL(bid openrtb2.Bid) string {
	return ""
}

func getVideoClickThroughURL(bid openrtb2.Bid) string {
	videoCreative := strings.TrimSpace(bid.AdM)
	doc := etree.NewDocument()

	if err := doc.ReadFromString(videoCreative); err != nil {
		glog.Errorf("[googlesdk] video_creative:[%s] error:[%s]", videoCreative, err.Error())
		return ""
	}

	clickThrough := doc.Element.FindElement(videoClickThroughTagPath)
	if clickThrough == nil {
		return ""
	}
	return clickThrough.Text()
}

func SetSDKRenderedAdID(app *openrtb2.App, endpoint string) string {
	if endpoint != models.EndpointGoogleSDK || app == nil || app.Ext == nil {
		return ""
	}

	if sdkRenderedAdID, err := jsonparser.GetString(app.Ext, "installed_sdk", "id"); err == nil {
		return sdkRenderedAdID
	}
	return ""
}
