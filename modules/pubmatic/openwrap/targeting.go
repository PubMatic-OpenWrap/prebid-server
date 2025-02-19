package openwrap

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/exchange"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
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
	if len(bid.DealID) != 0 {
		targeting[models.CreatePartnerKey(seat, models.PWT_DEALID)] = bid.DealID
	}
	var priceBucket string
	if priceGranularity != nil {
		priceBucket = exchange.GetPriceBucketOW(bid.Price, *priceGranularity)
	}
	if priceBucket != "" {
		targeting[models.CreatePartnerKey(seat, models.PwtPb)] = priceBucket
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
		if priceBucket != "" {
			targeting[models.PwtPb] = priceBucket
		}
	}
}

func (m OpenWrap) addPWTTargetingForBid(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) (droppedBids map[string][]openrtb2.Bid, warnings []string) {
	if !rctx.SendAllBids {
		droppedBids = make(map[string][]openrtb2.Bid)
	}

	//setTargeting needs a seperate loop as final winner would be decided after all the bids are processed by auction
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			impId, _ := models.GetImpressionID(bid.ImpID)
			bidId := bid.ID

			impCtx, ok := rctx.ImpBidCtx[impId]
			if !ok {
				continue
			}

			isWinningBid := false
			if rctx.WinningBids.IsWinningBid(impId, bidId) {
				isWinningBid = true
			}

			if !(isWinningBid || rctx.SendAllBids) {
				droppedBids[seatBid.Seat] = append(droppedBids[seatBid.Seat], bid)
			}

			bidCtx, ok := impCtx.BidCtx[bidId]
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

			if rctx.IsCTVRequest && rctx.Endpoint == models.EndpointJson {
				if bidCtx.AdPod == nil {
					bidCtx.AdPod = &models.AdpodBidExt{}
				}
				if impCtx.AdpodConfig != nil {
					bidCtx.AdPod.IsAdpodBid = true
				}
				bidCtx.AdPod.Targeting = GetTargettingForAdpod(bid, rctx.PartnerConfigMap[models.VersionLevelConfigID], impCtx, bidCtx, seatBid.Seat)
				if rctx.Debug {
					bidCtx.AdPod.Debug.Targeting = GetTargettingForDebug(rctx, bid.ID, impCtx.TagID, bidCtx)
				}
			}

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
			impCtx.BidCtx[bidId] = bidCtx
			rctx.ImpBidCtx[impId] = impCtx
		}
	}
	return
}

func GetTargettingForDebug(rctx models.RequestCtx, bidID, tagID string, bidCtx models.BidCtx) map[string]string {
	targeting := make(map[string]string)

	targeting[models.PwtBidID] = utils.GetOriginalBidId(bidID)
	targeting[models.PWT_CACHE_PATH] = models.AMP_CACHE_PATH
	targeting[models.PWT_ECPM] = fmt.Sprintf("%.2f", bidCtx.NetECPM)
	targeting[models.PWT_PUBID] = rctx.PubIDStr
	targeting[models.PWT_SLOTID] = tagID
	targeting[models.PWT_PROFILEID] = rctx.ProfileIDStr

	if targeting[models.PWT_ECPM] == "" {
		targeting[models.PWT_ECPM] = "0"
	}

	versionID := fmt.Sprint(rctx.DisplayID)
	if versionID != "0" {
		targeting[models.PWT_VERSIONID] = versionID
	}

	for k, v := range bidCtx.Prebid.Targeting {
		targeting[k] = v
	}

	if !rctx.SupportDeals {
		delete(targeting, models.PwtPbCatDur)
	}

	return targeting
}

func GetTargettingForAdpod(bid openrtb2.Bid, partnerConfig map[string]string, impCtx models.ImpCtx, bidCtx models.BidCtx, seat string) map[string]string {
	targetingKeyValMap := make(map[string]string)
	targetingKeyValMap[models.PWT_PARTNERID] = seat

	if bidCtx.Prebid != nil {
		if bidCtx.Prebid.Video != nil && bidCtx.Prebid.Video.Duration > 0 {
			targetingKeyValMap[models.PWT_DURATION] = strconv.Itoa(bidCtx.Prebid.Video.Duration)
		}

		prefix, _, _, err := jsonparser.Get(impCtx.NewExt, "prebid", "bidder", seat, "dealtier", "prefix")
		if bidCtx.Prebid.DealTierSatisfied && partnerConfig[models.DealTierLineItemSetup] == "1" && err == nil && len(prefix) > 0 {
			targetingKeyValMap[models.PwtDT] = fmt.Sprintf("%s%d", string(prefix), bidCtx.Prebid.DealPriority)
		} else if len(bid.DealID) > 0 && partnerConfig[models.DealIDLineItemSetup] == "1" {
			targetingKeyValMap[models.PWT_DEALID] = bid.DealID
		} else {
			priceBucket, ok := bidCtx.Prebid.Targeting[models.PwtPb]
			if ok {
				targetingKeyValMap[models.PwtPb] = priceBucket
			}
		}

		catDur, ok := bidCtx.Prebid.Targeting[models.PwtPbCatDur]
		if ok {
			cat, dur := getCatAndDurFromPwtCatDur(catDur)
			if len(cat) > 0 {
				targetingKeyValMap[models.PwtCat] = cat
			}

			if len(dur) > 0 && targetingKeyValMap[models.PWT_DURATION] == "" {
				targetingKeyValMap[models.PWT_DURATION] = dur
			}
		}
	}

	return targetingKeyValMap
}

func getCatAndDurFromPwtCatDur(pwtCatDur string) (string, string) {
	arr := strings.Split(pwtCatDur, "_")
	if len(arr) == 2 {
		return "", TrimRightByte(arr[1], 's')
	}
	if len(arr) == 3 {
		return arr[1], TrimRightByte(arr[2], 's')
	}
	return "", ""
}

func TrimRightByte(s string, b byte) string {
	if s[len(s)-1] == b {
		return s[:len(s)-1]
	}
	return s
}
