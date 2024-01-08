package auction

import (
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/gofrs/uuid"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
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

func FormAdpodBidsAndPerformExclusion(response *openrtb2.BidResponse, rctx models.RequestCtx) (map[string][]string, []error) {
	var errs []error

	if len(response.SeatBid) == 0 {
		return nil, errs
	}

	impAdpodBidsMap, _ := generateAdpodBids(response.SeatBid, rctx.ImpBidCtx)
	adpodBids, errs := doAdPodExclusions(impAdpodBidsMap, rctx.ImpBidCtx)
	if len(errs) > 0 {
		return nil, errs
	}

	// Record APRC for bids
	collectAPRC(impAdpodBidsMap, rctx.ImpBidCtx)

	winningBidIds, err := GetWinningBidsIds(adpodBids, rctx.ImpBidCtx)
	if err != nil {
		return nil, []error{err}
	}

	return winningBidIds, nil
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

func generateAdpodBids(seatBids []openrtb2.SeatBid, impCtx map[string]models.ImpCtx) (map[string]*AdPodBid, []openrtb2.SeatBid) {
	impAdpodBidsMap := make(map[string]*AdPodBid)
	videoSeatBids := make([]openrtb2.SeatBid, 0)

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
				//filter invalid bids
				continue
			}

			impId, sequence := models.GetImpressionID(bid.ImpID)
			eachImpCtx, ok := impCtx[impId]
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
			if eachImpCtx.AdpodConfig == nil {
				videoBids = append(videoBids, *bid)
				continue
			}

			ext := openrtb_ext.ExtBid{}
			if bid.Ext != nil {
				json.Unmarshal(bid.Ext, &ext)
			}

			// if deps.cfg.GenerateBidID == false {
			// 	//making unique bid.id's per impression
			// 	bid.ID = util.GetUniqueBidID(bid.ID, len(impBids.Bids)+1)
			// }

			//get duration of creative
			duration, status := getBidDuration(bid, *eachImpCtx.AdpodConfig, eachImpCtx.ImpAdPodCfg, sequence)

			eachImpBid := Bid{
				Bid:               bid,
				ExtBid:            ext,
				Status:            status,
				Duration:          int(duration),
				DealTierSatisfied: GetDealTierSatisfied(&ext),
				Seat:              seat.Seat,
			}

			//Adding adpod bids
			impBids, ok := impAdpodBidsMap[impId]
			if !ok {
				impBids = &AdPodBid{
					OriginalImpID: impId,
					SeatName:      string(models.BidderOWPrebidCTV),
				}
				impAdpodBidsMap[impId] = impBids
			}

			impBids.Bids = append(impBids.Bids, &eachImpBid)

		}
		if len(videoBids) > 0 {
			videoSeatBids = append(videoSeatBids, openrtb2.SeatBid{
				Bid:   videoBids,
				Seat:  seat.Seat,
				Group: seat.Group,
				Ext:   seat.Ext,
			})
		}
	}

	//Sort the adpod bids
	for _, v := range impAdpodBidsMap {
		sort.Slice(v.Bids, func(i, j int) bool { return v.Bids[i].Price > v.Bids[j].Price })
	}

	return impAdpodBidsMap, videoSeatBids
}

/*
getBidDuration determines the duration of video ad from given bid.
it will try to get the actual ad duration returned by the bidder using prebid.video.duration
if prebid.video.duration not present then uses defaultDuration passed as an argument
if video lengths matching policy is present for request then it will validate and update duration based on policy
*/
func getBidDuration(bid *openrtb2.Bid, adpodConfig models.AdPod, config []*models.ImpAdPodConfig, sequence int) (int64, int64) {

	// C1: Read it from bid.ext.prebid.video.duration field
	duration, err := jsonparser.GetInt(bid.Ext, "prebid", "video", "duration")
	if nil != err || duration <= 0 {
		var defaultDuration int64
		for i := range config {
			if sequence == int(config[i].SequenceNumber) {
				defaultDuration = config[i].MaxDuration
			}
		}
		// incase if duration is not present use impression duration directly as it is
		return defaultDuration, models.StatusOK
	}

	// C2: Based on video lengths matching policy validate and return duration
	if len(adpodConfig.VideoAdDurationMatching) > 0 {
		return getDurationBasedOnDurationMatchingPolicy(duration, adpodConfig.VideoAdDurationMatching, config)
	}

	//default return duration which is present in bid.ext.prebid.vide.duration field
	return duration, models.StatusOK
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
