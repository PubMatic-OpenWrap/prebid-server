package bidderparams

import (
	"encoding/json"
	"fmt"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

func PreparePubMaticParamsV25(rctx models.RequestCtx, cache cache.Cache, imp openrtb2.Imp, impExt models.ImpExtension, partnerID int) (string, string, bool, []byte, error) {
	extImpPubMatic := openrtb_ext.ExtImpPubmatic{
		PublisherId: getPubMaticPublisherID(rctx, partnerID),
		WrapExt:     getPubMaticWrapperExt(rctx, partnerID),
		Keywords:    getImpExtPubMaticKeyWords(impExt, rctx.PartnerConfigMap[partnerID][models.BidderCode]),
		Floors:      models.GetMultiFloors(rctx.MultiFloors, imp.ID),
		OWSDK:       impExt.OWSDK,
	}

	slots, slotMap, slotMappingInfo, _ := getSlotMeta(rctx, cache, imp, impExt, partnerID)

	var err error
	var matchedSlot, matchedPattern string
	var isRegexSlot, isRegexKGP bool

	kgp := rctx.PartnerConfigMap[partnerID][models.KEY_GEN_PATTERN]
	if kgp == models.REGEX_KGP || kgp == models.ADUNIT_SIZE_REGEX_KGP {
		isRegexKGP = true
	}

	if rctx.IsTestRequest == models.TestValueTwo {
		if len(slots) > 0 {
			extImpPubMatic.AdSlot = slots[0]
		}
		params, err := json.Marshal(extImpPubMatic)
		return extImpPubMatic.AdSlot, "", false, params, err
	}

	// simple+regex key match
	matchedSlot, matchedPattern, isRegexSlot = getMatchingSlotAndPattern(rctx, cache, slots, slotMap, slotMappingInfo, isRegexKGP, isRegexSlot, partnerID, &extImpPubMatic, imp)

	if paramMap := getSlotMappings(matchedSlot, matchedPattern, slotMap); paramMap != nil {
		if matchedPattern == "" {
			// use alternate names defined in DB for this slot if selection is non-regex
			// use owSlotName to addres case insensitive slotname.
			// Ex: slot="/43743431/DMDEMO@300x250" and owSlotName="/43743431/DMDemo@300x250"
			if v, ok := paramMap[models.KEY_OW_SLOT_NAME]; ok {
				if owSlotName, ok := v.(string); ok {
					extImpPubMatic.AdSlot = owSlotName
				}
			}
		}

		// Update slot key for PubMatic secondary flow
		if v, ok := paramMap[models.KEY_SLOT_NAME]; ok {
			if secondarySlotName, ok := v.(string); ok {
				extImpPubMatic.AdSlot = secondarySlotName
			}
		}
	}

	// last resort: send slotname w/o size to translator
	if extImpPubMatic.AdSlot == "" {
		var div string
		if impExt.Wrapper != nil {
			div = impExt.Wrapper.Div
		}
		unmappedKPG := getDefaultMappingKGP(kgp)
		extImpPubMatic.AdSlot = models.GenerateSlotName(0, 0, unmappedKPG, imp.TagID, div, rctx.Source)
		if len(slots) != 0 { // reuse this field for wt and wl in combination with isRegex
			matchedPattern = slots[0]
		}
	}

	params, err := json.Marshal(extImpPubMatic)
	return matchedSlot, matchedPattern, isRegexSlot, params, err
}

func getDealTier(impExt models.ImpExtension, bidderCode string) *openrtb_ext.DealTier {
	if len(impExt.Bidder) != 0 {
		if bidderExt, ok := impExt.Bidder[bidderCode]; ok && bidderExt != nil && bidderExt.DealTier != nil {
			return bidderExt.DealTier
		}
	}
	return nil
}

func getImpExtPubMaticKeyWords(impExt models.ImpExtension, bidderCode string) []*openrtb_ext.ExtImpPubmaticKeyVal {
	if len(impExt.Bidder) != 0 {
		if bidderExt, ok := impExt.Bidder[bidderCode]; ok && bidderExt != nil && len(bidderExt.KeyWords) != 0 {
			keywords := make([]*openrtb_ext.ExtImpPubmaticKeyVal, 0)
			for _, keyVal := range bidderExt.KeyWords {
				//ignore key values pair with no values
				if len(keyVal.Values) == 0 {
					continue
				}
				keyValPair := openrtb_ext.ExtImpPubmaticKeyVal{
					Key:    keyVal.Key,
					Values: keyVal.Values,
				}
				keywords = append(keywords, &keyValPair)
			}
			if len(keywords) != 0 {
				return keywords
			}
		}
	}
	return nil
}

func getMatchingSlotAndPattern(rctx models.RequestCtx, cache cache.Cache, slots []string, slotMap map[string]models.SlotMapping, slotMappingInfo models.SlotMappingInfo, isRegexKGP, isRegexSlot bool, partnerID int, extImpPubMatic *openrtb_ext.ExtImpPubmatic, imp openrtb2.Imp) (string, string, bool) {

	hash := ""
	var matchedSlot, matchedPattern string
	for _, slot := range slots {
		matchedSlot, matchedPattern = GetMatchingSlot(rctx, cache, slot, slotMap, slotMappingInfo, isRegexKGP, partnerID)
		if matchedSlot != "" {
			extImpPubMatic.AdSlot = matchedSlot

			if matchedPattern != "" {
				isRegexSlot = true
				// imp.TagID = hash
				// TODO: handle kgpv case sensitivity in hashvaluemap
				if slotMappingInfo.HashValueMap != nil {
					if v, ok := slotMappingInfo.HashValueMap[matchedPattern]; ok {
						extImpPubMatic.AdSlot = v
						imp.TagID = hash // TODO, make imp pointer. But do other bidders accept hash as TagID?
					}
				}
			}
			break
		}
	}
	return matchedSlot, matchedPattern, isRegexSlot
}

func getPubMaticPublisherID(rctx models.RequestCtx, partnerID int) string {
	//Pubmatic secondary flow send account level pubID
	if pubID, ok := rctx.PartnerConfigMap[partnerID][models.KEY_PUBLISHER_ID]; ok {
		return pubID
	}
	//PubMatic alias flow
	if pubID, ok := rctx.PartnerConfigMap[partnerID][models.PubID]; ok {
		return pubID
	}
	//PubMatic primary flow
	return rctx.PubIDStr
}

func getPubMaticWrapperExt(rctx models.RequestCtx, partnerID int) json.RawMessage {
	wrapExt := fmt.Sprintf(`{"%s":%d,"%s":%d}`, models.SS_PM_VERSION_ID, rctx.DisplayID, models.SS_PM_PROFILE_ID, rctx.ProfileID)

	// change profile id for pubmatic2
	if secondaryProfileID, ok := rctx.PartnerConfigMap[partnerID][models.KEY_PROFILE_ID]; ok {
		wrapExt = fmt.Sprintf(`{"%s":%s}`, models.SS_PM_PROFILE_ID, secondaryProfileID)
	}

	//Pubmatic alias flow do not send wrapExt
	if isAlias, ok := rctx.PartnerConfigMap[partnerID][models.IsAlias]; ok && isAlias == "1" {
		if pubID, ok := rctx.PartnerConfigMap[partnerID][models.PubID]; ok && pubID != rctx.PubIDStr {
			return nil
		}
	}
	return json.RawMessage(wrapExt)
}
