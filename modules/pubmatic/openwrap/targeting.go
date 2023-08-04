package openwrap

import (
	"fmt"
	"strings"

	"github.com/PubMatic-OpenWrap/prebid-server/modules/pubmatic/openwrap/fullscreenclickability"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// whitelist of prebid targeting keys
var prebidTargetingKeysWhitelist = map[string]struct{}{
	string(openrtb_ext.HbpbConstantKey): {},
	models.HbBuyIdPubmaticConstantKey:   {},
	// OTT - 18 Deal priortization support
	// this key required to send deal prefix and priority
	string(openrtb_ext.HbCategoryDurationKey): {},
}

// check if prebid targeting keys are whitelisted
func allowTargetingKey(key string) bool {
	if _, ok := prebidTargetingKeysWhitelist[key]; ok {
		return true
	}
	return strings.HasPrefix(key, models.HbBuyIdPrefix)
}

func addPWTTargetingForBid(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) (droppedBids map[string][]openrtb2.Bid, warnings []string) {
	if rctx.Platform != models.PLATFORM_APP {
		return
	}

	if !rctx.SendAllBids {
		droppedBids = make(map[string][]openrtb2.Bid)
	}

	//setTargeting needs a seperate loop as final winner would be decided after all the bids are processed by auction
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			impCtx, ok := rctx.ImpBidCtx[bid.ImpID]
			if !ok {
				continue
			}

			isWinningBid := false
			if b, ok := rctx.WinningBids[bid.ImpID]; ok && b.ID == bid.ID {
				isWinningBid = true
			}

			if !(isWinningBid || rctx.SendAllBids) {
				droppedBids[seatBid.Seat] = append(droppedBids[seatBid.Seat], bid)
			}

			bidCtx, ok := impCtx.BidCtx[bid.ID]
			if !ok {
				continue
			}

			newTargeting := make(map[string]string)
			for key, value := range bidCtx.Prebid.Targeting {
				if allowTargetingKey(key) {
					updatedKey := key
					if strings.HasPrefix(key, models.PrebidTargetingKeyPrefix) {
						updatedKey = strings.Replace(key, models.PrebidTargetingKeyPrefix, models.OWTargetingKeyPrefix, 1)
					}
					newTargeting[updatedKey] = value
				}
				delete(bidCtx.Prebid.Targeting, key)
			}

			bidCtx.Prebid.Targeting = newTargeting
			bidCtx.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_SLOTID)] = bid.ID
			bidCtx.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_SZ)] = models.GetSize(bid.W, bid.H)
			bidCtx.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_PARTNERID)] = seatBid.Seat
			bidCtx.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_ECPM)] = fmt.Sprintf("%.2f", bidCtx.NetECPM)
			bidCtx.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_PLATFORM)] = getPlatformName(rctx.Platform)
			bidCtx.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_BIDSTATUS)] = "1"
			if len(bid.DealID) != 0 {
				bidCtx.Prebid.Targeting[models.CreatePartnerKey(seatBid.Seat, models.PWT_DEALID)] = bid.DealID
			}

			if isWinningBid {
				if rctx.SendAllBids {
					bidCtx.Winner = 1
				}
				if fullscreenclickability.IsFscApplicable(rctx.PubID, seatBid.Seat, bidCtx.DspId) {
					bidCtx.Fsc = 1
				}
				bidCtx.Prebid.Targeting[models.PWT_SLOTID] = bid.ID
				bidCtx.Prebid.Targeting[models.PWT_BIDSTATUS] = "1"
				bidCtx.Prebid.Targeting[models.PWT_SZ] = models.GetSize(bid.W, bid.H)
				bidCtx.Prebid.Targeting[models.PWT_PARTNERID] = seatBid.Seat
				bidCtx.Prebid.Targeting[models.PWT_ECPM] = fmt.Sprintf("%.2f", bidCtx.NetECPM)
				bidCtx.Prebid.Targeting[models.PWT_PLATFORM] = getPlatformName(rctx.Platform)
				if len(bid.DealID) != 0 {
					bidCtx.Prebid.Targeting[models.PWT_DEALID] = bid.DealID
				}
			} else if !rctx.SendAllBids {
				warnings = append(warnings, "dropping bid "+bid.ID+" as sendAllBids is disabled")
			}

			// cache for bid details for logger and tracker
			if impCtx.BidCtx == nil {
				impCtx.BidCtx = make(map[string]models.BidCtx)
			}
			impCtx.BidCtx[bid.ID] = bidCtx
			rctx.ImpBidCtx[bid.ImpID] = impCtx
		}
	}
	return
}
