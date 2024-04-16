package bidderparams

import (
	"errors"
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

func PrepareAdapterParamsV25(rctx models.RequestCtx, cache cache.Cache, bidRequest openrtb2.BidRequest, imp openrtb2.Imp, impExt models.ImpExtension, partnerID int) (string, string, bool, []byte, error) {
	partnerConfig, ok := rctx.PartnerConfigMap[partnerID]
	if !ok {
		return "", "", false, nil, errors.New("ErrBidderParamsValidationError")
	}

	var isRegexSlot bool
	var matchedSlot, matchedPattern string

	isRegexKGP := rctx.PartnerConfigMap[partnerID][models.KEY_GEN_PATTERN] == models.REGEX_KGP
	slots, slotMap, slotMappingInfo, hw := getSlotMeta(rctx, cache, bidRequest, imp, impExt, partnerID)

	for i, slot := range slots {
		matchedSlot, matchedPattern = GetMatchingSlot(rctx, cache, slot, slotMap, slotMappingInfo, isRegexKGP, partnerID)
		if matchedSlot == "" {
			continue
		}

		slotMappingObj, ok := slotMap[strings.ToLower(matchedSlot)]
		if !ok {
			slotMappingObj = slotMap[strings.ToLower(matchedPattern)]
			isRegexSlot = true
		}

		bidderParams := make(map[string]interface{}, len(slotMappingObj.SlotMappings))
		for k, v := range slotMappingObj.SlotMappings {
			bidderParams[k] = v
		}

		for key, value := range partnerConfig {
			if !ignoreKeys[key] {
				bidderParams[key] = value
			}
		}

		h := hw[i][0]
		w := hw[i][1]
		params, err := adapters.PrepareBidParamJSONForPartner(&w, &h, bidderParams, slot, partnerConfig[models.PREBID_PARTNER_NAME], partnerConfig[models.BidderCode], &impExt)
		if err != nil || params == nil {
			continue
		}
		return matchedSlot, matchedPattern, isRegexSlot, params, nil
	}

	return "", "", false, nil, nil
}
