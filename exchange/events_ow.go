package exchange

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/endpoints/events"
	"github.com/prebid/prebid-server/v3/metrics"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

type OpenWrapEventTracking struct {
	enabledVideoEvents bool
	me                 metrics.MetricsEngine
}

func (ev *eventTracking) injectVideoEvents(
	bidRequest *openrtb2.BidRequest,
	bid *openrtb2.Bid,
	vastXML, trackerURL, bidID, requestingBidder, bidderCoreName string) {

	if !ev.enabledVideoEvents {
		return
	}

	imp := openrtb_ext.GetImpressionID(bidRequest, bid.ImpID)
	if imp == nil || imp.Video == nil {
		return
	}

	eventURLMap := events.GetVideoEventTracking(bidRequest, imp, bid, trackerURL, bidID, requestingBidder, bidderCoreName, ev.auctionTimestampMs)
	if len(eventURLMap) == 0 {
		return
	}

	adm := strings.TrimSpace(bid.AdM)
	nurlPresent := (adm == "" || strings.HasPrefix(adm, "http"))
	eventInjector := events.GetXMLEventInjector(nurlPresent, imp.Video.Linearity)

	_startTime := time.Now()
	response, err := eventInjector.Inject(vastXML, eventURLMap)
	if err != nil {
		openrtb_ext.XMLLogf(openrtb_ext.XMLLogFormat, eventInjector.Name(), "vcr", base64.StdEncoding.EncodeToString([]byte(vastXML)))
		ev.me.RecordXMLParserError(eventInjector.Name(), "vcr", "")
		return
	}

	ev.me.RecordXMLParserProcessingTime(eventInjector.Name(), "vcr", "", time.Since(_startTime))
	bid.AdM = response
}
