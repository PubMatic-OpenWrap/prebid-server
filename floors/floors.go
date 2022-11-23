package floors

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/currency"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type Price struct {
	FloorMin    float64
	FloorMinCur string
}

const (
	defaultDelimiter string = "|"
	catchAll         string = "*"
	skipRateMin      int    = 0
	skipRateMax      int    = 100
	modelWeightMax   int    = 100
	modelWeightMin   int    = 1
	enforceRateMin   int    = 0
	enforceRateMax   int    = 100
)

func EnrichWithPriceFloors(bidRequestWrapper *openrtb_ext.RequestWrapper, account config.Account, conversions currency.Conversions) []error {
	err := []error{}
	if bidRequestWrapper == nil || bidRequestWrapper.BidRequest == nil {
		return []error{fmt.Errorf("Empty bidrequest")}
	}

	if isPriceFloorsDisabled(account, bidRequestWrapper) {
		return []error{fmt.Errorf("Floors feature is disabled at account level or request")}
	}

	floors, err := resolveFloors(account, bidRequestWrapper, conversions)
	if len(err) == 0 {
		err = updateBidRequestWithFloors(floors, bidRequestWrapper.BidRequest, conversions)
	}
	return err
}

// updateBidRequestWithFloors will update imp.bidfloor and imp.bidfloorcur based on rules matching
func updateBidRequestWithFloors(extFloorRules *openrtb_ext.PriceFloorRules, request *openrtb2.BidRequest, conversions currency.Conversions) []error {
	var (
		floorErrList      []error
		floorModelErrList []error
		floorVal          float64
	)

	if extFloorRules == nil || extFloorRules.Data == nil || len(extFloorRules.Data.ModelGroups) == 0 {
		return []error{}
	}

	if !extFloorRules.GetEnabled() {
		return []error{fmt.Errorf("Floors disabled in request")}
	}

	floorData := extFloorRules.Data
	modelGroup := floorData.ModelGroups[0]
	if modelGroup.Schema.Delimiter == "" {
		modelGroup.Schema.Delimiter = defaultDelimiter
	}

	extFloorRules.Skipped = new(bool)
	if shouldSkipFloors(extFloorRules.Data.ModelGroups[0].SkipRate, extFloorRules.Data.SkipRate, extFloorRules.SkipRate, rand.Intn) {
		*extFloorRules.Skipped = true
		floorData.ModelGroups = nil
		return floorModelErrList
	}

	floorErrList = validateFloorRulesAndLowerValidRuleKey(modelGroup.Schema, modelGroup.Schema.Delimiter, modelGroup.Values)
	if len(modelGroup.Values) > 0 {
		for i := 0; i < len(request.Imp); i++ {
			desiredRuleKey := createRuleKey(modelGroup.Schema, request, request.Imp[i])
			matchedRule, isRuleMatched := findRule(modelGroup.Values, modelGroup.Schema.Delimiter, desiredRuleKey, len(modelGroup.Schema.Fields))

			floorVal = modelGroup.Default
			if isRuleMatched {
				floorVal = modelGroup.Values[matchedRule]
			}

			floorMinVal, floorCur, err := getMinFloorValue(extFloorRules, conversions)
			if err == nil {
				floorVal = math.Round(floorVal*10000) / 10000
				bidFloor := floorVal
				if floorMinVal > float64(0) && floorVal < floorMinVal {
					bidFloor = floorMinVal
				}

				if bidFloor > float64(0) {
					request.Imp[i].BidFloor = bidFloor
					request.Imp[i].BidFloorCur = floorCur
				}
				if isRuleMatched {
					updateImpExtWithFloorDetails(&request.Imp[i], matchedRule, floorVal, bidFloor)
				}
			} else {
				floorModelErrList = append(floorModelErrList, fmt.Errorf("Error in getting FloorMin value : '%v'", err.Error()))
			}

		}
	}
	floorModelErrList = append(floorModelErrList, floorErrList...)
	return floorModelErrList
}

func isPriceFloorsDisabled(account config.Account, bidRequestWrapper *openrtb_ext.RequestWrapper) bool {
	return isPriceFloorsDisabledForAccount(account) || isPriceFloorsDisabledForRequest(bidRequestWrapper)
}

func isPriceFloorsDisabledForAccount(account config.Account) bool {
	return !account.PriceFloors.Enabled
}

func isPriceFloorsDisabledForRequest(bidRequestWrapper *openrtb_ext.RequestWrapper) bool {
	requestExt, err := bidRequestWrapper.GetRequestExt()
	if err == nil {
		if prebidExt := requestExt.GetPrebid(); prebidExt != nil && prebidExt.Floors != nil && !prebidExt.Floors.GetEnabled() {
			return true
		}
	}
	return false
}

func resolveFloors(account config.Account, bidRequestWrapper *openrtb_ext.RequestWrapper, conversions currency.Conversions) (*openrtb_ext.PriceFloorRules, []error) {
	var errlist []error
	var floorsJson *openrtb_ext.PriceFloorRules

	reqFloor := extractFloorsFromRequest(bidRequestWrapper)
	fetchResult := fetchAccountFloors(account)

	if shouldUseDynamicFetchedFloor(account) && fetchResult != nil && fetchResult.fetchStatus == openrtb_ext.FetchSuccess {
		mergedFloor := mergeFloors(reqFloor, fetchResult.priceFloors, conversions)
		floorsJson, errlist = createFloorsFrom(mergedFloor, fetchResult.fetchStatus, openrtb_ext.FetchLocation)
	} else if reqFloor != nil {
		floorsJson, errlist = createFloorsFrom(reqFloor, openrtb_ext.FetchNone, openrtb_ext.RequestLocation)
	} else {
		floorsJson, errlist = createFloorsFrom(nil, openrtb_ext.FetchNone, openrtb_ext.NoDataLocation)
	}
	updateFloorsInRequest(bidRequestWrapper, floorsJson)
	return floorsJson, errlist
}

func createFloorsFrom(floors *openrtb_ext.PriceFloorRules, fetchStatus, floorLocation string) (*openrtb_ext.PriceFloorRules, []error) {

	var floorModelErrList []error
	if floors != nil && floors.Data != nil {
		floorData := floors.Data

		floorSkipRateErr := validateFloorParams(floors)
		if floorSkipRateErr != nil {
			return floors, append(floorModelErrList, floorSkipRateErr)
		}

		floorData.ModelGroups, floorModelErrList = selectValidFloorModelGroups(floorData.ModelGroups)
		if len(floorData.ModelGroups) == 0 {
			return floors, floorModelErrList
		} else if len(floorData.ModelGroups) > 1 {
			floorData.ModelGroups = selectFloorModelGroup(floorData.ModelGroups, rand.Intn)
		}

		modelGroup := floorData.ModelGroups[0]
		if modelGroup.Schema.Delimiter == "" {
			modelGroup.Schema.Delimiter = defaultDelimiter
		}
	} else if floors == nil {
		floors = new(openrtb_ext.PriceFloorRules)
	}
	floors.FetchStatus = fetchStatus
	floors.PriceFloorLocation = floorLocation
	return floors, floorModelErrList
}

func mergeFloors(reqFloors *openrtb_ext.PriceFloorRules, fetchFloors openrtb_ext.PriceFloorRules, conversions currency.Conversions) *openrtb_ext.PriceFloorRules {

	var enforceRate int

	floorsEnabledByRequest := reqFloors.GetEnabled()
	floorMinPrice := resolveFloorMin(reqFloors, fetchFloors, conversions)

	if reqFloors != nil && reqFloors.Enforcement != nil {
		enforceRate = reqFloors.Enforcement.EnforceRate
	}

	if floorsEnabledByRequest || enforceRate > 0 || floorMinPrice.FloorMin > float64(0) {
		floorsEnabledByProvider := getFloorsEnabledFlag(fetchFloors)
		floorsProviderEnforcement := fetchFloors.Enforcement

		if fetchFloors.Enabled == nil {
			fetchFloors.Enabled = new(bool)
		}
		*fetchFloors.Enabled = floorsEnabledByProvider && floorsEnabledByRequest
		fetchFloors.Enforcement = resolveEnforcement(floorsProviderEnforcement, enforceRate)
		if floorMinPrice.FloorMin > float64(0) {
			fetchFloors.FloorMin = floorMinPrice.FloorMin
			fetchFloors.FloorMinCur = floorMinPrice.FloorMinCur
		}
	}
	return &fetchFloors
}

func resolveEnforcement(enforcement *openrtb_ext.PriceFloorEnforcement, enforceRate int) *openrtb_ext.PriceFloorEnforcement {
	if enforcement == nil {
		enforcement = new(openrtb_ext.PriceFloorEnforcement)
	}
	enforcement.EnforceRate = enforceRate
	return enforcement
}

func getFloorsEnabledFlag(reqFloors openrtb_ext.PriceFloorRules) bool {
	if reqFloors.Enabled != nil {
		return *reqFloors.Enabled
	}
	return true
}

func resolveFloorMin(reqFloors *openrtb_ext.PriceFloorRules, fetchFloors openrtb_ext.PriceFloorRules, conversions currency.Conversions) Price {

	var floorCur, reqFloorMinCur string
	var reqFloorMin float64
	if reqFloors != nil {
		floorCur = getFloorCurrency(reqFloors)
		reqFloorMin = reqFloors.FloorMin
		reqFloorMinCur = reqFloors.FloorMinCur
	}

	if len(reqFloorMinCur) == 0 {
		reqFloorMinCur = floorCur
	}

	provFloorMinCur := fetchFloors.FloorMinCur
	provFloorMin := fetchFloors.FloorMin

	if len(reqFloorMinCur) > 0 {
		if reqFloorMin > float64(0) {
			return Price{FloorMin: reqFloorMin, FloorMinCur: reqFloorMinCur}
		} else if provFloorMin > float64(0) {
			if len(provFloorMinCur) == 0 || strings.Compare(reqFloorMinCur, provFloorMinCur) == 0 {
				return Price{FloorMin: provFloorMin, FloorMinCur: reqFloorMinCur}
			}
			rate, err := conversions.GetRate(provFloorMinCur, reqFloorMinCur)
			if err == nil {
				return Price{FloorMinCur: reqFloorMinCur,
					FloorMin: math.Round(rate*provFloorMin*10000) / 10000}
			}
		}
	}

	if len(provFloorMinCur) > 0 {
		if provFloorMin > float64(0) {
			return Price{FloorMin: provFloorMin, FloorMinCur: provFloorMinCur}
		} else if reqFloorMin > float64(0) {
			return Price{FloorMin: reqFloorMin, FloorMinCur: provFloorMinCur}
		}
	}
	return Price{FloorMin: 0.0, FloorMinCur: floorCur}
}

func shouldUseDynamicFetchedFloor(Account config.Account) bool {
	return Account.PriceFloors.UseDynamicData
}

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

func updateFloorsInRequest(bidRequestWrapper *openrtb_ext.RequestWrapper, priceFloors *openrtb_ext.PriceFloorRules) {
	requestExt, err := bidRequestWrapper.GetRequestExt()
	if err == nil {
		prebidExt := requestExt.GetPrebid()
		if prebidExt != nil {
			prebidExt.Floors = priceFloors
			requestExt.SetPrebid(prebidExt)
		}
	}
}
