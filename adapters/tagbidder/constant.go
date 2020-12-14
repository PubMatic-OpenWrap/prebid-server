package tagbidder

const (
	intBase = 10
	comma   = `,`
)

//List of Tag Bidder Macros
const (
	//Request
	MacroTest              = `MacroTest`
	MacroTimeout           = `MacroTimeout`
	MacroWhitelistSeat     = `MacroWhitelistSeat`
	MacroWhitelistLang     = `MacroWhitelistLang`
	MacroBlockedseat       = `MacroBlockedseat`
	MacroCurrency          = `MacroCurrency`
	MacroBlockedCategory   = `MacroBlockedCategory`
	MacroBlockedAdvertiser = `MacroBlockedAdvertiser`
	MacroBlockedApp        = `MacroBlockedApp`

	//Source
	MacroFD             = `MacroFD`
	MacroTransactionID  = `MacroTransactionID`
	MacroPaymentIDChain = `MacroPaymentIDChain`

	//Regs
	MacroCoppa = `MacroCoppa`

	//Impression
	MacroDisplayManager        = `MacroDisplayManager`
	MacroDisplayManagerVersion = `MacroDisplayManagerVersion`
	MacroInterstitial          = `MacroInterstitial`
	MacroTagID                 = `MacroTagID`
	MacroBidFloor              = `MacroBidFloor`
	MacroBidFloorCurrency      = `MacroBidFloorCurrency`
	MacroSecure                = `MacroSecure`
	MacroPMP                   = `MacroPMP`

	//Video
	MacroVideoMIMES            = `MacroVideoMIMES`
	MacroVideoMinimumDuration  = `MacroVideoMinimumDuration`
	MacroVideoMaximumDuration  = `MacroVideoMaximumDuration`
	MacroVideoProtocols        = `MacroVideoProtocols`
	MacroVideoPlayerWidth      = `MacroVideoPlayerWidth`
	MacroVideoPlayerHeight     = `MacroVideoPlayerHeight`
	MacroVideoStartDelay       = `MacroVideoStartDelay`
	MacroVideoPlacement        = `MacroVideoPlacement`
	MacroVideoLinearity        = `MacroVideoLinearity`
	MacroVideoSkip             = `MacroVideoSkip`
	MacroVideoSkipMinimum      = `MacroVideoSkipMinimum`
	MacroVideoSkipAfter        = `MacroVideoSkipAfter`
	MacroVideoSequence         = `MacroVideoSequence`
	MacroVideoBlockedAttribute = `MacroVideoBlockedAttribute`
	MacroVideoMaximumExtended  = `MacroVideoMaximumExtended`
	MacroVideoMinimumBitRate   = `MacroVideoMinimumBitRate`
	MacroVideoMaximumBitRate   = `MacroVideoMaximumBitRate`
	MacroVideoBoxing           = `MacroVideoBoxing`
	MacroVideoPlaybackMethod   = `MacroVideoPlaybackMethod`
	MacroVideoDelivery         = `MacroVideoDelivery`
	MacroVideoPosition         = `MacroVideoPosition`
	MacroVideoAPI              = `MacroVideoAPI`

	//Site
	MacroSiteID       = `MacroSiteID`
	MacroSiteName     = `MacroSiteName`
	MacroSitePage     = `MacroSitePage`
	MacroSiteReferrer = `MacroSiteReferrer`
	MacroSiteSearch   = `MacroSiteSearch`
	MacroSiteMobile   = `MacroSiteMobile`

	//App
	MacroAppID       = `MacroAppID`
	MacroAppName     = `MacroAppName`
	MacroAppBundle   = `MacroAppBundle`
	MacroAppStoreURL = `MacroAppStoreURL`
	MacroAppVersion  = `MacroAppVersion`
	MacroAppPaid     = `MacroAppPaid`

	//SiteAppCommon
	MacroCategory        = `MacroCategory`
	MacroDomain          = `MacroDomain`
	MacroSectionCategory = `MacroSectionCategory`
	MacroPageCategory    = `MacroPageCategory`
	MacroPrivacyPolicy   = `MacroPrivacyPolicy`
	MacroKeywords        = `MacroKeywords`

	//Publisher
	MacroPubID     = `MacroPubID`
	MacroPubName   = `MacroPubName`
	MacroPubDomain = `MacroPubDomain`

	//Content
	MacroContentID                = `MacroContentID`
	MacroContentEpisode           = `MacroContentEpisode`
	MacroContentTitle             = `MacroContentTitle`
	MacroContentSeries            = `MacroContentSeries`
	MacroContentSeason            = `MacroContentSeason`
	MacroContentArtist            = `MacroContentArtist`
	MacroContentGenre             = `MacroContentGenre`
	MacroContentAlbum             = `MacroContentAlbum`
	MacroContentISrc              = `MacroContentISrc`
	MacroContentURL               = `MacroContentURL`
	MacroContentCategory          = `MacroContentCategory`
	MacroContentProductionQuality = `MacroContentProductionQuality`
	MacroContentVideoQuality      = `MacroContentVideoQuality`
	MacroContentContext           = `MacroContentContext`

	//Producer
	MacroProducerID   = `MacroProducerID`
	MacroProducerName = `MacroProducerName`

	//Device
	MacroUserAgent       = `MacroUserAgent`
	MacroDNT             = `MacroDNT`
	MacroLMT             = `MacroLMT`
	MacroIP              = `MacroIP`
	MacroDeviceType      = `MacroDeviceType`
	MacroMake            = `MacroMake`
	MacroModel           = `MacroModel`
	MacroDeviceOS        = `MacroDeviceOS`
	MacroDeviceOSVersion = `MacroDeviceOSVersion`
	MacroDeviceWidth     = `MacroDeviceWidth`
	MacroDeviceHeight    = `MacroDeviceHeight`
	MacroDeviceJS        = `MacroDeviceJS`
	MacroDeviceLanguage  = `MacroDeviceLanguage`
	MacroDeviceIFA       = `MacroDeviceIFA`
	MacroDeviceDIDSHA1   = `MacroDeviceDIDSHA1`
	MacroDeviceDIDMD5    = `MacroDeviceDIDMD5`
	MacroDeviceDPIDSHA1  = `MacroDeviceDPIDSHA1`
	MacroDeviceDPIDMD5   = `MacroDeviceDPIDMD5`
	MacroDeviceMACSHA1   = `MacroDeviceMACSHA1`
	MacroDeviceMACMD5    = `MacroDeviceMACMD5`

	//Geo
	MacroLatitude  = `MacroLatitude`
	MacroLongitude = `MacroLongitude`
	MacroCountry   = `MacroCountry`
	MacroRegion    = `MacroRegion`
	MacroCity      = `MacroCity`
	MacroZip       = `MacroZip`
	MacroUTCOffset = `MacroUTCOffset`

	//User
	MacroUserID      = `MacroUserID`
	MacroYearOfBirth = `MacroYearOfBirth`
	MacroGender      = `MacroGender`

	//Extension
	MacroGDPRConsent = `MacroGDPRConsent`
	MacroGDPR        = `MacroGDPR`
	MacroUSPrivacy   = `MacroUSPrivacy`

	//Additional
	MacroCacheBuster = `MacroCacheBuster`
)

//MacroKeyType types of macro keys
type MacroKeyType string

const (
	UnkownMacroKeys       MacroKeyType = ``
	StandardORTBMacroKeys MacroKeyType = `standard`
	CustomORTBMacroKeys   MacroKeyType = `custom`
)
