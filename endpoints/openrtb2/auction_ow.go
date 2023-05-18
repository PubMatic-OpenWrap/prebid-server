package openrtb2

import (
	"strconv"

	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/metrics"
)

// recordRejectedBids records the rejected bids and respective rejection reason code
func recordRejectedBids(pubID string, rejBids []analytics.RejectedBid, metricEngine metrics.MetricsEngine) {

	var found bool
	var codeLabel string
	reasonCodeMap := make(map[openrtb3.NonBidStatusCode]string)

	for _, bid := range rejBids {
		if codeLabel, found = reasonCodeMap[bid.RejectionReason]; !found {
			codeLabel = strconv.FormatInt(int64(bid.RejectionReason), 10)
			reasonCodeMap[bid.RejectionReason] = codeLabel
		}
		metricEngine.RecordRejectedBids(pubID, bid.Seat, codeLabel)
	}
}
