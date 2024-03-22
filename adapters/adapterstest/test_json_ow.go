package adapterstest

import (
	"strings"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// updateRequestInfoForOW updates the reqInfo as per OW requirement
func updateRequestInfoForOW(reqInfo adapters.ExtraRequestInfo, fileName string) adapters.ExtraRequestInfo {
	// for oRTB bidders (having prefix as 'ortb_', set the bidderName from file)
	if strings.HasPrefix(fileName, "ortb_") {
		files := strings.Split(fileName, "/")
		if len(files) > 0 {
			reqInfo.BidderCoreName = openrtb_ext.BidderName(files[0])
		}
	}
	return reqInfo
}
