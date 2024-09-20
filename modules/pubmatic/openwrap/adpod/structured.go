package adpod

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

type StructuredAdpod struct {
	models.AdpodCtx
	ImpBidMap          map[string][]*models.Bid
	WinningBid         map[string]*models.Bid
	SelectedCategories map[string]bool
	SelectedDomains    map[string]bool
}

func NewStructuredAdpod(podId string, impCtx models.ImpCtx, profileConfigs *models.AdpodProfileConfig) *StructuredAdpod {
	adpod := &StructuredAdpod{}
	return adpod
}

func (sa *StructuredAdpod) GetPodType() models.PodType {
	return models.Structured
}

func (sa *StructuredAdpod) AddImpressions(imp openrtb2.Imp) {
	sa.Imps = append(sa.Imps, imp)
}
