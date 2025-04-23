package pubmatic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/adapters"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

const (
	dsaKey                       = "dsa"
	transparencyKey              = "transparency"
	multiFloors                  = "_mf"
	appLovinMaxImpressionPattern = `_mf[0-9]+$`
	multiBidMultiFloorValueKey   = "mbmfv"
	billingId                    = "billing_id"
	publisherSettingListId       = "publisher_setting_list_id"
	excludedCreatives            = "excluded_creatives"
	isAppOpenAd                  = "is_app_open_ad"
	allowedRestrictedCategory    = "allowed_restricted_category"
	creativeEnforcementSettings  = "creative_enforcement_settings"
)

var (
	paramKey    = []byte(`"params"`)
	dsaParamKey = []byte(`"dsaparams"`)
)

var appLovinMaxImpressionRegex = regexp.MustCompile(appLovinMaxImpressionPattern)

func getTargetingKeys(bidExt json.RawMessage, bidderName string) map[string]string {
	targets := map[string]string{}
	if bidExt != nil {
		bidExtMap := make(map[string]interface{})
		err := json.Unmarshal(bidExt, &bidExtMap)
		if err == nil && bidExtMap[buyId] != nil {
			targets[buyIdTargetingKey+bidderName], _ = bidExtMap[buyId].(string)
		}
	}
	return targets
}

func copySBExtToBidExt(sbExt json.RawMessage, bidExt json.RawMessage) json.RawMessage {
	if sbExt != nil {
		sbExtMap := getMapFromJSON(sbExt)
		bidExtMap := make(map[string]interface{})
		if bidExt != nil {
			bidExtMap = getMapFromJSON(bidExt)
		}
		if bidExtMap != nil && sbExtMap != nil {
			if sbExtMap[buyId] != nil && bidExtMap[buyId] == nil {
				bidExtMap[buyId] = sbExtMap[buyId]
			}
		}
		byteAra, _ := json.Marshal(bidExtMap)
		return json.RawMessage(byteAra)
	}
	return bidExt
}

// prepareMetaObject prepares the Meta structure using Bid Response
func prepareMetaObject(bid openrtb2.Bid, bidExt *pubmaticBidExt, seat string) *openrtb_ext.ExtBidPrebidMeta {

	meta := &openrtb_ext.ExtBidPrebidMeta{
		NetworkID:    bidExt.DspId,
		AdvertiserID: bidExt.AdvertiserID,
		MediaType:    string(getBidType(bidExt)),
	}

	if meta.NetworkID != 0 {
		meta.DemandSource = strconv.Itoa(meta.NetworkID)
	}

	if len(seat) > 0 {
		meta.AdvertiserID, _ = strconv.Atoi(seat)
	}

	meta.AgencyID = meta.AdvertiserID

	if len(bid.Cat) > 0 {
		meta.PrimaryCategoryID = bid.Cat[0]
		meta.SecondaryCategoryIDs = bid.Cat
	}

	// NOTE: We will not recieve below fields from the translator response also not sure on what will be the key names for these in the response,
	// when we needed we can add it back.
	// New fields added, assignee fields name may change
	// Assign meta.BrandId to bidExt.ADomain[0]  //BrandID is of Type int and ADomain values if string type like "mystartab.com"
	// meta.NetworkName = bidExt.NetworkName;
	// meta.AdvertiserName = bidExt.AdvertiserName;
	// meta.AgencyName = bidExt.AgencyName;
	// meta.BrandName = bidExt.BrandName;
	// meta.DChain = bidExt.DChain;

	return meta
}

// renameTransparencyParamsKey renames the bid.ext.dsa.transparency.params key to bid.ext.dsa.transparency.dsaparams
func renameTransparencyParamsKey(bidExt []byte) []byte {
	transparencyObjectCnt := 0
	jsonparser.ArrayEach(bidExt, func(transparencyObject []byte, dataType jsonparser.ValueType, offset int, err error) {
		transparencyObject = bytes.Replace(transparencyObject, paramKey, dsaParamKey, 1)
		bidExt, err = jsonparser.Set(bidExt, transparencyObject, dsaKey, transparencyKey, fmt.Sprintf("[%d]", transparencyObjectCnt))
		if err != nil {
			return
		}
		transparencyObjectCnt++
	}, dsaKey, transparencyKey)

	return bidExt
}

// buildMultiFloorRequests builds multiple requests for each floor value
func (a *PubmaticAdapter) buildMultiFloorRequests(request *openrtb2.BidRequest, impFloorsMap map[string][]float64, cookies []string) ([]*adapters.RequestData, []error) {
	requestData := make([]*adapters.RequestData, 0, MAX_MULTIFLOORS_PUBMATIC*len(request.Imp))
	errs := make([]error, 0, MAX_MULTIFLOORS_PUBMATIC*len(request.Imp))

	for i := 0; i < MAX_MULTIFLOORS_PUBMATIC; i++ {
		isFloorsUpdated := false
		newImps := make([]openrtb2.Imp, len(request.Imp))
		copy(newImps, request.Imp)
		//TODO-AK: Remove the imp from the request if the floor is not present except for the first floor
		for j := range newImps {
			floors, ok := impFloorsMap[request.Imp[j].ID]
			if !ok || len(floors) <= i {
				continue
			}
			isFloorsUpdated = true
			newImps[j].BidFloor = floors[i]
			newImps[j].ID = fmt.Sprintf("%s"+multiFloors+"%d", newImps[j].ID, i+1)
		}

		if !isFloorsUpdated {
			continue
		}

		newRequest := *request
		newRequest.Imp = newImps

		newRequestData, errData := a.buildAdapterRequest(&newRequest, cookies)
		if errData != nil {
			errs = append(errs, errData)
		}
		if len(newRequestData) > 0 {
			requestData = append(requestData, newRequestData...)
		}
	}
	return requestData, errs
}

func trimSuffixWithPattern(input string) string {
	return appLovinMaxImpressionRegex.ReplaceAllString(input, "")
}

func updateBidExtWithMultiFloor(bidImpID string, bidExt, reqBody []byte) []byte {
	var externalRequest openrtb2.BidRequest

	if err := json.Unmarshal(reqBody, &externalRequest); err != nil {
		return bidExt
	}

	updatedBidExt := bidExt
	if bidExt == nil {
		updatedBidExt = json.RawMessage(`{}`)
	}

	for _, imp := range externalRequest.Imp {
		if imp.ID != bidImpID {
			continue
		}
		if imp.BidFloor <= 0 {
			continue
		}
		var err error
		updatedBidExt, err = jsonparser.Set(updatedBidExt, []byte(fmt.Sprintf("%f", imp.BidFloor)), multiBidMultiFloorValueKey)
		if err != nil {
			return bidExt
		}
	}

	if string(updatedBidExt) != "{}" {
		return updatedBidExt
	}
	return bidExt
}

func addGoogleSDKParamsToBidExt(bidExtMap map[string]interface{}, bidderExt ExtImpBidderPubmatic) {
	// Google SDK Params
	if len(bidderExt.BillingIds) > 0 {
		bidExtMap[billingId] = bidderExt.BillingIds
	}
	if len(bidderExt.PublisherSettingListIds) > 0 {
		bidExtMap[publisherSettingListId] = bidderExt.PublisherSettingListIds
	}
	if len(bidderExt.ExcludedCreatives) > 0 {
		bidExtMap[excludedCreatives] = bidderExt.ExcludedCreatives
	}
	if bidderExt.IsAppOpenAd {
		bidExtMap[isAppOpenAd] = bidderExt.IsAppOpenAd
	}
	if bidderExt.AllowedRestrictedCategory != 0 {
		bidExtMap[allowedRestrictedCategory] = bidderExt.AllowedRestrictedCategory
	}
	if bidderExt.CreativeEnforcementSettings != nil {
		bidExtMap[creativeEnforcementSettings] = bidderExt.CreativeEnforcementSettings
	}
}
