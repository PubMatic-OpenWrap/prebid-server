package exchange

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/endpoints/events"
	"github.com/prebid/prebid-server/v3/metrics"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

type OpenWrapEventTracking struct {
	enabledVideoEvents       bool
	fastXMLEnabledPercentage int
	me                       metrics.MetricsEngine
}

func (ev *eventTracking) injectVideoEvents(
	bidRequest *openrtb2.BidRequest,
	bid *openrtb2.Bid,
	vastXML, trackerURL, bidID, requestingBidder, bidderCoreName string) {

	if !ev.enabledVideoEvents {
		return
	}

	// always inject event  trackers without checkign isModifyingVASTXMLAllowed
	newVastXML, metrics, err := events.InjectVideoEventTrackers(bidRequest, bid, vastXML, trackerURL, bidID, requestingBidder, bidderCoreName, ev.auctionTimestampMs, openrtb_ext.IsFastXMLEnabled(ev.fastXMLEnabledPercentage))
	if err == nil {
		bid.AdM = newVastXML
	}

	if metrics != nil && ev.me != nil {
		recordFastXMLMetrics(ev.me, "vcr", "0", metrics)
	}
}
