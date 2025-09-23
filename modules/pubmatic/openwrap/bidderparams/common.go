package bidderparams

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

var ignoreKeys = map[string]bool{
	models.PARTNER_ACCOUNT_NAME: true,
	models.ADAPTER_NAME:         true,
	models.ADAPTER_ID:           true,
	models.TIMEOUT:              true,
	models.KEY_GEN_PATTERN:      true,
	models.PREBID_PARTNER_NAME:  true,
	models.PROTOCOL:             true,
	models.SERVER_SIDE_FLAG:     true,
	models.LEVEL:                true,
	models.PARTNER_ID:           true,
	models.REVSHARE:             true,
	models.THROTTLE:             true,
	models.BidderCode:           true,
	models.IsAlias:              true,
	models.VENDORID:             true,
	models.BidderFilters:        true,
}

func getSlotMeta(rctx models.RequestCtx, cache cache.Cache, imp openrtb2.Imp, impExt models.ImpExtension, partnerID int) ([]string, map[string]models.SlotMapping, models.SlotMappingInfo, [][2]int64) {
	var slotMap map[string]models.SlotMapping
	var slotMappingInfo models.SlotMappingInfo

	//don't read mappings from cache in case of test=2
	if !(rctx.IsTestRequest == models.TestValueTwo && rctx.PartnerConfigMap[partnerID][models.BidderCode] == models.BidderPubMatic) {
		slotMap = cache.GetMappingsFromCacheV25(rctx, partnerID)
		if slotMap == nil {
			return nil, nil, models.SlotMappingInfo{}, nil
		}
		slotMappingInfo = cache.GetSlotToHashValueMapFromCacheV25(rctx, partnerID)
		if len(slotMappingInfo.OrderedSlotList) == 0 {
			return nil, nil, models.SlotMappingInfo{}, nil
		}
	}

	var hw [][2]int64
	if imp.Banner != nil {
		if imp.Banner.W != nil && imp.Banner.H != nil {
			hw = append(hw, [2]int64{*imp.Banner.H, *imp.Banner.W})
		}

		for _, format := range imp.Banner.Format {
			hw = append(hw, [2]int64{format.H, format.W})
		}
	}

	if imp.Video != nil {
		hw = append(hw, [2]int64{0, 0})
	}

	if imp.Native != nil {
		hw = append(hw, [2]int64{1, 1})
	}

	kgp := rctx.PartnerConfigMap[partnerID][models.KEY_GEN_PATTERN]

	var div string
	if impExt.Wrapper != nil {
		div = impExt.Wrapper.Div
	}

	var slots []string
	for _, format := range hw {
		// TODO fix the param sequence. make it consistent. HxW
		slot := models.GenerateSlotName(format[0], format[1], kgp, imp.TagID, div, rctx.Source)
		if slot != "" {
			slots = append(slots, slot)
			// NYC_TODO: break at i=0 for pubmatic?
		}
	}

	// NYC_TODO wh is returned temporarily
	return slots, slotMap, slotMappingInfo, hw
}

/*
formSlotForDefaultMapping: In this method, we are removing wxh from the kgp because
pubmatic adapter sets wxh that we send in imp.ext.pubmatic.adslot as primary size while calling translator.
In case of default mappings, since all sizes are unmapped, we don't want to treat any size as primary
thats why we are removing size from kgp
*/
func getDefaultMappingKGP(keyGenPattern string) string {
	if strings.Contains(keyGenPattern, "@_W_x_H_") {
		return strings.ReplaceAll(keyGenPattern, "@_W_x_H_", "")
	}
	return keyGenPattern
}

// getSlotMappings will returns slotMapping from map based on slotKey
func getSlotMappings(matchedSlot, matchedPattern string, slotMap map[string]models.SlotMapping) map[string]interface{} {
	slotKey := matchedSlot
	if matchedPattern != "" {
		slotKey = matchedPattern
	}

	if slotMappingObj, ok := slotMap[strings.ToLower(slotKey)]; ok {
		return slotMappingObj.SlotMappings
	}

	return nil
}

func GetMatchingSlot(rctx models.RequestCtx, cache cache.Cache, slot string, slotMap map[string]models.SlotMapping, slotMappingInfo models.SlotMappingInfo, isRegexKGP bool, partnerID int) (string, string) {
	if _, ok := slotMap[strings.ToLower(slot)]; ok {
		return slot, ""
	}

	if isRegexKGP {
		if matchedSlot, regexPattern := GetRegexMatchingSlot(rctx, cache, slot, slotMap, slotMappingInfo, partnerID); matchedSlot != "" {
			return matchedSlot, regexPattern
		}
	}

	return "", ""
}

const pubSlotRegex = "psregex_%d_%d_%d_%d_%s" // slot and its matching regex info at publisher, profile, display version and adapter level

type regexSlotEntry struct {
	SlotName     string
	RegexPattern string
}

// TODO: handle this db injection correctly
func GetRegexMatchingSlot(rctx models.RequestCtx, cache cache.Cache, slot string, slotMap map[string]models.SlotMapping, slotMappingInfo models.SlotMappingInfo, partnerID int) (string, string) {
	// Ex. "psregex_5890_56777_1_8_/43743431/DMDemo1@@728x90"
	cacheKey := fmt.Sprintf(pubSlotRegex, rctx.PubID, rctx.ProfileID, rctx.DisplayID, partnerID, slot)
	if v, ok := cache.Get(cacheKey); ok {
		if rse, ok := v.(regexSlotEntry); ok {
			return rse.SlotName, rse.RegexPattern
		}
	}

	//Flags passed to regexp.Compile
	regexFlags := "(?i)" // case in-sensitive match

	// if matching regex is not found in cache, run checks for the regex patterns in DB
	for _, slotname := range slotMappingInfo.OrderedSlotList {
		slotnameMatched := false
		dbSlotNameParts := strings.Split(slotname, "@")
		requestSlotKeyParts := strings.Split(slot, "@")
		if len(dbSlotNameParts) == len(requestSlotKeyParts) {
			for i, dbPart := range dbSlotNameParts {
				re, err := regexp.Compile(regexFlags + dbPart)
				if err != nil {
					// If an invalid regex pattern is encountered, check further entries intead of returning immediately
					break
				}
				matchingPart := re.FindString(requestSlotKeyParts[i])
				if matchingPart == "" && requestSlotKeyParts[i] != "" {
					// request slot key did not match the Regex pattern
					// check the next regex pattern from the DB
					break
				}
				if i == len(dbSlotNameParts)-1 {
					slotnameMatched = true
				}
			}
		}

		if slotnameMatched {
			cache.Set(cacheKey, regexSlotEntry{SlotName: slot, RegexPattern: slotname})
			return slot, slotname
		}
	}

	return "", ""
}
