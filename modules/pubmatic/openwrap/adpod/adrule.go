package adpod

import (
	"errors"
	"fmt"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/ortb"
)

const (
	MaxAdrules = 30
)

const (
	ImpressionIDFormat = "%s-%s-%d"
)

func ApplyAdruleConfigs(rctx *models.RequestCtx, bidRequest *openrtb2.BidRequest) error {
	if !rctx.AdruleFlag {
		return nil
	}

	imps := make([]openrtb2.Imp, 0)
	for _, imp := range bidRequest.Imp {
		if imp.Video == nil {
			continue
		}

		impCtx, ok := rctx.ImpBidCtx[imp.ID]
		if !ok {
			continue
		}

		if impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig == nil || len(impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig.Adrule) == 0 {
			continue
		}

		if err := validateAdrule(impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig.Adrule); err != nil {
			return err
		}

		imps = append(imps, createImpressions(rctx, &impCtx, &imp, impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig.Adrule)...)
	}

	if len(imps) > 0 {
		bidRequest.Imp = imps
		// Create adpod Ctx
		if rctx.AdpodCtx == nil {
			rctx.AdpodCtx = make(map[string]models.AdpodConfig)
		}
		for _, imp := range imps {
			rctx.AdpodCtx.AddAdpodConfig(&imp)
		}
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

func createImpressions(rctx *models.RequestCtx, impCtx *models.ImpCtx, imp *openrtb2.Imp, adrule []*openrtb2.Video) []openrtb2.Imp {
	imps := make([]openrtb2.Imp, 0)
	for i, v := range adrule {
		impCopy := ortb.DeepCloneImpression(imp)
		impCopy.ID = fmt.Sprintf(ImpressionIDFormat, v.PodID, imp.ID, i)
		applyAdRule(impCopy.Video, v)

		impCtxCopy := impCtx.DeepCopy()
		impCtxCopy.ImpID = impCopy.ID
		impCtxCopy.Video = impCopy.Video
		rctx.ImpBidCtx[impCopy.ID] = impCtxCopy
		imps = append(imps, *impCopy)
	}
	return imps
}

func applyAdRule(video *openrtb2.Video, adrule *openrtb2.Video) {
	video.PodDur = adrule.PodDur
	video.MaxSeq = adrule.MaxSeq
	video.MaxDuration = adrule.MaxDuration
	video.MinDuration = adrule.MinDuration
	video.PodID = adrule.PodID
	video.StartDelay = adrule.StartDelay
	video.RqdDurs = adrule.RqdDurs
}
