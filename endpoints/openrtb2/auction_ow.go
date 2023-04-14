package openrtb2

import (
	"encoding/json"
	"strconv"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/openrtb/v17/openrtb3"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/analytics/openwrap"
	"github.com/prebid/prebid-server/metrics"
	"github.com/prebid/prebid-server/openrtb_ext"
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

func UpdateResponseExtOW(response *openrtb2.BidResponse, ao analytics.AuctionObject) {
	if response == nil {
		return
	}

	extBidResponse := openrtb_ext.ExtBidResponse{}
	if len(response.Ext) != 0 {
		if err := json.Unmarshal(response.Ext, &extBidResponse); err != nil {
			return
		}
	}

	rCtx := openwrap.GetRequestCtx(ao.HookExecutionOutcome)
	if rCtx == nil {
		return
	}

	if rCtx.LogInfoFlag == 1 {
		extBidResponse.LogInfo.Logger = openwrap.GetLogAuctionObjectAsURL(ao, true)
	}

	if seatNonBids := updateSeatNoBid(ao); len(seatNonBids) != 0 {
		if extBidResponse.Prebid == nil {
			extBidResponse.Prebid = &openrtb_ext.ExtResponsePrebid{}
		}
		extBidResponse.Prebid.SeatNonBid = seatNonBids
	}

	extBidResponse.Logger = openwrap.GetLogAuctionObjectAsURL(ao, false)

	response.Ext, _ = json.Marshal(extBidResponse)
}

func updateSeatNoBid(ao analytics.AuctionObject) []openrtb_ext.SeatNonBid {
	seatNonBids := make([]openrtb_ext.SeatNonBid, 0, len(ao.RejectedBids))

	seatNoBids := make(map[string][]analytics.RejectedBid)
	for _, rejectedBid := range ao.RejectedBids {
		seatNoBids[rejectedBid.Seat] = append(seatNoBids[rejectedBid.Seat], rejectedBid)
	}

	for seat, rejectedBids := range seatNoBids {
		extSeatNoBid := openrtb_ext.SeatNonBid{
			Seat:    seat,
			NonBids: make([]openrtb_ext.NonBid, 0, len(rejectedBids)),
		}

		for _, rejectedBid := range rejectedBids {
			extSeatNoBid.NonBids = append(extSeatNoBid.NonBids, openrtb_ext.NonBid{
				ImpId:      rejectedBid.Bid.Bid.ImpID,
				StatusCode: rejectedBid.RejectionReason,
				Ext: openrtb_ext.NonBidExt{
					Prebid: openrtb_ext.ExtResponseNonBidPrebid{
						Bid: openrtb_ext.Bid{
							Bid:            *rejectedBid.Bid.Bid,
							OriginalBidCPM: rejectedBid.Bid.OriginalBidCPM,
							OriginalBidCur: rejectedBid.Bid.OriginalBidCur,
						},
					},
				},
			})
		}

		seatNonBids = append(seatNonBids, extSeatNoBid)
	}

	return seatNonBids
}
