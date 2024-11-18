package adpod

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/constant"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/util"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/adpod/impressions"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/ortb"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
)

const (
	impressionIDFormat = `%v::%v`
)

type DynamicAdpod struct {
	models.AdpodCtx
	MinPodDuration       int64
	MaxPodDuration       int64
	MaxExtended          int64
	Imp                  openrtb2.Imp
	AdpodConfig          *models.AdPod
	GeneratedSlotConfigs []models.GeneratedSlotConfig
	AdpodBid             *models.AdPodBid
	WinningBids          *models.AdPodBid
	Error                error
}

func NewDynamicAdpod(podId string, imp openrtb2.Imp, impCtx models.ImpCtx, profileConfigs *models.AdpodProfileConfig, requestAdPodExt *models.ExtRequestAdPod) *DynamicAdpod {
	var (
		maxPodDuration int64
		adpodCfg       *models.AdPod
	)
	exclusion := getExclusionConfigs(podId, requestAdPodExt)
	video := impCtx.Video
	newProfileConfig := new(models.AdpodProfileConfig)
	*newProfileConfig = *profileConfigs
	if video.PodDur > 0 {
		maxPodDuration = video.PodDur
		adpodCfg = &models.AdPod{
			MinAds:                      1,
			MaxAds:                      int(video.MaxSeq),
			MinDuration:                 int(video.MinDuration),
			MaxDuration:                 int(video.MaxDuration),
			AdvertiserExclusionPercent:  ptrutil.ToPtr(100),
			IABCategoryExclusionPercent: ptrutil.ToPtr(100),
		}

		if len(video.RqdDurs) > 0 {
			durs := make([]int, 0)
			minDur := video.RqdDurs[0]
			maxDur := video.RqdDurs[0]
			for _, dur := range video.RqdDurs {
				if dur < minDur {
					minDur = dur
				}
				if dur > maxDur {
					maxDur = dur
				}
				durs = append(durs, int(dur))
			}
			adpodCfg.MinDuration = int(minDur)
			adpodCfg.MaxDuration = int(maxDur)
			newProfileConfig.AdserverCreativeDurationMatchingPolicy = openrtb_ext.OWExactVideoAdDurationMatching
			newProfileConfig.AdserverCreativeDurations = durs
		}
		// if exclusion.AdvertiserDomainExclusion {
		// 	adpodCfg.AdvertiserExclusionPercent = ptrutil.ToPtr(100)
		// }
		// if exclusion.IABCategoryExclusion {
		// 	adpodCfg.IABCategoryExclusionPercent = ptrutil.ToPtr(100)
		// }
	} else {
		maxPodDuration = video.MaxDuration
		adpodCfg = impCtx.AdpodConfig
	}

	return &DynamicAdpod{
		MinPodDuration: video.MinDuration,
		MaxPodDuration: maxPodDuration,
		AdpodCtx: models.AdpodCtx{
			PodId:          podId,
			Type:           models.Dynamic,
			ProfileConfigs: newProfileConfig,
			Exclusion:      exclusion,
		},
		AdpodConfig: adpodCfg,
		Imp:         imp,
	}
}

// Fix exclusion support
func getExclusionConfigs(podId string, adpodExt *models.ExtRequestAdPod) models.Exclusion {
	var exclusion models.Exclusion

	if adpodExt != nil && adpodExt.Exclusion != nil {
		var iabCategory, advertiserDomain bool
		for i := range adpodExt.Exclusion.IABCategory {
			if adpodExt.Exclusion.IABCategory[i] == podId {
				iabCategory = true
				break
			}
		}

		for i := range adpodExt.Exclusion.AdvertiserDomain {
			if adpodExt.Exclusion.AdvertiserDomain[i] == podId {
				advertiserDomain = true
				break
			}
		}

		exclusion.IABCategoryExclusion = iabCategory
		exclusion.AdvertiserDomainExclusion = advertiserDomain
	}

	return exclusion
}

func (da *DynamicAdpod) GetImpressions() []*openrtb_ext.ImpWrapper {
	err := da.getAdPodImpConfigs()
	if err != nil {
		da.Error = err
		return nil
	}

	var imps []*openrtb_ext.ImpWrapper
	for _, config := range da.GeneratedSlotConfigs {
		impCopy := ortb.DeepCloneImpression(&da.Imp)
		impCopy.ID = config.ImpID
		impCopy.Video.MinDuration = config.MinDuration
		impCopy.Video.MaxDuration = config.MaxDuration
		impCopy.Video.Sequence = config.SequenceNumber
		impCopy.Video.Ext = jsonparser.Delete(impCopy.Video.Ext, "adpod")
		impCopy.Video.Ext = jsonparser.Delete(impCopy.Video.Ext, "offset")
		if string(impCopy.Video.Ext) == "{}" {
			impCopy.Video.Ext = nil
		}
		imps = append(imps, &openrtb_ext.ImpWrapper{Imp: impCopy})
	}

	return imps
}

// getAdPodImpsConfigs will return number of impressions configurations within adpod
func (da *DynamicAdpod) getAdPodImpConfigs() error {
	selectedAlgorithm := impressions.SelectAlgorithm(da.AdpodConfig, da.AdpodCtx.ProfileConfigs)
	impGen := impressions.NewImpressions(da.MinPodDuration, da.MaxPodDuration, da.AdpodConfig, da.AdpodCtx.ProfileConfigs, selectedAlgorithm)
	impRanges := impGen.Get()

	// check if algorithm has generated impressions
	if len(impRanges) == 0 {
		return errors.New("unable to generate impressions for adpod for impression: " + da.Imp.ID)
	}

	config := make([]models.GeneratedSlotConfig, len(impRanges))
	for i, value := range impRanges {
		config[i] = models.GeneratedSlotConfig{
			ImpID:          generateImpressionID(da.Imp.ID, i+1),
			MinDuration:    value[0],
			MaxDuration:    value[1],
			SequenceNumber: int8(i + 1), /* Must be starting with 1 */
		}
	}

	da.GeneratedSlotConfigs = config
	return nil
}
func generateImpressionID(impID string, seqNo int) string {
	return fmt.Sprintf(impressionIDFormat, impID, seqNo)
}

func (da *DynamicAdpod) CollectBid(bid *openrtb2.Bid, seat string) {
	originalImpId, sequence := DecodeImpressionID(bid.ImpID)

	if da.AdpodBid == nil {
		da.AdpodBid = &models.AdPodBid{
			Bids:          make([]*models.Bid, 0),
			OriginalImpID: originalImpId,
			SeatName:      string(openrtb_ext.BidderOWPrebidCTV),
		}
	}

	ext := openrtb_ext.ExtBid{}
	if bid.Ext != nil {
		json.Unmarshal(bid.Ext, &ext)
	}

	//get duration of creative
	duration, status := getBidDuration(bid, da.ProfileConfigs, da.GeneratedSlotConfigs, da.GeneratedSlotConfigs[sequence-1].MaxDuration)

	da.AdpodBid.Bids = append(da.AdpodBid.Bids, &models.Bid{
		Bid:               bid,
		ExtBid:            ext,
		Status:            status,
		Duration:          int(duration),
		DealTierSatisfied: util.GetDealTierSatisfied(&ext),
		Seat:              seat,
	})
}

/*
getBidDuration determines the duration of video ad from given bid.
it will try to get the actual ad duration returned by the bidder using prebid.video.duration
if prebid.video.duration not present then uses defaultDuration passed as an argument
if video lengths matching policy is present for request then it will validate and update duration based on policy
*/
func getBidDuration(bid *openrtb2.Bid, profileConfigs *models.AdpodProfileConfig, config []models.GeneratedSlotConfig, defaultDuration int64) (int64, constant.BidStatus) {

	// C1: Read it from bid.ext.prebid.video.duration field
	duration, err := jsonparser.GetInt(bid.Ext, "prebid", "video", "duration")
	if nil != err || duration <= 0 {
		// incase if duration is not present use impression duration directly as it is
		return defaultDuration, constant.StatusOK
	}

	// C2: Based on video lengths matching policy validate and return duration
	if nil != profileConfigs && len(profileConfigs.AdserverCreativeDurations) > 0 {
		return getDurationBasedOnDurationMatchingPolicy(duration, profileConfigs.AdserverCreativeDurationMatchingPolicy, config)
	}

	//default return duration which is present in bid.ext.prebid.vide.duration field
	return duration, constant.StatusOK
}

// getDurationBasedOnDurationMatchingPolicy will return duration based on durationmatching policy
func getDurationBasedOnDurationMatchingPolicy(duration int64, policy openrtb_ext.OWVideoAdDurationMatchingPolicy, config []models.GeneratedSlotConfig) (int64, constant.BidStatus) {
	switch policy {
	case openrtb_ext.OWExactVideoAdDurationMatching:
		tmp := getNearestDuration(duration, config)
		if tmp != duration {
			return duration, constant.StatusDurationMismatch
		}
		//its and valid duration return it with StatusOK

	case openrtb_ext.OWRoundupVideoAdDurationMatching:
		tmp := getNearestDuration(duration, config)
		if tmp == -1 {
			return duration, constant.StatusDurationMismatch
		}
		//update duration with nearest one duration
		duration = tmp
		//its and valid duration return it with StatusOK
	}

	return duration, constant.StatusOK
}

// GetNearestDuration will return nearest duration value present in ImpAdPodConfig objects
// it will return -1 if it doesn't found any match
func getNearestDuration(duration int64, config []models.GeneratedSlotConfig) int64 {
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

func (da *DynamicAdpod) HoldAuction() {
	if da.AdpodBid == nil || len(da.AdpodBid.Bids) == 0 {
		return
	}

	// Check if we need sorting
	// sort.Slice(da.AdpodBid.Bids, func(i, j int) bool { return da.AdpodBid.Bids[i].Price > da.AdpodBid.Bids[j].Price })

	buckets := GetDurationWiseBidsBucket(da.AdpodBid.Bids)
	if len(buckets) == 0 {
		// da.Error = util.DurationMismatchWarning
		return
	}

	//combination generator
	//comb := combination.NewCombination(buckets, uint64(da.MinPodDuration), uint64(da.MaxPodDuration), da.VideoExt.AdPod)
	//combination generator
	comb := NewCombination(buckets, uint64(da.MinPodDuration), uint64(da.MaxPodDuration), da.AdpodConfig)
	//adpod generator
	adpodGenerator := NewAdPodGenerator(buckets, comb, da.AdpodConfig)
	adpodBid := adpodGenerator.GetAdPodBids()
	if adpodBid == nil {
		// da.Error = util.UnableToGenerateAdPodWarning
		return
	}
	adpodBid.OriginalImpID = da.AdpodBid.OriginalImpID
	adpodBid.SeatName = da.AdpodBid.SeatName

	da.WinningBids = adpodBid
}

func (da *DynamicAdpod) CollectAPRC(rctx models.RequestCtx) {
	if len(da.AdpodBid.Bids) == 0 {
		return
	}
	impCtx, ok := rctx.ImpBidCtx[da.AdpodBid.OriginalImpID]
	if !ok {
		return
	}
	bidIdToAprcMap := make(map[string]int64)
	for _, bid := range da.AdpodBid.Bids {
		bidIdToAprcMap[bid.ID] = bid.Status
	}

	impCtx.BidIDToAPRC = bidIdToAprcMap
	rctx.ImpBidCtx[da.AdpodBid.OriginalImpID] = impCtx
}

func (da *DynamicAdpod) GetWinningBidsIds(rctx models.RequestCtx, winningBidIds map[string][]string) {
	if len(da.WinningBids.Bids) == 0 {
		return
	}
	impCtx, ok := rctx.ImpBidCtx[da.AdpodBid.OriginalImpID]
	if !ok {
		return
	}
	for _, bid := range da.WinningBids.Bids {
		if len(bid.AdM) == 0 {
			continue
		}
		winningBidIds[da.AdpodBid.OriginalImpID] = append(winningBidIds[da.AdpodBid.OriginalImpID], bid.ID)
		impCtx.BidIDToAPRC[bid.ID] = models.StatusWinningBid
	}
	rctx.ImpBidCtx[da.AdpodBid.OriginalImpID] = impCtx
}
