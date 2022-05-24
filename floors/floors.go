package floors

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/buger/jsonparser"
	"github.com/mxmCherry/openrtb/v15/openrtb2"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type RequestType string

const (
	SiteDomain string = "siteDomain"
	PubDomain  string = "pubDomain"
	Domain     string = "domain"
	Bundle     string = "bundle"
	Channel    string = "channel"
	MediaType  string = "mediaType"
	Size       string = "size"
	GptSlot    string = "gptSlot"
	PbAdSlot   string = "pbAdSlot"
	Country    string = "country"
	DeviceType string = "deviceType"
	Tablet     string = "tablet"
	Phone      string = "phone"
)

const (
	DEFAULT_DELIMITER      string = "|"
	CATCH_ALL              string = "*"
	SKIP_RATE_MIN          int    = 0
	SKIP_RATE_MAX          int    = 100
	MODEL_WEIGHT_MAX_VALUE int    = 1000000
	MODEL_WEIGHT_MIN_VALUE int    = 0
	ENFORCE_RATE_MIN       int    = 0
	ENFORCE_RATE_MAX       int    = 100
)

type FloorConfig struct {
	FloorEnabled      bool
	EnforceRate       int
	EnforceDealFloors bool
)

func (fc *FloorConfig) Enabled() bool {
	return fc.FloorEnabled
}

func (fc *FloorConfig) GetEnforceRate() int {
	return fc.EnforceRate
}

func (fc *FloorConfig) EnforceDealFloor() bool {
	return fc.EnforceDealFloors
}

type Floor interface {
	Enabled() bool
	GetEnforceRate() int
	EnforceDealFloor() bool

}

// IsRequestEnabledWithFloor will check if floors is enabled in request
func IsRequestEnabledWithFloor(Floors *openrtb_ext.PriceFloorRules) bool {
	return Floors != nil && Floors.Enabled != nil && *Floors.Enabled
}

// UpdateImpsWithFloors will validate floor rules, based on request and rules prepares various combinations
// to match with floor rules and selects appripariate floor rule and update imp.bidfloor and imp.bidfloorcur
func UpdateImpsWithFloors(floorExt *openrtb_ext.PriceFloorRules, request *openrtb2.BidRequest) []error {
	var floorErrList []error
	var floorModelErrList []error
	var floorVal float64
	floorData := floorExt.Data

	floorData.ModelGroups, floorModelErrList = validateFloorModelGroups(floorData.ModelGroups)
	if len(floorData.ModelGroups) == 0 {
		return floorModelErrList
	} else if len(floorData.ModelGroups) > 1 {
		selectFloorModelGroup(floorData.ModelGroups, rand.Intn)
	}

	if floorData.ModelGroups[0].Schema.Delimiter == "" {
		floorData.ModelGroups[0].Schema.Delimiter = DEFAULT_DELIMITER
	}

	floorExt.Skipped = new(bool)
	if shouldSkipFloors(floorExt.Data.ModelGroups[0].SkipRate, floorExt.Data.SkipRate, floorExt.SkipRate, rand.Intn) {
		*floorExt.Skipped = true
		floorExt.Enforcement.EnforcePBS = false
		return floorModelErrList
	}

	floorErrList = validateFloorRules(floorData.ModelGroups[0].Schema, floorData.ModelGroups[0].Schema.Delimiter, floorData.ModelGroups[0].Values)
	if len(floorData.ModelGroups[0].Values) > 0 {
		for i := 0; i < len(request.Imp); i++ {
			desiredRuleKey := createRuleKey(floorData.ModelGroups[0].Schema, request, request.Imp[i])
			matchedRule := findRule(floorData.ModelGroups[0].Values, floorData.ModelGroups[0].Schema.Delimiter, desiredRuleKey, len(floorData.ModelGroups[0].Schema.Fields))

			floorVal = floorData.ModelGroups[0].Default
			if matchedRule != "" {
				floorVal = floorData.ModelGroups[0].Values[matchedRule]
			}

			request.Imp[i].BidFloor = floorVal
			if floorExt.FloorMin != 0.0 && floorVal < floorExt.FloorMin {
				request.Imp[i].BidFloor = floorExt.FloorMin
			}
			request.Imp[i].BidFloorCur = "USD"

			updateImpExtWithFloorDetails(matchedRule, &request.Imp[i])
		}
	}
	floorModelErrList = append(floorModelErrList, floorErrList...)
	return floorModelErrList
}

func updateImpExtWithFloorDetails(matchedRule string, imp *openrtb2.Imp) {
	imp.Ext, _ = jsonparser.Set(imp.Ext, []byte(`"`+matchedRule+`"`), "prebid", "floors", "floorRule")
	imp.Ext, _ = jsonparser.Set(imp.Ext, []byte(fmt.Sprintf("%.4f", imp.BidFloor)), "prebid", "floors", "floorRuleValue")
}

func selectFloorModelGroup(modelGroups []openrtb_ext.PriceFloorModelGroup, f func(int) int) {
	totalModelWeight := 0

	for i := 0; i < len(modelGroups); i++ {
		totalModelWeight += modelGroups[i].ModelWeight
	}

	sort.SliceStable(modelGroups, func(i, j int) bool {
		return modelGroups[i].ModelWeight < modelGroups[j].ModelWeight
	})

	winWeight := f(totalModelWeight + 1)
	for i, modelGroup := range modelGroups {
		winWeight -= modelGroup.ModelWeight
		if winWeight <= 0 {
			modelGroups[0], modelGroups[i] = modelGroups[i], modelGroups[0]
			return
		}
	}
}
