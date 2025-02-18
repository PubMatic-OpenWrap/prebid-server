package tracker

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
	"golang.org/x/exp/slices"
)

func InjectTrackers(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) (*openrtb2.BidResponse, error) {
	if rctx.TrackerDisabled {
		return bidResponse, nil
	}

	var errs error
	for i, seatBid := range bidResponse.SeatBid {
		for j, bid := range seatBid.Bid {
			var (
				errMsg string
				err    error
			)
			tracker := rctx.Trackers[bid.ID]
			adformat := tracker.BidType
			if rctx.Platform == models.PLATFORM_VIDEO {
				adformat = "video"
			}
			pixels := getUniversalPixels(rctx, adformat, seatBid.Seat)

			switch adformat {
			case models.Banner:
				bidResponse.SeatBid[i].Bid[j].AdM, bidResponse.SeatBid[i].Bid[j].BURL = injectBannerTracker(rctx, tracker, bid, seatBid.Seat, pixels)
				if tracker.IsOMEnabled {
					bidResponse.SeatBid[i].Bid[j].Ext, err = jsonparser.Set(bid.Ext, []byte(`1`), models.ImpCountingMethod)
				}
			case models.Video:
				trackers := []models.OWTracker{tracker}
				bidResponse.SeatBid[i].Bid[j].AdM, bidResponse.SeatBid[i].Bid[j].BURL, err = injectVideoCreativeTrackers(rctx, bid, trackers)
			case models.Native:
				if impBidCtx, ok := rctx.ImpBidCtx[bid.ImpID]; ok {
					bidResponse.SeatBid[i].Bid[j].AdM, bidResponse.SeatBid[i].Bid[j].BURL, err = injectNativeCreativeTrackers(impBidCtx.Native, bid, tracker, rctx.Endpoint)
				} else {
					errMsg = fmt.Sprintf("native obj not found for impid %s", bid.ImpID)
				}
			default:
				errMsg = fmt.Sprintf("Invalid adformat %s for bidid %s", adformat, bid.ID)
			}

			if err != nil {
				errMsg = fmt.Sprintf("failed to inject tracker for bidid %s with error %s", bid.ID, err.Error())
			}
			if errMsg != "" {
				rctx.MetricsEngine.RecordInjectTrackerErrorCount(adformat, rctx.PubIDStr, seatBid.Seat)
				errs = models.ErrorWrap(errs, errors.New(errMsg))
				glog.Errorf("[InjectTrackers] [PubID]: %d [ProfileID]: %d [Partner]: %v [Error]: %v", rctx.PubID, rctx.ProfileID, seatBid.Seat, errMsg)
			}

		}
	}
	return bidResponse, errs
}

func getUniversalPixels(rctx models.RequestCtx, adformat string, bidderCode string) []adunitconfig.UniversalPixel {
	var pixels, upixels []adunitconfig.UniversalPixel
	if rctx.AdUnitConfig != nil && rctx.AdUnitConfig.Config != nil {
		if defaultAdUnitConfig, ok := rctx.AdUnitConfig.Config[models.AdunitConfigDefaultKey]; ok && defaultAdUnitConfig != nil && len(defaultAdUnitConfig.UniversalPixel) != 0 {
			upixels = defaultAdUnitConfig.UniversalPixel
		}
	}
	for _, pixelVal := range upixels {
		if pixelVal.MediaType != adformat {
			continue
		}
		if len(pixelVal.Partners) > 0 && !slices.Contains(pixelVal.Partners, bidderCode) {
			continue
		}
		pixel := pixelVal.Pixel // for pixelType `js`
		if pixelVal.PixelType == models.PixelTypeUrl {
			pixel = strings.Replace(models.UniversalPixelMacroForUrl, "${pixelUrl}", pixelVal.Pixel, 1)
		}
		pixelVal.Pixel = pixel
		pixels = append(pixels, pixelVal)
	}
	return pixels
}

func getBURL(burl, trackerURL string) string {
	if trackerURL == "" {
		return burl
	}

	if burl == "" {
		return trackerURL
	}

	escapedBurl := url.QueryEscape(burl)
	return trackerURL + "&" + models.OwSspBurl + "=" + escapedBurl
}
