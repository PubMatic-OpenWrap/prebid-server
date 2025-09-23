package bidderparams

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func PrepareVASTBidderParams(rctx models.RequestCtx, cache cache.Cache, imp openrtb2.Imp, impExt models.ImpExtension, partnerID int) (string, json.RawMessage, []string, error) {
	if imp.Video == nil {
		return "", nil, nil, nil
	}

	slots, slotMap, _, _ := getSlotMeta(rctx, cache, imp, impExt, partnerID)
	if len(slots) == 0 {
		return "", nil, nil, nil
	}

	pubVASTTags := cache.GetPublisherVASTTagsFromCache(rctx.PubID)
	if len(pubVASTTags) == 0 {
		return "", nil, nil, nil
	}

	matchedSlotKeys, err := getVASTBidderSlotKeys(&imp, slots[0], slotMap, pubVASTTags, rctx.AdpodCtx)
	if len(matchedSlotKeys) == 0 {
		return "", nil, nil, err
	}

	// NYC_TODO:
	//setting flagmap
	// bidderWrapper := &BidderWrapper{VASTagFlags: make(map[string]bool)}
	// for _, key := range matchedSlotKeys {
	// 	bidderWrapper.VASTagFlags[key] = false
	// }
	// impWrapper.Bidder[bidderCode] = bidderWrapper
	var bidParams json.RawMessage
	if imp.Video != nil {
		bidParams = adapters.PrepareVASTBidderParamJSON(pubVASTTags, matchedSlotKeys, slotMap)
	} else {
		bidParams = nil
	}

	/*
		Sample Values
			//slotkey:/15671365/DMDemo1@com.pubmatic.openbid.app@
			//slotKeys:[/15671365/DMDemo1@com.pubmatic.openbid.app@101]
			//slotMap:map[/15671365/DMDemo1@com.pubmatic.openbid.app@101:map[param1:6005 param2:test param3:example]]
			//Ext:{"tags":[{"tagid":"101","url":"sample_url_1","dur":15,"price":"15","params":{"param1":"6005","param2":"test","param3":"example"}}]}
	*/
	return slots[0], bidParams, matchedSlotKeys, nil
}

// getVASTBidderSlotKeys returns all slot keys which are matching to vast tag slot key
func getVASTBidderSlotKeys(imp *openrtb2.Imp,
	slotKey string,
	slotMap map[string]models.SlotMapping,
	pubVASTTags models.PublisherVASTTags,
	adpodCtx models.AdpodCtx) ([]string, error) {

	//TODO: Optimize this function
	var (
		result, defaultMapping []string
		validationErr          error
		isValidationError      bool
	)

	for _, sm := range slotMap {
		//formCaseInsensitiveVASTSlotKey forms slotKey for case in-sensitive comparison.
		//It converts AdUnit part of slot key to lower case for comparison.
		//Currently it only supports slot-keys of the format <AdUnit>@<bundle-id>@<VAST ID>
		//This function needs to be enhanced to support different slot-key formats.
		key := formCaseInsensitiveVASTSlotKey(sm.SlotName)
		tempSlotKey := formCaseInsensitiveVASTSlotKey(slotKey)
		isDefaultMappingSelected := false

		index := strings.Index(key, "@@")
		if index != -1 {
			//prefix check only for `/15671365/MG_VideoAdUnit@`
			if !strings.HasPrefix(tempSlotKey, key[:index+1]) {
				continue
			}

			//getting slot key `/15671365/MG_VideoAdUnit@@`
			tempSlotKey = key[:index+2]
			isDefaultMappingSelected = true
		} else if !strings.HasPrefix(key, tempSlotKey) {
			continue
		}

		//get vast tag id and slotkey
		vastTagID, _ := strconv.Atoi(key[len(tempSlotKey):])
		if vastTagID == 0 {
			continue
		}

		//check pubvasttag details
		vastTag, ok := pubVASTTags[vastTagID]
		if !ok {
			continue
		}

		podId := imp.Video.PodID
		if podId == "" {
			podId = imp.ID
		}

		var podDur int64
		adpodConfig, ok := adpodCtx[podId]
		if ok {
			for _, slot := range adpodConfig.Slots {
				if slot.Flexible {
					podDur = slot.PodDur
				}
			}
		}

		//validate vast tag details
		if err := validateVASTTag(vastTag, imp.Video.MinDuration, imp.Video.MaxDuration, podDur); err != nil {
			isValidationError = true
			continue
		}

		if isDefaultMappingSelected {
			defaultMapping = append(defaultMapping, sm.SlotName)
		} else {
			result = append(result, sm.SlotName)
		}
	}

	if len(result) == 0 && len(defaultMapping) == 0 && isValidationError {
		validationErr = errors.New("ErrInvalidVastTag")
	}

	if len(result) == 0 {
		return defaultMapping, validationErr
	}

	return result, validationErr
}

// formCaseInsensitiveVASTSlotKey forms slotKey for case in-sensitive comparison.
// It converts AdUnit part of slot key to lower case for comparison.
// Currently it only supports slot-keys of the format <AdUnit>@<bundle-id>@<VAST ID>
// This function needs to be enhanced to support different slot-key formats.
func formCaseInsensitiveVASTSlotKey(key string) string {
	index := strings.Index(key, "@")
	caseInsensitiveSlotKey := key
	if index != -1 {
		caseInsensitiveSlotKey = strings.ToLower(key[:index]) + key[index:]
	}
	return caseInsensitiveSlotKey
}

func validateVASTTag(
	vastTag *models.VASTTag,
	videoMinDuration, videoMaxDuration int64,
	podDur int64) error {

	if vastTag == nil {
		return fmt.Errorf("Empty vast tag")
	}

	//TODO: adding checks for Duration and URL
	if len(vastTag.URL) == 0 {
		return fmt.Errorf("VAST tag mandatory parameter 'url' missing: %v", vastTag.ID)
	}

	if vastTag.Duration <= 0 {
		return fmt.Errorf("VAST tag mandatory parameter 'duration' missing: %v", vastTag.ID)
	}

	if videoMaxDuration != 0 && vastTag.Duration > int(videoMaxDuration) {
		return fmt.Errorf("VAST tag 'duration' validation failed 'tag.duration > video.maxduration' vastTagID:%v, tag.duration:%v, video.maxduration:%v", vastTag.ID, vastTag.Duration, videoMaxDuration)
	}

	if podDur == 0 {
		//non-adpod request
		if videoMinDuration != 0 && vastTag.Duration < int(videoMinDuration) {
			return fmt.Errorf("VAST tag 'duration' validation failed 'tag.duration < video.minduration' vastTagID:%v, tag.duration:%v, video.minduration:%v", vastTag.ID, vastTag.Duration, videoMinDuration)
		}

	} else {
		//adpod request
		if vastTag.Duration > int(podDur) {
			return fmt.Errorf("VAST tag 'duration' validation failed 'tag.duration > imp.video.PodDur' vastTagID:%v, tag.duration:%v, imp.video.PodDur:%v", vastTag.ID, vastTag.Duration, podDur)
		}
	}

	return nil
}
