package rtbbidder

import (
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type RTBAwareBidder struct {
	adapters.Bidder
}

// RTBAwareBidder wraps a bidder to change internal infoaware bidder
func BuildRTBAwareBidder(bidder adapters.Bidder, info config.BidderInfo) adapters.Bidder {
	return &RTBAwareBidder{
		Bidder: bidder,
	}
}

func (r *RTBAwareBidder) MakeRequests(request *openrtb2.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	// for each request, we prepare separate instance of reqInfo
	// reqInfo.BidderName --> magnite , magnite-1
	if value, ok := GetSyncer().AliasMap[string(reqInfo.BidderName)]; ok {
		reqInfo.BidderName = string(openrtb_ext.BidderName(value)) // magnite-1 --> magnite
	}
	infoAwareBidder := GetSyncer().InfoAwareBidders[string(reqInfo.BidderName)]
	return infoAwareBidder.MakeRequests(request, reqInfo)
}
