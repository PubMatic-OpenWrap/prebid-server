package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

// PrepareBidParamJSONForPartner preparing bid params json for partner
func PrepareBidParamJSONForPartner(width *int64, height *int64, fieldMap map[string]interface{}, slotKey, adapterName, bidderCode string, impExt *models.ImpExtension) (json.RawMessage, error) {
	params := BidderParameters{
		AdapterName: adapterName,
		BidderCode:  bidderCode,
		ImpExt:      impExt,
		FieldMap:    fieldMap,
		Width:       width,
		Height:      height,
		SlotKey:     slotKey,
	}

	//get callback function and execute it
	callback := getBuilder(params.AdapterName)
	return callback(params)
}

// defaultBuilder for building json object for all other bidder
func defaultBuilder(params BidderParameters) (json.RawMessage, error) {
	//check if ResolveOWBidder is required or not
	params.AdapterName = ResolveOWBidder(params.AdapterName)
	return prepareBidParamJSONDefault(params)
}

// builderPubMatic for building json object for all other bidder
func builderPubMatic(params BidderParameters) (json.RawMessage, error) {
	jsonStr := bytes.Buffer{}
	jsonStr.WriteByte('{')

	//UOE-5744: Adding custom changes for hybrid profiles
	//publisherID
	publisherID, _ := getString(params.FieldMap["publisherId"])
	fmt.Fprintf(&jsonStr, `"publisherId":"%s"`, publisherID)

	//adSlot
	if adSlot, ok := getString(params.FieldMap["adSlot"]); ok {
		fmt.Fprintf(&jsonStr, `,"adSlot":"%s"`, adSlot)
	}

	//pmzoneid
	if pmzoneid, ok := getString(params.FieldMap["pmzoneid"]); ok {
		fmt.Fprintf(&jsonStr, `,"pmzoneid":"%s"`, pmzoneid)
	}

	//dctr
	if dctr, ok := getString(params.FieldMap["dctr"]); ok {
		fmt.Fprintf(&jsonStr, `,"dctr":"%s"`, dctr)
	}

	//kadfloor
	if kadfloor, ok := getString(params.FieldMap["kadfloor"]); ok {
		fmt.Fprintf(&jsonStr, `,"kadfloor":"%s"`, kadfloor)
	}

	//wrapper object
	if value, ok := params.FieldMap["wrapper"]; ok {
		if wrapper, ok := value.(map[string]interface{}); ok {
			fmt.Fprintf(&jsonStr, `,"wrapper":{`)

			//profile
			profile, _ := getInt(wrapper["profile"])
			fmt.Fprintf(&jsonStr, `"profile":%d`, profile)

			//version
			version, _ := getInt(wrapper["version"])
			fmt.Fprintf(&jsonStr, `,"version":%d`, version)

			jsonStr.WriteByte('}')
		}
	}

	//keywords
	if value, ok := params.FieldMap["keywords"]; ok {
		if keywords, err := json.Marshal(value); err == nil {
			fmt.Fprintf(&jsonStr, `,"keywords":%s`, string(keywords))
		}
	}

	//bidViewability Object
	if value, ok := params.FieldMap["bidViewability"]; ok {
		if bvsJson, err := json.Marshal(value); err == nil {
			fmt.Fprintf(&jsonStr, `,"bidViewability":%s`, string(bvsJson))
		}
	}

	jsonStr.WriteByte('}')

	return jsonStr.Bytes(), nil
}

// builderAppNexus for building json object for AppNexus bidder
func builderAppNexus(params BidderParameters) (json.RawMessage, error) {
	jsonStr := bytes.Buffer{}
	jsonStr.WriteByte('{')

	//incase if placementId not present then fallback to placement_id else log 0
	placementID, ok := getInt(params.FieldMap["placementId"])
	if !ok {
		placementID, _ = getInt(params.FieldMap["placement_id"])
	}
	fmt.Fprintf(&jsonStr, `"placementId":%d`, placementID)

	//reserve parameter
	if reserve, ok := getFloat64(params.FieldMap["reserve"]); ok {
		fmt.Fprintf(&jsonStr, `,"reserve":%.3f`, reserve)
	}

	//use_pmt_rule parameter
	usePaymentRule, ok := getBool(params.FieldMap["usePaymentRule"])
	if !ok {
		usePaymentRule, ok = getBool(params.FieldMap["use_pmt_rule"])
	}
	if ok {
		fmt.Fprintf(&jsonStr, `,"use_pmt_rule":%t`, usePaymentRule)
	}

	//anyone invcode and member
	invCode, ok := getString(params.FieldMap["invCode"])
	if !ok {
		invCode, ok = getString(params.FieldMap["inv_code"])
	}
	if ok {
		fmt.Fprintf(&jsonStr, `,"invCode":"%s"`, invCode)
	} else {
		if member, ok := getString(params.FieldMap["member"]); ok {
			fmt.Fprintf(&jsonStr, `,"member":"%s"`, member)
		}
	}

	//keywords
	if val, ok := params.FieldMap["keywords"]; ok {
		//UOE-5744: Adding custom changes for hybrid profiles
		if keywords, _ := json.Marshal(val); len(keywords) > 0 {
			fmt.Fprintf(&jsonStr, `,"keywords":%s`, string(keywords))
		}
	} else if keywords := getKeywordStringForPartner(params.ImpExt, params.BidderCode); keywords != "" {
		fmt.Fprintf(&jsonStr, `,"keywords":%s`, keywords)
	}

	//generate_ad_pod_id
	if generateAdPodID, ok := getBool(params.FieldMap["generate_ad_pod_id"]); ok {
		fmt.Fprintf(&jsonStr, `,"generate_ad_pod_id":%t`, generateAdPodID)
	}

	//other parameters
	for key, val := range params.FieldMap {
		if ignoreAppnexusKeys[key] {
			continue
		}
		if strVal, ok := getString(val); ok {
			fmt.Fprintf(&jsonStr, `,"%s":"%s"`, key, strVal)
		}
	}

	jsonStr.WriteByte('}')

	return jsonStr.Bytes(), nil
}

// builderIndex for building json object for Index bidder
func builderIndex(params BidderParameters) (json.RawMessage, error) {
	jsonStr := bytes.Buffer{}

	siteID, ok := getString(params.FieldMap["siteID"])
	if !ok {
		//UOE-5744: Adding custom changes for hybrid profiles
		if siteID, ok = getString(params.FieldMap["siteId"]); !ok {
			return nil, fmt.Errorf(errMandatoryParameterMissingFormat, params.AdapterName, "siteID")
		}
	}

	width, height := params.Width, params.Height
	if width == nil || height == nil {
		//UOE-5744: Adding custom changes for hybrid profiles
		size, ok := getIntArray(params.FieldMap["size"])
		if len(size) != 2 || !ok {
			return nil, fmt.Errorf(errMandatoryParameterMissingFormat, params.AdapterName, "size")
		}
		w := int64(size[0])
		h := int64(size[1])
		width, height = &w, &h
	}

	fmt.Fprintf(&jsonStr, `{"siteId":"%s","size":[%d,%d]}`, siteID, *width, *height)
	return jsonStr.Bytes(), nil
}

// builderPulsePoint for building json object for PulsePoint bidder
func builderPulsePoint(params BidderParameters) (json.RawMessage, error) {
	jsonStr := bytes.Buffer{}

	cp, _ := getInt(params.FieldMap["cp"])
	ct, _ := getInt(params.FieldMap["ct"])
	cf, _ := getString(params.FieldMap["cf"])
	//[UOE-5744]: read adsize from fieldmap itself

	if len(cf) == 0 {
		cf = "0x0"
		adSlot := strings.Split(params.SlotKey, "@")
		if len(adSlot) == 2 && adSlot[0] != "" && adSlot[1] != "" {
			cf = adSlot[1]
		}
	}

	fmt.Fprintf(&jsonStr, `{"cp":%d,"ct":%d,"cf":"%s"}`, cp, ct, cf)
	return jsonStr.Bytes(), nil
}

// builderRubicon for building json object for Rubicon bidder
func builderRubicon(params BidderParameters) (json.RawMessage, error) {
	jsonStr := bytes.Buffer{}
	jsonStr.WriteByte('{')

	if accountID, ok := getInt(params.FieldMap["accountId"]); ok {
		fmt.Fprintf(&jsonStr, `"accountId":%d,`, accountID)
	}

	if siteID, ok := getInt(params.FieldMap["siteId"]); ok {
		fmt.Fprintf(&jsonStr, `"siteId":%d,`, siteID)
	}

	if zoneID, ok := getInt(params.FieldMap["zoneId"]); ok {
		fmt.Fprintf(&jsonStr, `"zoneId":%d,`, zoneID)
	}

	if _, ok := params.FieldMap["video"]; ok {
		if videoMap, ok := (params.FieldMap["video"]).(map[string]interface{}); ok {
			jsonStr.WriteString(`"video":{`)

			if width, ok := getInt(videoMap["playerWidth"]); ok {
				fmt.Fprintf(&jsonStr, `"playerWidth":%d,`, width)
			}

			if height, ok := getInt(videoMap["playerHeight"]); ok {
				fmt.Fprintf(&jsonStr, `"playerHeight":%d,`, height)
			}

			if sizeID, ok := getInt(videoMap["size_id"]); ok {
				fmt.Fprintf(&jsonStr, `"size_id":%d,`, sizeID)
			}

			if lang, ok := getString(videoMap["language"]); ok {
				fmt.Fprintf(&jsonStr, `"language":"%s",`, lang)
			}

			trimComma(&jsonStr)
			jsonStr.WriteString(`},`)
		}
	}

	trimComma(&jsonStr)
	jsonStr.WriteByte('}')

	return jsonStr.Bytes(), nil
}

// builderOpenx for building json object for Openx bidder
func builderOpenx(params BidderParameters) (json.RawMessage, error) {
	jsonStr := bytes.Buffer{}
	jsonStr.WriteByte('{')

	if delDomain, ok := getString(params.FieldMap["delDomain"]); ok {
		fmt.Fprintf(&jsonStr, `"delDomain":"%s",`, delDomain)
	} else {
	}

	if unit, ok := getString(params.FieldMap["unit"]); ok {
		fmt.Fprintf(&jsonStr, `"unit":"%s"`, unit)
	} else {
	}

	trimComma(&jsonStr)
	jsonStr.WriteByte('}')
	return jsonStr.Bytes(), nil
}

// builderSovrn for building json object for Sovrn bidder
func builderSovrn(params BidderParameters) (json.RawMessage, error) {
	jsonStr := bytes.Buffer{}
	jsonStr.WriteByte('{')

	if tagID, ok := getString(params.FieldMap["tagid"]); ok {
		fmt.Fprintf(&jsonStr, `"tagid":"%s",`, tagID)
	} else {
	}

	if bidFloor, ok := getFloat64(params.FieldMap["bidfloor"]); ok {
		fmt.Fprintf(&jsonStr, `"bidfloor":%f`, bidFloor)
	}

	trimComma(&jsonStr)
	jsonStr.WriteByte('}')

	return jsonStr.Bytes(), nil
}

// builderImproveDigital for building json object for ImproveDigital bidder
func builderImproveDigital(params BidderParameters) (json.RawMessage, error) {
	jsonStr := bytes.Buffer{}
	jsonStr.WriteByte('{')

	if placementID, ok := getInt(params.FieldMap["placementId"]); ok {
		fmt.Fprintf(&jsonStr, `"placementId":%d`, placementID)
	} else {
		publisherID, ok1 := getInt(params.FieldMap["publisherId"])
		placement, ok2 := getString(params.FieldMap["placementKey"])
		if !ok1 || !ok2 {
			return nil, fmt.Errorf(errMandatoryParameterMissingFormat, params.AdapterName, "['placementId'] or ['publisherId', 'placementKey']")
		}
		fmt.Fprintf(&jsonStr, `"publisherId":%d,"placementKey":"%s"`, publisherID, placement)
	}

	width, height := params.Width, params.Height
	////UOE-5744: Adding custom changes for hybrid profiles
	if val, ok := params.FieldMap["size"]; ok {
		if size, ok := val.(map[string]interface{}); ok {
			w, ok1 := getInt(size["w"])
			h, ok2 := getInt(size["h"])
			if ok1 && ok2 {
				_w := int64(w)
				_h := int64(h)
				width = &(_w)
				height = &(_h)
			}
		}
	}
	if width != nil && height != nil {
		fmt.Fprintf(&jsonStr, `,"size":{"w":%d,"h":%d}`, *width, *height)
	}

	jsonStr.WriteByte('}')
	return jsonStr.Bytes(), nil
}

// builderBeachfront for building json object for Beachfront bidder
func builderBeachfront(params BidderParameters) (json.RawMessage, error) {
	jsonStr := bytes.Buffer{}
	jsonStr.WriteByte('{')

	if appID, ok := getString(params.FieldMap["appId"]); !ok {
		return nil, fmt.Errorf(errMandatoryParameterMissingFormat, params.AdapterName, "appId")
	} else {
		fmt.Fprintf(&jsonStr, `"appId":"%s",`, appID)
	}

	if bidfloor, ok := getFloat64(params.FieldMap["bidfloor"]); !ok {
		return nil, fmt.Errorf(errMandatoryParameterMissingFormat, params.AdapterName, "bidfloor")
	} else {
		fmt.Fprintf(&jsonStr, `"bidfloor":%f`, bidfloor)
	}

	//As per beachfront bidder parameter documentation, by default the video response will be a nurl URL.
	//OpenWrap platform currently only consumes 'adm' responses so setting hardcoded value 'adm' for videoResponseType.
	jsonStr.WriteString(`,"videoResponseType":"adm"`)

	jsonStr.WriteByte('}')
	return jsonStr.Bytes(), nil
}

// builderSmaato for building json object for Smaato bidder
func builderSmaato(params BidderParameters) (json.RawMessage, error) {
	jsonStr := bytes.Buffer{}
	jsonStr.WriteByte('{')

	if publisherID, ok := getString(params.FieldMap["publisherId"]); !ok {
		return nil, fmt.Errorf(errMandatoryParameterMissingFormat, params.AdapterName, "publisherId")
	} else {
		fmt.Fprintf(&jsonStr, `"publisherId":"%s",`, publisherID)
	}

	if adspaceID, ok := getString(params.FieldMap["adspaceId"]); !ok {
		return nil, fmt.Errorf(errMandatoryParameterMissingFormat, params.AdapterName, "adspaceId")
	} else {
		fmt.Fprintf(&jsonStr, `"adspaceId":"%s"`, adspaceID)
	}

	jsonStr.WriteByte('}')
	return jsonStr.Bytes(), nil
}

// builderSmartAdServer for building json object for SmartAdServer bidder
func builderSmartAdServer(params BidderParameters) (json.RawMessage, error) {
	jsonStr := bytes.Buffer{}
	jsonStr.WriteByte('{')

	if networkID, ok := getInt(params.FieldMap["networkId"]); !ok {
		return nil, fmt.Errorf(errMandatoryParameterMissingFormat, params.AdapterName, "networkId")
	} else {
		fmt.Fprintf(&jsonStr, `"networkId":%d`, networkID)
	}

	// siteId, pageId and formatId are dependent on each other and hence need to be sent only when all three are present
	siteID, isSiteIDPresent := getInt(params.FieldMap["siteId"])
	pageID, isPageIDPresent := getInt(params.FieldMap["pageId"])
	formatID, isFormatIDPresent := getInt(params.FieldMap["formatId"])

	if isSiteIDPresent && isPageIDPresent && isFormatIDPresent {
		// all three are valid integers
		fmt.Fprintf(&jsonStr, `,"siteId":%d,"pageId":%d,"formatId":%d`, siteID, pageID, formatID)
	} else {
	}

	jsonStr.WriteByte('}')
	return jsonStr.Bytes(), nil
}

// builderGumGum for building json object for GumGum bidder
func builderGumGum(params BidderParameters) (json.RawMessage, error) {
	jsonStr := bytes.Buffer{}

	if zone, ok := getString(params.FieldMap["zone"]); !ok {
		return nil, fmt.Errorf(errMandatoryParameterMissingFormat, params.AdapterName, "zone")
	} else {
		fmt.Fprintf(&jsonStr, `{"zone":"%s"}`, zone)
	}

	return jsonStr.Bytes(), nil
}

// builderPangle for building json object for Pangle bidder
func builderPangle(params BidderParameters) (json.RawMessage, error) {
	jsonStr := bytes.Buffer{}

	token, ok := getString(params.FieldMap["token"])
	if !ok {
		return nil, fmt.Errorf(errMandatoryParameterMissingFormat, params.AdapterName, "token")
	}

	appID, appIDPresent := getString(params.FieldMap["appid"])
	placementID, placementIDPresent := getString(params.FieldMap["placementid"])

	if appIDPresent && !placementIDPresent {
		return nil, fmt.Errorf(errMandatoryParameterMissingFormat, params.AdapterName, "placementid")
	} else if !appIDPresent && placementIDPresent {
		return nil, fmt.Errorf(errMandatoryParameterMissingFormat, params.AdapterName, "appid")
	}

	if appIDPresent && placementIDPresent {
		fmt.Fprintf(&jsonStr, `{"token":"%s","placementid":"%s","appid":"%s"}`, token, placementID, appID)
	} else {
		fmt.Fprintf(&jsonStr, `{"token":"%s"}`, token)
	}

	return jsonStr.Bytes(), nil
}

// builderSonobi for building json object for Sonobi bidder
func builderSonobi(params BidderParameters) (json.RawMessage, error) {
	jsonStr := bytes.Buffer{}

	tagID, _ := getString(params.FieldMap["ad_unit"]) //checking with ad_unit value
	if len(tagID) == 0 {
		tagID, _ = getString(params.FieldMap["placement_id"]) //checking with placement_id
	}

	if len(tagID) == 0 {
		return nil, fmt.Errorf(errMandatoryParameterMissingFormat, params.AdapterName, "'ad_unit' or 'placement_id'")
	}

	fmt.Fprintf(&jsonStr, `{"TagID":"%s"}`, tagID)
	return jsonStr.Bytes(), nil
}

// builderAdform for building json object for Adform bidder
func builderAdform(params BidderParameters) (json.RawMessage, error) {
	jsonStr := bytes.Buffer{}

	if mid, ok := getInt(params.FieldMap["mid"]); ok {
		fmt.Fprintf(&jsonStr, `{"mid":%d}`, mid)
	} else {
		inv, invPresent := getInt(params.FieldMap["inv"])
		mname, mnamePresent := getString(params.FieldMap["mname"])

		if !(invPresent && mnamePresent) {
			return nil, fmt.Errorf(errMandatoryParameterMissingFormat, params.AdapterName, "'mid' and 'inv'")
		}

		fmt.Fprintf(&jsonStr, `{"inv":%d,"mname":"%s"}`, inv, mname)
	}

	return jsonStr.Bytes(), nil
}

// builderCriteo for building json object for Criteo bidder
func builderCriteo(params BidderParameters) (json.RawMessage, error) {
	jsonStr := bytes.Buffer{}

	anyOf := []string{"zoneId", "networkId"} // not checking zoneid and networkid as client side uses only zoneId and networkId
	for _, param := range anyOf {
		if val, ok := getInt(params.FieldMap[param]); ok {
			fmt.Fprintf(&jsonStr, `{"%s":%d}`, param, val)
			break
		}
	}

	if jsonStr.Len() == 0 {
		return nil, fmt.Errorf(errMandatoryParameterMissingFormat, params.AdapterName, anyOf)
	}

	return jsonStr.Bytes(), nil
}

// builderOutbrain for building json object for Outbrain bidder
func builderOutbrain(params BidderParameters) (json.RawMessage, error) {
	jsonStr := bytes.Buffer{}
	publisherMap, ok := params.FieldMap["publisher"]
	if !ok {
		return nil, nil
	}

	publisher, ok := publisherMap.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	id, ok := getString(publisher["id"])
	if !ok {
		return nil, nil
	}

	fmt.Fprintf(&jsonStr, `{"publisher":{"id":"%s"}}`, id)
	return jsonStr.Bytes(), nil
}

func builderApacdex(params BidderParameters) (json.RawMessage, error) {
	jsonStr := bytes.Buffer{}
	jsonStr.WriteByte('{')
	anyOf := []string{BidderParamApacdex_siteId, BidderParamApacdex_placementId}
	for _, param := range anyOf {
		if key, ok := getString(params.FieldMap[param]); ok {
			fmt.Fprintf(&jsonStr, `"%s":"%s"`, param, key)
			break
		}
	}
	//  len=1 (no mandatory params present)
	if jsonStr.Len() == 1 {
		return nil, fmt.Errorf(errMandatoryParameterMissingFormat, params.AdapterName, anyOf)
	}
	if floorPrice, ok := getFloat64(params.FieldMap[BidderParamApacdex_floorPrice]); ok {
		fmt.Fprintf(&jsonStr, `,"%s":%g`, BidderParamApacdex_floorPrice, floorPrice)
	}
	//geo object(hybrid param)
	if value, ok := params.FieldMap[BidderParamApacdex_geo]; ok {
		if geoJson, err := json.Marshal(value); err == nil {
			fmt.Fprintf(&jsonStr, `,"%s":%s`, BidderParamApacdex_geo, string(geoJson))
		}
	}
	jsonStr.WriteByte('}')
	return jsonStr.Bytes(), nil
}
