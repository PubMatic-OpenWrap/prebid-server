package tracker

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
)

func InjectTrackers(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) (*openrtb2.BidResponse, error) {
	var errs error
	for i, seatBid := range bidResponse.SeatBid {
		for j, bid := range seatBid.Bid {
			var errMsg string
			var err error
			tracker := rctx.Trackers[bid.ID]
			adformat := tracker.BidType
			if rctx.Platform == models.PLATFORM_VIDEO {
				adformat = "video"
			}
			pixels := getUniversalPixels(rctx, adformat, seatBid.Seat)

			switch adformat {
			case models.Banner:
				bidResponse.SeatBid[i].Bid[j].AdM = injectBannerTracker(rctx, tracker, bid, seatBid.Seat, pixels)
			case models.Video:
				trackers := []models.OWTracker{tracker}
				bidResponse.SeatBid[i].Bid[j].AdM, err = injectVideoCreativeTrackers(bid, trackers)
			case models.Native:
				if impBidCtx, ok := rctx.ImpBidCtx[bid.ImpID]; ok {
					bidResponse.SeatBid[i].Bid[j].AdM, err = injectNativeCreativeTrackers(impBidCtx.Native, bid.AdM, tracker)
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
