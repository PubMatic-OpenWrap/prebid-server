package adpod

import (
	"errors"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

const (
	MaxAdrules = 30
)

const (
	ImpressionIDFormat = "%s-%s-%d"
)

func checkV26AdpodConfigs(configs []models.PodConfig) bool {
	var podIdPresent bool
	for i := range configs {
		if len(configs[i].PodID) > 0 {
			podIdPresent = true
			break
		}
	}
	return podIdPresent
}

func getAdpodConfigsFromAdrule(adrules []*openrtb2.Video) []models.PodConfig {
	configs := make([]models.PodConfig, 0)

	for _, rule := range adrules {
		config := models.PodConfig{
			PodID:       rule.PodID,
			PodDur:      rule.PodDur,
			MaxSeq:      rule.MaxSeq,
			MinDuration: rule.MinDuration,
			MaxDuration: rule.MaxDuration,
			RqdDurs:     rule.RqdDurs,
			StartDelay:  rule.StartDelay,
		}
		configs = append(configs, config)
	}
	return configs
}

func ApplyAdruleAdpodConfigs(rctx *models.RequestCtx, bidRequest *openrtb2.BidRequest) error {
	if !rctx.AdruleFlag {
		return nil
	}

	for _, imp := range bidRequest.Imp {
		if imp.Video == nil {
			continue
		}

		impCtx, ok := rctx.ImpBidCtx[imp.ID]
		if !ok {
			continue
		}

		if impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig == nil || len(impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig.Adrule) == 0 {
			rctx.AdruleFlag = false
			continue
		}

		if err := validateAdrule(impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig.Adrule); err != nil {
			rctx.AdruleFlag = false
			return err
		}

		podConfigs, ok := rctx.ImpAdPodConfig[imp.ID]
		if !ok {
			rctx.ImpAdPodConfig[imp.ID] = getAdpodConfigsFromAdrule(impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig.Adrule)
			continue
		}

		if checkV26AdpodConfigs(podConfigs) {
			rctx.AdruleFlag = false
			continue
		}

		rctx.ImpAdPodConfig[imp.ID] = getAdpodConfigsFromAdrule(impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig.Adrule)
	}

	return nil
}

func validateAdrule(adrule []*openrtb2.Video) error {
	var totalAdRules int
	if len(adrule) > MaxAdrules {
		return errors.New("Number of adrules exceeds the limit")
	}

	for _, v := range adrule {
		totalAdRules += 1
		if v.MinDuration < 0 {
			return errors.New("Invalid Adrule MinDuration")
		}
		if v.MaxDuration <= 0 {
			return errors.New("Invalid Adrule MaxDuration")
		}
		if v.MinDuration > v.MaxDuration {
			return errors.New("Invalid Adrule Min and Max Duration")
		}
		if v.MaxSeq > 0 {
			totalAdRules += int(v.MaxSeq) - 1
		}
	}

	if totalAdRules > MaxAdrules {
		return errors.New("Number of adrules exceeds the limit")
	}

	return nil
}
