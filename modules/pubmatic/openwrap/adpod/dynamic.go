package adpod

import (
	"errors"
	"fmt"

	"github.com/PubMatic-OpenWrap/prebid-server/v2/util/ptrutil"
	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/adpod/impressions"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/utils/ortb"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
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
func NewDynamicAdpod(podId string, impCtx models.ImpCtx, profileConfigs *models.AdpodProfileConfig, requestAdPodExt *models.ExtRequestAdPod) *DynamicAdpod {
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
		AdpodV25:    adpodCfgV25,
		AdpodBid:    &models.AdPodBid{},
		WinningBids: &models.AdPodBid{},
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
