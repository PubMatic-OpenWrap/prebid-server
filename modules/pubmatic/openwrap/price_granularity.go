package openwrap

import (
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/util/ptrutil"
)

func computePriceGranularity(rctx models.RequestCtx) (openrtb_ext.PriceGranularity, error) {
	var priceGranularity string

	//Get the value of priceGranularity from config otherwise set "auto"
	if priceGranularity = models.GetVersionLevelPropertyFromPartnerConfig(rctx.PartnerConfigMap, models.PriceGranularityKey); priceGranularity == "" {
		priceGranularity = "auto"
	}

	//incase of test request
	if rctx.IsTestRequest > 0 {
		priceGranularity = "testpg"
	}

	//Get custom price granularity object
	if priceGranularity == models.PriceGranularityCustom {
		customPriceGranularityValue := models.GetVersionLevelPropertyFromPartnerConfig(rctx.PartnerConfigMap, models.PriceGranularityCustomConfig)
		pgObject, err := newCustomPriceGranuality(customPriceGranularityValue)
		return pgObject, err
	}

	//compute pg obj based on legacy string (auto, med/medium, low, high, dense, ow-ctv-med, testpg)
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
