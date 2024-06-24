package ctv

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

const (
	MaxAdrules = 30
)

func ApplyAdruleConfigs(rctx models.RequestCtx, bidRequest *openrtb2.BidRequest) (*openrtb2.BidRequest, error) {
	if bidRequest == nil || len(bidRequest.Imp) == 0 {
		return bidRequest, nil
	}

	var imps []openrtb2.Imp
	for _, imp := range bidRequest.Imp {
		if imp.Video == nil {
			imps = append(imps, imp)
			continue
		}

		adUnitCfg := rctx.ImpBidCtx[imp.ID].VideoAdUnitCtx.AppliedSlotAdUnitConfig
		if adUnitCfg == nil {
			imps = append(imps, imp)
			continue
		}

		if adUnitCfg.Adrule == nil {
			imps = append(imps, imp)
			continue
		}

		for i, rule := range adUnitCfg.Adrule {
			impCopy, _ := deepCloneImpression(imp)
			impCopy.ID = fmt.Sprintf("%s-%s-%d", impCopy.ID, rule.PodID, i)
			impCopy.Video.PodID = rule.PodID
			impCopy.Video.PodDur = rule.PodDur
			impCopy.Video.MaxSeq = rule.MaxSeq
			impCopy.Video.MinDuration = rule.MinDuration
			impCopy.Video.MaxDuration = rule.MaxDuration
			impCopy.Video.RqdDurs = rule.RqdDurs
			impCopy.Video.StartDelay = rule.StartDelay

			imps = append(imps, impCopy)
		}

	}

	bidRequest.Imp = imps
	return bidRequest, nil
}

func deepCloneImpression(imp openrtb2.Imp) (openrtb2.Imp, error) {
	data, err := json.Marshal(imp)
	if err != nil {
		return imp, err
	}
	var clonedImp openrtb2.Imp
	if err = json.Unmarshal(data, &clonedImp); err != nil {
		return imp, err
	}
	return clonedImp, nil
}

func validateAdrule(adrule []*openrtb2.Video) error {
	for _, v := range adrule {
		if v.MinDuration < 0 {
			return errors.New("invalid minduration configured in adrule")
		}
		if v.MaxDuration <= 0 {
			return errors.New("invalid maxduration configured in adrule")
		}
		if v.MinDuration > v.MaxDuration {
			return errors.New("minduration is greater than maxduration in adrule")
		}
	}

	return nil
}
