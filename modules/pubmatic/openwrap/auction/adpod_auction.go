package auction

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	ctvlegacy "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adpod/legacy/auction"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adpod/auction"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func AdpodAuction(rCtx *models.RequestCtx, result *hookstage.HookResult[hookstage.AuctionResponsePayload], bidresponse *openrtb2.BidResponse) {
	if len(rCtx.AdpodCtx) == 0 || len(bidresponse.SeatBid) == 0 {
		return
	}

	for _, podConfig := range rCtx.AdpodCtx {
		switch podConfig.PodType {
		case models.PodTypeDynamic:
			errs := dynamicAdpodAuction(rCtx, podConfig, bidresponse)
			if len(errs) > 0 {
				for _, err := range errs {
					result.Errors = append(result.Errors, err.Error())
				}
			}
		case models.PodTypeStructured:
			errs := structuredAdpodAuction(rCtx, podConfig, bidresponse)
			if len(errs) > 0 {
				for _, err := range errs {
					result.Errors = append(result.Errors, err.Error())
				}
			}
		case models.PodTypeHybrid:
			errs := hybridAdpodAuction(rCtx, podConfig, bidresponse)
			if len(errs) > 0 {
				for _, err := range errs {
					result.Errors = append(result.Errors, err.Error())
				}
			}
		}
	}
}

func dynamicAdpodAuction(rCtx *models.RequestCtx, podConfig models.AdpodConfig, bidresponse *openrtb2.BidResponse) []error {
	// Legacy adpod auction
	errs := ctvlegacy.DynamicAdpodAuction(rCtx, bidresponse, podConfig)
	return errs
}

func structuredAdpodAuction(rCtx *models.RequestCtx, podConfig models.AdpodConfig, bidresponse *openrtb2.BidResponse) []error {
	errs := auction.StructuredAdpodAuction(rCtx, podConfig, bidresponse)
	return errs
}

func hybridAdpodAuction(rCtx *models.RequestCtx, podConfig models.AdpodConfig, bidresponse *openrtb2.BidResponse) []error {
	return nil
}
