package openrtb2

import (
	"encoding/json"
	"net/http"
	"runtime/debug"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/analytics"
	"github.com/prebid/prebid-server/v3/analytics/pubmatic"
	"github.com/prebid/prebid-server/v3/metrics"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

// recordRejectedBids records the rejected bids and respective rejection reason code
func recordRejectedBids(pubID string, seatNonBids []openrtb_ext.SeatNonBid, metricEngine metrics.MetricsEngine) {

	var found bool
	var codeLabel string
	reasonCodeMap := make(map[openrtb3.NoBidReason]string)

	for _, seatNonbid := range seatNonBids {
		for _, nonBid := range seatNonbid.NonBid {
			if codeLabel, found = reasonCodeMap[openrtb3.NoBidReason(nonBid.StatusCode)]; !found {
				codeLabel = strconv.FormatInt(int64(nonBid.StatusCode), 10)
				reasonCodeMap[openrtb3.NoBidReason(nonBid.StatusCode)] = codeLabel
			}
			metricEngine.RecordRejectedBids(pubID, seatNonbid.Seat, codeLabel)
		}
	}
}

func UpdateResponseExtOW(w http.ResponseWriter, bidResponse *openrtb2.BidResponse, ao analytics.AuctionObject) {
	defer func() {
		if r := recover(); r != nil {
			response, err := json.Marshal(bidResponse)
			if err != nil {
				glog.Error("response:" + string(response) + ". err: " + err.Error() + ". stacktrace:" + string(debug.Stack()))
				return
			}
			glog.Error("response:" + string(response) + ". stacktrace:" + string(debug.Stack()))
		}
	}()

	if bidResponse == nil {
		return
	}

	rCtx := pubmatic.GetRequestCtx(ao.HookExecutionOutcome)
	if rCtx == nil {
		return
	}

	//Send owlogger in response only in case of debug mode
	if rCtx.Debug && !rCtx.LoggerDisabled {
		var orignalMaxBidResponse *openrtb2.BidResponse
		if rCtx.Endpoint == models.EndpointAppLovinMax {
			orignalMaxBidResponse = new(openrtb2.BidResponse)
			*orignalMaxBidResponse = *bidResponse
			pubmatic.RestoreBidResponse(rCtx, ao)
		}

		owlogger, _ := pubmatic.GetLogAuctionObjectAsURL(ao, rCtx, false, true)
		if rCtx.Endpoint == models.EndpointAppLovinMax {
			*bidResponse = *orignalMaxBidResponse
		}
		if len(bidResponse.Ext) == 0 {
			bidResponse.Ext = []byte("{}")
		}
		if updatedExt, err := jsonparser.Set([]byte(bidResponse.Ext), []byte(strconv.Quote(owlogger)), "owlogger"); err == nil {
			bidResponse.Ext = updatedExt
		}
	} else if rCtx.Endpoint == models.EndpointAppLovinMax {
		bidResponse.Ext = nil
		if rCtx.AppLovinMax.Reject {
			w.WriteHeader(http.StatusNoContent)
		}
	}

	if rCtx.WakandaDebug.IsEnable() {
		rCtx.WakandaDebug.SetHTTPResponseWriter(w)
	}
}
