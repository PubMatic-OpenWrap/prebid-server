package googlesdk

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

// LogModifiedRequest logs the Google SDK request after entrypoint modifications.
func LogModifiedRequest(rctx models.RequestCtx, reqID string, wrapperData *wrapperData, hasSignal bool, requestBody []byte) {
	ctx := logContext(rctx, reqID)
	wrapper := "missing"
	if wrapperData != nil {
		wrapper = fmt.Sprintf("publisher_id:%s profile_id:%s tag_id:%s",
			wrapperData.PublisherId, wrapperData.ProfileId, wrapperData.TagId)
	}
	glog.Infof("[GoogleSDK] stage:[request_modified] %s signal:[%t] wrapper:[%s] body:[%s]",
		ctx, hasSignal, wrapper, string(requestBody))
}

// LogRequestMappingFailed logs when ad_unit_mapping extraction fails and the request is left unmodified.
func LogRequestMappingFailed(reqID string, err error, requestBody []byte) {
	glog.Infof("[GoogleSDK] stage:[request_mapping_failed] reqid:[%s] err:[%v] body:[%s]",
		reqID, err, string(requestBody))
}

// LogAuctionResponse logs the auction response from Pubmatic before Google SDK transforms.
func LogAuctionResponse(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) {
	glog.Infof("[GoogleSDK] stage:[pubmatic_auction_response] %s summary:[%s] response:[%s]",
		logContext(rctx, bidResponse.ID), bidResponseSummary(bidResponse), marshalForLog(bidResponse))
}

// LogResponseRejectDecision logs whether the Google SDK response will be rejected.
func LogResponseRejectDecision(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse, reject bool) {
	glog.Infof("[GoogleSDK] stage:[response_reject_decision] %s reject:[%t] summary:[%s] nbr:[%s]",
		logContext(rctx, bidResponse.ID), reject, bidResponseSummary(bidResponse), nbrString(bidResponse.NBR))
}

// LogTransformedResponse logs the bid response after ApplyGoogleSDKResponse.
func LogTransformedResponse(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) {
	glog.Infof("[GoogleSDK] stage:[response_after_transform] %s summary:[%s] response:[%s]",
		logContext(rctx, bidResponse.ID), bidResponseSummary(bidResponse), marshalForLog(bidResponse))
}

// LogFinalResponse logs the response returned to the Google SDK client after rejection stripping.
func LogFinalResponse(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse, stripped bool) {
	glog.Infof("[GoogleSDK] stage:[response_final] %s stripped:[%t] summary:[%s] response:[%s]",
		logContext(rctx, bidResponse.ID), stripped, bidResponseSummary(bidResponse), marshalForLog(bidResponse))
}

// LogCustomizeBidRejected logs when a bid is dropped during Google SDK response customization.
func LogCustomizeBidRejected(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse, reason string) {
	glog.Infof("[GoogleSDK] stage:[response_customize_rejected] %s reason:[%s] summary:[%s] nbr:[%s]",
		logContext(rctx, bidResponse.ID), reason, bidResponseSummary(bidResponse), nbrString(bidResponse.NBR))
}
