package gocache

import (
	"github.com/pkg/errors"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

// validation check for Universal Pixels Object
func validUPixels(upixel []adunitconfig.UniversalPixel) []adunitconfig.UniversalPixel {

	var validPixels []adunitconfig.UniversalPixel
	for index, pixel := range upixel {
		if len(pixel.Pixel) == 0 {
			continue
		}
		if len(pixel.PixelType) == 0 || (pixel.PixelType != models.PixelTypeJS && pixel.PixelType != models.PixelTypeUrl) {
			continue
		}
		if pixel.Pos != "" && pixel.Pos != models.PixelPosAbove && pixel.Pos != models.PixelPosBelow {
			continue
		}
		validPixels = append(validPixels, upixel[index])
	}
	return validPixels
}

// wraps error with error msg
func errorWrap(cErr, nErr error) error {
	if cErr == nil {
		return nErr
	}

	return errors.Wrap(cErr, nErr.Error())
}
