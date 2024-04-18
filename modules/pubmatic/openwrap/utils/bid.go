package utils

import (
	"regexp"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

var bidIDRegx = regexp.MustCompile("(" + models.BidIdSeparator + ")")

func GetOriginalBidId(bidID string) string {
	return bidIDRegx.Split(bidID, -1)[0]
}

func SetUniqueBidID(originalBidID, generatedBidID string) string {
	return originalBidID + models.BidIdSeparator + generatedBidID
}
