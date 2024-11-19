package pubmatic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

const (
	dsaKey                       = "dsa"
	transparencyKey              = "transparency"
	multiFloors                  = "_mf"
	appLovinMaxImpressionPattern = "_mf.*"
	multiBidMultiFloorKey        = "mbmfv"
)

var (
	paramKey    = []byte(`"params"`)
	dsaParamKey = []byte(`"dsaparams"`)
)

var re = regexp.MustCompile(appLovinMaxImpressionPattern)

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
	requestData := []*adapters.RequestData{}
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
	return re.ReplaceAllString(input, "")
}

func updateBidExtWithMultiFloor(bidImpID string, bidExt, reqBody []byte) []byte {
	reqBodyMap := getMapFromJSON(reqBody)
	var err error

	if reqBodyMap == nil {
		return bidExt
	}

	if imp, ok := reqBodyMap["imp"]; ok {
		impMap := imp.([]interface{})
		for _, impObj := range impMap {
			impObjMap := impObj.(map[string]interface{})
			if reqImpID, ok := impObjMap["id"]; ok {
				if floor, ok := impObjMap["bidfloor"]; ok && reqImpID.(string) == bidImpID {
					floorValue := floor.(float64)
					if floorValue > 0 {
						if bidExt, err = jsonparser.Set(bidExt, []byte(fmt.Sprintf("%f", floorValue)), multiBidMultiFloorKey); err != nil {
							return bidExt
						}
					}
				}
			}
		}
	}
	return bidExt
}
