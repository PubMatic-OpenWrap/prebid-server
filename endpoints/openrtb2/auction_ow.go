package openrtb2

import (
	"strconv"

	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/metrics"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// recordRejectedBids records the rejected bids and respective rejection reason code
func recordRejectedBids(pubID string, seatNonBids []openrtb_ext.SeatNonBid, metricEngine metrics.MetricsEngine) {

	var found bool
	var codeLabel string
	reasonCodeMap := make(map[openrtb3.NonBidStatusCode]string)

	for _, seatNonbid := range seatNonBids {
		for _, nonBid := range seatNonbid.NonBid {
			if codeLabel, found = reasonCodeMap[openrtb3.NonBidStatusCode(nonBid.StatusCode)]; !found {
				codeLabel = strconv.FormatInt(int64(nonBid.StatusCode), 10)
				reasonCodeMap[openrtb3.NonBidStatusCode(nonBid.StatusCode)] = codeLabel
			}
			metricEngine.RecordRejectedBids(pubID, seatNonbid.Seat, codeLabel)
		}
	}
}
