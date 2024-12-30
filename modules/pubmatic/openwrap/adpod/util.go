package adpod

import (
	"strconv"
	"strings"

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

func DecodeImpressionID(id string) (string, int) {
	index := strings.LastIndex(id, "::")
	if index == -1 {
		return id, 0
	}

	sequence, err := strconv.Atoi(id[index+2:])
	if nil != err || 0 == sequence {
		return id, 0
	}

	return id[:index], sequence
}
