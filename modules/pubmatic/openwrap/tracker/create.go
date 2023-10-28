package tracker

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/bidderparams"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func CreateTrackers(rctx models.RequestCtx, bidResponse *openrtb2.BidResponse) map[string]models.OWTracker {
	trackers := make(map[string]models.OWTracker)

	// pubmatic's KGP details per impression
	type pubmaticMarketplaceMeta struct {
		PubmaticKGP, PubmaticKGPV, PubmaticKGPSV string
	}
	pmMkt := make(map[string]pubmaticMarketplaceMeta)

	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			impId, _ := models.GetImpressionID(bid.ImpID)
			bidId := bid.ID
			tracker := models.Tracker{
				PubID:     rctx.PubID,
				ProfileID: fmt.Sprintf("%d", rctx.ProfileID),
				VersionID: fmt.Sprintf("%d", rctx.DisplayID),
				PageURL:   rctx.PageURL,
				Timestamp: rctx.StartTime,
				IID:       rctx.LoggerImpressionID,
				Platform:  int(rctx.DevicePlatform),
				SSAI:      rctx.SSAI,
				ImpID:     impId,
			}

			tagid := ""
			netECPM := float64(0)
			matchedSlot := ""
			price := bid.Price
			isRewardInventory := 0
			partnerID := seatBid.Seat
			bidType := "banner"
			var dspId int

			var isRegex bool
			var kgp, kgpv, kgpsv string

			if impCtx, ok := rctx.ImpBidCtx[impId]; ok {
				if bidderMeta, ok := impCtx.Bidders[seatBid.Seat]; ok {
					matchedSlot = bidderMeta.MatchedSlot
					partnerID = bidderMeta.PrebidBidderCode
				}

				if bidCtx, ok := impCtx.BidCtx[bidId]; ok {
					if bidResponse.Cur != "USD" {
						price = bidCtx.OriginalBidCPMUSD
					}
					netECPM = bidCtx.NetECPM

					// TODO do most calculation in wt
					// marketplace/alternatebiddercodes feature
					bidExt := bidCtx.BidExt
					if bidExt.Prebid != nil && bidExt.Prebid.Meta != nil && len(bidExt.Prebid.Meta.AdapterCode) != 0 && seatBid.Seat != bidExt.Prebid.Meta.AdapterCode {
						partnerID = bidExt.Prebid.Meta.AdapterCode

						if aliasSeat, ok := rctx.PrebidBidderCode[partnerID]; ok {
							if bidderMeta, ok := impCtx.Bidders[aliasSeat]; ok {
								matchedSlot = bidderMeta.MatchedSlot
							}
						}
					}
					bidType = bidCtx.CreativeType
					dspId = bidCtx.DspId
				}

				_ = matchedSlot
				// --------------------------------------------------------------------------------------------------
				// Move this code to a function. Confirm the kgp, kgpv, kgpsv relation in wt and wl.
				// --------------------------------------------------------------------------------------------------
				// var kgp, kgpv, kgpsv string

				if bidderMeta, ok := impCtx.Bidders[seatBid.Seat]; ok {
					partnerID = bidderMeta.PrebidBidderCode
					kgp = bidderMeta.KGP
					kgpv = bidderMeta.KGPV
					kgpsv = bidderMeta.MatchedSlot
					isRegex = bidderMeta.IsRegex
				}

				// 1. nobid
				if bid.Price == 0 && bid.H == 0 && bid.W == 0 {
					//NOTE: kgpsv = bidderMeta.MatchedSlot above. Use the same
					if !isRegex && kgpv != "" { // unmapped pubmatic's slot
						kgpsv = kgpv
					} else if !isRegex {
						kgpv = kgpsv
					}
				} else if !isRegex {
					if kgpv != "" { // unmapped pubmatic's slot
						kgpsv = kgpv
					} else if bid.H != 0 && bid.W != 0 { // Check when bid.H and bid.W will be zero with Price !=0. Ex: MobileInApp-MultiFormat-OnlyBannerMapping_Criteo_Partner_Validaton
						// 2. valid bid
						// kgpv has regex, do not generate slotName again
						// kgpsv could be unmapped or mapped slot, generate slotName again based on bid.H and bid.W
						kgpsv := bidderparams.GenerateSlotName(bid.H, bid.W, kgp, impCtx.TagID, impCtx.Div, rctx.Source)
						kgpv = kgpsv
					}
				}

				if kgpv == "" {
					kgpv = kgpsv
				}
				// --------------------------------------------------------------------------------------------------

				tagid = impCtx.TagID
				tracker.Secure = impCtx.Secure
				isRewardInventory = getRewardedInventoryFlag(rctx.ImpBidCtx[impId].IsRewardInventory)
			}

			if seatBid.Seat == "pubmatic" {
				pmMkt[impId] = pubmaticMarketplaceMeta{
					PubmaticKGP:   kgp,
					PubmaticKGPV:  kgpv,
					PubmaticKGPSV: kgpsv,
				}
			}

			tracker.Adunit = tagid
			tracker.SlotID = fmt.Sprintf("%s_%s", impId, tagid)
			tracker.RewardedInventory = isRewardInventory
			tracker.PartnerInfo = models.Partner{
				PartnerID:  partnerID,
				BidderCode: seatBid.Seat,
				BidID:      bidId,
				OrigBidID:  bid.ID,
				KGPV:       kgpv,
				NetECPM:    float64(netECPM),
				GrossECPM:  models.GetGrossEcpm(price),
			}

			if len(bid.ADomain) != 0 {
				if domain, err := models.ExtractDomain(bid.ADomain[0]); err == nil {
					tracker.Advertiser = domain
				}
			}

			var finalTrackerURL string
			trackerURL := ConstructTrackerURL(rctx, tracker)
			trackURL, err := url.Parse(trackerURL)
			if err == nil {
				trackURL.Scheme = models.HTTPSProtocol
				finalTrackerURL = trackURL.String()
			}

			trackers[bidId] = models.OWTracker{
				Tracker:       tracker,
				TrackerURL:    finalTrackerURL,
				Price:         price,
				PriceModel:    models.VideoPricingModelCPM,
				PriceCurrency: bidResponse.Cur,
				ErrorURL:      ConstructVideoErrorURL(rctx, rctx.VideoErrorTrackerEndpoint, bid, tracker),
				BidType:       bidType,
				DspId:         dspId,
			}
		}
	}

	// overwrite marketplace bid details with that of parent bidder
	for bidID, tracker := range trackers {
		if _, ok := rctx.MarketPlaceBidders[tracker.Tracker.PartnerInfo.BidderCode]; ok {
			if v, ok := pmMkt[tracker.Tracker.ImpID]; ok {
				tracker.Tracker.PartnerInfo.PartnerID = "pubmatic"
				tracker.Tracker.PartnerInfo.KGPV = v.PubmaticKGPV
			}
		}

		var finalTrackerURL string
		trackerURL := ConstructTrackerURL(rctx, tracker.Tracker)
		trackURL, err := url.Parse(trackerURL)
		if err == nil {
			trackURL.Scheme = models.HTTPSProtocol
			finalTrackerURL = trackURL.String()
		}
		tracker.TrackerURL = finalTrackerURL

		trackers[bidID] = tracker
	}

	return trackers
}

func getRewardedInventoryFlag(reward *int8) int {
	if reward != nil {
		return int(*reward)
	}
	return 0
}

// ConstructTrackerURL constructing tracker url for impression
func ConstructTrackerURL(rctx models.RequestCtx, tracker models.Tracker) string {
	trackerURL, err := url.Parse(rctx.TrackerEndpoint)
	if err != nil {
		return ""
	}

	v := url.Values{}
	v.Set(models.TRKPubID, strconv.Itoa(tracker.PubID))
	v.Set(models.TRKPageURL, tracker.PageURL)
	v.Set(models.TRKTimestamp, strconv.FormatInt(tracker.Timestamp, 10))
	v.Set(models.TRKIID, tracker.IID)
	v.Set(models.TRKProfileID, tracker.ProfileID)
	v.Set(models.TRKVersionID, tracker.VersionID)
	v.Set(models.TRKSlotID, tracker.SlotID)
	v.Set(models.TRKAdunit, tracker.Adunit)
	if tracker.RewardedInventory == 1 {
		v.Set(models.TRKRewardedInventory, strconv.Itoa(tracker.RewardedInventory))
	}
	partner := tracker.PartnerInfo
	v.Set(models.TRKPartnerID, partner.PartnerID)
	v.Set(models.TRKBidderCode, partner.BidderCode)
	v.Set(models.TRKKGPV, partner.KGPV)
	v.Set(models.TRKGrossECPM, fmt.Sprint(partner.GrossECPM))
	v.Set(models.TRKNetECPM, fmt.Sprint(partner.NetECPM))
	v.Set(models.TRKBidID, partner.BidID)
	if tracker.SSAI != "" {
		v.Set(models.TRKSSAI, tracker.SSAI)
	}
	v.Set(models.TRKOrigBidID, partner.OrigBidID)
	queryString := v.Encode()

	//Code for making tracker call http/https based on secure flag for in-app platform
	//TODO change platform to models.PLATFORM_APP once in-app platform starts populating from wrapper UI
	if rctx.Platform == models.PLATFORM_DISPLAY {
		if tracker.Secure == 1 {
			trackerURL.Scheme = "https"
		} else {
			trackerURL.Scheme = "http"
		}

	}
	trackerQueryStr := trackerURL.String() + models.TRKQMARK + queryString
	return trackerQueryStr
}

// ConstructVideoErrorURL constructing video error url for video impressions
func ConstructVideoErrorURL(rctx models.RequestCtx, errorURLString string, bid openrtb2.Bid, tracker models.Tracker) string {
	if len(errorURLString) == 0 {
		return ""
	}

	errorURL, err := url.Parse(errorURLString)
	if err != nil {
		return ""
	}

	errorURL.Scheme = models.HTTPSProtocol
	tracker.SURL = rctx.OriginCookie

	//operId Note: It should be first parameter in url otherwise it will get failed at analytics side.
	if len(errorURL.RawQuery) > 0 {
		errorURL.RawQuery = models.ERROperIDParam + models.TRKAmpersand + errorURL.RawQuery
	} else {
		errorURL.RawQuery = models.ERROperIDParam
	}

	v := url.Values{}
	v.Set(models.ERRPubID, strconv.Itoa(tracker.PubID))                  //pubId
	v.Set(models.ERRProfileID, tracker.ProfileID)                        //profileId
	v.Set(models.ERRVersionID, tracker.VersionID)                        //versionId
	v.Set(models.ERRTimestamp, strconv.FormatInt(tracker.Timestamp, 10)) //ts
	v.Set(models.ERRPartnerID, tracker.PartnerInfo.PartnerID)            //pid
	v.Set(models.ERRBidderCode, tracker.PartnerInfo.BidderCode)          //bc
	v.Set(models.ERRAdunit, tracker.Adunit)                              //au
	v.Set(models.ERRSUrl, tracker.SURL)                                  // sURL
	v.Set(models.ERRPlatform, strconv.Itoa(tracker.Platform))            // pfi
	v.Set(models.ERRAdvertiser, tracker.Advertiser)                      // adv

	if tracker.SSAI != "" {
		v.Set(models.ERRSSAI, tracker.SSAI) // ssai for video/json endpoint
	}

	if bid.CrID == "" {
		v.Set(models.ERRCreativeID, "-1")
	} else {
		v.Set(models.ERRCreativeID, bid.CrID) //creativeId
	}

	var out bytes.Buffer
	out.WriteString(errorURL.String())
	out.WriteString(models.TRKAmpersand)
	out.WriteString(v.Encode())
	out.WriteString(models.TRKAmpersand)
	out.WriteString(models.ERRErrorCodeParam) //ier

	//queryString +=
	errorURLQueryStr := out.String()

	return errorURLQueryStr
}
