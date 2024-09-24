package adpod

import (
	"errors"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/gofrs/uuid"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

func FormAdpodBidsAndPerformExclusion(rctx models.RequestCtx, response *openrtb2.BidResponse) (map[string]map[string]bool, []error) {
	var (
		errs             []error
		impToWinningBids = make(map[string]map[string]bool)
	)
	if len(response.SeatBid) == 0 {
		return nil, errs
	}
	generateAdpodBids(rctx, response.SeatBid)
	// adpodBids, errs := doAdPodExclusions(impAdpodBidsMap, rctx.ImpBidCtx)
	// if len(errs) > 0 {
	// 	return nil, errs
	// }

	for _, adpodCtx := range rctx.AdpodCtx {
		adpodCtx.HoldAuction()
		adpodCtx.CollectAPRC(rctx.ImpBidCtx)
		adpodCtx.GetWinningBidsIds(rctx.ImpBidCtx, impToWinningBids)
	}

	// Record APRC for bids
	// collectAPRC(impAdpodBidsMap, rctx.ImpBidCtx)

	return impToWinningBids, nil
}

// GetTargeting returns the value of targeting key associated with bidder
// it is expected that bid.Ext contains prebid.targeting map
// if value not present or any error occured empty value will be returned
// along with error.
func GetTargeting(key openrtb_ext.TargetingKey, bidder openrtb_ext.BidderName, bid openrtb2.Bid) (string, error) {
	bidderSpecificKey := key.BidderKey(openrtb_ext.BidderName(bidder), 20)
	return jsonparser.GetString(bid.Ext, "prebid", "targeting", bidderSpecificKey)
}

func addTargetingKey(bid *openrtb2.Bid, key openrtb_ext.TargetingKey, value string) error {
	if bid == nil {
		return errors.New("Invalid bid")
	}

	raw, err := jsonparser.Set(bid.Ext, []byte(strconv.Quote(value)), "prebid", "targeting", string(key))
	if err == nil {
		bid.Ext = raw
	}
	return err
}

func generateAdpodBids(rCtx models.RequestCtx, seatBids []openrtb2.SeatBid) {
	for i := range seatBids {
		seat := seatBids[i]
		videoBids := make([]openrtb2.Bid, 0)
		for j := range seat.Bid {
			bid := &seat.Bid[j]
			if len(bid.ID) == 0 {
				bidID, err := uuid.NewV4()
				if err != nil {
					continue
				}
				bid.ID = bidID.String()
			}

			if bid.Price == 0 {
				continue
			}

			impId, _ := models.GetImpressionID(bid.ImpID)
			_, ok := rCtx.ImpBidCtx[impId]
			if !ok {
				// Bid is rejected due to invalid imp id
				continue
			}

			value, err := GetTargeting(openrtb_ext.HbCategoryDurationKey, openrtb_ext.BidderName(seat.Seat), *bid)
			if err == nil {
				// ignore error
				addTargetingKey(bid, openrtb_ext.HbCategoryDurationKey, value)
			}

			value, err = GetTargeting(openrtb_ext.HbpbConstantKey, openrtb_ext.BidderName(seat.Seat), *bid)
			if err == nil {
				// ignore error
				addTargetingKey(bid, openrtb_ext.HbpbConstantKey, value)
			}
			podId, ok := rCtx.ImpToPodId[bid.ImpID]
			if !ok {
				videoBids = append(videoBids, *bid)
				continue
			}

			adpodCtx, ok := rCtx.AdpodCtx[podId]
			if !ok {
				continue
			}

			adpodCtx.CollectBid(bid, seat.Seat)

		}
		if len(videoBids) > 0 {
			// videoSeatBids = append(videoSeatBids, openrtb2.SeatBid{
			// 	Bid:   videoBids,
			// 	Seat:  seat.Seat,
			// 	Group: seat.Group,
			// 	Ext:   seat.Ext,
			// })
		}
	}
}

// /*
// getBidDuration determines the duration of video ad from given bid.
// it will try to get the actual ad duration returned by the bidder using prebid.video.duration
// if prebid.video.duration not present then uses defaultDuration passed as an argument
// if video lengths matching policy is present for request then it will validate and update duration based on policy
// */
// func getBidDuration(bid *openrtb2.Bid, adpodConfig models.AdPod, adpodProfileCfg *models.AdpodProfileConfig, config []*models.GeneratedSlotConfig, sequence int) (int64, int64) {

// 	// C1: Read it from bid.ext.prebid.video.duration field
// 	duration, err := jsonparser.GetInt(bid.Ext, "prebid", "video", "duration")
// 	if err != nil || duration <= 0 {
// 		var defaultDuration int64
// 		for i := range config {
// 			if sequence == int(config[i].SequenceNumber) {
// 				defaultDuration = config[i].MaxDuration
// 			}
// 		}
// 		// incase if duration is not present use impression duration directly as it is
// 		return defaultDuration, models.StatusOK
// 	}

// 	// C2: Based on video lengths matching policy validate and return duration
// 	if adpodProfileCfg != nil && len(adpodProfileCfg.AdserverCreativeDurationMatchingPolicy) > 0 {
// 		return getDurationBasedOnDurationMatchingPolicy(duration, adpodProfileCfg.AdserverCreativeDurationMatchingPolicy, config)
// 	}

// 	//default return duration which is present in bid.ext.prebid.vide.duration field
// 	return duration, models.StatusOK
// }

// // getDurationBasedOnDurationMatchingPolicy will return duration based on durationmatching policy
// func getDurationBasedOnDurationMatchingPolicy(duration int64, policy openrtb_ext.OWVideoAdDurationMatchingPolicy, config []*models.GeneratedSlotConfig) (int64, int64) {
// 	switch policy {
// 	case openrtb_ext.OWExactVideoAdDurationMatching:
// 		tmp := GetNearestDuration(duration, config)
// 		if tmp != duration {
// 			return duration, models.StatusDurationMismatch
// 		}
// 		//its and valid duration return it with StatusOK

// 	case openrtb_ext.OWRoundupVideoAdDurationMatching:
// 		tmp := GetNearestDuration(duration, config)
// 		if tmp == -1 {
// 			return duration, models.StatusDurationMismatch
// 		}
// 		//update duration with nearest one duration
// 		duration = tmp
// 		//its and valid duration return it with StatusOK
// 	}

// 	return duration, models.StatusOK
// }

// // GetDealTierSatisfied ...
// func GetDealTierSatisfied(ext *openrtb_ext.ExtBid) bool {
// 	return ext != nil && ext.Prebid != nil && ext.Prebid.DealTierSatisfied
// }

// // GetNearestDuration will return nearest duration value present in ImpAdPodConfig objects
// // it will return -1 if it doesn't found any match
// func GetNearestDuration(duration int64, config []*models.GeneratedSlotConfig) int64 {
// 	tmp := int64(-1)
// 	diff := int64(math.MaxInt64)
// 	for _, c := range config {
// 		tdiff := (c.MaxDuration - duration)
// 		if tdiff == 0 {
// 			tmp = c.MaxDuration
// 			break
// 		}
// 		if tdiff > 0 && tdiff <= diff {
// 			tmp = c.MaxDuration
// 			diff = tdiff
// 		}
// 	}
// 	return tmp
// }
