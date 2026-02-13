package auction

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	ctvlegacy "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adpod/legacy/auction"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adpod/auction"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/stage"
)

func AdpodAuction(
	rCtx *models.RequestCtx,
	bidresponse *openrtb2.BidResponse,
	result stage.AuctionResponseResult,
) stage.AuctionResponseResult {
	if len(rCtx.AdpodCtx) == 0 || len(bidresponse.SeatBid) == 0 {
		return result
	}

	for _, podConfig := range rCtx.AdpodCtx {
		switch podConfig.PodType {
		case models.PodTypeDynamic:
			errs := dynamicAdpodAuction(rCtx, bidresponse, podConfig)
			if len(errs) > 0 {
				for _, err := range errs {
					result.Errors = append(result.Errors, err.Error())
				}
			}
		case models.PodTypeStructured:
			errs := structuredAdpodAuction(rCtx, bidresponse, podConfig)
			if len(errs) > 0 {
				for _, err := range errs {
					result.Errors = append(result.Errors, err.Error())
				}
			}
		case models.PodTypeHybrid:
			errs := hybridAdpodAuction(rCtx, bidresponse, podConfig)
			if len(errs) > 0 {
				for _, err := range errs {
					result.Errors = append(result.Errors, err.Error())
				}
			}
		}
	}

	return result
}

func dynamicAdpodAuction(
	rCtx *models.RequestCtx,
	bidresponse *openrtb2.BidResponse,
	podConfig models.AdpodConfig,
) []error {
	// Legacy adpod auction
	errs := ctvlegacy.DynamicAdpodAuction(rCtx, bidresponse, podConfig)
	return errs
}

func structuredAdpodAuction(
	rCtx *models.RequestCtx,
	bidresponse *openrtb2.BidResponse,
	podConfig models.AdpodConfig,
) []error {
	errs := auction.StructuredAdpodAuction(rCtx, bidresponse, podConfig)
	return errs
}

func hybridAdpodAuction(
	rCtx *models.RequestCtx,
	bidresponse *openrtb2.BidResponse,
	podConfig models.AdpodConfig,
) []error {
	return nil
}
