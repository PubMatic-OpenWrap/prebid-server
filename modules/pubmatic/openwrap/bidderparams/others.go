package bidderparams

import (
	"errors"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func PrepareAdapterParamsV25(rctx models.RequestCtx, cache cache.Cache, bidRequest openrtb2.BidRequest, imp openrtb2.Imp, impExt models.ImpExtension, partnerID int, begin time.Time, prebidBidderCode string) (string, string, bool, []byte, error) {
	start := time.Now()
	stageDur := make(map[string]int64)

	partnerConfig, ok := rctx.PartnerConfigMap[partnerID]
	if !ok {
		return "", "", false, nil, errors.New("ErrBidderParamsValidationError")
	}

	var isRegexSlot, isRegexKGP bool
	var matchedSlot, matchedPattern string

	if kgp := rctx.PartnerConfigMap[partnerID][models.KEY_GEN_PATTERN]; kgp == models.REGEX_KGP || kgp == models.ADUNIT_SIZE_REGEX_KGP {
		isRegexKGP = true
	}
	t := time.Now()
	label := getLabel("getSlotMeta_PrepareAdapterParamsV25", imp.ID, prebidBidderCode, "")
	slots, slotMap, slotMappingInfo, hw := getSlotMeta(rctx, cache, bidRequest, imp, impExt, partnerID)
	stageDur[label] = time.Since(t).Milliseconds()
	timing(label, bidRequest.ID, imp.ID, t, begin)

	if len(slots) == 0 || slotMap == nil {
		return "", "", false, nil, nil
	}

	for i, slot := range slots {
		glog.V(3).Infof("PrepareAdapterParamsV25: slot: %v", slot)
		t = time.Now()
		label = getLabel("getMatchingSlot_PrepareAdapterParamsV25_slot", imp.ID, prebidBidderCode, slot)
		matchedSlot, matchedPattern = GetMatchingSlot(rctx, cache, slot, slotMap, slotMappingInfo, isRegexKGP, partnerID)
		stageDur[label] = time.Since(t).Milliseconds()
		timing(label, bidRequest.ID, imp.ID, t, begin)
		if matchedSlot == "" {
			continue
		}

		lowerSlot := strings.ToLower(matchedSlot)
		slotMappingObj, ok := slotMap[lowerSlot]
		if !ok {
			slotMappingObj = slotMap[strings.ToLower(matchedPattern)]
			isRegexSlot = true
		}

		bidderParams := make(map[string]interface{}, len(slotMappingObj.SlotMappings)+len(partnerConfig))
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
		t = time.Now()
		label = getLabel("prepareBidParamJSONForPartner_PrepareAdapterParamsV25_slot", imp.ID, prebidBidderCode, slot)
		params, err := adapters.PrepareBidParamJSONForPartner(&w, &h, bidderParams, slot, partnerConfig[models.PREBID_PARTNER_NAME], partnerConfig[models.BidderCode], &impExt)
		timing(label, bidRequest.ID, imp.ID, t, begin)
		stageDur[label] = time.Since(t).Milliseconds()
		if err != nil || params == nil {
			continue
		}
		total := time.Since(begin).Milliseconds()
		end := time.Since(start).Milliseconds()
		glog.Infof("[PrepareAdapterParamsV25] req:%s total from beforevalidation:%dms total prepareadapterparams:%dms stages:%v", bidRequest.ID, total, end, stageDur)
		return matchedSlot, matchedPattern, isRegexSlot, params, nil
	}
	total := time.Since(begin).Milliseconds()
	end := time.Since(start).Milliseconds()
	glog.Infof("[PrepareAdapterParamsV25] req:%s total from beforevalidation:%dms total prepareadapterparams:%dms stages:%v", bidRequest.ID, total, end, stageDur)
	return "", "", false, nil, nil
}
