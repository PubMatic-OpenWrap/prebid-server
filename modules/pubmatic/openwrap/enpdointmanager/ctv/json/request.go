package ctvjson

import (
	"slices"
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adpod"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/stage"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
)

func filterImpsWithInvalidAdserverURL(rCtx *models.RequestCtx, bidRequest *openrtb2.BidRequest) {
	var invalidImpIds []string
	for impId, impCtx := range rCtx.ImpBidCtx {
		if len(impCtx.AdserverURL) == 0 {
			continue
		}

		impCtx.AdserverURL = strings.TrimSpace(impCtx.AdserverURL)
		if !utils.IsValidURL(impCtx.AdserverURL) {
			invalidImpIds = append(invalidImpIds, impId)
		}
	}

	// Remove Invalid Imps
	for _, impId := range invalidImpIds {
		delete(rCtx.ImpBidCtx, impId)
	}

	validImps := make([]openrtb2.Imp, 0, len(rCtx.ImpBidCtx))
	for _, imp := range bidRequest.Imp {
		if slices.Contains(invalidImpIds, imp.ID) {
			continue
		}
		validImps = append(validImps, imp)
	}

	bidRequest.Imp = validImps
}

func processRedirectURL(rCtx *models.RequestCtx, result stage.BeforeValidationResult) (stage.BeforeValidationResult, bool) {
	if len(rCtx.RedirectURL) > 0 {
		rCtx.RedirectURL = strings.TrimSpace(rCtx.RedirectURL)
		if rCtx.ResponseFormat == models.ResponseFormatRedirect && !utils.IsValidURL(rCtx.RedirectURL) {
			result.NbrCode = int(nbr.InvalidRedirectURL)
			result.Errors = append(result.Errors, "Invalid redirect URL")
			return result, false
		}
	}

	if rCtx.ResponseFormat == models.ResponseFormatRedirect && len(rCtx.RedirectURL) == 0 {
		result.NbrCode = int(nbr.MissingOWRedirectURL)
		result.Errors = append(result.Errors, "owRedirectURL is missing")
		return result, false
	}

	return result, true
}

func updateAdpodConfigs(rCtx *models.RequestCtx, bidRequest *openrtb2.BidRequest) []error {
	var errs []error

	// Apply Adrule settings
	err := adpod.ApplyAdruleAdpodConfigs(rCtx, bidRequest)
	if err != nil {
		errs = append(errs, err)
	}

	return errs
}
