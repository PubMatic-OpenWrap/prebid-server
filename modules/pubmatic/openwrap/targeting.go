package openwrap

import (
	"fmt"
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/exchange"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/utils"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
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

func addInAppTargettingKeys(targeting map[string]string, seat string, ecpm float64, bid *openrtb2.Bid, isWinningBid bool, priceGranularity *openrtb_ext.PriceGranularity) {
	targeting[models.CreatePartnerKey(seat, models.PWT_SLOTID)] = utils.GetOriginalBidId(bid.ID)
	targeting[models.CreatePartnerKey(seat, models.PWT_SZ)] = models.GetSize(bid.W, bid.H)
	targeting[models.CreatePartnerKey(seat, models.PWT_PARTNERID)] = seat
	targeting[models.CreatePartnerKey(seat, models.PWT_ECPM)] = fmt.Sprintf("%.2f", ecpm)
	targeting[models.CreatePartnerKey(seat, models.PWT_PLATFORM)] = getPlatformName(models.PLATFORM_APP)
	targeting[models.CreatePartnerKey(seat, models.PWT_BIDSTATUS)] = "1"
	var priceBucketKey string
	if priceGranularity != nil {
		priceBucketKey = exchange.GetPriceBucketOW(bid.Price, *priceGranularity)
	}
	targeting[models.CreatePartnerKey(seat, models.PwtPb)] = priceBucketKey
	if len(bid.DealID) != 0 {
		targeting[models.CreatePartnerKey(seat, models.PWT_DEALID)] = bid.DealID
	}

	if isWinningBid {
		targeting[models.PWT_SLOTID] = utils.GetOriginalBidId(bid.ID)
		targeting[models.PWT_BIDSTATUS] = "1"
		targeting[models.PWT_SZ] = models.GetSize(bid.W, bid.H)
		targeting[models.PWT_PARTNERID] = seat
		targeting[models.PWT_ECPM] = fmt.Sprintf("%.2f", ecpm)
		targeting[models.PWT_PLATFORM] = getPlatformName(models.PLATFORM_APP)
		if len(bid.DealID) != 0 {
			targeting[models.PWT_DEALID] = bid.DealID
		}
		targeting[models.PwtPb] = priceBucketKey
	}
}

func (m OpenWrap) addPWTTargetingForBid(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) (droppedBids map[string][]openrtb2.Bid, warnings []string) {
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
			if bidCtx.Prebid == nil {
				bidCtx.Prebid = new(openrtb_ext.ExtBidPrebid)
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

			if rctx.Platform == models.PLATFORM_APP {
				addInAppTargettingKeys(newTargeting, seatBid.Seat, bidCtx.NetECPM, &bid, isWinningBid, rctx.PriceGranularity)
			}
			for key, value := range rctx.CustomDimensions {
				//append cds key-val if sendToGAM is true or not present
				if value.SendToGAM == nil || (value.SendToGAM != nil && *value.SendToGAM) {
					newTargeting[key] = value.Value
				}
			}
			bidCtx.Prebid.Targeting = newTargeting

			if isWinningBid {
				if rctx.SendAllBids {
					bidCtx.Winner = 1
				}
				if m.pubFeatures.IsFscApplicable(rctx.PubID, seatBid.Seat, bidCtx.DspId) {
					bidCtx.Fsc = 1
				}
			} else if !rctx.SendAllBids {
				warnings = append(warnings, "dropping bid "+utils.GetOriginalBidId(bid.ID)+" as sendAllBids is disabled")
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
