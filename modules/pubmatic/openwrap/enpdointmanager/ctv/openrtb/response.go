package ctvopenrtb

import (
	"encoding/base64"

	"github.com/gofrs/uuid"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/vastbuilder"
)

func formResponse(rCtx *models.RequestCtx, bidResponse *openrtb2.BidResponse) *openrtb2.BidResponse {
	if bidResponse == nil || bidResponse.SeatBid == nil {
		return nil
	}

	seatBids := make([]openrtb2.SeatBid, 0)
	bidsByPod := make(map[string][]openrtb2.Bid)
	for _, seatBid := range bidResponse.SeatBid {
		//Copy seatBid and reset its bids
		videoSeatBid := seatBid
		videoSeatBid.Bid = nil
		for _, bid := range seatBid.Bid {
			if len(bid.AdM) == 0 || bid.Price <= 0 {
				continue
			}

			if !rCtx.AdpodCtx.IsAdpodSlot(bid.ImpID) {
				videoSeatBid.Bid = append(videoSeatBid.Bid, bid)
			}

			impCtx, ok := rCtx.ImpBidCtx[bid.ImpID]
			if !ok {
				continue
			}

			podId := bid.ImpID
			if impCtx.Video != nil && impCtx.Video.PodID != "" {
				podId = impCtx.Video.PodID
			}

			bidsByPod[podId] = append(bidsByPod[podId], bid)
		}

		if len(videoSeatBid.Bid) > 0 {
			seatBids = append(seatBids, videoSeatBid)
		}
	}

	if len(seatBids) == 0 && len(bidsByPod) == 0 {
		bidResponse.NBR = openrtb3.NoBidUnknownError.Ptr()
		bidResponse.SeatBid = nil
		return bidResponse
	}

	seatBids = append(seatBids, getPrebidCTVSeatBid(bidsByPod)...)
	bidResponse.SeatBid = seatBids

	return bidResponse
}

func getPrebidCTVSeatBid(bidsMap map[string][]openrtb2.Bid) []openrtb2.SeatBid {
	seatBids := []openrtb2.SeatBid{}

	for podId, bids := range bidsMap {
		bid := openrtb2.Bid{
			ImpID: podId,
		}

		bidID, err := uuid.NewV4()
		if err == nil {
			bid.ID = bidID.String()
		} else {
			bid.ID = bids[0].ID
		}

		// Get Categories and ad domain
		builder := vastbuilder.GetVastBuilder()
		category := make(map[string]struct{})
		addomain := make(map[string]struct{})

		for _, eachBid := range bids {
			err := builder.Append(&eachBid)
			if err != nil {
				glog.Errorf("[CTV] type:[adpod_builder_append] parser:[%s] error:[%s] creative:[%s]", builder.Name(), err.Error(), base64.StdEncoding.EncodeToString([]byte(eachBid.AdM)))
				continue
			}

			bid.Price += eachBid.Price
			for _, cat := range eachBid.Cat {
				category[cat] = struct{}{}
			}

			for _, domain := range eachBid.ADomain {
				addomain[domain] = struct{}{}
			}
		}

		bid.AdM, err = builder.Build()
		if err != nil {
			glog.Errorf("[CTV] type:[adpod_builder_build] parser:[%s] error:[%s]", builder.Name(), err.Error())
			continue
		}

		bid.Cat = map2array(category)
		bid.ADomain = map2array(addomain)

		seatBid := openrtb2.SeatBid{}
		seatBid.Seat = models.BidderOWPrebidCTV
		seatBid.Bid = append(seatBid.Bid, bid)

		seatBids = append(seatBids, seatBid)
	}

	return seatBids
}

func map2array(m map[string]struct{}) []string {
	var result []string
	for key := range m {
		result = append(result, key)
	}
	return result[:]
}
