package adpod

import (
	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
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
				seatBid.Bids[i].Bid.Dur = getBidDuration(seatBid.Bids[i].Bid, impCtx.ImpAdPodCfg, sequence)
			}
		}
	}
}

/*
getBidDuration determines the duration of video ad from given bid.
it will try to get the actual ad duration returned by the bidder using prebid.video.duration
if prebid.video.duration not present then uses defaultDuration passed as an argument
if video lengths matching policy is present for request then it will validate and update duration based on policy
*/
func getBidDuration(bid *openrtb2.Bid, config []*models.ImpAdPodConfig, sequence int) int64 {
	// C1: Read it from bid.ext.prebid.video.duration field
	duration, err := jsonparser.GetInt(bid.Ext, "prebid", "video", "duration")
	if err != nil || duration <= 0 {
		var defaultDuration int64
		for i := range config {
			if sequence == int(config[i].SequenceNumber) {
				defaultDuration = config[i].MaxDuration
			}
		}
		// incase if duration is not present use impression duration directly as it is
		return defaultDuration
	}

	//default return duration which is present in bid.ext.prebid.vide.duration field
	return duration
}
