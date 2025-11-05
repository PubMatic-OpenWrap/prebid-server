package ctvjson

import (
	"slices"
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
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
