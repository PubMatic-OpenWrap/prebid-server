package utils

import (
	"regexp"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

var bidIdRegx = regexp.MustCompile("[" + models.BidIdSeparator + "]")

func GetOriginalBidId(bidId string) string {
	return bidIdRegx.Split(bidId, -1)[0]
}
