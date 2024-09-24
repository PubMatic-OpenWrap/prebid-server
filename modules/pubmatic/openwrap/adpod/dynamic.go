package adpod

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/util"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/adpod/impressions"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/utils/ortb"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
)

const (
	ctvImpressionIDSeparator = "::"
)

type DynamicAdpod struct {
	models.AdpodCtx
	MinPodDuration       int64
	MaxPodDuration       int64
	MaxExtended          int64
	Imp                  openrtb2.Imp
	AdpodV25             *models.AdPod
	GeneratedSlotConfigs []models.GeneratedSlotConfig
	AdpodBid             *models.AdPodBid
	WinningBids          *models.AdPodBid
	Error                error
}

// TODO: Set exclusion config using request configs.
// TODO: Set ReqDurs config using request configs.
// TODO: Different execlusion config handling
func NewDynamicAdpod(podId string, imp openrtb2.Imp, impCtx models.ImpCtx, profileConfigs *models.AdpodProfileConfig, requestAdPodExt *models.ExtRequestAdPod) *DynamicAdpod {
	var (
		maxPodDuration int64
		adpodCfgV25    *models.AdPod
	)
	exclusion := getExclusionConfigs(podId, requestAdPodExt)
	video := impCtx.Video

	if video.PodDur > 0 {
		maxPodDuration = video.PodDur
		adpodCfgV25 = &models.AdPod{
			MinAds:                      1,
			MaxAds:                      int(video.MaxSeq),
			MinDuration:                 int(video.MinDuration),
			MaxDuration:                 int(video.MaxDuration),
			AdvertiserExclusionPercent:  ptrutil.ToPtr(0),
			IABCategoryExclusionPercent: ptrutil.ToPtr(0),
		}
		if exclusion.AdvertiserDomainExclusion {
			adpodCfgV25.AdvertiserExclusionPercent = ptrutil.ToPtr(100)
		}
		if exclusion.IABCategoryExclusion {
			adpodCfgV25.IABCategoryExclusionPercent = ptrutil.ToPtr(100)
		}
	} else {
		maxPodDuration = video.MaxDuration
		adpodCfgV25 = impCtx.AdpodConfig
	}

	return &DynamicAdpod{
		MinPodDuration: video.MinDuration,
		MaxPodDuration: maxPodDuration,
		AdpodCtx: models.AdpodCtx{
			PodId:          podId,
			Type:           models.Dynamic,
			ProfileConfigs: profileConfigs,
			Exclusion:      exclusion,
		},
		AdpodV25: adpodCfgV25,
		Imp:      imp,
	}
}

func (da *DynamicAdpod) GetPodType() models.PodType {
	return models.Dynamic
}

func (da *DynamicAdpod) AddImpressions(imp openrtb2.Imp) {
	da.Imps = append(da.Imps, imp)
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

/***************************** Dynamic adpod processing method ************************************/

func generateImpressionID(impID string, seqNo int) string {
	return fmt.Sprintf(impressions.ImpressionIDFormat, impID, seqNo)
}

// Function to retrieve the original impression ID and sequence number
func retrieveImpressionIDAndSeq(combinedID string) (string, int) {
	parts := strings.SplitN(combinedID, ctvImpressionIDSeparator, 2)
	if len(parts) != 2 {
		return combinedID, 0
	}

	seqNo, err := strconv.Atoi(parts[1])
	if err != nil {
		return parts[0], 0
	}

	return parts[0], seqNo
}

// getAdPodImpsConfigs will return number of impressions configurations within adpod
func (da *DynamicAdpod) getAdPodImpConfigs() error {
	selectedAlgorithm := impressions.SelectAlgorithm(da.AdpodV25, da.AdpodCtx.ProfileConfigs)
	impGen := impressions.NewImpressions(da.MinPodDuration, da.MaxPodDuration, da.AdpodV25, da.AdpodCtx.ProfileConfigs, selectedAlgorithm)
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

func (da *DynamicAdpod) CollectBid(bid *openrtb2.Bid, seat string) {
	originalImpId, sequence := retrieveImpressionIDAndSeq(bid.ImpID)

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

	duration, status := getBidDuration(bid, da.AdpodV25, da.AdpodCtx.ProfileConfigs, da.GeneratedSlotConfigs, sequence)

	da.AdpodBid.Bids = append(da.AdpodBid.Bids, &models.Bid{
		Bid:               bid,
		ExtBid:            ext,
		Status:            status,
		Duration:          int(duration),
		DealTierSatisfied: util.GetDealTierSatisfied(&ext),
		Seat:              string(models.BidderOWPrebidCTV),
	})
}

/*
getBidDuration determines the duration of video ad from given bid.
it will try to get the actual ad duration returned by the bidder using prebid.video.duration
if prebid.video.duration not present then uses defaultDuration passed as an argument
if video lengths matching policy is present for request then it will validate and update duration based on policy
*/
func getBidDuration(bid *openrtb2.Bid, adpodConfig *models.AdPod, adpodProfileCfg *models.AdpodProfileConfig, config []models.GeneratedSlotConfig, sequence int) (int64, int64) {

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
		return defaultDuration, models.StatusOK
	}

	// C2: Based on video lengths matching policy validate and return duration
	if adpodProfileCfg != nil && len(adpodProfileCfg.AdserverCreativeDurationMatchingPolicy) > 0 {
		return getDurationBasedOnDurationMatchingPolicy(duration, adpodProfileCfg.AdserverCreativeDurationMatchingPolicy, config)
	}

	//default return duration which is present in bid.ext.prebid.vide.duration field
	return duration, models.StatusOK
}

// getDurationBasedOnDurationMatchingPolicy will return duration based on durationmatching policy
func getDurationBasedOnDurationMatchingPolicy(duration int64, policy openrtb_ext.OWVideoAdDurationMatchingPolicy, config []models.GeneratedSlotConfig) (int64, int64) {
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
func GetNearestDuration(duration int64, config []models.GeneratedSlotConfig) int64 {
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
		da.Error = errors.New("prebid_ctv all bids filtered while matching lineitem duration")
		return
	}

	comb := NewCombination(
		buckets,
		uint64(da.MinPodDuration),
		uint64(da.MaxPodDuration),
		da.AdpodV25)

	//adpod generator
	adpodGenerator := NewAdPodGenerator(buckets, comb, da.AdpodV25)

	adpodBid := adpodGenerator.GetAdPodBids()
	if adpodBid == nil {
		da.Error = errors.New("prebid_ctv unable to generate adpod from bids combinations")
		return
	}
	adpodBid.OriginalImpID = da.AdpodBid.OriginalImpID
	adpodBid.SeatName = da.AdpodBid.SeatName

	da.WinningBids = adpodBid
}

func (da *DynamicAdpod) CollectAPRC(impCtxMap map[string]models.ImpCtx) {
	if len(da.AdpodBid.Bids) == 0 {
		return
	}
	impCtx := impCtxMap[da.AdpodBid.OriginalImpID]
	bidIdToAprc := make(map[string]int64)
	for _, bid := range da.AdpodBid.Bids {
		bidIdToAprc[bid.ID] = bid.Status
	}
	impCtx.BidIDToAPRC = bidIdToAprc
	impCtxMap[da.AdpodBid.OriginalImpID] = impCtx
}

func (da *DynamicAdpod) GetWinningBidsIds(impCtxMap map[string]models.ImpCtx, ImpToWinningBids map[string]map[string]bool) {
	if len(da.WinningBids.Bids) == 0 {
		return
	}
	impCtx := impCtxMap[da.AdpodBid.OriginalImpID]

	winningBids := make(map[string]bool)
	for _, bid := range da.WinningBids.Bids {
		winningBids[bid.ID] = true
		impCtx.BidIDToAPRC[bid.ID] = models.StatusWinningBid
	}
	ImpToWinningBids[da.AdpodBid.OriginalImpID] = winningBids
}

type BidsBuckets map[int][]*models.Bid

func GetDurationWiseBidsBucket(bids []*models.Bid) BidsBuckets {
	result := BidsBuckets{}

	for i, bid := range bids {
		if bid.Status == models.StatusOK {
			result[bid.Duration] = append(result[bid.Duration], bids[i])
		}
	}

	for k, v := range result {
		//sort.Slice(v[:], func(i, j int) bool { return v[i].Price > v[j].Price })
		sortBids(v)
		result[k] = v
	}

	return result
}

func sortBids(bids []*models.Bid) {
	sort.Slice(bids, func(i, j int) bool {
		if bids[i].DealTierSatisfied == bids[j].DealTierSatisfied {
			return bids[i].Price > bids[j].Price
		}
		return bids[i].DealTierSatisfied
	})
}
