package adpod

import (
	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

func ConvertDownTo25(r *openrtb_ext.RequestWrapper) {
	var imps []*openrtb_ext.ImpWrapper
	for _, imp := range r.GetImp() {
		if imp.Video == nil {
			continue
		}

		if imp.Video.PodID == "" {
			continue
		}

		if imp.Video.PodDur > 0 || imp.Video.MaxSeq > 0 { //dynamic adpod or dynamic part of hybrid adpod
			imps = append(imps, convertDynamicAdPodDownTo25(imp)...)
		} else { // structured adpod or structured part of hybrid adpod
			imps = append(imps, convertStructuredAdPodDownTo25(imp)...)
		}
	}

	if len(imps) > 0 {
		r.SetImp(imps)
	}
}

func convertDynamicAdPodDownTo25(imp *openrtb_ext.ImpWrapper) []*openrtb_ext.ImpWrapper {
	var impressionCount int
	reqdDurs := imp.Video.RqdDurs

	if imp.Video.MaxSeq > 0 {
		impressionCount = int(imp.Video.MaxSeq)
	} else if len(reqdDurs) > 0 {
		impressionCount = int(len(reqdDurs))
	} else if imp.Video.MinDuration > 0 {
		impressionCount = int(imp.Video.PodDur / imp.Video.MinDuration)
	} else if imp.Video.MaxDuration > 0 {
		impressionCount = int(imp.Video.PodDur / imp.Video.MaxDuration)
	} else {
		impressionCount = 1
	}

	var imps []*openrtb_ext.ImpWrapper
	for i := 0; i < impressionCount; i++ {
		impCopy := imp.DeepClone()
		if l := len(reqdDurs); l > 0 {
			if i >= l {
				break
			}
			impCopy.Video.MinDuration = reqdDurs[i]
			impCopy.Video.MaxDuration = reqdDurs[i]
			impCopy.Video.RqdDurs = nil
		}
		impCopy.ID = utils.PopulateV25ImpID(impCopy.Video.PodID, imp.ID, i)
		impCopy.Video.PodID = ""
		impCopy.Video.SlotInPod = adcom1.SlotPositionInPod(0)
		impCopy.Video.PodDur = 0
		impCopy.Video.MaxSeq = 0
		imps = append(imps, impCopy)
	}
	return imps
}

func convertStructuredAdPodDownTo25(imp *openrtb_ext.ImpWrapper) []*openrtb_ext.ImpWrapper {
	if len(imp.Video.RqdDurs) == 0 {
		impCopy := imp.DeepClone()
		impCopy.ID = utils.PopulateV25ImpID(impCopy.Video.PodID, imp.ID, 0)
		impCopy.Video.PodID = ""
		impCopy.Video.SlotInPod = adcom1.SlotPositionInPod(0)
		return []*openrtb_ext.ImpWrapper{impCopy}
	}

	var imps []*openrtb_ext.ImpWrapper
	for i := range imp.Video.RqdDurs {
		impCopy := imp.DeepClone()
		impCopy.Video.MinDuration = imp.Video.RqdDurs[i]
		impCopy.Video.MaxDuration = imp.Video.RqdDurs[i]
		impCopy.Video.RqdDurs = nil
		impCopy.ID = utils.PopulateV25ImpID(impCopy.Video.PodID, imp.ID, i)
		impCopy.Video.PodID = ""
		impCopy.Video.SlotInPod = adcom1.SlotPositionInPod(0)
		imps = append(imps, impCopy)
	}
	return imps
}
