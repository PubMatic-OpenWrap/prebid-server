package exchange

import (
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/endpoints/events"
	"github.com/prebid/prebid-server/v2/metrics"
)

type OpenWrapEventTracking struct {
	enabledVideoEvents bool
	enableFastXML      bool
	me                 metrics.MetricsEngine
}

func (ev *eventTracking) injectVideoEvents(
	bidRequest *openrtb2.BidRequest,
	bid *openrtb2.Bid,
	vastXML, trackerURL, bidID, requestingBidder, bidderCoreName string) {

	if !ev.enabledVideoEvents {
		return
	}

	// always inject event  trackers without checkign isModifyingVASTXMLAllowed
	newVastXML, metrics, err := events.InjectVideoEventTrackers(bidRequest, bid, vastXML, trackerURL, bidID, requestingBidder, bidderCoreName, ev.auctionTimestampMs, ev.enableFastXML)
	if err == nil {
		bid.AdM = newVastXML
	}

	if metrics != nil && ev.me != nil {
		recordFastXMLMetrics(ev.me, "vast_events_injection", requestingBidder, metrics)
		if metrics.IsRespMismatch {
			glog.V(2).Infof("\n[XML_PARSER_TEST] method:[vcr] creative:[%s]", vastXML)
		}
	}
}
