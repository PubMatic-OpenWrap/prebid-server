package adpod

import (
	"github.com/prebid/prebid-server/v3/exchange/entities"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

func ConvertUpTo26(rCtx models.RequestCtx, responses map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid) {
	for _, seatBid := range responses {
		for i := range seatBid.Bids {
			_, impId, sequence := utils.DecodeV25ImpID(seatBid.Bids[i].Bid.ImpID)
			seatBid.Bids[i].Bid.ImpID = impId

			impCtx, ok := rCtx.ImpBidCtx[impId]
			if !ok {
				continue
			}

			if seatBid.Bids[i].Bid.Dur == 0 {
				seatBid.Bids[i].Bid.Dur = getBidDuration(seatBid.Bids[i], impCtx, sequence)
			}
		}
	}
}

func getBidDuration(bid *entities.PbsOrtbBid, impCtx models.ImpCtx, sequence int) int64 {
	var duration int64
	if bid.BidVideo != nil {
		duration = int64(bid.BidVideo.Duration)
	}

	if duration <= 0 {
		var defaultDuration int64
		if impCtx.Video != nil {
			defaultDuration = int64(impCtx.Video.MaxDuration)
		}

		for i := range impCtx.ImpAdPodCfg {
			if sequence == int(impCtx.ImpAdPodCfg[i].SequenceNumber) {
				defaultDuration = impCtx.ImpAdPodCfg[i].MaxDuration
			}
		}
		return defaultDuration
	}

	return duration
}
