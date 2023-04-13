package openrtb2

import (
	"encoding/json"
	"strconv"

	"github.com/prebid/openrtb/v17/openrtb3"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/analytics/openwrap"
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

func getLogInfo(requestExt, responseExt []byte, ao *analytics.AuctionObject) []byte {
	// isLogInfo, err := jsonparser.GetInt(requestExt, "wrapper", "loginfo")
	// if err == nil && isLogInfo == 1 {
	if rCtx := openwrap.GetRequestCtx(ao.HookExecutionOutcome); rCtx != nil && rCtx.LogInfoFlag == 1 {
		responseExtMap := make(map[string]interface{})
		if err := json.Unmarshal(responseExt, &responseExtMap); err == nil && responseExtMap["loginfo"] != nil {
			if logInfo, ok := responseExtMap["loginfo"].(map[string]interface{}); ok {
				logInfo["logger"] = openwrap.GetLogAuctionObjectAsURL(ao, true)
			}

			responseExtMap["seatnonbid"] = ao.RejectedBids

			responseExt, _ = json.Marshal(responseExtMap)
		}
		return responseExt
	}
	return responseExt
}
