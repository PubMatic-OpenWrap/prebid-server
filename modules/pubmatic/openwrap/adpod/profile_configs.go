package adpod

import (
	"errors"
	"fmt"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

const MaxAdrules = 30

func ApplyAdpodProfileConfigs(rctx models.RequestCtx, bidrequest *openrtb2.BidRequest) (*openrtb2.BidRequest, error) {
	// Check for adrule
	if rctx.AdruleFlag {
		imps := make([]openrtb2.Imp, 0)
		for index, rule := range rctx.VideoConfigs {
			impCopy := bidrequest.Imp[0]
			impCopy.ID = fmt.Sprintf("%s-%s-%d", impCopy.ID, rule.PodID, index)
			applyAdRule(impCopy.Video, rule)
			imps = append(imps, impCopy)
		}

		bidrequest.Imp = imps
	}

	return bidrequest, nil
}

func applyAdRule(video *openrtb2.Video, adrule *openrtb2.Video) {
	video.PodDur = adrule.PodDur
	video.MaxSeq = adrule.MaxSeq
	video.MaxDuration = adrule.MaxDuration
	video.MinDuration = adrule.MinDuration
	video.PodID = adrule.PodID
	video.StartDelay = adrule.StartDelay
	// video.RqdDurs = adrule.RqdDurs

	return
}

func ValidateAdrule(adrule []*openrtb2.Video) error {
	var totalAdRules int
	if len(adrule) > MaxAdrules {
		return errors.New("Adrule count exceeds the limit")
	}
	for _, v := range adrule {
		totalAdRules += 1
		if v.MinDuration < 0 {
			return errors.New("Invalid MinDuration provided for adrule")
		}
		if v.MaxDuration <= 0 {
			return errors.New("Invalid Maxduration provided for adrule")
		}
		if v.MinDuration > v.MaxDuration {
			return errors.New("MinDuration should be less than MaxDuration for adrule")
		}
		if v.MaxSeq > 0 {
			totalAdRules += int(v.MaxSeq) - 1
		}
	}
	if totalAdRules > MaxAdrules {
		return errors.New("Adrule count exceeds the limit")
	}
	return nil
}
