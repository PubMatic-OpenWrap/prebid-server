package floors

import (
	"errors"
	"math"
	"math/rand"
	"strings"

	"github.com/prebid/prebid-server/v3/config"
	"github.com/prebid/prebid-server/v3/currency"
	"github.com/prebid/prebid-server/v3/metrics"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
)

type Price struct {
	FloorMin    float64
	FloorMinCur string
}

const (
	defaultCurrency  string  = "USD"
	defaultDelimiter string  = "|"
	catchAll         string  = "*"
	skipRateMin      int     = 0
	skipRateMax      int     = 100
	modelWeightMax   int     = 100
	modelWeightMin   int     = 1
	enforceRateMin   int     = 0
	enforceRateMax   int     = 100
	dataRateMin      int     = 0
	dataRateMax      int     = 100
	floorPrecision   float64 = 0.01
)

const (
	fetchFailure     = "1"
	unmarshalFailure = "2"
	invalidFloors    = "3"
	floorsSkipped    = "4"
	zeroFloorValue   = "5"
	highFloorValue   = "6"
	setMaxFloorValue = "7"
	ZERO_FLOOR_VALUE = 0
	HIGH_FLOOR_VALUE = 200
)

// EnrichWithPriceFloors checks for floors enabled in account and request and selects floors data from dynamic fetched if present
// else selects floors data from req.ext.prebid.floors and update request with selected floors details
func EnrichWithPriceFloors(bidRequestWrapper *openrtb_ext.RequestWrapper, account config.Account, conversions currency.Conversions, priceFloorFetcher FloorFetcher, metricsEngine metrics.MetricsEngine) []error {
	if bidRequestWrapper == nil || bidRequestWrapper.BidRequest == nil {
		return []error{errors.New("Empty bidrequest")}
	}

	if !isPriceFloorsEnabled(account, bidRequestWrapper) {
		return []error{errors.New("Floors feature is disabled at account or in the request")}
	}
	floors, err := resolveFloors(account, bidRequestWrapper, conversions, priceFloorFetcher, metricsEngine)

	updateReqErrs := updateBidRequestWithFloors(floors, bidRequestWrapper, conversions, metricsEngine, account.ID)
	updateFloorsInRequest(bidRequestWrapper, floors)
	return append(err, updateReqErrs...)
}

// updateBidRequestWithFloors will update imp.bidfloor and imp.bidfloorcur based on rules matching
func updateBidRequestWithFloors(extFloorRules *openrtb_ext.PriceFloorRules, request *openrtb_ext.RequestWrapper, conversions currency.Conversions, metricEngine metrics.MetricsEngine, accountID string) []error {
	var (
		floorErrList []error
		floorVal     float64
	)

	if extFloorRules == nil || extFloorRules.Data == nil || len(extFloorRules.Data.ModelGroups) == 0 {
		return []error{}
	}

	modelGroup := extFloorRules.Data.ModelGroups[0]
	if modelGroup.Schema.Delimiter == "" {
		modelGroup.Schema.Delimiter = defaultDelimiter
	}

	extFloorRules.Skipped = new(bool)
	if shouldSkipFloors(modelGroup.SkipRate, extFloorRules.Data.SkipRate, extFloorRules.SkipRate, rand.Intn) {
		*extFloorRules.Skipped = true
		metricEngine.RecordFloorStatus(accountID, extFloorRules.PriceFloorLocation, floorsSkipped)
		return []error{}
	}

	floorErrList = validateFloorRulesAndLowerValidRuleKey(modelGroup.Schema, modelGroup.Schema.Delimiter, modelGroup.Values)
	if len(modelGroup.Values) > 0 {
		for _, imp := range request.GetImp() {
			desiredRuleKey := createRuleKey(modelGroup.Schema, request, imp)
			matchedRule, isRuleMatched := findRule(modelGroup.Values, modelGroup.Schema.Delimiter, desiredRuleKey)
			floorVal = modelGroup.Default
			if isRuleMatched {
				floorVal = modelGroup.Values[matchedRule]
			}

			// No rule is matched or no default value provided or non-zero bidfloor not provided
			if floorVal == 0.0 {
				continue
			}

			floorMinVal, floorCur, err := getMinFloorValue(extFloorRules, imp, conversions)
			if err == nil {
				floorVal = roundToFourDecimals(floorVal)
				bidFloor := floorVal
				if floorMinVal > 0.0 && floorVal < floorMinVal {
					bidFloor = floorMinVal
				}
				if bidFloor < ZERO_FLOOR_VALUE {
					metricEngine.RecordFloorStatus(accountID, extFloorRules.PriceFloorLocation, zeroFloorValue)
				}
				if bidFloor > HIGH_FLOOR_VALUE {
					metricEngine.RecordFloorStatus(accountID, extFloorRules.PriceFloorLocation, highFloorValue)
				}

				if extFloorRules.SetMaxFloor {
					floorVal, floorCur, floorLoc := GetMaxFloorValue(imp.BidFloor, imp.BidFloorCur, bidFloor, floorCur, conversions)
					imp.BidFloor = floorVal
					imp.BidFloorCur = floorCur
					if floorLoc != "" {
						extFloorRules.PriceFloorLocation = openrtb_ext.RequestLocation
						metricEngine.RecordFloorStatus(accountID, extFloorRules.PriceFloorLocation, setMaxFloorValue)
					}
				} else {
					imp.BidFloor = bidFloor
					imp.BidFloorCur = floorCur
				}

				err = updateImpExtWithFloorDetails(imp, matchedRule, floorVal, imp.BidFloor)
				if err != nil {
					floorErrList = append(floorErrList, err)
				}
			} else {
				floorErrList = append(floorErrList, err)
			}
		}
	}
	return floorErrList
}

// roundToFourDecimals retuns given value to 4 decimal points
func roundToFourDecimals(in float64) float64 {
	return math.Round(in*10000) / 10000
}

// isPriceFloorsEnabled check for floors are enabled at account and request level
func isPriceFloorsEnabled(account config.Account, bidRequestWrapper *openrtb_ext.RequestWrapper) bool {
	return isPriceFloorsEnabledForAccount(account) && isPriceFloorsEnabledForRequest(bidRequestWrapper)
}

// isPriceFloorsEnabledForAccount check for floors enabled flag in account config
func isPriceFloorsEnabledForAccount(account config.Account) bool {
	return account.PriceFloors.Enabled
}

// isPriceFloorsEnabledForRequest check for floors are enabled flag in request
func isPriceFloorsEnabledForRequest(bidRequestWrapper *openrtb_ext.RequestWrapper) bool {
	requestExt, err := bidRequestWrapper.GetRequestExt()
	if err == nil {
		if prebidExt := requestExt.GetPrebid(); prebidExt != nil && prebidExt.Floors != nil {
			return prebidExt.Floors.GetEnabled()
		}
	}
	return true
}

// useFetchedData will check if to use fetched data or request data
func useFetchedData(rate *int) bool {
	if rate == nil {
		return true
	}
	randomNumber := rand.Intn(dataRateMax)
	return randomNumber < *rate
}

// resolveFloors does selection of floors fields from request data and dynamic fetched data if dynamic fetch is enabled
func resolveFloors(account config.Account, bidRequestWrapper *openrtb_ext.RequestWrapper, conversions currency.Conversions, priceFloorFetcher FloorFetcher, metricsEngine metrics.MetricsEngine) (*openrtb_ext.PriceFloorRules, []error) {
	var (
		errList     []error
		floorRules  *openrtb_ext.PriceFloorRules
		fetchResult *openrtb_ext.PriceFloorRules
		fetchStatus string
	)

	reqFloor := extractFloorsFromRequest(bidRequestWrapper)
	if reqFloor != nil && reqFloor.Location != nil && len(reqFloor.Location.URL) > 0 {
		account.PriceFloors.Fetcher.URL = reqFloor.Location.URL
	}
	account.PriceFloors.Fetcher.AccountID = account.ID

	if priceFloorFetcher != nil && account.PriceFloors.UseDynamicData {
		fetchResult, fetchStatus = priceFloorFetcher.Fetch(account.PriceFloors)
	}

	if fetchResult != nil && fetchStatus == openrtb_ext.FetchSuccess && useFetchedData(fetchResult.Data.UseFetchDataRate) {
		mergedFloor := mergeFloors(reqFloor, fetchResult, conversions)
		floorRules, errList = createFloorsFrom(mergedFloor, account, fetchStatus, openrtb_ext.FetchLocation, metricsEngine)
	} else if reqFloor != nil {
		floorRules, errList = createFloorsFrom(reqFloor, account, fetchStatus, openrtb_ext.RequestLocation, metricsEngine)
	} else {
		floorRules, errList = createFloorsFrom(nil, account, fetchStatus, openrtb_ext.NoDataLocation, metricsEngine)
	}
	return floorRules, errList
}

// createFloorsFrom does preparation of floors data which shall be used for further processing
func createFloorsFrom(floors *openrtb_ext.PriceFloorRules, account config.Account, fetchStatus, floorLocation string, metricsEngine metrics.MetricsEngine) (*openrtb_ext.PriceFloorRules, []error) {
	var floorModelErrList []error
	finalFloors := &openrtb_ext.PriceFloorRules{
		FetchStatus:        fetchStatus,
		PriceFloorLocation: floorLocation,
	}

	if floors != nil {
		floorValidationErr := validateFloorParams(floors)
		if floorValidationErr != nil {
			metricsEngine.RecordFloorStatus(account.ID, floorLocation, invalidFloors)
			return finalFloors, append(floorModelErrList, floorValidationErr)
		}

		finalFloors.Enforcement = floors.Enforcement
		if floors.Data != nil {
			validModelGroups, floorModelErrList := selectValidFloorModelGroups(floors.Data.ModelGroups, account)
			if len(floorModelErrList) > 0 {
				metricsEngine.RecordFloorStatus(account.ID, floorLocation, invalidFloors)
			}
			if len(validModelGroups) == 0 {
				return finalFloors, floorModelErrList
			} else {
				*finalFloors = *floors
				finalFloors.Data = new(openrtb_ext.PriceFloorData)
				*finalFloors.Data = *floors.Data
				finalFloors.PriceFloorLocation = floorLocation
				finalFloors.FetchStatus = fetchStatus
				if len(validModelGroups) > 1 {
					validModelGroups = selectFloorModelGroup(validModelGroups, rand.Intn)
				}
				finalFloors.Data.ModelGroups = []openrtb_ext.PriceFloorModelGroup{validModelGroups[0].Copy()}
			}
		}
	}

	if floorLocation == openrtb_ext.RequestLocation && finalFloors.Data == nil {
		finalFloors.PriceFloorLocation = openrtb_ext.NoDataLocation
	}

	return finalFloors, floorModelErrList
}

// resolveEnforcement does retrieval of enforceRate from request
func resolveEnforcement(enforcement *openrtb_ext.PriceFloorEnforcement, enforceRate int) *openrtb_ext.PriceFloorEnforcement {
	if enforcement == nil {
		enforcement = new(openrtb_ext.PriceFloorEnforcement)
	}
	enforcement.EnforceRate = enforceRate
	return enforcement
}

// getFloorsEnabledFlag gets floors enabled flag from request
func getFloorsEnabledFlag(reqFloors openrtb_ext.PriceFloorRules) bool {
	if reqFloors.Enabled != nil {
		return *reqFloors.Enabled
	}
	return true
}

// shouldUseDynamicFetchedFloor gets UseDynamicData flag from account level config
func shouldUseDynamicFetchedFloor(Account config.Account) bool {
	return Account.PriceFloors.UseDynamicData
}

// extractFloorsFromRequest gets floors data from req.ext.prebid.floors
func extractFloorsFromRequest(bidRequestWrapper *openrtb_ext.RequestWrapper) *openrtb_ext.PriceFloorRules {
	requestExt, err := bidRequestWrapper.GetRequestExt()
	if err == nil {
		prebidExt := requestExt.GetPrebid()
		if prebidExt != nil && prebidExt.Floors != nil {
			return prebidExt.Floors
		}
	}
	return nil
}

// updateFloorsInRequest updates req.ext.prebid.floors with floors data
func updateFloorsInRequest(bidRequestWrapper *openrtb_ext.RequestWrapper, priceFloors *openrtb_ext.PriceFloorRules) {
	requestExt, err := bidRequestWrapper.GetRequestExt()
	if err == nil {
		prebidExt := requestExt.GetPrebid()
		if prebidExt == nil {
			prebidExt = &openrtb_ext.ExtRequestPrebid{}
		}
		prebidExt.Floors = priceFloors
		requestExt.SetPrebid(prebidExt)
		bidRequestWrapper.RebuildRequest()
	}
}

// resolveFloorMin gets floorMin value from request and dynamic fetched data
func resolveFloorMin(reqFloors *openrtb_ext.PriceFloorRules, fetchFloors *openrtb_ext.PriceFloorRules, conversions currency.Conversions) Price {
	var requestFloorMinCur, providerFloorMinCur string
	var requestFloorMin, providerFloorMin float64

	if reqFloors != nil {
		requestFloorMin = reqFloors.FloorMin
		requestFloorMinCur = reqFloors.FloorMinCur
		if len(requestFloorMinCur) == 0 && reqFloors.Data != nil {
			requestFloorMinCur = reqFloors.Data.Currency
		}
	}

	if fetchFloors != nil {
		providerFloorMin = fetchFloors.FloorMin
		providerFloorMinCur = fetchFloors.FloorMinCur
		if len(providerFloorMinCur) == 0 && fetchFloors.Data != nil {
			providerFloorMinCur = fetchFloors.Data.Currency
		}
	}

	if len(requestFloorMinCur) > 0 {
		if requestFloorMin > 0 {
			return Price{FloorMin: requestFloorMin, FloorMinCur: requestFloorMinCur}
		}

		if providerFloorMin > 0 {
			if strings.Compare(providerFloorMinCur, requestFloorMinCur) == 0 || len(providerFloorMinCur) == 0 {
				return Price{FloorMin: providerFloorMin, FloorMinCur: requestFloorMinCur}
			}
			rate, err := conversions.GetRate(providerFloorMinCur, requestFloorMinCur)
			if err != nil {
				return Price{FloorMin: 0, FloorMinCur: requestFloorMinCur}
			}
			return Price{FloorMin: roundToFourDecimals(rate * providerFloorMin), FloorMinCur: requestFloorMinCur}
		}
	}

	if len(providerFloorMinCur) > 0 {
		if providerFloorMin > 0 {
			return Price{FloorMin: providerFloorMin, FloorMinCur: providerFloorMinCur}
		}
		if requestFloorMin > 0 {
			return Price{FloorMin: requestFloorMin, FloorMinCur: providerFloorMinCur}
		}
	}

	return Price{FloorMin: requestFloorMin, FloorMinCur: requestFloorMinCur}

}

// mergeFloors does merging for floors data from request and dynamic fetch
func mergeFloors(reqFloors *openrtb_ext.PriceFloorRules, fetchFloors *openrtb_ext.PriceFloorRules, conversions currency.Conversions) *openrtb_ext.PriceFloorRules {
	mergedFloors := fetchFloors.DeepCopy()
	if mergedFloors.Enabled == nil {
		mergedFloors.Enabled = new(bool)
	}
	*mergedFloors.Enabled = fetchFloors.GetEnabled() && reqFloors.GetEnabled()

	if reqFloors == nil {
		return mergedFloors
	}

	if reqFloors.Enforcement != nil {
		mergedFloors.Enforcement = reqFloors.Enforcement.DeepCopy()
	}

	floorMinPrice := resolveFloorMin(reqFloors, fetchFloors, conversions)
	if floorMinPrice.FloorMin > 0 {
		mergedFloors.FloorMin = floorMinPrice.FloorMin
		mergedFloors.FloorMinCur = floorMinPrice.FloorMinCur
	}

	if reqFloors != nil && reqFloors.Location != nil && reqFloors.Location.URL != "" {
		mergedFloors.Location = ptrutil.Clone(reqFloors.Location)
	}

	return mergedFloors
}
