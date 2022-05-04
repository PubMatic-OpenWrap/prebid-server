package floors

import (
	"encoding/json"
	"fmt"
	"math/bits"
	"math/rand"
	"sort"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/mxmCherry/openrtb/v15/openrtb2"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type RequestType string

const (
	DEFAULT_DELIMITER      string = "|"
	CATCH_ALL              string = "*"
	SKIP_RATE_MIN          int    = 0
	SKIP_RATE_MAX          int    = 100
	MODEL_WEIGHT_MAX_VALUE int    = 1000000
	MODEL_WEIGHT_MIN_VALUE int    = 0
)

func IsRequestEnabledWithFloor(requestExt *openrtb_ext.ExtRequest) bool {
	if requestExt.Prebid.Floors != nil && requestExt.Prebid.Floors.Enabled != nil && *requestExt.Prebid.Floors.Enabled == true {
		return true
	} else {
		return false
	}
}

func shouldSkipFloors(floorData *openrtb_ext.PriceFloorData) bool {
	var skipRate int

	if floorData.ModelGroups[0].SkipRate > 0 {
		skipRate = floorData.ModelGroups[0].SkipRate
	} else {
		skipRate = floorData.SkipRate
	}

	if skipRate > 0 && skipRate > rand.Intn(SKIP_RATE_MAX) {
		return true
	} else {
		return false
	}
}

func UpdateImpsWithFloors(floorExt *openrtb_ext.PriceFloorRules, request *openrtb2.BidRequest) []error {
	var floorErrList []error
	var floorVal float64

	floorData := floorExt.Data

	floorModelErrList := validateFloorModelGroups(floorData.ModelGroups)

	if len(floorData.ModelGroups) > 1 {
		selectFloorModelGroup(floorData.ModelGroups)
	}

	if floorData.ModelGroups[0].Schema.Delimiter == "" {
		floorData.ModelGroups[0].Schema.Delimiter = DEFAULT_DELIMITER
	}

	floorExt.Skipped = new(bool)
	if shouldSkipFloors(floorData) == true {
		*floorExt.Skipped = true
		return floorModelErrList
	} else {
		*floorExt.Skipped = false
	}

	floorErrList = validateFloorRules(floorData.ModelGroups[0].Schema, floorData.ModelGroups[0].Schema.Delimiter, floorData.ModelGroups[0].Values)
	for i := 0; i < len(request.Imp); i++ {
		desiredRuleKey := CreateRuleKey(floorData.ModelGroups[0].Schema, request, request.Imp[i])
		matchedRule := findRule(floorData.ModelGroups[0].Values, floorData.ModelGroups[0].Schema.Delimiter, desiredRuleKey, len(floorData.ModelGroups[0].Schema.Fields))

		if matchedRule == "" {
			floorVal = floorData.ModelGroups[0].Default
		} else {
			floorVal = floorData.ModelGroups[0].Values[matchedRule]
		}

		if floorExt.FloorMin != 0.0 && floorVal < floorExt.FloorMin {
			request.Imp[i].BidFloor = floorExt.FloorMin
		} else {
			request.Imp[i].BidFloor = floorVal
		}
		request.Imp[i].BidFloorCur = "USD"

		updateImpExtWithFloorDetails(matchedRule, request, i)
	}
	floorModelErrList = append(floorModelErrList, floorErrList...)
	return floorModelErrList
}

func updateImpExtWithFloorDetails(matchedRule string, request *openrtb2.BidRequest, i int) {
	request.Imp[i].Ext, _ = jsonparser.Set(request.Imp[i].Ext, []byte(`"`+matchedRule+`"`), "prebid", "floors", "floorRule")
	request.Imp[i].Ext, _ = jsonparser.Set(request.Imp[i].Ext, []byte(fmt.Sprintf("%.4f", request.Imp[i].BidFloor)), "prebid", "floors", "floorRuleValue")
}

func validateFloorModelGroups(modelGroups []openrtb_ext.PriceFloorModelGroup) []error {
	var floorModelErrList []error

	for i, modelGroup := range modelGroups {

		if modelGroup.SkipRate < SKIP_RATE_MIN || modelGroup.SkipRate > SKIP_RATE_MAX {
			floorModelErrList = append(floorModelErrList, fmt.Errorf("Invalid Floor Model = '%v' due to SkipRate = '%v'", modelGroup.ModelVersion, modelGroup.SkipRate))
			modelGroups = append(modelGroups[:i], modelGroups[i+1:]...)
			continue
		}

		if modelGroup.ModelWeight < MODEL_WEIGHT_MIN_VALUE || modelGroup.ModelWeight > MODEL_WEIGHT_MAX_VALUE {
			floorModelErrList = append(floorModelErrList, fmt.Errorf("Invalid Floor Model = '%v' due to ModelWeight = '%v'", modelGroup.ModelVersion, modelGroup.ModelWeight))
			modelGroups = append(modelGroups[:i], modelGroups[i+1:]...)
			continue
		}
	}
	return floorModelErrList

}

func selectFloorModelGroup(modelGroups []openrtb_ext.PriceFloorModelGroup) {
	totalModelWeight := 0

	for i := 0; i < len(modelGroups); i++ {
		totalModelWeight += modelGroups[i].ModelWeight
	}

	sort.SliceStable(modelGroups, func(i, j int) bool {
		return modelGroups[i].ModelWeight < modelGroups[j].ModelWeight
	})

	winWeight := rand.Intn(totalModelWeight)

	for i, modelGroup := range modelGroups {
		winWeight -= modelGroup.ModelWeight
		if winWeight <= 0 {
			modelGroups = append([]openrtb_ext.PriceFloorModelGroup{modelGroup}, append((modelGroups)[:i], (modelGroups)[i+1:]...)...)
			return
		}
	}
}

func validateFloorRules(Schema openrtb_ext.PriceFloorSchema, delimiter string, RuleValues map[string]float64) []error {
	var floorErrList []error

	schemaLen := len(Schema.Fields)

	for key, _ := range RuleValues {
		parsedKey := strings.Split(key, delimiter)
		if len(parsedKey) != schemaLen {
			// Number of fields are not matching
			floorErrList = append(floorErrList, fmt.Errorf("Invalid Floor Rule = '%s' for Schema Fields = '%v'", key, Schema.Fields))
			delete(RuleValues, key)
		}
	}

	return floorErrList
}

func findRule(RuleValues map[string]float64, delimiter string, desiredRuleKey []string, numFields int) string {

	ruleKeys := PrepareRuleCombinations(desiredRuleKey, numFields, delimiter)
	for i := 0; i < len(ruleKeys); i++ {
		if _, ok := RuleValues[ruleKeys[i]]; ok {
			return ruleKeys[i]
		}
	}
	return ""
}

func getSizeValue(imp openrtb2.Imp) string {
	var size string
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
	} else {
		size = CATCH_ALL
	}

	return size
}

func extractChanelNameFromBidRequestExt(bidRequest *openrtb2.BidRequest) string {
	requestExt := &openrtb_ext.ExtRequest{}

	if bidRequest == nil {
		return ""
	}

	if len(bidRequest.Ext) > 0 {
		err := json.Unmarshal(bidRequest.Ext, &requestExt)
		if err != nil {
			return ""
		}
	}

	if requestExt.Prebid.Channel != nil {
		return requestExt.Prebid.Channel.Name
	} else {
		return ""
	}
}

func getpbadslot(imp openrtb2.Imp) string {
	var value string
	pbAdSlot, err := jsonparser.GetString(imp.Ext, "data", "pbadslot")
	if err == nil {
		value = pbAdSlot
	} else {
		value = CATCH_ALL
	}
	return value
}

func CreateRuleKey(floorSchema openrtb_ext.PriceFloorSchema, request *openrtb2.BidRequest, imp openrtb2.Imp) []string {
	var ruleKeys []string

	for _, field := range floorSchema.Fields {
		value := ""
		switch field {
		case "mediaType":
			if imp.Banner != nil {
				value = "banner"
			} else if imp.Video != nil {
				value = "video"
			} else if imp.Audio != nil {
				value = "audio"
			} else if imp.Native != nil {
				value = "native"
			} else {
				value = CATCH_ALL
			}
		case "size":
			value = getSizeValue(imp)
		case "domain":
			if request.Site != nil {
				if len(request.Site.Domain) > 0 {
					value = request.Site.Domain
				} else {
					value = request.Site.Publisher.Domain
				}
			} else {
				if len(request.App.Domain) > 0 {
					value = request.App.Domain
				} else {
					value = request.App.Publisher.Domain
				}
			}
		case "siteDomain":
			if request.Site != nil {
				value = request.Site.Domain
			} else {
				value = request.App.Domain
			}
		case "bundle":
			if request.App != nil {
				value = request.App.Bundle
			} else {
				value = CATCH_ALL
			}
		case "pubDomain":
			if request.Site != nil {
				value = request.Site.Publisher.Domain
			} else {
				value = request.App.Publisher.Domain
			}
		case "country":
			if request.Device != nil && request.Device.Geo != nil {
				value = request.Device.Geo.Country
			} else {
				value = CATCH_ALL
			}
		case "deviceType":
			if request.Device != nil && len(request.Device.UA) > 0 {
				if strings.Contains(request.Device.UA, "Phone") ||
					strings.Contains(request.Device.UA, "iPhone") ||
					strings.Contains(request.Device.UA, "Android") ||
					strings.Contains(request.Device.UA, "Mobile") {
					value = "phone"
				} else if strings.Contains(request.Device.UA, "tablet") ||
					strings.Contains(request.Device.UA, "iPad") ||
					strings.Contains(request.Device.UA, "Windows NT") {
					value = "tablet"
				} else {
					value = CATCH_ALL
				}
			} else {
				value = CATCH_ALL
			}
		case "channel":
			channel := extractChanelNameFromBidRequestExt(request)
			if channel != "" {
				value = channel
			} else {
				value = CATCH_ALL
			}
		case "gptSlot":
			adsname, err := jsonparser.GetString(imp.Ext, "data", "adserver", "name")
			if err == nil && adsname == "gam" {
				gptSlot, _ := jsonparser.GetString(imp.Ext, "data", "adserver", "adslot")
				if gptSlot != "" {
					value = gptSlot
				} else {
					value = CATCH_ALL
				}
			} else {
				value = getpbadslot(imp)
			}
		case "pbAdSlot":
			value = getpbadslot(imp)
		}

		ruleKeys = append(ruleKeys, value)
	}
	return ruleKeys
}

func PrepareRuleCombinations(keys []string, numSchemaFields int, delimiter string) []string {
	var subset []string
	var comb []int
	var desiredkeys [][]string
	var ruleKeys []string

	segNum := 1 << numSchemaFields
	for i := 0; i < numSchemaFields; i++ {
		subset = append(subset, keys[i])
		comb = append(comb, i)
	}
	desiredkeys = append(desiredkeys, subset)

	for numWildCart := 1; numWildCart <= numSchemaFields; numWildCart++ {
		newComb := GenerateCombinations(comb, numWildCart, segNum)
		for i := 0; i < len(newComb); i++ {
			eachSet := make([]string, len(desiredkeys[0]))
			_ = copy(eachSet, desiredkeys[0])
			for j := 0; j < len(newComb[i]); j++ {
				eachSet[newComb[i][j]] = CATCH_ALL
			}
			desiredkeys = append(desiredkeys, eachSet)
		}
	}

	for i := 0; i < len(desiredkeys); i++ {
		subset := desiredkeys[i][0]
		for j := 1; j < len(desiredkeys[i]); j++ {
			subset += delimiter + desiredkeys[i][j]
		}
		ruleKeys = append(ruleKeys, subset)
	}
	return ruleKeys
}

func GenerateCombinations(set []int, numWildCart int, segNum int) (comb [][]int) {
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
