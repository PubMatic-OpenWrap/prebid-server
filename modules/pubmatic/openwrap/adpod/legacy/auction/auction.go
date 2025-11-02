package ctvlegacy

import (
	"math"
	"sort"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

type Bid struct {
	*openrtb2.Bid
	openrtb_ext.ExtBid
	Duration          int
	Status            int64
	DealTierSatisfied bool
	Seat              string
}

type AdPodBid struct {
	Bids          []*Bid
	Price         float64
	Cat           []string
	ADomain       []string
	OriginalImpID string
	SeatName      string
}

// getDurationBasedOnDurationMatchingPolicy will return duration based on durationmatching policy
func getDurationBasedOnDurationMatchingPolicy(duration int64, policy openrtb_ext.OWVideoAdDurationMatchingPolicy, config []*models.ImpAdPodConfig) (int64, int64) {
	switch policy {
	case openrtb_ext.OWExactVideoAdDurationMatching:
		tmp := GetNearestDuration(duration, config)
		if tmp != duration {
			return duration, models.StatusDurationMismatch
		}
		//its and valid duration return it with StatusOK

	case openrtb_ext.OWRoundupVideoAdDurationMatching:
		tmp := GetNearestDuration(duration, config)
		if tmp == -1 {
			return duration, models.StatusDurationMismatch
		}
		//update duration with nearest one duration
		duration = tmp
		//its and valid duration return it with StatusOK
	}

	return duration, models.StatusOK
}

// GetDealTierSatisfied ...
func GetDealTierSatisfied(ext *openrtb_ext.ExtBid) bool {
	return ext != nil && ext.Prebid != nil && ext.Prebid.DealTierSatisfied
}

// GetNearestDuration will return nearest duration value present in ImpAdPodConfig objects
// it will return -1 if it doesn't found any match
func GetNearestDuration(duration int64, config []*models.ImpAdPodConfig) int64 {
	tmp := int64(-1)
	diff := int64(math.MaxInt64)
	for _, c := range config {
		tdiff := (c.MaxDuration - duration)
		if tdiff == 0 {
			tmp = c.MaxDuration
			break
		}
		if tdiff > 0 && tdiff <= diff {
			tmp = c.MaxDuration
			diff = tdiff
		}
	}
	return tmp
}

func DynamicAdpodAuction(rctx *models.RequestCtx, response *openrtb2.BidResponse, podId string, podConfig models.AdpodConfig) []error {
	impAdpodBids := getAdpodBid(rctx, response, podId, podConfig)
	winningAdpodBids, errs := doAuctionAndExclusion(impAdpodBids, podConfig)
	if len(errs) > 0 {
		return errs
	}

	// Record APRC for bids
	bidIdToAprcMap := getAprc(impAdpodBids)

	// Update winning bids and impctx
	updateWinningBids(rctx, podId, winningAdpodBids, bidIdToAprcMap)

	return nil
}

func getAdpodBid(rctx *models.RequestCtx, response *openrtb2.BidResponse, podId string, podConfig models.AdpodConfig) *AdPodBid {
	adpodBid := &AdPodBid{
		OriginalImpID: podId,
		SeatName:      string(models.BidderOWPrebidCTV),
	}
	for i := range response.SeatBid {
		seat := response.SeatBid[i]
		for j := range seat.Bid {
			bid := &seat.Bid[j]
			if bid.Price == 0 {
				continue
			}

			if bid.ImpID != podId {
				continue
			}

			eachImpCtx, ok := rctx.ImpBidCtx[bid.ImpID]
			if !ok {
				continue
			}

			bidExt := eachImpCtx.BidCtx[bid.ID].ExtBid

			duration, status := validateBidDuration(int64(bid.Dur), rctx.AdpodProfileConfig, eachImpCtx.ImpAdPodCfg)
			eachImpBid := Bid{
				Bid:               bid,
				ExtBid:            bidExt,
				Status:            status,
				Duration:          int(duration),
				DealTierSatisfied: GetDealTierSatisfied(&bidExt),
				Seat:              seat.Seat,
			}

			adpodBid.Bids = append(adpodBid.Bids, &eachImpBid)
		}
	}

	//Sort the adpod bids
	sort.Slice(adpodBid.Bids, func(i, j int) bool { return adpodBid.Bids[i].Price > adpodBid.Bids[j].Price })

	return adpodBid
}

func validateBidDuration(duration int64, adpodProfileCfg *models.AdpodProfileConfig, config []*models.ImpAdPodConfig) (int64, int64) {
	if adpodProfileCfg == nil || len(adpodProfileCfg.AdserverCreativeDurationMatchingPolicy) == 0 {
		return duration, models.StatusOK
	}

	return getDurationBasedOnDurationMatchingPolicy(duration, adpodProfileCfg.AdserverCreativeDurationMatchingPolicy, config)
}

func updateWinningBids(rctx *models.RequestCtx, podId string, winningAdpodBid *AdPodBid, bidIdToAprcMap map[string]int64) {
	impCtx, ok := rctx.ImpBidCtx[podId]
	if !ok {
		return
	}

	// Update winning bids
	var winningOwBids []*models.OwBid
	for _, bid := range winningAdpodBid.Bids {
		bidCtx := impCtx.BidCtx[bid.ID]
		owBid := &models.OwBid{
			ID:                   bid.ID,
			NetEcpm:              bidCtx.EN,
			BidDealTierSatisfied: bid.DealTierSatisfied,
		}
		winningOwBids = append(winningOwBids, owBid)
		bidIdToAprcMap[bid.ID] = models.StatusWinningBid
	}
	rctx.WinningBids[podId] = winningOwBids

	// update aprc in the impCtx
	impCtx.BidIDToAPRC = bidIdToAprcMap

	// Update NBR for the bids
	for bidId, aprc := range bidIdToAprcMap {
		bidCtx := impCtx.BidCtx[bidId]
		bidCtx.Nbr = ConvertAPRCToNBR(aprc)
		impCtx.BidCtx[bidId] = bidCtx
	}
	rctx.ImpBidCtx[podId] = impCtx
}
