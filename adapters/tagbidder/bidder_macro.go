package tagbidder

import "github.com/PubMatic-OpenWrap/openrtb"

//BidderMacro default implementation
type BidderMacro struct {
	IBidderMacro
	request   *openrtb.BidRequest
	isApp     bool
	imp       *openrtb.Imp
	publisher *openrtb.Publisher
	content   *openrtb.Content
}

//NewBidderMacro contains definition for all openrtb macro's
func NewBidderMacro(request *openrtb.BidRequest) *BidderMacro {
	bidder := &BidderMacro{
		request: request,
	}
	bidder.init()
	return bidder
}

func (tag *BidderMacro) init() {
	if nil != tag.request.App {
		tag.isApp = true
		tag.publisher = tag.request.App.Publisher
		tag.content = tag.request.App.Content
	} else {
		tag.publisher = tag.request.Site.Publisher
		tag.content = tag.request.Site.Content
	}
}

//LoadImpression will set current imp
func (tag *BidderMacro) LoadImpression(imp *openrtb.Imp) error {
	tag.imp = imp
	return nil
}

/********************* Request *********************/

//MacroTest contains definition for Test Parameter
func (tag *BidderMacro) MacroTest(key string) string {
	return ""
}

//MacroTimeout contains definition for Timeout Parameter
func (tag *BidderMacro) MacroTimeout(key string) string {
	return ""
}

//MacroWhitelistSeat contains definition for WhitelistSeat Parameter
func (tag *BidderMacro) MacroWhitelistSeat(key string) string {
	return ""
}

//MacroWhitelistLang contains definition for WhitelistLang Parameter
func (tag *BidderMacro) MacroWhitelistLang(key string) string {
	return ""
}

//MacroBlockedseat contains definition for Blockedseat Parameter
func (tag *BidderMacro) MacroBlockedseat(key string) string {
	return ""
}

//MacroCurrency contains definition for Currency Parameter
func (tag *BidderMacro) MacroCurrency(key string) string {
	return ""
}

//MacroBlockedCategory contains definition for BlockedCategory Parameter
func (tag *BidderMacro) MacroBlockedCategory(key string) string {
	return ""
}

//MacroBlockedAdvertiser contains definition for BlockedAdvertiser Parameter
func (tag *BidderMacro) MacroBlockedAdvertiser(key string) string {
	return ""
}

//MacroBlockedApp contains definition for BlockedApp Parameter
func (tag *BidderMacro) MacroBlockedApp(key string) string {
	return ""
}

/********************* Source *********************/

//MacroFD contains definition for FD Parameter
func (tag *BidderMacro) MacroFD(key string) string {
	return ""
}

//MacroTransactionID contains definition for TransactionID Parameter
func (tag *BidderMacro) MacroTransactionID(key string) string {
	return ""
}

//MacroPaymentIDChain contains definition for PaymentIDChain Parameter
func (tag *BidderMacro) MacroPaymentIDChain(key string) string {
	return ""
}

/********************* Regs *********************/

//MacroCoppa contains definition for Coppa Parameter
func (tag *BidderMacro) MacroCoppa(key string) string {
	return ""
}

/********************* Impression *********************/

//MacroDisplayManager contains definition for DisplayManager Parameter
func (tag *BidderMacro) MacroDisplayManager(key string) string {
	return ""
}

//MacroDisplayManagerVersion contains definition for DisplayManagerVersion Parameter
func (tag *BidderMacro) MacroDisplayManagerVersion(key string) string {
	return ""
}

//MacroInterstitial contains definition for Interstitial Parameter
func (tag *BidderMacro) MacroInterstitial(key string) string {
	return ""
}

//MacroTagID contains definition for TagID Parameter
func (tag *BidderMacro) MacroTagID(key string) string {
	return ""
}

//MacroBidFloor contains definition for BidFloor Parameter
func (tag *BidderMacro) MacroBidFloor(key string) string {
	return ""
}

//MacroBidFloorCurrency contains definition for BidFloorCurrency Parameter
func (tag *BidderMacro) MacroBidFloorCurrency(key string) string {
	return ""
}

//MacroSecure contains definition for Secure Parameter
func (tag *BidderMacro) MacroSecure(key string) string {
	return ""
}

//MacroPMP contains definition for PMP Parameter
func (tag *BidderMacro) MacroPMP(key string) string {
	return ""
}

/********************* Video *********************/

//MacroVideoMIMES contains definition for VideoMIMES Parameter
func (tag *BidderMacro) MacroVideoMIMES(key string) string {
	return ""
}

//MacroVideoMinimumDuration contains definition for VideoMinimumDuration Parameter
func (tag *BidderMacro) MacroVideoMinimumDuration(key string) string {
	return ""
}

//MacroVideoMaximumDuration contains definition for VideoMaximumDuration Parameter
func (tag *BidderMacro) MacroVideoMaximumDuration(key string) string {
	return ""
}

//MacroVideoProtocols contains definition for VideoProtocols Parameter
func (tag *BidderMacro) MacroVideoProtocols(key string) string {
	return ""
}

//MacroVideoPlayerWidth contains definition for VideoPlayerWidth Parameter
func (tag *BidderMacro) MacroVideoPlayerWidth(key string) string {
	return ""
}

//MacroVideoPlayerHeight contains definition for VideoPlayerHeight Parameter
func (tag *BidderMacro) MacroVideoPlayerHeight(key string) string {
	return ""
}

//MacroVideoStartDelay contains definition for VideoStartDelay Parameter
func (tag *BidderMacro) MacroVideoStartDelay(key string) string {
	return ""
}

//MacroVideoPlacement contains definition for VideoPlacement Parameter
func (tag *BidderMacro) MacroVideoPlacement(key string) string {
	return ""
}

//MacroVideoLinearity contains definition for VideoLinearity Parameter
func (tag *BidderMacro) MacroVideoLinearity(key string) string {
	return ""
}

//MacroVideoSkip contains definition for VideoSkip Parameter
func (tag *BidderMacro) MacroVideoSkip(key string) string {
	return ""
}

//MacroVideoSkipMinimum contains definition for VideoSkipMinimum Parameter
func (tag *BidderMacro) MacroVideoSkipMinimum(key string) string {
	return ""
}

//MacroVideoSkipAfter contains definition for VideoSkipAfter Parameter
func (tag *BidderMacro) MacroVideoSkipAfter(key string) string {
	return ""
}

//MacroVideoSequence contains definition for VideoSequence Parameter
func (tag *BidderMacro) MacroVideoSequence(key string) string {
	return ""
}

//MacroVideoBlockedAttribute contains definition for VideoBlockedAttribute Parameter
func (tag *BidderMacro) MacroVideoBlockedAttribute(key string) string {
	return ""
}

//MacroVideoMaximumExtended contains definition for VideoMaximumExtended Parameter
func (tag *BidderMacro) MacroVideoMaximumExtended(key string) string {
	return ""
}

//MacroVideoMinimumBitRate contains definition for VideoMinimumBitRate Parameter
func (tag *BidderMacro) MacroVideoMinimumBitRate(key string) string {
	return ""
}

//MacroVideoMaximumBitRate contains definition for VideoMaximumBitRate Parameter
func (tag *BidderMacro) MacroVideoMaximumBitRate(key string) string {
	return ""
}

//MacroVideoBoxing contains definition for VideoBoxing Parameter
func (tag *BidderMacro) MacroVideoBoxing(key string) string {
	return ""
}

//MacroVideoPlaybackMethod contains definition for VideoPlaybackMethod Parameter
func (tag *BidderMacro) MacroVideoPlaybackMethod(key string) string {
	return ""
}

//MacroVideoDelivery contains definition for VideoDelivery Parameter
func (tag *BidderMacro) MacroVideoDelivery(key string) string {
	return ""
}

//MacroVideoPosition contains definition for VideoPosition Parameter
func (tag *BidderMacro) MacroVideoPosition(key string) string {
	return ""
}

//MacroVideoAPI contains definition for VideoAPI Parameter
func (tag *BidderMacro) MacroVideoAPI(key string) string {
	return ""
}

/********************* Site *********************/

//MacroSiteID contains definition for SiteID Parameter
func (tag *BidderMacro) MacroSiteID(key string) string {
	return ""
}

//MacroSiteName contains definition for SiteName Parameter
func (tag *BidderMacro) MacroSiteName(key string) string {
	return ""
}

//MacroSitePage contains definition for SitePage Parameter
func (tag *BidderMacro) MacroSitePage(key string) string {
	return ""
}

//MacroSiteReferrer contains definition for SiteReferrer Parameter
func (tag *BidderMacro) MacroSiteReferrer(key string) string {
	return ""
}

//MacroSiteSearch contains definition for SiteSearch Parameter
func (tag *BidderMacro) MacroSiteSearch(key string) string {
	return ""
}

//MacroSiteMobile contains definition for SiteMobile Parameter
func (tag *BidderMacro) MacroSiteMobile(key string) string {
	return ""
}

/********************* App *********************/

//MacroAppID contains definition for AppID Parameter
func (tag *BidderMacro) MacroAppID(key string) string {
	return ""
}

//MacroAppName contains definition for AppName Parameter
func (tag *BidderMacro) MacroAppName(key string) string {
	return ""
}

//MacroAppBundle contains definition for AppBundle Parameter
func (tag *BidderMacro) MacroAppBundle(key string) string {
	return ""
}

//MacroAppStoreURL contains definition for AppStoreURL Parameter
func (tag *BidderMacro) MacroAppStoreURL(key string) string {
	return ""
}

//MacroAppVersion contains definition for AppVersion Parameter
func (tag *BidderMacro) MacroAppVersion(key string) string {
	return ""
}

//MacroAppPaid contains definition for AppPaid Parameter
func (tag *BidderMacro) MacroAppPaid(key string) string {
	return ""
}

/********************* Site/App Common *********************/

//MacroCategory contains definition for Category Parameter
func (tag *BidderMacro) MacroCategory(key string) string {
	return ""
}

//MacroDomain contains definition for Domain Parameter
func (tag *BidderMacro) MacroDomain(key string) string {
	return ""
}

//MacroSectionCategory contains definition for SectionCategory Parameter
func (tag *BidderMacro) MacroSectionCategory(key string) string {
	return ""
}

//MacroPageCategory contains definition for PageCategory Parameter
func (tag *BidderMacro) MacroPageCategory(key string) string {
	return ""
}

//MacroPrivacyPolicy contains definition for PrivacyPolicy Parameter
func (tag *BidderMacro) MacroPrivacyPolicy(key string) string {
	return ""
}

//MacroKeywords contains definition for Keywords Parameter
func (tag *BidderMacro) MacroKeywords(key string) string {
	return ""
}

/********************* Publisher *********************/

//MacroPubID contains definition for PubID Parameter
func (tag *BidderMacro) MacroPubID(key string) string {
	return ""
}

//MacroPubName contains definition for PubName Parameter
func (tag *BidderMacro) MacroPubName(key string) string {
	return ""
}

//MacroPubDomain contains definition for PubDomain Parameter
func (tag *BidderMacro) MacroPubDomain(key string) string {
	return ""
}

/********************* Content *********************/

//MacroContentID contains definition for ContentID Parameter
func (tag *BidderMacro) MacroContentID(key string) string {
	return ""
}

//MacroContentEpisode contains definition for ContentEpisode Parameter
func (tag *BidderMacro) MacroContentEpisode(key string) string {
	return ""
}

//MacroContentTitle contains definition for ContentTitle Parameter
func (tag *BidderMacro) MacroContentTitle(key string) string {
	return ""
}

//MacroContentSeries contains definition for ContentSeries Parameter
func (tag *BidderMacro) MacroContentSeries(key string) string {
	return ""
}

//MacroContentSeason contains definition for ContentSeason Parameter
func (tag *BidderMacro) MacroContentSeason(key string) string {
	return ""
}

//MacroContentArtist contains definition for ContentArtist Parameter
func (tag *BidderMacro) MacroContentArtist(key string) string {
	return ""
}

//MacroContentGenre contains definition for ContentGenre Parameter
func (tag *BidderMacro) MacroContentGenre(key string) string {
	return ""
}

//MacroContentAlbum contains definition for ContentAlbum Parameter
func (tag *BidderMacro) MacroContentAlbum(key string) string {
	return ""
}

//MacroContentISrc contains definition for ContentISrc Parameter
func (tag *BidderMacro) MacroContentISrc(key string) string {
	return ""
}

//MacroContentURL contains definition for ContentURL Parameter
func (tag *BidderMacro) MacroContentURL(key string) string {
	return ""
}

//MacroContentCategory contains definition for ContentCategory Parameter
func (tag *BidderMacro) MacroContentCategory(key string) string {
	return ""
}

//MacroContentProductionQuality contains definition for ContentProductionQuality Parameter
func (tag *BidderMacro) MacroContentProductionQuality(key string) string {
	return ""
}

//MacroContentVideoQuality contains definition for ContentVideoQuality Parameter
func (tag *BidderMacro) MacroContentVideoQuality(key string) string {
	return ""
}

//MacroContentContext contains definition for ContentContext Parameter
func (tag *BidderMacro) MacroContentContext(key string) string {
	return ""
}

/********************* Producer *********************/

//MacroProducerID contains definition for ProducerID Parameter
func (tag *BidderMacro) MacroProducerID(key string) string {
	return ""
}

//MacroProducerName contains definition for ProducerName Parameter
func (tag *BidderMacro) MacroProducerName(key string) string {
	return ""
}

/********************* Device *********************/

//MacroUserAgent contains definition for UserAgent Parameter
func (tag *BidderMacro) MacroUserAgent(key string) string {
	return ""
}

//MacroDNT contains definition for DNT Parameter
func (tag *BidderMacro) MacroDNT(key string) string {
	return ""
}

//MacroLMT contains definition for LMT Parameter
func (tag *BidderMacro) MacroLMT(key string) string {
	return ""
}

//MacroIP contains definition for IP Parameter
func (tag *BidderMacro) MacroIP(key string) string {
	return ""
}

//MacroDeviceType contains definition for DeviceType Parameter
func (tag *BidderMacro) MacroDeviceType(key string) string {
	return ""
}

//MacroMake contains definition for Make Parameter
func (tag *BidderMacro) MacroMake(key string) string {
	return ""
}

//MacroModel contains definition for Model Parameter
func (tag *BidderMacro) MacroModel(key string) string {
	return ""
}

//MacroDeviceOS contains definition for DeviceOS Parameter
func (tag *BidderMacro) MacroDeviceOS(key string) string {
	return ""
}

//MacroDeviceOSVersion contains definition for DeviceOSVersion Parameter
func (tag *BidderMacro) MacroDeviceOSVersion(key string) string {
	return ""
}

//MacroDeviceWidth contains definition for DeviceWidth Parameter
func (tag *BidderMacro) MacroDeviceWidth(key string) string {
	return ""
}

//MacroDeviceHeight contains definition for DeviceHeight Parameter
func (tag *BidderMacro) MacroDeviceHeight(key string) string {
	return ""
}

//MacroDeviceJS contains definition for DeviceJS Parameter
func (tag *BidderMacro) MacroDeviceJS(key string) string {
	return ""
}

//MacroDeviceLanguage contains definition for DeviceLanguage Parameter
func (tag *BidderMacro) MacroDeviceLanguage(key string) string {
	return ""
}

//MacroDeviceIFA contains definition for DeviceIFA Parameter
func (tag *BidderMacro) MacroDeviceIFA(key string) string {
	return ""
}

//MacroDeviceDIDSHA1 contains definition for DeviceDIDSHA1 Parameter
func (tag *BidderMacro) MacroDeviceDIDSHA1(key string) string {
	return ""
}

//MacroDeviceDIDMD5 contains definition for DeviceDIDMD5 Parameter
func (tag *BidderMacro) MacroDeviceDIDMD5(key string) string {
	return ""
}

//MacroDeviceDPIDSHA1 contains definition for DeviceDPIDSHA1 Parameter
func (tag *BidderMacro) MacroDeviceDPIDSHA1(key string) string {
	return ""
}

//MacroDeviceDPIDMD5 contains definition for DeviceDPIDMD5 Parameter
func (tag *BidderMacro) MacroDeviceDPIDMD5(key string) string {
	return ""
}

//MacroDeviceMACSHA1 contains definition for DeviceMACSHA1 Parameter
func (tag *BidderMacro) MacroDeviceMACSHA1(key string) string {
	return ""
}

//MacroDeviceMACMD5 contains definition for DeviceMACMD5 Parameter
func (tag *BidderMacro) MacroDeviceMACMD5(key string) string {
	return ""
}

/********************* Geo *********************/

//MacroLatitude contains definition for Latitude Parameter
func (tag *BidderMacro) MacroLatitude(key string) string {
	return ""
}

//MacroLongitude contains definition for Longitude Parameter
func (tag *BidderMacro) MacroLongitude(key string) string {
	return ""
}

//MacroCountry contains definition for Country Parameter
func (tag *BidderMacro) MacroCountry(key string) string {
	return ""
}

//MacroRegion contains definition for Region Parameter
func (tag *BidderMacro) MacroRegion(key string) string {
	return ""
}

//MacroCity contains definition for City Parameter
func (tag *BidderMacro) MacroCity(key string) string {
	return ""
}

//MacroZip contains definition for Zip Parameter
func (tag *BidderMacro) MacroZip(key string) string {
	return ""
}

//MacroUTCOffset contains definition for UTCOffset Parameter
func (tag *BidderMacro) MacroUTCOffset(key string) string {
	return ""
}

/********************* User *********************/

//MacroUserID contains definition for UserID Parameter
func (tag *BidderMacro) MacroUserID(key string) string {
	return ""
}

//MacroYearOfBirth contains definition for YearOfBirth Parameter
func (tag *BidderMacro) MacroYearOfBirth(key string) string {
	return ""
}

//MacroGender contains definition for Gender Parameter
func (tag *BidderMacro) MacroGender(key string) string {
	return ""
}

/********************* Extension *********************/

//MacroGDPRConsent contains definition for GDPRConsent Parameter
func (tag *BidderMacro) MacroGDPRConsent(key string) string {
	return ""
}

//MacroGDPR contains definition for GDPR Parameter
func (tag *BidderMacro) MacroGDPR(key string) string {
	return ""
}

//MacroUSPrivacy contains definition for USPrivacy Parameter
func (tag *BidderMacro) MacroUSPrivacy(key string) string {
	return ""
}

/********************* Additional *********************/

//MacroCacheBuster contains definition for CacheBuster Parameter
func (tag *BidderMacro) MacroCacheBuster(key string) string {
	return ""
}

//Custom contains definition for CacheBuster Parameter
func (tag *BidderMacro) Custom(key string) string {
	return ""
}
