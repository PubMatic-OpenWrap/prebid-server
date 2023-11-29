package v25

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/prebid/openrtb/v19/adcom1"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/v2/hooks/hookstage"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

func ConvertVideoToAuctionRequest(payload hookstage.EntrypointPayload, result *hookstage.HookResult[hookstage.EntrypointPayload]) (models.RequestExtWrapper, error) {
	values := payload.Request.URL.Query()

	pubID := values.Get(models.PUBID_KEY)
	profileID := values.Get(models.PROFILEID_KEY)

	owRedirectURLStr := values.Get(models.OWRedirectURLKey)
	// mimeTypesStr := values.Get(models.MimeTypes)
	gdprFlag := values.Get(models.GDPRFlag)
	ccpa := values.Get(models.CCPAUSPrivacyKey)
	eids := values.Get(models.OWUserEids)
	consentString := values.Get(models.ConsentString)
	appReq := values.Get(models.AppRequest)
	responseFormat := values.Get(models.ResponseFormatKey)

	if owRedirectURLStr == "" && responseFormat != models.ResponseFormatJSON {
		return models.RequestExtWrapper{}, errors.New(models.OWRedirectURLKey + " missing in request")
	}

	redirectURL, err := url.Parse(owRedirectURLStr)
	if err != nil {
		return models.RequestExtWrapper{}, errors.New(models.OWRedirectURLKey + "url parsing failed")
	}
	redirectQueryParams := redirectURL.Query()
	// Replace macro values in DFP URL - NYC TODO: Do we still need to trim the macro prefix?
	for k := range values {
		if strings.HasPrefix(k, models.MacroPrefix) {
			paramName := strings.TrimPrefix(k, models.MacroPrefix)
			redirectQueryParams.Set(paramName, values.Get(k))
		}
	}

	bidRequest := openrtb2.BidRequest{
		Imp: []openrtb2.Imp{
			{
				Video: &openrtb2.Video{
					MIMEs:          GetStringArr(GetValueFromRequest(values, redirectQueryParams, models.MimeORTBParam)),
					MaxDuration:    GetCustomAtoI64(GetString(GetValueFromRequest(values, redirectQueryParams, models.MaxDurationORTBParam))),
					MinDuration:    GetCustomAtoI64(GetString(GetValueFromRequest(values, redirectQueryParams, models.MinDurationORTBParam))),
					Protocols:      models.GetProtocol(GetIntArr(GetValueFromRequest(values, redirectQueryParams, models.ProtocolsORTBParam))),
					Skip:           GetCustomAtoI8(GetString(GetValueFromRequest(values, redirectQueryParams, models.SkipORTBParam))),
					SkipMin:        GetCustomAtoI64(GetString(GetValueFromRequest(values, redirectQueryParams, models.SkipMinORTBParam))),
					SkipAfter:      GetCustomAtoI64(GetString(GetValueFromRequest(values, redirectQueryParams, models.SkipAfterORTBParam))),
					BAttr:          models.GetCreativeAttributes(GetIntArr(GetValueFromRequest(values, redirectQueryParams, models.BAttrORTBParam))),
					MaxExtended:    GetCustomAtoI64(GetString(GetValueFromRequest(values, redirectQueryParams, models.MaxExtendedORTBParam))),
					MinBitRate:     GetCustomAtoI64(GetString(GetValueFromRequest(values, redirectQueryParams, models.MinBitrateORTBParam))),
					MaxBitRate:     GetCustomAtoI64(GetString(GetValueFromRequest(values, redirectQueryParams, models.MaxBitrateORTBParam))),
					PlaybackMethod: models.GetPlaybackMethod(GetIntArr(GetValueFromRequest(values, redirectQueryParams, models.PlaybackMethodORTBParam))),
					Delivery:       models.GetDeliveryMethod(GetIntArr(GetValueFromRequest(values, redirectQueryParams, models.DeliveryORTBParam))),
					API:            models.GetAPIFramework((GetIntArr(GetValueFromRequest(values, redirectQueryParams, models.APIORTBParam)))),
				},
			},
		},
	}

	if sequence := GetCustomAtoI8(GetString(GetValueFromRequest(values, redirectQueryParams, models.SequenceORTBParam))); sequence != nil {
		bidRequest.Imp[0].Video.Sequence = *sequence
	}
	if boxingAllowed := GetCustomAtoI8(GetString(GetValueFromRequest(values, redirectQueryParams, models.BoxingAllowedORTBParam))); boxingAllowed != nil {
		bidRequest.Imp[0].Video.BoxingAllowed = *boxingAllowed
	}
	if prctl := GetCustomAtoI8(GetString(GetValueFromRequest(values, redirectQueryParams, models.ProtocolORTBParam))); prctl != nil {
		bidRequest.Imp[0].Video.Protocol = adcom1.MediaCreativeSubtype(*prctl)
	}
	if strtDelay := GetCustomAtoI64(GetString(GetValueFromRequest(values, redirectQueryParams, models.StartDelayORTBParam))); strtDelay != 0 {
		st := adcom1.StartDelay(strtDelay)
		bidRequest.Imp[0].Video.StartDelay = &st
	}
	if placementValue := GetCustomAtoI64(GetString(GetValueFromRequest(values, redirectQueryParams, models.PlacementORTBParam))); placementValue != 0 {
		bidRequest.Imp[0].Video.Placement = adcom1.VideoPlacementSubtype(placementValue)
	}
	if linearityValue := GetCustomAtoI64(GetString(GetValueFromRequest(values, redirectQueryParams, models.LinearityORTBParam))); linearityValue != 0 {
		bidRequest.Imp[0].Video.Linearity = adcom1.LinearityMode(linearityValue)
	}
	if pos := GetCustomAtoI8(GetString(GetValueFromRequest(values, redirectQueryParams, models.PosORTBParam))); pos != nil {
		pos := adcom1.PlacementPosition(*pos)
		bidRequest.Imp[0].Video.Pos = &pos
	}

	size := GetString(GetValueFromRequest(values, redirectQueryParams, models.SizeORTBParam))
	if size != "" && strings.Split(size, "x") != nil {
		sizeValues := strings.Split(size, "x")
		bidRequest.Imp[0].Video.W, _ = strconv.ParseInt(sizeValues[0], 10, 64)
		bidRequest.Imp[0].Video.H, _ = strconv.ParseInt(sizeValues[1], 10, 64)
	}

	slot := redirectQueryParams.Get(models.InventoryUnitKey)
	if slot == "" && responseFormat == models.ResponseFormatJSON {
		slot = values.Get(models.InventoryUnitMacroKey)
	}

	validationFailed := false
	if slot == "" {
		validationFailed = true
	}

	// TODO NYC: do we need this??
	// if mimeTypesStr == "" {
	// 	validationFailed = true
	// } else {
	// 	mimeStrArr := strings.Split(values.Get(models.OWMimeTypes), models.MimesSeparator)
	// 	if len(mimeStrArr) == 0 {
	// 		validationFailed = true
	// 	} else {
	// 		for _, mime := range mimeStrArr {
	// 			if models.MimeIDToValueMap[mime] == "" {
	// 				validationFailed = true
	// 				break
	// 			}
	// 		}
	// 	}
	// }

	// if gdprFlag != "" && gdprFlag != "0" && gdprFlag != "1" {
	// 	validationFailed = true
	// }

	// if ccpa != "" && len(ccpa) != 4 {
	// 	validationFailed = true
	// }

	// request is for Mobile App, perform necessary validations
	if appReq == "1" && (models.CheckIfValidQueryParamFlag(values, models.DeviceLMT) || models.CheckIfValidQueryParamFlag(values, models.DeviceDNT)) {
		validationFailed = true
	}

	if validationFailed {
		return models.RequestExtWrapper{}, errors.New("validation failure")
	}

	if uuid, err := uuid.NewV4(); err == nil {
		bidRequest.ID = uuid.String()
	}

	if uuid, err := uuid.NewV4(); err == nil {
		bidRequest.Imp[0].ID = uuid.String()
	}
	bidRequest.Imp[0].TagID = slot
	bidRequest.Imp[0].BidFloor, _ = models.Atof(values.Get(models.FloorValue), 4)
	bidRequest.Imp[0].BidFloorCur = values.Get(models.FloorCurrency)

	content := &openrtb2.Content{
		Genre: GetString(GetValueFromRequest(values, redirectQueryParams, models.ContentGenreORTBParam)),
		Title: GetString(GetValueFromRequest(values, redirectQueryParams, models.ContentTitleORTBParam)),
	}
	if content.Genre == "" && content.Title == "" {
		content = nil
	}

	if appReq == "1" {
		bidRequest.App = &openrtb2.App{
			ID:       GetString(GetValueFromRequest(values, redirectQueryParams, models.AppIDORTBParam)),
			Name:     GetString(GetValueFromRequest(values, redirectQueryParams, models.AppNameORTBParam)),
			Bundle:   GetString(GetValueFromRequest(values, redirectQueryParams, models.AppBundleORTBParam)),
			StoreURL: GetString(GetValueFromRequest(values, redirectQueryParams, models.AppStoreURLORTBParam)),
			Domain:   GetString(GetValueFromRequest(values, redirectQueryParams, models.AppDomainORTBParam)),
			Keywords: GetString(GetValueFromRequest(values, redirectQueryParams, models.OwAppKeywords)),
			Cat:      GetStringArr(GetValueFromRequest(values, redirectQueryParams, models.AppCatORTBParam)),
			Publisher: &openrtb2.Publisher{
				ID: pubID,
			},
			Content: content,
		}
		bidRequest.Device = &openrtb2.Device{
			Lmt:      GetCustomAtoI8(GetString(GetValueFromRequest(values, redirectQueryParams, models.DeviceLMTORTBParam))),
			DNT:      GetCustomAtoI8(GetString(GetValueFromRequest(values, redirectQueryParams, models.DeviceDNTORTBParam))),
			IFA:      GetString(GetValueFromRequest(values, redirectQueryParams, models.DeviceIfaORTBParam)),
			DIDSHA1:  GetString(GetValueFromRequest(values, redirectQueryParams, models.DeviceDidsha1ORTBParam)),
			DIDMD5:   GetString(GetValueFromRequest(values, redirectQueryParams, models.DeviceDidmd5ORTBParam)),
			DPIDSHA1: GetString(GetValueFromRequest(values, redirectQueryParams, models.DeviceDpidsha1ORTBParam)),
			DPIDMD5:  GetString(GetValueFromRequest(values, redirectQueryParams, models.DeviceDpidmd5ORTBParam)),
			MACSHA1:  GetString(GetValueFromRequest(values, redirectQueryParams, models.DeviceMacsha1ORTBParam)),
			MACMD5:   GetString(GetValueFromRequest(values, redirectQueryParams, models.DeviceMacmd5ORTBParam)),
			Geo: &openrtb2.Geo{
				Lat:       GetCustomStrToFloat(GetString(GetValueFromRequest(values, redirectQueryParams, models.GeoLatORTBParam))),
				Lon:       GetCustomStrToFloat(GetString(GetValueFromRequest(values, redirectQueryParams, models.GeoLonORTBParam))),
				Country:   GetString(GetValueFromRequest(values, redirectQueryParams, models.GeoCountryORTBParam)),
				City:      GetString(GetValueFromRequest(values, redirectQueryParams, models.GeoCityORTBParam)),
				Metro:     GetString(GetValueFromRequest(values, redirectQueryParams, models.GeoMetroORTBParam)),
				ZIP:       GetString(GetValueFromRequest(values, redirectQueryParams, models.GeoZipORTBParam)),
				UTCOffset: GetCustomAtoI64(GetString(GetValueFromRequest(values, redirectQueryParams, models.GeoUTOffsetORTBParam))),
			},
		}

		paid := GetCustomAtoI8(GetString(GetValueFromRequest(values, redirectQueryParams, models.AppPaidORTBParam)))
		if paid != nil {
			bidRequest.App.Paid = *paid
		}

		js := GetCustomAtoI8(GetString(GetValueFromRequest(values, redirectQueryParams, models.DeviceJSORTBParam)))
		if js != nil {
			bidRequest.Device.JS = *js
		}

		if locationTypeValue := GetCustomAtoI8(GetString(GetValueFromRequest(values, redirectQueryParams, models.GeoTypeORTBParam))); locationTypeValue != nil {
			bidRequest.Device.Geo.Type = adcom1.LocationType(*locationTypeValue)
		}

		var deviceExt models.ExtDevice

		if session_id := GetValueFromRequest(values, redirectQueryParams, models.DeviceExtSessionID); session_id != nil {
			deviceExt.SessionID = GetString(session_id)
		}

		if ifaType := GetValueFromRequest(values, redirectQueryParams, models.DeviceExtIfaType); ifaType != nil {
			deviceExt.ExtDevice = &openrtb_ext.ExtDevice{
				IFAType: GetString(ifaType),
			}
		}
		bidRequest.Device.Ext, _ = json.Marshal(deviceExt)
	} else {
		url := redirectQueryParams.Get(models.URLKey)
		if url == "" {
			url = redirectQueryParams.Get(models.DescriptionURLKey)
		}
		if url == "" {
			url = payload.Request.Header.Get(models.PAGE_URL_HEADER)
		}

		bidRequest.Site = &openrtb2.Site{
			Publisher: &openrtb2.Publisher{
				ID: pubID,
			},
			Content: content,
			Page:    url,
		}
	}

	bidderParams := GetString(GetValueFromRequest(values, redirectQueryParams, models.BidderParams))
	impPrebidExt := GetString(GetValueFromRequest(values, redirectQueryParams, models.ImpPrebidExt))
	updatedImpExt := false
	impExt := "{"
	if bidderParams != "" {
		impExt += "bidder" + bidderParams
		updatedImpExt = true
	}
	if impPrebidExt != "" {
		if updatedImpExt {
			impExt += ","
		}
		impExt += "prebid" + impPrebidExt
		updatedImpExt = true
	}
	impExt += "}"
	bidRequest.Imp[0].Ext = json.RawMessage(impExt)

	if validationFailed {
		return models.RequestExtWrapper{}, errors.New("validation failed")
	}

	if gdprFlag != "" || ccpa != "" {
		bidRequest.Regs = &openrtb2.Regs{}
		regsExt := openrtb_ext.ExtRegs{}

		if gdprFlag != "" {
			gdprInt, _ := strconv.ParseInt(gdprFlag, 10, 8)
			gdprInt8 := int8(gdprInt)
			regsExt.GDPR = &gdprInt8
		}

		if ccpa != "" {
			regsExt.USPrivacy = ccpa
		}

		bidRequest.Regs.Ext, _ = json.Marshal(regsExt)
	}

	bidRequest.User = &openrtb2.User{
		ID:     GetString(GetValueFromRequest(values, redirectQueryParams, models.UserIDORTBParam)),
		Gender: GetString(GetValueFromRequest(values, redirectQueryParams, models.UserGenderORTBParam)),
		Yob:    GetCustomAtoI64(GetString(GetValueFromRequest(values, redirectQueryParams, models.UserYobORTBParam))),
	}

	if consentString != "" || eids != "" {
		userExt := openrtb_ext.ExtUser{
			Consent: consentString,
		}

		var eidList []openrtb2.EID
		if err := json.Unmarshal([]byte(eids), &eidList); err == nil {
			userExt.Eids = eidList
		}

		bidRequest.User.Ext, _ = json.Marshal(userExt)
	}

	sourceExt := models.ExtSource{}
	if omidpv := GetString(GetValueFromRequest(values, redirectQueryParams, models.SourceOmidpvORTBParam)); omidpv != "" {
		sourceExt.OMIDPV = omidpv
	}
	if omidpn := GetString(GetValueFromRequest(values, redirectQueryParams, models.SourceOmidpnORTBParam)); omidpn != "" {
		sourceExt.OMIDPN = omidpn
	}
	bidRequest.Source = &openrtb2.Source{}
	bidRequest.Source.Ext, _ = json.Marshal(sourceExt)

	profileId, _ := strconv.Atoi(profileID)
	displayVersion := 0
	if val := getValueForKeyFromParams(models.VERSION_KEY, appReq, values, redirectQueryParams); val != "" {
		displayVersion, _ = strconv.Atoi(val)
	}

	contentTransparency := values.Get(models.ContentTransparency)
	var transparency map[string]openrtb_ext.TransparencyRule
	if contentTransparency != "" {
		_ = json.Unmarshal([]byte(contentTransparency), &transparency)
	}

	requestExtWrapper := models.RequestExtWrapper{
		ProfileId:        profileId,
		VersionId:        displayVersion,
		SumryDisableFlag: 1,
		SSAuctionFlag:    1,
	}
	requestExt := models.RequestExt{
		Wrapper: &requestExtWrapper,
		ExtRequest: openrtb_ext.ExtRequest{
			Prebid: openrtb_ext.ExtRequestPrebid{
				Debug: getValueForKeyFromParams(models.DEBUG_KEY, appReq, values, redirectQueryParams) == "1",
				Transparency: &openrtb_ext.TransparencyExt{
					Content: transparency,
				},
			},
		},
	}
	bidRequest.Ext, _ = json.Marshal(requestExt)

	// Replace macro values in DFP URL
	for k := range values {
		if strings.HasPrefix(k, models.MacroPrefix) {
			paramName := strings.TrimPrefix(k, models.MacroPrefix)
			redirectQueryParams.Set(paramName, values.Get(k))
		}
	}
	DFPControllerValue := rand.Int()
	redirectQueryParams.Set(models.Correlator, strconv.Itoa(DFPControllerValue))
	redirectURL.RawQuery = redirectQueryParams.Encode()
	owRedirectURLStr = redirectURL.String()
	//Update Original HTTP Request with updated value of 'owredirect'
	values.Set(models.OWRedirectURLKey, owRedirectURLStr)
	rawQuery := values.Encode()

	body, err := json.Marshal(bidRequest)
	if err != nil {
		return requestExtWrapper, err
	}

	result.ChangeSet.AddMutation(func(ep hookstage.EntrypointPayload) (hookstage.EntrypointPayload, error) {
		ep.Request.URL.RawQuery = rawQuery
		// ep.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		ep.Body = body
		return ep, nil
	}, hookstage.MutationUpdate, "entrypoint-update-amp-redirect-url")

	return requestExtWrapper, nil
}

// GetValueFromRequest contains logic to get value for parameter identified by 'key'
func GetValueFromRequest(requestQueryParams, redirectQueryParams url.Values, key string) interface{} {
	values := requestQueryParams
	switch key {
	case models.MimeORTBParam:
		// DFP currently does not have a parameter defined for Mime types
		if values.Get(models.OWMimeTypes) != "" {
			mimeStrArr := strings.Split(values.Get(models.OWMimeTypes), models.MimesSeparator)
			mimeValueArr := make([]string, 0)
			for _, mime := range mimeStrArr {
				if models.MimeIDToValueMap[mime] != "" {
					mimeValueArr = append(mimeValueArr, models.MimeIDToValueMap[mime])
				}
			}
			return mimeValueArr
		}
	case models.OwAppKeywords:
		if values.Get(models.OwAppKeywords) != "" {
			return values.Get(models.OwAppKeywords)
		}
	case models.MaxDurationORTBParam:
		return getValue(models.MaxDurationORTBParam, values, redirectQueryParams, models.DFPMaxAdDuration, models.OWMaxAdDuration)
	case models.MinDurationORTBParam:
		return getValue(models.MinDurationORTBParam, values, redirectQueryParams, models.DFPMinAdDuration, models.OWMinAdDuration)
	case models.StartDelayORTBParam:
		if values.Get(models.OWStartDelay) != "" {
			return values.Get(models.OWStartDelay)
		} else if redirectQueryParams.Get(models.DFPVPos) != "" {
			posStr := redirectQueryParams.Get(models.DFPVPos)
			return models.VideoPositionToStartDelayMap[posStr]
		}
	case models.PlaybackMethodORTBParam:
		var pbStr string
		if values.Get(models.OWPlaybackMethod) != "" {
			pbStr = values.Get(models.OWPlaybackMethod)
		} else if redirectQueryParams.Get(models.DFPVpmute) != "" || redirectQueryParams.Get(models.DFPVpa) != "" {
			if redirectQueryParams.Get(models.DFPVpmute) == "1" && redirectQueryParams.Get(models.DFPVpa) == "0" {
				pbStr = "2,6"
			} else if redirectQueryParams.Get(models.DFPVpmute) == "0" && redirectQueryParams.Get(models.DFPVpa) == "1" {
				pbStr = "1,2"
			} else if redirectQueryParams.Get(models.DFPVpmute) == "1" && redirectQueryParams.Get(models.DFPVpa) == "1" {
				pbStr = "2"
			}
		}
		if pbStr != "" {
			pbIntArr := make([]int, 0)
			for _, pb := range strings.Split(pbStr, ",") {
				pbInt, _ := strconv.Atoi(pb)
				pbIntArr = append(pbIntArr, pbInt)
			}
			return pbIntArr
		}
	case models.APIORTBParam:
		return getValueInArray(models.APIORTBParam, values, redirectQueryParams, "", models.OWAPI)
	case models.DeliveryORTBParam:
		return getValueInArray(models.DeliveryORTBParam, values, redirectQueryParams, "", models.OWDelivery)
	case models.ProtocolsORTBParam:
		return getValueInArray(models.ProtocolsORTBParam, values, redirectQueryParams, "", models.OWProtocols)
	case models.BAttrORTBParam:
		return getValueInArray(models.BAttrORTBParam, values, redirectQueryParams, "", models.OWBAttr)
	case models.LinearityORTBParam:
		if values.Get(models.OWLinearity) != "" {
			return values.Get(models.OWLinearity)
		} else if redirectQueryParams.Get(models.DFPVAdType) != "" {
			adtypeStr := redirectQueryParams.Get(models.DFPVAdType)
			return models.LinearityMap[adtypeStr]
		}
	case models.PlacementORTBParam:
		return getValue(models.PlacementORTBParam, values, redirectQueryParams, "", models.OWPlacement)
	case models.MinBitrateORTBParam:
		return getValue(models.MinBitrateORTBParam, values, redirectQueryParams, "", models.OWMinBitrate)
	case models.MaxBitrateORTBParam:
		return getValue(models.MaxBitrateORTBParam, values, redirectQueryParams, "", models.OWMaxBitrate)
	case models.SkipORTBParam:
		return getValue(models.SkipORTBParam, values, redirectQueryParams, "", models.OWSkippable)
	case models.SkipMinORTBParam:
		return getValue(models.SkipMinORTBParam, values, redirectQueryParams, "", models.OWSkipMin)
	case models.SkipAfterORTBParam:
		return getValue(models.SkipAfterORTBParam, values, redirectQueryParams, "", models.OWSkipAfter)
	case models.SequenceORTBParam:
		return getValue(models.SequenceORTBParam, values, redirectQueryParams, "", models.OWSequence)
	case models.BoxingAllowedORTBParam:
		return getValue(models.BoxingAllowedORTBParam, values, redirectQueryParams, "", models.OWBoxingAllowed)
	case models.MaxExtendedORTBParam:
		return getValue(models.MaxExtendedORTBParam, values, redirectQueryParams, "", models.OWMaxExtended)
	case models.ProtocolORTBParam:
		return getValue(models.ProtocolORTBParam, values, redirectQueryParams, "", models.OWProtocol)
	case models.PosORTBParam:
		return getValue(models.PosORTBParam, values, redirectQueryParams, "", models.OWPos)
	case models.AppIDORTBParam:
		return getValue(models.AppIDORTBParam, values, redirectQueryParams, "", models.OWAppId)
	case models.AppNameORTBParam:
		return getValue(models.AppNameORTBParam, values, redirectQueryParams, "", models.OWAppName)
	case models.AppBundleORTBParam:
		return getValue(models.AppBundleORTBParam, values, redirectQueryParams, "", models.OWAppBundle)
	case models.AppDomainORTBParam:
		return getValue(models.AppDomainORTBParam, values, redirectQueryParams, "", models.OWAppDomain)
	case models.AppStoreURLORTBParam:
		return getValue(models.AppStoreURLORTBParam, values, redirectQueryParams, "", models.OWAppStoreURL)
	case models.AppCatORTBParam:
		if values.Get(models.OWAppCat) != "" {
			catStrArr := strings.Split(values.Get(models.OWAppCat), models.Comma)
			return catStrArr
		}
	case models.AppPaidORTBParam:
		return getValue(models.AppPaidORTBParam, values, redirectQueryParams, "", models.OWAppPaid)
	case models.DeviceUAORTBParam:
		return getValue(models.DeviceUAORTBParam, values, redirectQueryParams, "", models.OWDeviceUA)
	case models.DeviceIPORTBParam:
		return getValue(models.DeviceIPORTBParam, values, redirectQueryParams, "", models.OWDeviceIP)
	case models.DeviceLMTORTBParam:
		return getValue(models.DeviceLMTORTBParam, values, redirectQueryParams, "", models.OWDeviceLMT)
	case models.DeviceDNTORTBParam:
		return getValue(models.DeviceDNTORTBParam, values, redirectQueryParams, "", models.OWDeviceDNT)
	case models.DeviceJSORTBParam:
		return getValue(models.DeviceJSORTBParam, values, redirectQueryParams, "", models.OWDeviceJS)
	case models.GeoLatORTBParam:
		return getValue(models.GeoLatORTBParam, values, redirectQueryParams, "", models.OWGeoLat)
	case models.GeoLonORTBParam:
		return getValue(models.GeoLonORTBParam, values, redirectQueryParams, "", models.OWGeoLon)
	case models.GeoTypeORTBParam:
		return getValue(models.GeoTypeORTBParam, values, redirectQueryParams, "", models.OWGeoType)
	case models.GeoCountryORTBParam:
		return getValue(models.GeoCountryORTBParam, values, redirectQueryParams, "", models.OWGeoCountry)
	case models.GeoCityORTBParam:
		return getValue(models.GeoCityORTBParam, values, redirectQueryParams, "", models.OWGeoCity)
	case models.GeoMetroORTBParam:
		return getValue(models.GeoMetroORTBParam, values, redirectQueryParams, "", models.OWGeoMetro)
	case models.GeoZipORTBParam:
		return getValue(models.GeoZipORTBParam, values, redirectQueryParams, "", models.OWGeoZip)
	case models.GeoUTOffsetORTBParam:
		return getValue(models.GeoUTOffsetORTBParam, values, redirectQueryParams, "", models.OWUTOffset)
	case models.DeviceIfaORTBParam:
		return getValue(models.DeviceIfaORTBParam, values, redirectQueryParams, "", models.OWDeviceIfa)
	case models.DeviceDidsha1ORTBParam:
		return getValue(models.DeviceDidsha1ORTBParam, values, redirectQueryParams, "", models.OWDeviceDidsha1)
	case models.DeviceDidmd5ORTBParam:
		return getValue(models.DeviceDidmd5ORTBParam, values, redirectQueryParams, "", models.OWDeviceDidmd5)
	case models.DeviceDpidsha1ORTBParam:
		return getValue(models.DeviceDpidsha1ORTBParam, values, redirectQueryParams, "", models.OWDeviceDpidsha1)
	case models.DeviceDpidmd5ORTBParam:
		return getValue(models.DeviceDpidmd5ORTBParam, values, redirectQueryParams, "", models.OWDeviceDpidmd5)
	case models.DeviceMacsha1ORTBParam:
		return getValue(models.DeviceMacsha1ORTBParam, values, redirectQueryParams, "", models.OWDeviceMacsha1)
	case models.DeviceMacmd5ORTBParam:
		return getValue(models.DeviceMacmd5ORTBParam, values, redirectQueryParams, "", models.OWDeviceMacmd5)
	case models.UserIDORTBParam:
		return getValue(models.UserIDORTBParam, values, redirectQueryParams, "", models.OWUserID)
	case models.SizeORTBParam:
		if values.Get(models.OWSize) != "" {
			return values.Get(models.OWSize)
		} else if redirectQueryParams.Get(models.DFPSize) != "" {
			// If multiple sizes are passed in DFP parameter, we will consider only the first
			DFPSizeStr := strings.Split(redirectQueryParams.Get(models.DFPSize), models.MultipleSizeSeparator)
			return DFPSizeStr[0]
		}
	case models.ContentGenreORTBParam:
		return getValue(models.ContentGenreORTBParam, values, redirectQueryParams, "", models.OWContentGenre)
	case models.ContentTitleORTBParam:
		return getValue(models.ContentTitleORTBParam, values, redirectQueryParams, "", models.OWContentTitle)
	case models.UserGenderORTBParam:
		return getValue(models.UserGenderORTBParam, values, redirectQueryParams, "", models.OWUserGender)
	case models.UserYobORTBParam:
		return getValue(models.UserYobORTBParam, values, redirectQueryParams, "", models.OWUserYob)
	case models.SourceOmidpvORTBParam:
		return getValue(models.SourceOmidpvORTBParam, values, redirectQueryParams, "", models.OWSourceOmidPv)
	case models.SourceOmidpnORTBParam:
		return getValue(models.SourceOmidpnORTBParam, values, redirectQueryParams, "", models.OWSourceOmidPn)
	case models.BidderParams:
		return getValue(models.BidderParams, values, redirectQueryParams, "", models.OWBidderParams)
	case models.DeviceExtSessionID:
		if _, ok := values[models.OWDeviceExtSessionID]; ok {
			return values.Get(models.OWDeviceExtSessionID)
		}
	case models.DeviceExtIfaType:
		if _, ok := values[models.OWDeviceExtIfaType]; ok {
			return values.Get(models.OWDeviceExtIfaType)
		}
	case models.FloorValue:
		if _, ok := values[models.FloorValue]; ok {
			return values.Get(models.FloorValue)
		}
	case models.FloorCurrency:
		if _, ok := values[models.FloorCurrency]; ok {
			return values.Get(models.FloorCurrency)
		}
	case models.ImpPrebidExt:
		return getValue(models.ImpPrebidExt, values, redirectQueryParams, "", models.OWImpPrebidExt)
	}
	return nil

}

func getValue(oRTBParamName string, values url.Values, redirectQueryParams url.Values, DFPParamName string, OWParamName string) interface{} {
	paramArr := models.ORTBToDFPOWMap[oRTBParamName]
	if paramArr == nil {
		return nil
	}

	if values.Get(OWParamName) != "" {
		return values.Get(OWParamName)
	} else if paramArr[1] != "" && DFPParamName != "" && redirectQueryParams.Get(DFPParamName) != "" {
		return redirectQueryParams.Get(DFPParamName)
	}

	return nil
}

func getValueInArray(oRTBParamName string, values url.Values, redirectQueryParams url.Values, DFPParamName string, OWParamName string) interface{} {
	valStr := GetString(getValue(oRTBParamName, values, redirectQueryParams, DFPParamName, OWParamName))
	if valStr != "" {
		valIntArr := make([]int, 0)
		for _, val := range strings.Split(valStr, ",") {
			valInt, _ := strconv.Atoi(val)
			valIntArr = append(valIntArr, valInt)
		}
		return valIntArr
	}
	return nil
}

func GetString(val interface{}) string {
	var result string
	if val != nil {
		result, ok := val.(string)
		if ok {
			return result
		}
	}
	return result
}

func GetStringArr(val interface{}) []string {
	var result []string
	if val != nil {
		result, ok := val.([]string)
		if ok {
			return result
		}
	}
	return result
}

func GetIntArr(val interface{}) []int {
	var result []int
	if val != nil {
		result, ok := val.([]int)
		if ok {
			return result
		}
	}
	return result
}

func GetInt(val interface{}) int {
	var result int
	if val != nil {
		result, ok := val.(int)
		if ok {
			return result
		}
	}
	return result
}

func GetCustomAtoI8(s string) *int8 {
	if s == "" {
		return nil
	}
	i, ok := strconv.Atoi(s)
	if ok == nil {
		i8 := int8(i)
		return &i8
	}
	return nil
}

func GetCustomAtoI64(s string) int64 {
	if s == "" {
		return 0
	}
	i, ok := strconv.ParseInt(s, 10, 64)
	if ok == nil {
		return i
	}
	return 0
}

func GetCustomStrToFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, ok := strconv.ParseFloat(s, 64)
	if ok == nil {
		return f
	}
	return 0
}

// getValueForKeyFromParams returns value for a key from the request params, if not present in request params
// then it checks url/description_url in the redirect URL
func getValueForKeyFromParams(key string, appReq string, requestParams, redirectURLParams url.Values) string {
	var value string

	//first check from request query params
	if val := requestParams.Get(key); val != "" {
		return val
	}

	//else check it in url/description_url in redirect url query params
	urlStr := getURLfromRedirectURL(redirectURLParams, appReq)

	if urlStr != "" {
		if urlObj, urlErr := url.Parse(urlStr); urlErr == nil {
			URLQueryParams := urlObj.Query()
			if val := URLQueryParams.Get(key); val != "" {
				return val
			}
		}
	}
	return value

}

// getURLfromRedirectURL return 'url' from redirectURL and if url is not present it returns desc URL for web request
func getURLfromRedirectURL(redirectQueryParams url.Values, appReq string) string {
	var URL string

	//check for 'url' query param
	if urlStr := redirectQueryParams.Get(models.URLKey); urlStr != "" {
		return urlStr
	}

	//if 'url' is not present, check for 'description_url' key
	if appReq != "1" {
		if descURL := redirectQueryParams.Get(models.DescriptionURLKey); descURL != "" {
			return descURL
		}
	}
	return URL
}
