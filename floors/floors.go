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
	modelWeightMin   int    = 0
	enforceRateMin   int    = 0
	enforceRateMax   int    = 100
)

// ModifyImpsWithFloors will validate floor rules, based on request and rules prepares various combinations
// to match with floor rules and selects appripariate floor rule and update imp.bidfloor and imp.bidfloorcur
func ModifyImpsWithFloors(floorExt *openrtb_ext.PriceFloorRules, request *openrtb2.BidRequest, conversions currency.Conversions) []error {
	var (
		floorErrList      []error
		floorModelErrList []error
		floorVal          float64
	)

	if floorExt == nil || floorExt.Data == nil {
		return nil
	}

	floorData := floorExt.Data
	floorSkipRateErr := validateFloorSkipRates(floorExt)
	if floorSkipRateErr != nil {
		return append(floorModelErrList, floorSkipRateErr)
	}

	floorData.ModelGroups, floorModelErrList = selectValidFloorModelGroups(floorData.ModelGroups)
	if len(floorData.ModelGroups) == 0 {
		return floorModelErrList
	} else if len(floorData.ModelGroups) > 1 {
		floorData.ModelGroups = selectFloorModelGroup(floorData.ModelGroups, rand.Intn)
	}

	modelGroup := floorData.ModelGroups[0]
	if modelGroup.Schema.Delimiter == "" {
		modelGroup.Schema.Delimiter = defaultDelimiter
	}

	floorExt.Skipped = new(bool)
	if shouldSkipFloors(floorExt.Data.ModelGroups[0].SkipRate, floorExt.Data.SkipRate, floorExt.SkipRate, rand.Intn) {
		*floorExt.Skipped = true
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

			floorMinVal, floorCur, err := getMinFloorValue(floorExt, conversions)
			if err == nil {
				bidFloor := floorVal
				if floorMinVal > 0.0 && floorVal < floorMinVal {
					bidFloor = floorMinVal
				}

				if bidFloor > 0.0 {
					request.Imp[i].BidFloor = math.Round(bidFloor*10000) / 10000
					request.Imp[i].BidFloorCur = floorCur
				}
				if isRuleMatched {
					updateImpExtWithFloorDetails(matchedRule, &request.Imp[i], modelGroup.Values[matchedRule])
				}
			} else {
				floorModelErrList = append(floorModelErrList, fmt.Errorf("Error in getting FloorMin value : '%v'", err.Error()))
			}

		}
	}
	floorModelErrList = append(floorModelErrList, floorErrList...)
	return floorModelErrList
}

func EnrichWithPriceFloors(bidRequestWrapper *openrtb_ext.RequestWrapper, account config.Account, conversions currency.Conversions) []error {

	err := []error{}
	if bidRequestWrapper == nil || bidRequestWrapper.BidRequest == nil || isPriceFloorsDisabled(account, bidRequestWrapper) {
		return err
	}

	floors, err := resolveFloors(account, bidRequestWrapper, conversions)
	if len(err) == 0 {
		err = updateBidRequestWithFloors(floors, bidRequestWrapper.BidRequest, conversions)
	}
	return err
}

// updateBidRequestWithFloors will update imp.bidfloor and imp.bidfloorcur based on rules matching
func updateBidRequestWithFloors(floorExt *openrtb_ext.PriceFloorRules, request *openrtb2.BidRequest, conversions currency.Conversions) []error {
	var (
		floorErrList      []error
		floorModelErrList []error
		floorVal          float64
	)

	if floorExt == nil || floorExt.Data == nil {
		return nil
	}

	floorData := floorExt.Data
	modelGroup := floorData.ModelGroups[0]
	if modelGroup.Schema.Delimiter == "" {
		modelGroup.Schema.Delimiter = defaultDelimiter
	}

	floorExt.Skipped = new(bool)
	if shouldSkipFloors(floorExt.Data.ModelGroups[0].SkipRate, floorExt.Data.SkipRate, floorExt.SkipRate, rand.Intn) {
		*floorExt.Skipped = true
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

			floorMinVal, floorCur, err := getMinFloorValue(floorExt, conversions)
			if err == nil {
				bidFloor := floorVal
				if floorMinVal > 0.0 && floorVal < floorMinVal {
					bidFloor = floorMinVal
				}

				if bidFloor > 0.0 {
					request.Imp[i].BidFloor = math.Round(bidFloor*10000) / 10000
					request.Imp[i].BidFloorCur = floorCur
				}
				if isRuleMatched {
					updateImpExtWithFloorDetails(matchedRule, &request.Imp[i], modelGroup.Values[matchedRule])
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
	reqFloor := extractFloorsFromRequest(bidRequestWrapper)
	fetchReult := fetchFloors(account)

	if reqFloor != nil && shouldUseDynamicFetched(account) && fetchReult != nil && fetchReult.fetchStatus == 0 {
		mergedFloor := mergeFloors(*reqFloor, fetchReult.priceFloors, conversions)
		return createFloorsFrom(&mergedFloor, fetchReult.fetchStatus, openrtb_ext.Fetch)
	}

	if reqFloor != nil {
		return createFloorsFrom(reqFloor, fetchReult.fetchStatus, openrtb_ext.Request)
	}

	return createFloorsFrom(nil, fetchReult.fetchStatus, openrtb_ext.NoData)
}

func createFloorsFrom(floors *openrtb_ext.PriceFloorRules, fetchStatus, floorLocation int) (*openrtb_ext.PriceFloorRules, []error) {

	var floorModelErrList []error
	if floors != nil && floors.Data != nil {
		floorData := floors.Data

		floorSkipRateErr := validateFloorSkipRates(floors)
		if floorSkipRateErr != nil {
			return floors, append(floorModelErrList, floorSkipRateErr)
		}

		floorData.ModelGroups, _ = selectValidFloorModelGroups(floorData.ModelGroups)
		if len(floorData.ModelGroups) == 0 {
			return floors, floorModelErrList
		} else if len(floorData.ModelGroups) > 1 {
			floorData.ModelGroups = selectFloorModelGroup(floorData.ModelGroups, rand.Intn)
		}

		modelGroup := floorData.ModelGroups[0]
		if modelGroup.Schema.Delimiter == "" {
			modelGroup.Schema.Delimiter = defaultDelimiter
		}
		floors.FetchStatus = fetchStatus
		floors.PriceFloorLocation = floorLocation

	}
	return floors, floorModelErrList
}

func mergeFloors(reqFloors openrtb_ext.PriceFloorRules, fetchFloors openrtb_ext.PriceFloorRules, conversions currency.Conversions) openrtb_ext.PriceFloorRules {
	var mergedFloors openrtb_ext.PriceFloorRules
	var enforceRate int

	floorsEnabledByRequest := getFloorsEnabledFlag(reqFloors)
	floorMinPrice := resolveFloorMin(reqFloors, fetchFloors, conversions)

	if reqFloors.Enforcement != nil {
		enforceRate = reqFloors.Enforcement.EnforceRate
	}

	if floorsEnabledByRequest || enforceRate > 0 || floorMinPrice.FloorMin > float64(0.0) {

		floorsEnabledByProvider := getFloorsEnabledFlag(fetchFloors)
		floorsProviderEnforcement := fetchFloors.Enforcement

		if fetchFloors.Enabled == nil {
			fetchFloors.Enabled = new(bool)
		}
		*fetchFloors.Enabled = floorsEnabledByProvider && floorsEnabledByRequest
		fetchFloors.Enforcement = resolveEnforcement(floorsProviderEnforcement, enforceRate)
		fetchFloors.FloorMin = floorMinPrice.FloorMin
		fetchFloors.FloorMinCur = floorMinPrice.FloorMinCur
	}
	return mergedFloors
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

func resolveFloorMin(reqFloors openrtb_ext.PriceFloorRules, fetchFloors openrtb_ext.PriceFloorRules, conversions currency.Conversions) Price {

	floorCur := getFloorCurrency(&reqFloors)
	reqFloorMinCur := reqFloors.FloorMinCur
	if len(reqFloorMinCur) > 0 {
		reqFloorMinCur = floorCur
	}
	reqFloorMin := reqFloors.FloorMin

	provFloorMinCur := fetchFloors.FloorMinCur
	provFloorMin := fetchFloors.FloorMin

	if len(reqFloorMinCur) > 0 {
		if reqFloorMin > float64(0.0) {
			return Price{FloorMin: reqFloorMin, FloorMinCur: reqFloorMinCur}
		} else if provFloorMin > float64(0.0) {
			if strings.Compare(reqFloorMinCur, provFloorMinCur) == 0 {
				return Price{FloorMin: provFloorMin, FloorMinCur: reqFloorMinCur}
			}
			rate, _ := conversions.GetRate(provFloorMinCur, reqFloorMinCur)
			return Price{FloorMinCur: reqFloorMinCur,
				FloorMin: rate * provFloorMin}
		}
	}

	if len(provFloorMinCur) > 0 {
		if reqFloorMin > float64(0.0) {
			rate, _ := conversions.GetRate(provFloorMinCur, reqFloorMinCur)
			return Price{FloorMinCur: reqFloorMinCur,
				FloorMin: rate * reqFloorMin}
		}
		return Price{FloorMin: provFloorMin, FloorMinCur: provFloorMinCur}
	}
	return Price{FloorMin: 0.0}
}

func shouldUseDynamicFetched(Account config.Account) bool {
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
