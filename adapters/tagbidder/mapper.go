package tagbidder

type macroCallBack struct {
	cached   bool
	callback func(IBidderMacro, string) string
}

type mapper map[string]*macroCallBack

var bidderMapper map[string]mapper

func (obj mapper) clone() mapper {
	cloned := make(mapper, len(obj))
	for k, v := range obj {
		newCallback := *v
		cloned[k] = &newCallback
	}
	return cloned
}

//SetCache value to specific key
func (obj *mapper) SetCache(key string, value bool) {
	if value, ok := (*obj)[key]; ok {
		value.cached = true
	}
}

//AddCustomMacro for adding custom macro whose definition will be present in IBidderMacro.Custom method
func (obj *mapper) AddCustomMacro(key string, isCached bool) {
	(*obj)[key] = &macroCallBack{cached: isCached, callback: IBidderMacro.Custom}
}

var _defaultMapper = mapper{
	//Request
	MacroTest:              &macroCallBack{cached: false, callback: IBidderMacro.MacroTest},
	MacroTimeout:           &macroCallBack{cached: false, callback: IBidderMacro.MacroTimeout},
	MacroWhitelistSeat:     &macroCallBack{cached: false, callback: IBidderMacro.MacroWhitelistSeat},
	MacroWhitelistLang:     &macroCallBack{cached: false, callback: IBidderMacro.MacroWhitelistLang},
	MacroBlockedseat:       &macroCallBack{cached: false, callback: IBidderMacro.MacroBlockedseat},
	MacroCurrency:          &macroCallBack{cached: false, callback: IBidderMacro.MacroCurrency},
	MacroBlockedCategory:   &macroCallBack{cached: false, callback: IBidderMacro.MacroBlockedCategory},
	MacroBlockedAdvertiser: &macroCallBack{cached: false, callback: IBidderMacro.MacroBlockedAdvertiser},
	MacroBlockedApp:        &macroCallBack{cached: false, callback: IBidderMacro.MacroBlockedApp},

	//Source
	MacroFD:             &macroCallBack{cached: false, callback: IBidderMacro.MacroFD},
	MacroTransactionID:  &macroCallBack{cached: false, callback: IBidderMacro.MacroTransactionID},
	MacroPaymentIDChain: &macroCallBack{cached: false, callback: IBidderMacro.MacroPaymentIDChain},

	//Regs
	MacroCoppa: &macroCallBack{cached: false, callback: IBidderMacro.MacroCoppa},

	//Impression
	MacroDisplayManager:        &macroCallBack{cached: false, callback: IBidderMacro.MacroDisplayManager},
	MacroDisplayManagerVersion: &macroCallBack{cached: false, callback: IBidderMacro.MacroDisplayManagerVersion},
	MacroInterstitial:          &macroCallBack{cached: false, callback: IBidderMacro.MacroInterstitial},
	MacroTagID:                 &macroCallBack{cached: false, callback: IBidderMacro.MacroTagID},
	MacroBidFloor:              &macroCallBack{cached: false, callback: IBidderMacro.MacroBidFloor},
	MacroBidFloorCurrency:      &macroCallBack{cached: false, callback: IBidderMacro.MacroBidFloorCurrency},
	MacroSecure:                &macroCallBack{cached: false, callback: IBidderMacro.MacroSecure},
	MacroPMP:                   &macroCallBack{cached: false, callback: IBidderMacro.MacroPMP},

	//Video
	MacroVideoMIMES:            &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoMIMES},
	MacroVideoMinimumDuration:  &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoMinimumDuration},
	MacroVideoMaximumDuration:  &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoMaximumDuration},
	MacroVideoProtocols:        &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoProtocols},
	MacroVideoPlayerWidth:      &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoPlayerWidth},
	MacroVideoPlayerHeight:     &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoPlayerHeight},
	MacroVideoStartDelay:       &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoStartDelay},
	MacroVideoPlacement:        &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoPlacement},
	MacroVideoLinearity:        &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoLinearity},
	MacroVideoSkip:             &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoSkip},
	MacroVideoSkipMinimum:      &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoSkipMinimum},
	MacroVideoSkipAfter:        &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoSkipAfter},
	MacroVideoSequence:         &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoSequence},
	MacroVideoBlockedAttribute: &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoBlockedAttribute},
	MacroVideoMaximumExtended:  &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoMaximumExtended},
	MacroVideoMinimumBitRate:   &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoMinimumBitRate},
	MacroVideoMaximumBitRate:   &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoMaximumBitRate},
	MacroVideoBoxing:           &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoBoxing},
	MacroVideoPlaybackMethod:   &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoPlaybackMethod},
	MacroVideoDelivery:         &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoDelivery},
	MacroVideoPosition:         &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoPosition},
	MacroVideoAPI:              &macroCallBack{cached: false, callback: IBidderMacro.MacroVideoAPI},

	//Site
	MacroSiteID:       &macroCallBack{cached: false, callback: IBidderMacro.MacroSiteID},
	MacroSiteName:     &macroCallBack{cached: false, callback: IBidderMacro.MacroSiteName},
	MacroSitePage:     &macroCallBack{cached: false, callback: IBidderMacro.MacroSitePage},
	MacroSiteReferrer: &macroCallBack{cached: false, callback: IBidderMacro.MacroSiteReferrer},
	MacroSiteSearch:   &macroCallBack{cached: false, callback: IBidderMacro.MacroSiteSearch},
	MacroSiteMobile:   &macroCallBack{cached: false, callback: IBidderMacro.MacroSiteMobile},

	//App
	MacroAppID:       &macroCallBack{cached: false, callback: IBidderMacro.MacroAppID},
	MacroAppName:     &macroCallBack{cached: false, callback: IBidderMacro.MacroAppName},
	MacroAppBundle:   &macroCallBack{cached: false, callback: IBidderMacro.MacroAppBundle},
	MacroAppStoreURL: &macroCallBack{cached: false, callback: IBidderMacro.MacroAppStoreURL},
	MacroAppVersion:  &macroCallBack{cached: false, callback: IBidderMacro.MacroAppVersion},
	MacroAppPaid:     &macroCallBack{cached: false, callback: IBidderMacro.MacroAppPaid},

	//SiteAppCommon
	MacroCategory:        &macroCallBack{cached: false, callback: IBidderMacro.MacroCategory},
	MacroDomain:          &macroCallBack{cached: false, callback: IBidderMacro.MacroDomain},
	MacroSectionCategory: &macroCallBack{cached: false, callback: IBidderMacro.MacroSectionCategory},
	MacroPageCategory:    &macroCallBack{cached: false, callback: IBidderMacro.MacroPageCategory},
	MacroPrivacyPolicy:   &macroCallBack{cached: false, callback: IBidderMacro.MacroPrivacyPolicy},
	MacroKeywords:        &macroCallBack{cached: false, callback: IBidderMacro.MacroKeywords},

	//Publisher
	MacroPubID:     &macroCallBack{cached: false, callback: IBidderMacro.MacroPubID},
	MacroPubName:   &macroCallBack{cached: false, callback: IBidderMacro.MacroPubName},
	MacroPubDomain: &macroCallBack{cached: false, callback: IBidderMacro.MacroPubDomain},

	//Content
	MacroContentID:                &macroCallBack{cached: false, callback: IBidderMacro.MacroContentID},
	MacroContentEpisode:           &macroCallBack{cached: false, callback: IBidderMacro.MacroContentEpisode},
	MacroContentTitle:             &macroCallBack{cached: false, callback: IBidderMacro.MacroContentTitle},
	MacroContentSeries:            &macroCallBack{cached: false, callback: IBidderMacro.MacroContentSeries},
	MacroContentSeason:            &macroCallBack{cached: false, callback: IBidderMacro.MacroContentSeason},
	MacroContentArtist:            &macroCallBack{cached: false, callback: IBidderMacro.MacroContentArtist},
	MacroContentGenre:             &macroCallBack{cached: false, callback: IBidderMacro.MacroContentGenre},
	MacroContentAlbum:             &macroCallBack{cached: false, callback: IBidderMacro.MacroContentAlbum},
	MacroContentISrc:              &macroCallBack{cached: false, callback: IBidderMacro.MacroContentISrc},
	MacroContentURL:               &macroCallBack{cached: false, callback: IBidderMacro.MacroContentURL},
	MacroContentCategory:          &macroCallBack{cached: false, callback: IBidderMacro.MacroContentCategory},
	MacroContentProductionQuality: &macroCallBack{cached: false, callback: IBidderMacro.MacroContentProductionQuality},
	MacroContentVideoQuality:      &macroCallBack{cached: false, callback: IBidderMacro.MacroContentVideoQuality},
	MacroContentContext:           &macroCallBack{cached: false, callback: IBidderMacro.MacroContentContext},

	//Producer
	MacroProducerID:   &macroCallBack{cached: false, callback: IBidderMacro.MacroProducerID},
	MacroProducerName: &macroCallBack{cached: false, callback: IBidderMacro.MacroProducerName},

	//Device
	MacroUserAgent:       &macroCallBack{cached: false, callback: IBidderMacro.MacroUserAgent},
	MacroDNT:             &macroCallBack{cached: false, callback: IBidderMacro.MacroDNT},
	MacroLMT:             &macroCallBack{cached: false, callback: IBidderMacro.MacroLMT},
	MacroIP:              &macroCallBack{cached: false, callback: IBidderMacro.MacroIP},
	MacroDeviceType:      &macroCallBack{cached: false, callback: IBidderMacro.MacroDeviceType},
	MacroMake:            &macroCallBack{cached: false, callback: IBidderMacro.MacroMake},
	MacroModel:           &macroCallBack{cached: false, callback: IBidderMacro.MacroModel},
	MacroDeviceOS:        &macroCallBack{cached: false, callback: IBidderMacro.MacroDeviceOS},
	MacroDeviceOSVersion: &macroCallBack{cached: false, callback: IBidderMacro.MacroDeviceOSVersion},
	MacroDeviceWidth:     &macroCallBack{cached: false, callback: IBidderMacro.MacroDeviceWidth},
	MacroDeviceHeight:    &macroCallBack{cached: false, callback: IBidderMacro.MacroDeviceHeight},
	MacroDeviceJS:        &macroCallBack{cached: false, callback: IBidderMacro.MacroDeviceJS},
	MacroDeviceLanguage:  &macroCallBack{cached: false, callback: IBidderMacro.MacroDeviceLanguage},
	MacroDeviceIFA:       &macroCallBack{cached: false, callback: IBidderMacro.MacroDeviceIFA},
	MacroDeviceDIDSHA1:   &macroCallBack{cached: false, callback: IBidderMacro.MacroDeviceDIDSHA1},
	MacroDeviceDIDMD5:    &macroCallBack{cached: false, callback: IBidderMacro.MacroDeviceDIDMD5},
	MacroDeviceDPIDSHA1:  &macroCallBack{cached: false, callback: IBidderMacro.MacroDeviceDPIDSHA1},
	MacroDeviceDPIDMD5:   &macroCallBack{cached: false, callback: IBidderMacro.MacroDeviceDPIDMD5},
	MacroDeviceMACSHA1:   &macroCallBack{cached: false, callback: IBidderMacro.MacroDeviceMACSHA1},
	MacroDeviceMACMD5:    &macroCallBack{cached: false, callback: IBidderMacro.MacroDeviceMACMD5},

	//Geo
	MacroLatitude:  &macroCallBack{cached: false, callback: IBidderMacro.MacroLatitude},
	MacroLongitude: &macroCallBack{cached: false, callback: IBidderMacro.MacroLongitude},
	MacroCountry:   &macroCallBack{cached: false, callback: IBidderMacro.MacroCountry},
	MacroRegion:    &macroCallBack{cached: false, callback: IBidderMacro.MacroRegion},
	MacroCity:      &macroCallBack{cached: false, callback: IBidderMacro.MacroCity},
	MacroZip:       &macroCallBack{cached: false, callback: IBidderMacro.MacroZip},
	MacroUTCOffset: &macroCallBack{cached: false, callback: IBidderMacro.MacroUTCOffset},

	//User
	MacroUserID:      &macroCallBack{cached: false, callback: IBidderMacro.MacroUserID},
	MacroYearOfBirth: &macroCallBack{cached: false, callback: IBidderMacro.MacroYearOfBirth},
	MacroGender:      &macroCallBack{cached: false, callback: IBidderMacro.MacroGender},

	//Extension
	MacroGDPRConsent: &macroCallBack{cached: false, callback: IBidderMacro.MacroGDPRConsent},
	MacroGDPR:        &macroCallBack{cached: false, callback: IBidderMacro.MacroGDPR},
	MacroUSPrivacy:   &macroCallBack{cached: false, callback: IBidderMacro.MacroUSPrivacy},

	//Additional
	MacroCacheBuster: &macroCallBack{cached: false, callback: IBidderMacro.MacroCacheBuster},
}

//GetNewDefaultMapper will return clone of default mapper function
func GetNewDefaultMapper() mapper {
	return _defaultMapper.clone()
}

//SetBidderMapper will be used by each bidder to set its respective macro mapper
func SetBidderMapper(bidder string, bidderMap mapper) {
	bidderMapper[bidder] = bidderMap
}

//GetBidderMapper will return mapper of specific bidder
func GetBidderMapper(bidder string) mapper {
	return bidderMapper[bidder]
}
