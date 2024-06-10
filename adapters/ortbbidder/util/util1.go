package util

import (
	"github.com/PubMatic-OpenWrap/prebid-server/v2/openrtb_ext"
	"github.com/prebid/openrtb/v20/openrtb2"
)

// getMediaTypeForBidFromMType returns the bidType from the MarkupType field
func GetMType(mtype openrtb2.MarkupType) openrtb_ext.BidType { // change name
	var bidType openrtb_ext.BidType
	switch mtype {
	case openrtb2.MarkupBanner:
		bidType = openrtb_ext.BidTypeBanner
	case openrtb2.MarkupVideo:
		bidType = openrtb_ext.BidTypeVideo
	case openrtb2.MarkupAudio:
		bidType = openrtb_ext.BidTypeAudio
	case openrtb2.MarkupNative:
		bidType = openrtb_ext.BidTypeNative
	}
	return bidType
}
