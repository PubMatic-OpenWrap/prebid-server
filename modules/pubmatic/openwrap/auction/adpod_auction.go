package auction

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	ctvlegacy "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adpod/legacy/auction"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func AdpodAuction(rCtx *models.RequestCtx, result hookstage.HookResult[hookstage.AuctionResponsePayload], bidresponse *openrtb2.BidResponse) (hookstage.HookResult[hookstage.AuctionResponsePayload], bool) {
	if len(rCtx.AdpodCtx) == 0 || len(bidresponse.SeatBid) == 0 {
		return result, true
	}

	for podId, podConfig := range rCtx.AdpodCtx {
		switch podConfig.PodType {
		case models.PodTypeDynamic:
			errs := dynamicAdpodAuction(rCtx, podId, podConfig, bidresponse)
			if len(errs) > 0 {
				for _, err := range errs {
					result.Errors = append(result.Errors, err.Error())
				}
				return result, false
			}
		case models.PodTypeStructured:
			errs := structuredAdpodAuction(rCtx, podId, podConfig, bidresponse)
			if len(errs) > 0 {
				for _, err := range errs {
					result.Errors = append(result.Errors, err.Error())
				}
				return result, false
			}
		case models.PodTypeHybrid:
			errs := hybridAdpodAuction(rCtx, podId, podConfig, bidresponse)
			if len(errs) > 0 {
				for _, err := range errs {
					result.Errors = append(result.Errors, err.Error())
				}
				return result, false
			}
		}
	}

	return result, true
}

func dynamicAdpodAuction(rCtx *models.RequestCtx, podId string, podConfig models.AdpodConfig, bidresponse *openrtb2.BidResponse) []error {
	// Legacy adpod auction
	errs := ctvlegacy.DynamicAdpodAuction(rCtx, bidresponse, podId, podConfig)
	if len(errs) > 0 {
		return errs
	}

	return nil
}

func structuredAdpodAuction(rCtx *models.RequestCtx, podId string, podConfig models.AdpodConfig, bidresponse *openrtb2.BidResponse) []error {
	return nil
}

func hybridAdpodAuction(rCtx *models.RequestCtx, podId string, podConfig models.AdpodConfig, bidresponse *openrtb2.BidResponse) []error {
	return nil
}
