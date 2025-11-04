package adapters

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"

	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

func PrepareVASTBidderParamJSON(pubVASTTags models.PublisherVASTTags, matchedSlotKeys []string, slotMap map[string]models.SlotMapping) json.RawMessage {
	bidderExt := openrtb_ext.ExtImpVASTBidder{}
	bidderExt.Tags = make([]*openrtb_ext.ExtImpVASTBidderTag, len(matchedSlotKeys))
	var tagIndex int = 0
	for _, slotKey := range matchedSlotKeys {
		vastTagID := getVASTTagID(slotKey)
		if vastTagID == 0 {
			continue
		}

		vastTag, ok := pubVASTTags[vastTagID]
		if !ok {
			continue
		}

		slotMappingObj, ok := slotMap[strings.ToLower(slotKey)]
		if !ok {
			continue
		}

		mapping := slotMappingObj.SlotMappings

		//adding mapping parameters as it is in ext.bidder
		params := mapping
		/*
			params := make(map[string]interface{})
			// Copy from the original map of for slot key to the target map
			for key, value := range mapping {
				params[key] = value
			}
		*/

		//prepare bidder ext json here
		bidderExt.Tags[tagIndex] = &openrtb_ext.ExtImpVASTBidderTag{
			//TagID:    strconv.Itoa(vastTag.ID),
			TagID:    slotKey,
			URL:      vastTag.URL,
			Duration: vastTag.Duration,
			Price:    vastTag.Price,
			Params:   params,
		}
		tagIndex++
	}

	if tagIndex > 0 {
		//If any vast tags found then create impression ext for vast bidder.
		bidderExt.Tags = bidderExt.Tags[:tagIndex]
		bidParamBuf, _ := json.Marshal(bidderExt)
		return bidParamBuf
	}
	return nil
}

// getVASTTagID returns VASTTag ID details from slot key
func getVASTTagID(key string) int {
	index := strings.LastIndex(key, "@")
	if index == -1 {
		return 0
	}
	id, _ := strconv.Atoi(key[index+1:])
	return id
}

func FilterImpsVastTagsByDuration(rCtx models.RequestCtx, request *openrtb_ext.RequestWrapper) {
	for _, imp := range request.GetImp() {
		// Decode Imp ID
		_, impId, _ := utils.DecodeV25ImpID(imp.ID)

		impCtx, ok := rCtx.ImpBidCtx[impId]
		if !ok {
			continue
		}

		impExt, err := imp.GetImpExt()
		if err != nil {
			continue
		}

		prebidExt := impExt.GetPrebid()
		if prebidExt == nil {
			continue
		}

		minDuration := imp.Video.MinDuration
		maxDuration := imp.Video.MaxDuration

		for bidder, partnerdata := range prebidExt.Bidder {
			var vastBidderExt openrtb_ext.ExtImpVASTBidder
			if err := json.Unmarshal(partnerdata, &vastBidderExt); err != nil {
				continue
			}

			if len(vastBidderExt.Tags) == 0 {
				continue
			}

			partnerData := impCtx.Bidders[bidder]
			if partnerData.VASTTagFlags == nil {
				partnerData.VASTTagFlags = make(map[string]bool)
			}

			compatibleTags := make([]*openrtb_ext.ExtImpVASTBidderTag, 0, len(vastBidderExt.Tags))
			for _, tag := range vastBidderExt.Tags {
				if minDuration <= int64(tag.Duration) && int64(tag.Duration) <= maxDuration {
					compatibleTags = append(compatibleTags, tag)
					partnerData.VASTTagFlags[tag.TagID] = false
				}
			}

			impCtx.Bidders[bidder] = partnerData

			vastBidderExt.Tags = compatibleTags
			newPartnerData, _ := json.Marshal(vastBidderExt)
			prebidExt.Bidder[bidder] = newPartnerData
		}

		// Set the updated prebid ext back to the imp ext
		impExt.SetPrebid(prebidExt)
	}
}
