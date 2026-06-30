package aps

import (
	"encoding/json"
	"fmt"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func logContext(rctx models.RequestCtx, reqID string) string {
	if reqID == "" {
		reqID = rctx.LoggerImpressionID
	}
	return fmt.Sprintf("reqid:[%s] pubid:[%s] profid:[%s] iid:[%s]",
		reqID, rctx.PubIDStr, rctx.ProfileIDStr, rctx.LoggerImpressionID)
}

func marshalForLog(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("<marshal error: %s>", err.Error())
	}
	return string(b)
}

func bidResponseSummary(bidResponse *openrtb2.BidResponse) string {
	if bidResponse == nil {
		return "bid_count:0 seatbid_count:0 nbr:none"
	}

	nbr := "none"
	if bidResponse.NBR != nil {
		nbr = fmt.Sprintf("%d", *bidResponse.NBR)
	}

	bidCount := 0
	for _, seatBid := range bidResponse.SeatBid {
		bidCount += len(seatBid.Bid)
	}

	return fmt.Sprintf("bid_count:%d seatbid_count:%d nbr:%s cur:%s",
		bidCount, len(bidResponse.SeatBid), nbr, bidResponse.Cur)
}

func nbrString(nbr *openrtb3.NoBidReason) string {
	if nbr == nil {
		return "none"
	}
	return fmt.Sprintf("%d", *nbr)
}

// LogModifiedRequest logs the APS request after entrypoint modifications.
func LogModifiedRequest(rctx models.RequestCtx, reqID, publisherID, profileID string, hasSignal bool, tagID string, requestBody []byte) {
	ctx := logContext(rctx, reqID)
	glog.Infof("[APS] stage:[request_modified] %s signal:[%t] publisher_id:[%s] profile_id:[%s] tagid:[%s] body:[%s]",
		ctx, hasSignal, publisherID, profileID, tagID, string(requestBody))
}

// LogSlotMappingFailed logs when APS slot UUID mapping fails at entrypoint.
func LogSlotMappingFailed(reqID, publisherID, errMsg string, requestBody []byte) {
	glog.Infof("[APS] stage:[slot_mapping_failed] reqid:[%s] pubid:[%s] err:[%s] body:[%s]",
		reqID, publisherID, errMsg, string(requestBody))
}

// LogAuctionResponse logs the auction response from Pubmatic before APS transforms.
func LogAuctionResponse(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) {
	glog.Infof("[APS] stage:[pubmatic_auction_response] %s summary:[%s] response:[%s]",
		logContext(rctx, bidResponse.ID), bidResponseSummary(bidResponse), marshalForLog(bidResponse))
}

// LogResponseRejectDecision logs whether the APS response will be rejected.
func LogResponseRejectDecision(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse, reject bool) {
	glog.Infof("[APS] stage:[response_reject_decision] %s reject:[%t] summary:[%s] nbr:[%s]",
		logContext(rctx, bidResponse.ID), reject, bidResponseSummary(bidResponse), nbrString(bidResponse.NBR))
}

// LogTransformedResponse logs the bid response after ApplyAPSResponse.
func LogTransformedResponse(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) {
	glog.Infof("[APS] stage:[response_after_transform] %s summary:[%s] response:[%s]",
		logContext(rctx, bidResponse.ID), bidResponseSummary(bidResponse), marshalForLog(bidResponse))
}

// LogFinalResponse logs the response returned to the APS client.
func LogFinalResponse(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse, reject bool) {
	glog.Infof("[APS] stage:[response_final] %s reject:[%t] summary:[%s] response:[%s]",
		logContext(rctx, bidResponse.ID), reject, bidResponseSummary(bidResponse), marshalForLog(bidResponse))
}
