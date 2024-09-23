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

func (sa *StructuredAdpod) CollectBid(bid *openrtb2.Bid, seat string) {

}

func (sa *StructuredAdpod) HoldAuction() {
}

func (sa *StructuredAdpod) CollectAPRC(impCtxMap map[string]models.ImpCtx) {
	// if len(sa.AdpodBid.Bids) == 0 {
	// 	return
	// }
	// impCtx := impCtxMap[da.AdpodBid.OriginalImpID]
	// bidIdToAprc := make(map[string]int64)
	// for _, bid := range da.AdpodBid.Bids {
	// 	bidIdToAprc[bid.ID] = bid.Status
	// }
	// impCtx.BidIDToAPRC = bidIdToAprc
	// impCtxMap[da.AdpodBid.OriginalImpID] = impCtx
}

func (sa *StructuredAdpod) GetWinningBidsIds(impCtxMap map[string]models.ImpCtx, ImpToWinningBids map[string]map[string]bool) {
}
