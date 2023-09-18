package openwrap

import (
	"github.com/PubMatic-OpenWrap/prebid-server/util/ptrutil"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func computePriceGranularity(rctx models.RequestCtx) (openrtb_ext.PriceGranularity, error) {
	//Get the value of priceGranularity from config
	priceGranularity := models.GetVersionLevelPropertyFromPartnerConfig(rctx.PartnerConfigMap, models.PriceGranularityKey)

	if priceGranularity == "" || (priceGranularity == models.PriceGranularityCustom && !rctx.IsCTVRequest) {
		// If it is empty then use default value as 'auto'
		// If it is custom but not CTV request then use default value as 'a
		priceGranularity = "auto"
	}
	if rctx.IsTestRequest > 0 && rctx.IsCTVRequest {
		//OTT-603: Adding test flag check
		priceGranularity = "testpg"
	}

	//  OTT-769: determine custom pg object based on customPriceGranularityValue config
	//  Expected that this check with be true iff platform is video / isCTVAPIRequest
	if priceGranularity == models.PriceGranularityCustom {
		customPriceGranularityValue := models.GetVersionLevelPropertyFromPartnerConfig(rctx.PartnerConfigMap, models.PriceGranularityCustomConfig)
		pgObject, err := newCustomPriceGranuality(customPriceGranularityValue)
		return pgObject, err
	}

	// OTT-769: (Backword compatibilty) compute based on legacy string (auto, med)
	pgObject, _ := openrtb_ext.NewPriceGranularityFromLegacyID(priceGranularity)

	return pgObject, nil
}

// newCustomPriceGranuality constructs the Custom PriceGranularity Object based on input
// customPGValue
// if pg ranges are not present inside customPGValue then this function by default
// returns Medium Price Granularity Object
// So, caller of this function must ensure that customPGValue has valid pg ranges
// Optimization (Not implemented) : we can think of - only do unmarshal once if haven't done before
func newCustomPriceGranuality(customPGValue string) (openrtb_ext.PriceGranularity, error) {
	// Assumptions
	// 1. customPriceGranularityValue will never be empty
	// 2. customPriceGranularityValue will not be legacy string viz. auto, dense
	// 3. ranges are specified inside customPriceGranularityValue
	pg := openrtb_ext.PriceGranularity{}
	err := pg.UnmarshalJSON([]byte(customPGValue))
	if err != nil {
		return pg, err
	}
	// Overwrite always to 2
	pg.Precision = ptrutil.ToPtr(2)
	return pg, nil
}
