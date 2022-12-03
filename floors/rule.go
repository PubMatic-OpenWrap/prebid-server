package floors

import (
	"encoding/json"
	"fmt"
	"math/bits"
	"regexp"
	"sort"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/currency"
	"github.com/prebid/prebid-server/openrtb_ext"
)

const (
	SiteDomain string = "siteDomain"
	PubDomain  string = "pubDomain"
	Domain     string = "domain"
	Bundle     string = "bundle"
	Channel    string = "channel"
	MediaType  string = "mediaType"
	Size       string = "size"
	GptSlot    string = "gptSlot"
	AdUnitCode string = "adUnitCode"
	Country    string = "country"
	DeviceType string = "deviceType"
	Tablet     string = "tablet"
	Phone      string = "phone"
)

func getFloorCurrency(floorExt *openrtb_ext.PriceFloorRules) string {
	floorCur := "USD"
	if floorExt == nil || floorExt.Data == nil {
		return floorCur
	}

	if floorExt.Data.Currency != "" {
		floorCur = floorExt.Data.Currency
	}

	if floorExt.Data.ModelGroups[0].Currency != "" {
		floorCur = floorExt.Data.ModelGroups[0].Currency
	}
	return floorCur
}

func getMinFloorValue(floorExt *openrtb_ext.PriceFloorRules, conversions currency.Conversions) (float64, string, error) {
	var err error
	var rate float64

	if floorExt == nil {
		return 0, "USD", err
	}
	floorMin := floorExt.FloorMin
	floorCur := getFloorCurrency(floorExt)

	if floorExt.FloorMin > 0.0 && floorExt.FloorMinCur != "" && floorCur != "" &&
		floorExt.FloorMinCur != floorCur {
		rate, err = conversions.GetRate(floorExt.FloorMinCur, floorCur)
		floorMin = rate * floorExt.FloorMin
	}
	return floorMin, floorCur, err
}

func updateImpExtWithFloorDetails(matchedRule string, imp *openrtb_ext.ImpWrapper, floorVal float64) {
	impExt, _ := imp.GetImpExt()
	extImpPrebid := impExt.GetPrebid()
	extImpPrebid.Floors = openrtb_ext.ExtImpPrebidFloors{
		FloorRule:      matchedRule,
		FloorRuleValue: fmt.Sprintf("%.4f", floorVal),
	}
	impExt.SetPrebid(extImpPrebid)
}

func selectFloorModelGroup(modelGroups []openrtb_ext.PriceFloorModelGroup, f func(int) int) []openrtb_ext.PriceFloorModelGroup {
	totalModelWeight := 0

	for i := 0; i < len(modelGroups); i++ {
		if modelGroups[i].ModelWeight == 0 {
			modelGroups[i].ModelWeight = 1
		}
		totalModelWeight += modelGroups[i].ModelWeight
	}

	sort.SliceStable(modelGroups, func(i, j int) bool {
		return modelGroups[i].ModelWeight < modelGroups[j].ModelWeight
	})

	winWeight := f(totalModelWeight + 1)
	debugWeight := winWeight
	for i, modelGroup := range modelGroups {
		winWeight -= modelGroup.ModelWeight
		if winWeight <= 0 {
			modelGroups[0], modelGroups[i] = modelGroups[i], modelGroups[0]
			modelGroups[0].DebugWeight = debugWeight
			return modelGroups[:1]
		}
	}
	return modelGroups[:1]
}

func shouldSkipFloors(ModelGroupsSkipRate, DataSkipRate, RootSkipRate int, f func(int) int) bool {
	skipRate := 0

	if ModelGroupsSkipRate > 0 {
		skipRate = ModelGroupsSkipRate
	} else if DataSkipRate > 0 {
		skipRate = DataSkipRate
	} else {
		skipRate = RootSkipRate
	}
	return skipRate >= f(skipRateMax+1)
}

func findRule(ruleValues map[string]float64, delimiter string, desiredRuleKey []string, numFields int) (string, bool) {

	ruleKeys := prepareRuleCombinations(desiredRuleKey, numFields, delimiter)
	for i := 0; i < len(ruleKeys); i++ {
		if _, ok := ruleValues[ruleKeys[i]]; ok {
			return ruleKeys[i], true
		}
	}
	return "", false
}

func createRuleKey(floorSchema openrtb_ext.PriceFloorSchema, request *openrtb2.BidRequest, imp openrtb2.Imp) []string {
	var ruleKeys []string

	for _, field := range floorSchema.Fields {
		value := catchAll
		switch field {
		case MediaType:
			value = getMediaType(imp)
		case Size:
			value = getSizeValue(imp)
		case Domain:
			value = getDomain(request)
		case SiteDomain:
			value = getSiteDomain(request)
		case Bundle:
			value = getBundle(request)
		case PubDomain:
			value = getPublisherDomain(request)
		case Country:
			value = getDeviceCountry(request)
		case DeviceType:
			value = getDeviceType(request)
		case Channel:
			value = extractChanelNameFromBidRequestExt(request)
		case GptSlot:
			value = getgptslot(imp)
		case AdUnitCode:
			value = getAdUnitCode(imp)
		}
		ruleKeys = append(ruleKeys, value)
	}
	return ruleKeys
}

func getDeviceType(request *openrtb2.BidRequest) string {
	value := catchAll
	if request.Device == nil || len(request.Device.UA) == 0 {
		return value
	}
	if isMobileDevice(request.Device.UA) {
		value = Phone
	} else if isTabletDevice(request.Device.UA) {
		value = Tablet
	}
	return value
}

func getDeviceCountry(request *openrtb2.BidRequest) string {
	value := catchAll
	if request.Device != nil && request.Device.Geo != nil {
		value = request.Device.Geo.Country
	}
	return value
}

func getMediaType(imp openrtb2.Imp) string {
	value := catchAll
	if imp.Banner != nil {
		value = string(openrtb_ext.BidTypeBanner)
	} else if imp.Video != nil {
		value = string(openrtb_ext.BidTypeVideo)
	} else if imp.Audio != nil {
		value = string(openrtb_ext.BidTypeAudio)
	} else if imp.Native != nil {
		value = string(openrtb_ext.BidTypeNative)
	}
	return value
}

func getSizeValue(imp openrtb2.Imp) string {
	size := catchAll
	width := int64(0)
	height := int64(0)
	if imp.Banner != nil {
		if len(imp.Banner.Format) > 0 {
			width = imp.Banner.Format[0].W
			height = imp.Banner.Format[0].H
		} else if imp.Banner.W != nil && imp.Banner.H != nil {
			width = *imp.Banner.W
			height = *imp.Banner.H
		}
	} else {
		width = imp.Video.W
		height = imp.Video.H
	}

	if width != 0 && height != 0 {
		size = fmt.Sprintf("%dx%d", width, height)
	}
	return size
}

func getDomain(request *openrtb2.BidRequest) string {
	value := catchAll
	if request.Site != nil {
		if len(request.Site.Domain) > 0 {
			value = request.Site.Domain
		} else if request.Site.Publisher != nil && len(request.Site.Publisher.Domain) > 0 {
			value = request.Site.Publisher.Domain
		}
	} else if request.App != nil {
		if len(request.App.Domain) > 0 {
			value = request.App.Domain
		} else if request.App.Publisher != nil && len(request.App.Publisher.Domain) > 0 {
			value = request.App.Publisher.Domain
		}
	}
	return value
}

func getSiteDomain(request *openrtb2.BidRequest) string {
	var value string
	if request.Site != nil {
		value = request.Site.Domain
	} else {
		value = request.App.Domain
	}
	return value
}

func getPublisherDomain(request *openrtb2.BidRequest) string {
	value := catchAll
	if request.Site != nil && request.Site.Publisher != nil && len(request.Site.Publisher.Domain) > 0 {
		value = request.Site.Publisher.Domain
	} else if request.App != nil && request.App.Publisher != nil && len(request.App.Publisher.Domain) > 0 {
		value = request.App.Publisher.Domain
	}
	return value
}

func getBundle(request *openrtb2.BidRequest) string {
	value := catchAll
	if request.App != nil && len(request.App.Bundle) > 0 {
		value = request.App.Bundle
	}
	return value
}

func getgptslot(imp openrtb2.Imp) string {
	value := catchAll
	adsname, err := jsonparser.GetString(imp.Ext, "data", "adserver", "name")
	if err == nil && adsname == "gam" {
		gptSlot, _ := jsonparser.GetString(imp.Ext, "data", "adserver", "adslot")
		if gptSlot != "" {
			value = gptSlot
		}
	} else {
		value = getpbadslot(imp)
	}
	return value
}

func extractChanelNameFromBidRequestExt(bidRequest *openrtb2.BidRequest) string {
	requestExt := &openrtb_ext.ExtRequest{}
	if bidRequest == nil {
		return catchAll
	}

	if len(bidRequest.Ext) > 0 {
		err := json.Unmarshal(bidRequest.Ext, &requestExt)
		if err != nil {
			return catchAll
		}
	}

	if requestExt.Prebid.Channel != nil {
		return requestExt.Prebid.Channel.Name
	}
	return catchAll
}

func getpbadslot(imp openrtb2.Imp) string {
	value := catchAll
	pbAdSlot, err := jsonparser.GetString(imp.Ext, "data", "pbadslot")
	if err == nil {
		value = pbAdSlot
	}
	return value
}

func getAdUnitCode(imp openrtb2.Imp) string {
	adUnitCode := catchAll
	gpId, err := jsonparser.GetString(imp.Ext, "gpid")
	if err == nil && gpId != "" {
		return gpId
	}

	if imp.TagID != "" {
		return imp.TagID
	}

	pbAdSlot, err := jsonparser.GetString(imp.Ext, "data", "pbadslot")
	if err == nil && pbAdSlot != "" {
		return pbAdSlot
	}

	storedrequestID, err := jsonparser.GetString(imp.Ext, "prebid", "storedrequest", "id")
	if err == nil && storedrequestID != "" {
		return storedrequestID
	}
	return adUnitCode
}

func isMobileDevice(userAgent string) bool {
	isMobile, err := regexp.MatchString("(?i)Phone|iPhone|Android|Mobile", userAgent)
	if err != nil {
		return false
	}
	return isMobile
}

func isTabletDevice(userAgent string) bool {
	isTablet, err := regexp.MatchString("(?i)tablet|iPad|Windows NT", userAgent)
	if err != nil {
		return false
	}
	return isTablet
}

func prepareRuleCombinations(keys []string, numSchemaFields int, delimiter string) []string {
	var subset []string
	var comb []int
	var desiredkeys [][]string
	var ruleKeys []string

	segNum := 1 << numSchemaFields
	for i := 0; i < numSchemaFields; i++ {
		subset = append(subset, strings.ToLower(keys[i]))
		comb = append(comb, i)
	}
	desiredkeys = append(desiredkeys, subset)
	for numWildCart := 1; numWildCart <= numSchemaFields; numWildCart++ {
		newComb := generateCombinations(comb, numWildCart, segNum)
		for i := 0; i < len(newComb); i++ {
			eachSet := make([]string, len(desiredkeys[0]))
			_ = copy(eachSet, desiredkeys[0])
			for j := 0; j < len(newComb[i]); j++ {
				eachSet[newComb[i][j]] = catchAll
			}
			desiredkeys = append(desiredkeys, eachSet)
		}
	}
	ruleKeys = prepareRuleKeys(desiredkeys, delimiter)
	return ruleKeys
}

func prepareRuleKeys(desiredkeys [][]string, delimiter string) []string {
	var ruleKeys []string
	for i := 0; i < len(desiredkeys); i++ {
		subset := desiredkeys[i][0]
		for j := 1; j < len(desiredkeys[i]); j++ {
			subset += delimiter + desiredkeys[i][j]
		}
		ruleKeys = append(ruleKeys, subset)
	}
	return ruleKeys
}

func generateCombinations(set []int, numWildCart int, segNum int) (comb [][]int) {
	length := uint(len(set))

	if numWildCart > len(set) {
		numWildCart = len(set)
	}

	for subsetBits := 1; subsetBits < (1 << length); subsetBits++ {
		if numWildCart > 0 && bits.OnesCount(uint(subsetBits)) != numWildCart {
			continue
		}
		var subset []int
		for object := uint(0); object < length; object++ {
			if (subsetBits>>object)&1 == 1 {
				subset = append(subset, set[object])
			}
		}
		comb = append(comb, subset)
	}

	// Sort combinations based on priority mentioned in https://docs.prebid.org/dev-docs/modules/floors.html#rule-selection-process
	sort.SliceStable(comb, func(i, j int) bool {
		wt1 := 0
		for k := 0; k < len(comb[i]); k++ {
			wt1 += 1 << (segNum - comb[i][k])
		}

		wt2 := 0
		for k := 0; k < len(comb[j]); k++ {
			wt2 += 1 << (segNum - comb[j][k])
		}
		return wt1 < wt2
	})
	return comb
}
