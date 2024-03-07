package openrtb2

import (
	"encoding/json"
	"runtime/debug"
	"strconv"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/analytics/pubmatic"
	"github.com/prebid/prebid-server/metrics"
	"github.com/prebid/prebid-server/openrtb_ext"
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

func UpdateResponseExtOW(bidResponse *openrtb2.BidResponse, ao analytics.AuctionObject) {
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

	extBidResponse := openrtb_ext.ExtBidResponse{}
	if len(bidResponse.Ext) != 0 {
		if err := json.Unmarshal(bidResponse.Ext, &extBidResponse); err != nil {
			return
		}
	}

	if rCtx.LogInfoFlag == 1 {
		extBidResponse.OwLogInfo.Logger, _ = pubmatic.GetLogAuctionObjectAsURL(ao, rCtx, true, true)
	}

	// TODO: uncomment after seatnonbid PR is merged https://github.com/prebid/prebid-server/pull/2505
	// if seatNonBids := updateSeatNoBid(rCtx, ao); len(seatNonBids) != 0 {
	// 	if extBidResponse.Prebid == nil {
	// 		extBidResponse.Prebid = &openrtb_ext.ExtResponsePrebid{}
	// 	}
	// 	extBidResponse.Prebid.SeatNonBid = seatNonBids
	// }

	if rCtx.Debug {
		extBidResponse.OwLogger, _ = pubmatic.GetLogAuctionObjectAsURL(ao, rCtx, false, true)
	}

	bidResponse.Ext, _ = json.Marshal(extBidResponse)
}

// TODO: uncomment after seatnonbid PR is merged https://github.com/prebid/prebid-server/pull/2505
// TODO: Move this to module once it gets []analytics.RejectedBid as param (submit it in vanilla)
// func updateSeatNoBid(rCtx *models.RequestCtx, ao analytics.AuctionObject) []openrtb_ext.SeatNonBid {
// 	seatNonBids := make([]openrtb_ext.SeatNonBid, 0, len(ao.RejectedBids))

// 	seatNoBids := make(map[string][]analytics.RejectedBid)
// 	for _, rejectedBid := range ao.RejectedBids {
// 		seatNoBids[rejectedBid.Seat] = append(seatNoBids[rejectedBid.Seat], rejectedBid)
// 	}

// 	for seat, rejectedBids := range seatNoBids {
// 		extSeatNoBid := openrtb_ext.SeatNonBid{
// 			Seat:    seat,
// 			NonBids: make([]openrtb_ext.NonBid, 0, len(rejectedBids)),
// 		}

// 		for _, rejectedBid := range rejectedBids {
// 			bid := *rejectedBid.Bid.Bid
// 			addClientConfig(rCtx, seat, &bid)
// 			extSeatNoBid.NonBids = append(extSeatNoBid.NonBids, openrtb_ext.NonBid{
// 				ImpId:      rejectedBid.Bid.Bid.ImpID,
// 				StatusCode: rejectedBid.RejectionReason,
// 				Ext: openrtb_ext.NonBidExt{
// 					Prebid: openrtb_ext.ExtResponseNonBidPrebid{
// 						Bid: openrtb_ext.Bid{
// 							Bid: bid,
// 						},
// 					},
// 				},
// 			})
// 		}

// 		seatNonBids = append(seatNonBids, extSeatNoBid)
// 	}

// 	return seatNonBids
// }

// func addClientConfig(rCtx *models.RequestCtx, seat string, bid *openrtb2.Bid) {
// 	if seatNoBidBySeat, ok := rCtx.NoSeatBids[bid.ImpID]; ok {
// 		if seatNoBids, ok := seatNoBidBySeat[seat]; ok {
// 			for _, seatNoBid := range seatNoBids {
// 				bidExt := models.BidExt{}
// 				if err := json.Unmarshal(seatNoBid.Ext, &bidExt); err != nil {
// 					continue
// 				}

// 				inBidExt := models.BidExt{}
// 				if err := json.Unmarshal(bid.Ext, &inBidExt); err != nil {
// 					continue
// 				}

// 				inBidExt.Banner = bidExt.Banner
// 				inBidExt.Video = bidExt.Video

// 				bid.Ext, _ = json.Marshal(inBidExt)
// 			}
// 		}
// 	}
// }
