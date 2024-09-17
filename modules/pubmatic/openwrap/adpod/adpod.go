package adpod

import (
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

func GetPodType(impCtx models.ImpCtx) models.PodType {
	if impCtx.AdpodConfig != nil {
		return models.Dynamic
	}

	if len(impCtx.Video.PodID) > 0 && impCtx.Video.PodDur > 0 {
		return models.Dynamic
	}

	if len(impCtx.Video.PodID) > 0 {
		return models.Structured
	}

	return models.NotAdpod
}
