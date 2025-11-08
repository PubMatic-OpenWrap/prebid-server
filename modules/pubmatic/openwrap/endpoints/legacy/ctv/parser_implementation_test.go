package ctv

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func GetHTTPTestRequest(method, path string, values url.Values, headers http.Header) *http.Request {
	request, _ := http.NewRequest(method, path+"?"+values.Encode(), nil)
	request.Header = headers
	return request
}

func getTestValues() url.Values {
	return url.Values{
		//BidRequest
		ORTBBidRequestID:      {"381d2e0b-548d-4f27-bfdd-e6e66f43557e"},
		ORTBBidRequestTest:    {"1"},
		ORTBBidRequestAt:      {"1"},
		ORTBBidRequestTmax:    {"120"},
		ORTBBidRequestWseat:   {"nike,puma,sketchers"},
		ORTBBidRequestAllImps: {"1"},
		ORTBBidRequestCur:     {"USD"},
		ORTBBidRequestBcat:    {"IAB1-1"},
		ORTBBidRequestBadv:    {"ford.com"},
		ORTBBidRequestBapp:    {"com.foo.mygame"},
		ORTBBidRequestWlang:   {"EN"},
		ORTBBidRequestBseat:   {"adserver"},

		//Source
		ORTBSourceFD:     {"1"},
		ORTBSourceTID:    {"edc7717c-ca43-4ad6-b2a1-354bd8b10f78"},
		ORTBSourcePChain: {"pchaintagid"},

		//Regs
		ORTBRegsCoppa:        {"1"},
		ORTBRegsExtGdpr:      {"1"},
		ORTBRegsExtUSPrivacy: {"1"},

		//Imp
		ORTBImpID:                {"381d2e0b-548d-4f27-bfdd-e6e66f43557e"},
		ORTBImpDisplayManager:    {"PubMaticSDK"},
		ORTBImpDisplayManagerVer: {"PubMaticSDK-1.0"},
		ORTBImpInstl:             {"1"},
		ORTBImpTagID:             {"/15671365/DMDemo1"},
		ORTBImpBidFloor:          {"1.1"},
		ORTBImpBidFloorCur:       {"USD"},
		ORTBImpClickBrowser:      {"0"},
		ORTBImpSecure:            {"0"},
		ORTBImpIframeBuster:      {"1"},
		ORTBImpExp:               {"1"},
		ORTBImpPmp:               {`{"private_auction":1,"deals":[{"id":"123","bidfloor":1.2,"bidfloorcur":"USD","at":1,"wseat":["IAB-1","IAB-2"],"wadomain":["WD1","WD2"]}]}`},

		//Video
		ORTBImpVideoMimes:          {"video/3gpp,video/mp4,video/webm"},
		ORTBImpVideoMinDuration:    {"5"},
		ORTBImpVideoMaxDuration:    {"120"},
		ORTBImpVideoProtocols:      {"2,3,5,6,7,8"},
		ORTBImpVideoPlayerWidth:    {"320"},
		ORTBImpVideoPlayerHeight:   {"480"},
		ORTBImpVideoStartDelay:     {"0"},
		ORTBImpVideoPlacement:      {"5"},
		ORTBImpVideoPlcmt:          {"2"},
		ORTBImpVideoLinearity:      {"1"},
		ORTBImpVideoSkip:           {"1"},
		ORTBImpVideoSkipMin:        {"1"},
		ORTBImpVideoSkipAfter:      {"1"},
		ORTBImpVideoSequence:       {"1"},
		ORTBImpVideoBAttr:          {"1,2,3"},
		ORTBImpVideoMaxExtended:    {"10"},
		ORTBImpVideoMinBitrate:     {"1200"},
		ORTBImpVideoMaxBitrate:     {"2000"},
		ORTBImpVideoBoxingAllowed:  {"1"},
		ORTBImpVideoPlaybackMethod: {"1"},
		ORTBImpVideoDelivery:       {"2"},
		ORTBImpVideoPos:            {"7"},
		ORTBImpVideoAPI:            {"2"},
		ORTBImpVideoCompanionType:  {"1,2,3"},

		//Site
		ORTBSiteID:            {"123"},
		ORTBSiteName:          {"EBay Shopping"},
		ORTBSiteDomain:        {"ebay.com"},
		ORTBSitePage:          {"http://ebay.com/inte/automation/s2s/pwt_parameter_validation_muti_slot_multi_size.html?pwtvc=1&pwtv=1&profileid=3277"},
		ORTBSiteRef:           {"http://ebay.com/home"},
		ORTBSiteSearch:        {"New Cloths"},
		ORTBSiteMobile:        {"1"},
		ORTBSiteCat:           {"IAB1-5,IAB1-6"},
		ORTBSiteSectionCat:    {"IAB1-5"},
		ORTBSitePageCat:       {"IAB1-6"},
		ORTBSitePrivacyPolicy: {"1"},
		ORTBSiteKeywords:      {"Clothes"},

		//Site.Publisher
		ORTBSitePublisherID:     {"5890"},
		ORTBSitePublisherName:   {"Test Publisher"},
		ORTBSitePublisherCat:    {"IAB1-5"},
		ORTBSitePublisherDomain: {"publisher.com"},

		//Site.Content
		ORTBSiteContentID:                 {"381d2e0b-548d-4f27-bfdd-e6e66f43557e"},
		ORTBSiteContentEpisode:            {"1"},
		ORTBSiteContentTitle:              {"Star Wars"},
		ORTBSiteContentSeries:             {"Star Wars"},
		ORTBSiteContentSeason:             {"Season 3"},
		ORTBSiteContentArtist:             {"George Lucas"},
		ORTBSiteContentGenre:              {"Action"},
		ORTBSiteContentAlbum:              {"Action"},
		ORTBSiteContentIsRc:               {"2"},
		ORTBSiteContentURL:                {"http://www.pubmatic.com/test/"},
		ORTBSiteContentCat:                {"IAB1-1,IAB1-6"},
		ORTBSiteContentProdQ:              {"1"},
		ORTBSiteContentVideoQuality:       {"1"},
		ORTBSiteContentContext:            {"1"},
		ORTBSiteContentContentRating:      {"MPAA"},
		ORTBSiteContentUserRating:         {"9-Stars"},
		ORTBSiteContentQaGmeDiarating:     {"1"},
		ORTBSiteContentKeywords:           {"Action Movies"},
		ORTBSiteContentLiveStream:         {"1"},
		ORTBSiteContentSourceRelationship: {"1"},
		ORTBSiteContentLen:                {"12000"},
		ORTBSiteContentLanguage:           {"en-US"},
		ORTBSiteContentEmbeddable:         {"1"},

		//Site.Content.Network level parameters
		ORTBSiteContentNetworkID:     {"Test Site Network ID"},
		ORTBSiteContentNetworkName:   {"Test Site Network Name"},
		ORTBSiteContentNetworkDomain: {"Test Site Network Domain"},

		//Site.Content.Channel level parameters
		ORTBSiteContentChannelID:     {"Test Site Channel ID"},
		ORTBSiteContentChannelName:   {"Test Site Channel Name"},
		ORTBSiteContentChannelDomain: {"Test Site Channel Domain"},

		//Site.Content.Producer
		ORTBSiteContentProducerID:     {"123"},
		ORTBSiteContentProducerName:   {"Gary Kurtz"},
		ORTBSiteContentProducerCat:    {"IAB1-5,IAB1-6"},
		ORTBSiteContentProducerDomain: {"producer.com"},

		// Source
		ORTBSourceSChain: {"1.0,1!ASI1,SID1,1,RID1,Name1,Domain1"},

		//App
		ORTBAppID:            {"1234"},
		ORTBAppName:          {"MyFooGame"},
		ORTBAppBundle:        {"com.foo.mygame"},
		ORTBAppDomain:        {"mygame.foo.com"},
		ORTBAppStoreURL:      {"https://play.google.com/store/apps/details?id=com.foo.mygame"},
		ORTBAppVer:           {"1.1"},
		ORTBAppPaid:          {"1"},
		ORTBAppCat:           {"IAB1-5,IAB1-6"},
		ORTBAppSectionCat:    {"IAB1-5"},
		ORTBAppPageCat:       {"IAB1-6"},
		ORTBAppPrivacyPolicy: {"1"},
		ORTBAppKeywords:      {"Games"},

		//App.Publisher
		ORTBAppPublisherID:     {"5890"},
		ORTBAppPublisherName:   {"Test Publisher"},
		ORTBAppPublisherCat:    {"IAB1-5"},
		ORTBAppPublisherDomain: {"publisher.com"},

		//App.Content
		ORTBAppContentID:                 {"381d2e0b-548d-4f27-bfdd-e6e66f43557e"},
		ORTBAppContentEpisode:            {"1"},
		ORTBAppContentTitle:              {"Star Wars"},
		ORTBAppContentSeries:             {"Star Wars"},
		ORTBAppContentSeason:             {"Season 3"},
		ORTBAppContentArtist:             {"George Lucas"},
		ORTBAppContentGenre:              {"Action"},
		ORTBAppContentAlbum:              {"Action"},
		ORTBAppContentIsRc:               {"2"},
		ORTBAppContentURL:                {"http://www.pubmatic.com/test/"},
		ORTBAppContentCat:                {"IAB1-1,IAB1-6"},
		ORTBAppContentProdQ:              {"1"},
		ORTBAppContentVideoQuality:       {"1"},
		ORTBAppContentContext:            {"1"},
		ORTBAppContentContentRating:      {"MPAA"},
		ORTBAppContentUserRating:         {"9-Stars"},
		ORTBAppContentQaGmeDiarating:     {"1"},
		ORTBAppContentKeywords:           {"Action Movies"},
		ORTBAppContentLiveStream:         {"1"},
		ORTBAppContentSourceRelationship: {"1"},
		ORTBAppContentLen:                {"12000"},
		ORTBAppContentLanguage:           {"en-US"},
		ORTBAppContentEmbeddable:         {"1"},

		//App.Content.Network level parameters
		ORTBAppContentNetworkID:     {"Test App Network ID"},
		ORTBAppContentNetworkName:   {"Test App Network Name"},
		ORTBAppContentNetworkDomain: {"Test App Network Domain"},

		//App.Content.Channel level parameters
		ORTBAppContentChannelID:     {"Test App Channel ID"},
		ORTBAppContentChannelName:   {"Test App Channel Name"},
		ORTBAppContentChannelDomain: {"Test App Channel Domain"},

		//App.Content.Producer
		ORTBAppContentProducerID:     {"123"},
		ORTBAppContentProducerName:   {"Gary Kurtz"},
		ORTBAppContentProducerCat:    {"IAB1-5,IAB1-6"},
		ORTBAppContentProducerDomain: {"producer.com"},

		//Device
		ORTBDeviceUserAgent:      {"Mozilla%2F5.0%20},Windows%20NT%206.1%3B%20Win64%3B%20x64%3B%20rv%3A47.0)%20Gecko%2F20100101%20Firefox%2F47.0"},
		ORTBDeviceDnt:            {"1"},
		ORTBDeviceLmt:            {"1"},
		ORTBDeviceIP:             {"127.0.0.1"},
		ORTBDeviceIpv6:           {"2001:db8::8a2e:370:7334"},
		ORTBDeviceDeviceType:     {"1"},
		ORTBDeviceMake:           {"Samsung"},
		ORTBDeviceModel:          {"Galaxy-A70S"},
		ORTBDeviceOs:             {"Android"},
		ORTBDeviceOsv:            {"MarshMellow"},
		ORTBDeviceHwv:            {"A70s"},
		ORTBDeviceWidth:          {"1366"},
		ORTBDeviceHeight:         {"768"},
		ORTBDevicePpi:            {"4096"},
		ORTBDevicePxRatio:        {"1.3"},
		ORTBDeviceJS:             {"1"},
		ORTBDeviceGeoFetch:       {"0"},
		ORTBDeviceFlashVer:       {"1.1"},
		ORTBDeviceLanguage:       {"en-US"},
		ORTBDeviceCarrier:        {"VERIZON"},
		ORTBDeviceMccmnc:         {"310-005"},
		ORTBDeviceConnectionType: {"2"},
		ORTBDeviceIfa:            {"EA7583CD-A667-48BC-B806-42ECB2B48606"},
		ORTBDeviceDidSha1:        {"EA7583CD-A667-48BC-B806-42ECB2B48606"},
		ORTBDeviceDidMd5:         {"EA7583CD-A667-48BC-B806-42ECB2B48606"},
		ORTBDeviceDpidSha1:       {"EA7583CD-A667-48BC-B806-42ECB2B48606"},
		ORTBDeviceDpidMd5:        {"EA7583CD-A667-48BC-B806-42ECB2B48606"},
		ORTBDeviceMacSha1:        {"EA7583CD-A667-48BC-B806-42ECB2B48606"},
		ORTBDeviceMacMd5:         {"EA7583CD-A667-48BC-B806-42ECB2B48606"},

		//Device.Geo
		ORTBDeviceGeoLat:           {"72.6"},
		ORTBDeviceGeoLon:           {"72.6"},
		ORTBDeviceGeoType:          {"1"},
		ORTBDeviceGeoAccuracy:      {"10"},
		ORTBDeviceGeoLastFix:       {"0"},
		ORTBDeviceGeoIPService:     {"1"},
		ORTBDeviceGeoCountry:       {"India"},
		ORTBDeviceGeoRegion:        {"Maharashtra"},
		ORTBDeviceGeoRegionFips104: {"MAHA"},
		ORTBDeviceGeoMetro:         {"Mumbai"},
		ORTBDeviceGeoCity:          {"Mumbai"},
		ORTBDeviceGeoZip:           {"123456"},
		ORTBDeviceGeoUtcOffset:     {"120"},

		//User
		ORTBUserID:         {"45067fec-eab7-4ca0-ad3a-87b01f21846a"},
		ORTBUserBuyerUID:   {"45067fec-eab7-4ca0-ad3a-87b01f21846a"},
		ORTBUserYob:        {"1990"},
		ORTBUserGender:     {"M"},
		ORTBUserKeywords:   {"Movies"},
		ORTBUserCustomData: {"Star Wars"},
		ORTBUserExtConsent: {"BOEFEAyOEFEAyAHABDENAI4AAAB9vABAASA"},

		//User.Geo
		ORTBUserGeoLat:           {"72.6"},
		ORTBUserGeoLon:           {"72.6"},
		ORTBUserGeoType:          {"1"},
		ORTBUserGeoAccuracy:      {"10"},
		ORTBUserGeoLastFix:       {"0"},
		ORTBUserGeoIPService:     {"1"},
		ORTBUserGeoCountry:       {"India"},
		ORTBUserGeoRegion:        {"Maharashtra"},
		ORTBUserGeoRegionFips104: {"MAHA"},
		ORTBUserGeoMetro:         {"Mumbai"},
		ORTBUserGeoCity:          {"Mumbai"},
		ORTBUserGeoZip:           {"123456"},
		ORTBUserGeoUtcOffset:     {"120"},

		//ReqWrapperExtension
		ORTBProfileID:            {"1567"},
		ORTBVersionID:            {"2"},
		ORTBSSAuctionFlag:        {"0"},
		ORTBSumryDisableFlag:     {"0"},
		ORTBClientConfigFlag:     {"1"},
		ORTBSupportDeals:         {"true"},
		ORTBIncludeBrandCategory: {"2"},
		ORTBSSAI:                 {"mediatailor"},

		//VideoExtension
		ORTBImpVideoExtOffset:                           {"1"},
		ORTBImpVideoExtAdPodMinAds:                      {"2"},
		ORTBImpVideoExtAdPodMaxAds:                      {"3"},
		ORTBImpVideoExtAdPodMinDuration:                 {"4"},
		ORTBImpVideoExtAdPodMaxDuration:                 {"5"},
		ORTBImpVideoExtAdPodAdvertiserExclusionPercent:  {"6"},
		ORTBImpVideoExtAdPodIABCategoryExclusionPercent: {"7"},

		//ReqAdPodExt
		ORTBRequestExtAdPodMinAds:                              {"8"},
		ORTBRequestExtAdPodMaxAds:                              {"9"},
		ORTBRequestExtAdPodMinDuration:                         {"10"},
		ORTBRequestExtAdPodMaxDuration:                         {"11"},
		ORTBRequestExtAdPodAdvertiserExclusionPercent:          {"12"},
		ORTBRequestExtAdPodIABCategoryExclusionPercent:         {"13"},
		ORTBRequestExtAdPodCrossPodAdvertiserExclusionPercent:  {"14"},
		ORTBRequestExtAdPodCrossPodIABCategoryExclusionPercent: {"15"},
		ORTBRequestExtAdPodIABCategoryExclusionWindow:          {"16"},
		ORTBRequestExtAdPodAdvertiserExclusionWindow:           {"17"},

		//Extensions
		ORTBBidRequestExt + ".k1.k2.key":          {"req.ext.value"},
		ORTBSourceExt + ".k1.k2.key":              {"src.ext.value"},
		ORTBRegsExt + ".k1.k2.key":                {"regs.ext.value"},
		ORTBImpExt + ".k1.k2.key":                 {"imp.ext.value"},
		ORTBImpVideoExt + ".k1.k2.key":            {"imp.vid.ext.value"},
		ORTBSiteExt + ".k1.k2.key":                {"site.ext.value"},
		ORTBAppExt + ".k1.k2.key":                 {"app.ext.value"},
		ORTBSitePublisherExt + ".k1.k2.key":       {"site.pub.ext.value"},
		ORTBSiteContentExt + ".k1.k2.key":         {"site.cnt.ext.value"},
		ORTBAppContentNetworkExt + ".k1.k2.key":   {"app.cnt.net.ext.value"},
		ORTBAppContentChannelExt + ".k1.k2.key":   {"app.cnt.chan.ext.value"},
		ORTBSiteContentNetworkExt + ".k1.k2.key":  {"site.cnt.net.ext.value"},
		ORTBSiteContentChannelExt + ".k1.k2.key":  {"site.cnt.chan.ext.value"},
		ORTBSiteContentProducerExt + ".k1.k2.key": {"site.cnt.prod.ext.value"},
		ORTBAppPublisherExt + ".k1.k2.key":        {"app.pub.ext.value"},
		ORTBAppContentExt + ".k1.k2.key":          {"app.cnt.ext.value"},
		ORTBAppContentProducerExt + ".k1.k2.key":  {"app.cnt.prod.ext.value"},
		ORTBDeviceExt + ".k1.k2.key":              {"dev.ext.value"},
		ORTBDeviceGeoExt + ".k1.k2.key":           {"dev.geo.ext.value"},
		ORTBUserExt + ".k1.k2.key":                {"user.ext.value"},
		ORTBUserGeoExt + ".k1.k2.key":             {"user.geo.ext.value"},

		//ContentObjectTransparency
		ORTBRequestExtPrebidTransparencyContent: {`{"pubmatic": {"include": 1}}`},

		//
		ORTBUserData:               {`[{"name":"publisher.com","ext":{"segtax":4},"segment":[{"id":"1"}]}]`},
		ORTBUserExtEIDS:            {`[{"source":"bvod.connect","uids":[{"id":"OztamSession-123456","atype":501},{"id":"7D92078A-8246-4BA4-AE5B-76104861E7DC","atype":2,"ext":{"seq":1,"demgid":"1234"}},{"id":"8D92078A-8246-4BA4-AE5B-76104861E7DC","atype":2,"ext":{"seq":2,"demgid":"2345"}}]}]`},
		ORTBUserExtSessionDuration: {"40"},
		ORTBUserExtImpDepth:        {"10"},

		ORTBDeviceExtIfaType:   {"ORTBDeviceExtIfaType"},
		ORTBDeviceExtSessionID: {"ORTBDeviceExtSessionID"},
		ORTBDeviceExtATTS:      {"1"},
	}
}

func TestParseORTBRequest(t *testing.T) {
	values := getTestValues()
	values.Add("key-not-present", "123")

	expectedRequest := `{"id":"381d2e0b-548d-4f27-bfdd-e6e66f43557e","imp":[{"id":"381d2e0b-548d-4f27-bfdd-e6e66f43557e","video":{"mimes":["video/3gpp","video/mp4","video/webm"],"minduration":5,"maxduration":120,"startdelay":0,"protocols":[2,3,5,6,7,8],"w":320,"h":480,"placement":5,"plcmt":2,"linearity":1,"skip":1,"skipmin":1,"skipafter":1,"sequence":1,"battr":[1,2,3],"maxextended":10,"minbitrate":1200,"maxbitrate":2000,"boxingallowed":1,"playbackmethod":[1],"delivery":[2],"pos":7,"api":[2],"companiontype":[1,2,3],"ext":{"adpod":{"admaxduration":5,"adminduration":4,"excladv":6,"excliabcat":7,"maxads":3,"minads":2},"k1":{"k2":{"key":"imp.vid.ext.value"}},"offset":1}},"pmp":{"private_auction":1,"deals":[{"id":"123","bidfloor":1.2,"bidfloorcur":"USD","at":1,"wseat":["IAB-1","IAB-2"],"wadomain":["WD1","WD2"]}]},"displaymanager":"PubMaticSDK","displaymanagerver":"PubMaticSDK-1.0","instl":1,"tagid":"/15671365/DMDemo1","bidfloor":1.1,"bidfloorcur":"USD","clickbrowser":0,"secure":0,"iframebuster":["1"],"exp":1,"ext":{"k1":{"k2":{"key":"imp.ext.value"}}}}],"site":{"id":"123","name":"EBay Shopping","domain":"ebay.com","cat":["IAB1-5","IAB1-6"],"sectioncat":["IAB1-5"],"pagecat":["IAB1-6"],"page":"http://ebay.com/inte/automation/s2s/pwt_parameter_validation_muti_slot_multi_size.html?pwtvc=1&pwtv=1&profileid=3277","ref":"http://ebay.com/home","search":"New Cloths","mobile":1,"privacypolicy":1,"publisher":{"id":"5890","name":"Test Publisher","cat":["IAB1-5"],"domain":"publisher.com","ext":{"k1":{"k2":{"key":"site.pub.ext.value"}}}},"content":{"id":"381d2e0b-548d-4f27-bfdd-e6e66f43557e","episode":1,"title":"Star Wars","series":"Star Wars","season":"Season 3","artist":"George Lucas","genre":"Action","album":"Action","isrc":"2","producer":{"id":"123","name":"Gary Kurtz","cat":["IAB1-5","IAB1-6"],"domain":"producer.com","ext":{"k1":{"k2":{"key":"site.cnt.prod.ext.value"}}}},"url":"http://www.pubmatic.com/test/","cat":["IAB1-1","IAB1-6"],"prodq":1,"videoquality":1,"context":1,"contentrating":"MPAA","userrating":"9-Stars","qagmediarating":1,"keywords":"Action Movies","livestream":1,"sourcerelationship":1,"len":12000,"language":"en-US","embeddable":1,"network":{"id":"Test Site Network ID","name":"Test Site Network Name","domain":"Test Site Network Domain","ext":{"k1":{"k2":{"key":"site.cnt.net.ext.value"}}}},"channel":{"id":"Test Site Channel ID","name":"Test Site Channel Name","domain":"Test Site Channel Domain","ext":{"k1":{"k2":{"key":"site.cnt.chan.ext.value"}}}},"ext":{"k1":{"k2":{"key":"site.cnt.ext.value"}}}},"keywords":"Clothes","ext":{"k1":{"k2":{"key":"site.ext.value"}}}},"app":{"id":"1234","name":"MyFooGame","bundle":"com.foo.mygame","domain":"mygame.foo.com","storeurl":"https://play.google.com/store/apps/details?id=com.foo.mygame","cat":["IAB1-5","IAB1-6"],"sectioncat":["IAB1-5"],"pagecat":["IAB1-6"],"ver":"1.1","privacypolicy":1,"paid":1,"publisher":{"id":"5890","name":"Test Publisher","cat":["IAB1-5"],"domain":"publisher.com","ext":{"k1":{"k2":{"key":"app.pub.ext.value"}}}},"content":{"id":"381d2e0b-548d-4f27-bfdd-e6e66f43557e","episode":1,"title":"Star Wars","series":"Star Wars","season":"Season 3","artist":"George Lucas","genre":"Action","album":"Action","isrc":"2","producer":{"id":"123","name":"Gary Kurtz","cat":["IAB1-5","IAB1-6"],"domain":"producer.com","ext":{"k1":{"k2":{"key":"app.cnt.prod.ext.value"}}}},"url":"http://www.pubmatic.com/test/","cat":["IAB1-1","IAB1-6"],"prodq":1,"videoquality":1,"context":1,"contentrating":"MPAA","userrating":"9-Stars","qagmediarating":1,"keywords":"Action Movies","livestream":1,"sourcerelationship":1,"len":12000,"language":"en-US","embeddable":1,"network":{"id":"Test App Network ID","name":"Test App Network Name","domain":"Test App Network Domain","ext":{"k1":{"k2":{"key":"app.cnt.net.ext.value"}}}},"channel":{"id":"Test App Channel ID","name":"Test App Channel Name","domain":"Test App Channel Domain","ext":{"k1":{"k2":{"key":"app.cnt.chan.ext.value"}}}},"ext":{"k1":{"k2":{"key":"app.cnt.ext.value"}}}},"keywords":"Games","ext":{"k1":{"k2":{"key":"app.ext.value"}}}},"device":{"geo":{"lat":72.6,"lon":72.6,"type":1,"accuracy":10,"ipservice":1,"country":"India","region":"Maharashtra","regionfips104":"MAHA","metro":"Mumbai","city":"Mumbai","zip":"123456","utcoffset":120,"ext":{"k1":{"k2":{"key":"dev.geo.ext.value"}}}},"geofetch":0,"dnt":1,"lmt":1,"ua":"Mozilla%2F5.0%20},Windows%20NT%206.1%3B%20Win64%3B%20x64%3B%20rv%3A47.0)%20Gecko%2F20100101%20Firefox%2F47.0","ip":"127.0.0.1","ipv6":"2001:db8::8a2e:370:7334","devicetype":1,"make":"Samsung","model":"Galaxy-A70S","os":"Android","osv":"MarshMellow","hwv":"A70s","h":768,"w":1366,"ppi":4096,"pxratio":1.3,"js":1,"flashver":"1.1","language":"en-US","carrier":"VERIZON","mccmnc":"310-005","connectiontype":2,"ifa":"EA7583CD-A667-48BC-B806-42ECB2B48606","didsha1":"EA7583CD-A667-48BC-B806-42ECB2B48606","didmd5":"EA7583CD-A667-48BC-B806-42ECB2B48606","dpidsha1":"EA7583CD-A667-48BC-B806-42ECB2B48606","dpidmd5":"EA7583CD-A667-48BC-B806-42ECB2B48606","macsha1":"EA7583CD-A667-48BC-B806-42ECB2B48606","macmd5":"EA7583CD-A667-48BC-B806-42ECB2B48606","ext":{"atts":1,"ifa_type":"ORTBDeviceExtIfaType","k1":{"k2":{"key":"dev.ext.value"}},"session_id":"ORTBDeviceExtSessionID"}},"user":{"id":"45067fec-eab7-4ca0-ad3a-87b01f21846a","buyeruid":"45067fec-eab7-4ca0-ad3a-87b01f21846a","yob":1990,"gender":"M","keywords":"Movies","customdata":"Star Wars","geo":{"lat":72.6,"lon":72.6,"type":1,"accuracy":10,"ipservice":1,"country":"India","region":"Maharashtra","regionfips104":"MAHA","metro":"Mumbai","city":"Mumbai","zip":"123456","utcoffset":120,"ext":{"k1":{"k2":{"key":"user.geo.ext.value"}}}},"data":[{"name":"publisher.com","segment":[{"id":"1"}],"ext":{"segtax":4}}],"ext":{"consent":"BOEFEAyOEFEAyAHABDENAI4AAAB9vABAASA","sessionduration":40,"impdepth":10,"eids":[{"source":"bvod.connect","uids":[{"atype":501,"id":"OztamSession-123456"},{"atype":2,"ext":{"demgid":"1234","seq":1},"id":"7D92078A-8246-4BA4-AE5B-76104861E7DC"},{"atype":2,"ext":{"demgid":"2345","seq":2},"id":"8D92078A-8246-4BA4-AE5B-76104861E7DC"}]}],"k1":{"k2":{"key":"user.ext.value"}}}},"test":1,"at":1,"tmax":120,"wseat":["nike","puma","sketchers"],"bseat":["adserver"],"allimps":1,"cur":["USD"],"wlang":["EN"],"bcat":["IAB1-1"],"badv":["ford.com"],"bapp":["com.foo.mygame"],"source":{"fd":1,"tid":"edc7717c-ca43-4ad6-b2a1-354bd8b10f78","pchain":"pchaintagid","schain":{"complete":1,"nodes":[{"asi":"ASI1","sid":"SID1","rid":"RID1","name":"Name1","domain":"Domain1","hp":1}],"ver":"1.0"},"ext":{"k1":{"k2":{"key":"src.ext.value"}}}},"regs":{"coppa":1,"ext":{"gdpr":1,"k1":{"k2":{"key":"regs.ext.value"}},"us_privacy":"1"},"gdpr":1,"us_privacy":"1"},"ext":{"adpod":{"admaxduration":11,"adminduration":10,"crosspodexcladv":14,"crosspodexcliabcat":15,"excladv":12,"excladvwindow":17,"excliabcat":13,"excliabcatwindow":16,"maxads":9,"minads":8},"k1":{"k2":{"key":"req.ext.value"}},"prebid":{"transparency":{"content":{"pubmatic":{"include":1}}}},"wrapper":{"clientconfig":1,"includebrandcategory":2,"profileid":1567,"ssai":"mediatailor","ssauction":0,"sumry_disable":0,"supportdeals":true,"versionid":2}}}`

	request := GetHTTPTestRequest("GET", "/ortb/vast", values, http.Header{
		"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36"},
		http.CanonicalHeaderKey("RLNCLIENTIPADDR"): {"172.16.8.74"},
		http.CanonicalHeaderKey("SOURCE_IP"):       {"172.16.8.74"},
	})

	var parser ORTBParser = NewOpenRTB(request)

	ortb, err := parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		actualRequest, err := json.Marshal(ortb)
		assert.NoError(t, err)
		assert.JSONEq(t, expectedRequest, string(actualRequest))
	}
}

func TestParseORTBRequestInvalid(t *testing.T) {
	values := url.Values{}
	values.Add(ORTBImpPmp, "{at:123}")

	for k, v := range values {
		request := GetHTTPTestRequest("GET", "/ortb/vast", url.Values{k: v}, http.Header{
			"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36"},
			http.CanonicalHeaderKey("RLNCLIENTIPADDR"): {"172.16.8.74"},
			http.CanonicalHeaderKey("SOURCE_IP"):       {"172.16.8.74"},
		})
		var parser ORTBParser = NewOpenRTB(request)
		ortb, err := parser.ParseORTBRequest(GetORTBParserMap())
		assert.NotNil(t, ortb)
		assert.Error(t, err)
	}
}

func TestParseORTBRequestParsingFailed(t *testing.T) {
	values := url.Values{}
	values.Add(ORTBBidRequestTest, "abc")

	for k, v := range values {
		request := GetHTTPTestRequest("GET", "/ortb/vast", url.Values{k: v}, http.Header{
			"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36"},
			http.CanonicalHeaderKey("RLNCLIENTIPADDR"): {"172.16.8.74"},
			http.CanonicalHeaderKey("SOURCE_IP"):       {"172.16.8.74"},
		})
		var parser ORTBParser = NewOpenRTB(request)
		ortb, err := parser.ParseORTBRequest(GetORTBParserMap())
		assert.NotNil(t, ortb)
		assert.Error(t, err)
	}
}

func TestParseORTBRequestEmptyFields(t *testing.T) {
	request := GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"src.fd": []string{"1"}}, http.Header{})
	parser := NewOpenRTB(request)
	ortb, err := parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["source"].(map[string]interface{})["fd"], "1")
	}

	request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"src.tid": []string{"1"}}, http.Header{})
	parser = NewOpenRTB(request)
	ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["source"].(map[string]interface{})["tid"], "1")
	}

	request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.id": []string{"site.id"}}, http.Header{})
	parser = NewOpenRTB(request)
	ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["site"].(map[string]interface{})["id"], "site.id")
	}

	request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.name": []string{"site.name"}}, http.Header{})
	parser = NewOpenRTB(request)
	ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["site"].(map[string]interface{})["name"], "site.name")
	}

	request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.domain": []string{"site.domain"}}, http.Header{})
	parser = NewOpenRTB(request)
	ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["site"].(map[string]interface{})["domain"], "site.domain")
	}

	request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.page": []string{"site.Page"}}, http.Header{})
	parser = NewOpenRTB(request)
	ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["site"].(map[string]interface{})["page"], "site.Page")
	}

	request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.ref": []string{"site.ref"}}, http.Header{})
	parser = NewOpenRTB(request)
	ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["site"].(map[string]interface{})["ref"], "site.ref")
	}

	request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.mobile": []string{"1"}}, http.Header{})
	parser = NewOpenRTB(request)
	ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["site"].(map[string]interface{})["mobile"], "1")
	}

	request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.search": []string{"site.Search"}}, http.Header{})
	parser = NewOpenRTB(request)
	ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["site"].(map[string]interface{})["search"], "site.Search")
	}

	request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cat": []string{"site,cat"}}, http.Header{})
	parser = NewOpenRTB(request)
	ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["site"].(map[string]interface{})["cat"], []string{"site", "cat"})
	}

	request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.sectioncat": []string{"site,sectioncat"}}, http.Header{})
	parser = NewOpenRTB(request)
	ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["site"].(map[string]interface{})["sectioncat"], []string{"site", "sectioncat"})
	}

	request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.privacypolicy": []string{"1"}}, http.Header{})
	parser = NewOpenRTB(request)
	ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["site"].(map[string]interface{})["privacypolicy"], "1")
	}

	request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.pagecat": []string{"site,pagecat"}}, http.Header{})
	parser = NewOpenRTB(request)
	ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["site"].(map[string]interface{})["pagecat"], []string{"site", "pagecat"})
	}

	request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.keywords": []string{"site.keywords"}}, http.Header{})
	parser = NewOpenRTB(request)
	ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["site"].(map[string]interface{})["keywords"], "site.keywords")
	}

	request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.pub.id": []string{"site.pub.id"}}, http.Header{})
	parser = NewOpenRTB(request)
	ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["site"].(map[string]interface{})["publisher"].(map[string]interface{})["id"], "site.pub.id")
	}

	request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.pub.name": []string{"site.pub.name"}}, http.Header{})
	parser = NewOpenRTB(request)
	ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["site"].(map[string]interface{})["publisher"].(map[string]interface{})["name"], "site.pub.name")
	}

	request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.pub.cat": []string{"site,pub,cat"}}, http.Header{})
	parser = NewOpenRTB(request)
	ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["site"].(map[string]interface{})["publisher"].(map[string]interface{})["cat"], []string{"site", "pub", "cat"})
	}

	request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.pub.domain": []string{"site.pub.domain"}}, http.Header{})
	parser = NewOpenRTB(request)
	ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["site"].(map[string]interface{})["publisher"].(map[string]interface{})["domain"], "site.pub.domain")
	}

	request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.id": []string{"site.cnt.id"}}, http.Header{})
	parser = NewOpenRTB(request)
	ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["site"].(map[string]interface{})["content"].(map[string]interface{})["id"], "site.cnt.id")
	}

	request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.episode": []string{"1"}}, http.Header{})
	parser = NewOpenRTB(request)
	ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	if assert.NoError(t, err) {
		assert.Equal(t, ortb["site"].(map[string]interface{})["content"].(map[string]interface{})["episode"], "1")
	}

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.title": []string{"site.cnt.title"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.Title, "site.cnt.title")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.series": []string{"site.cnt.series"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.Series, "site.cnt.series")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.season": []string{"site.cnt.Season"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.Season, "site.cnt.Season")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.artist": []string{"site.cnt.artist"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.Artist, "site.cnt.artist")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.genre": []string{"site.cnt.Genre"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.Genre, "site.cnt.Genre")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.isrc": []string{"site.cnt.ISRC"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.ISRC, "site.cnt.ISRC")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.album": []string{"site.cnt.album"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.Album, "site.cnt.album")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.url": []string{"site.cnt.url"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.URL, "site.cnt.url")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.cat": []string{"site,cnt,cat"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.Cat, []string{"site", "cnt", "cat"})
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.prodq": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.ProdQ, ptrutil.ToPtr(adcom1.ProductionQuality(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.videoquality": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.VideoQuality, ptrutil.ToPtr(adcom1.ProductionQuality(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.context": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.Context, adcom1.ContentContext(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.contentrating": []string{"site.cnt.contentrating"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.ContentRating, "site.cnt.contentrating")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.userrating": []string{"site.cnt.userrating"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.UserRating, "site.cnt.userrating")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.qagmediarating": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.QAGMediaRating, adcom1.MediaRating(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.keywords": []string{"site.cnt.keywords"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.Keywords, "site.cnt.keywords")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.livestream": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.LiveStream, ptrutil.ToPtr(int8(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.sourcerelationship": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.SourceRelationship, ptrutil.ToPtr(int8(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.len": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.Len, int64(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.len": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.Len, int64(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.language": []string{"site.cnt.language"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.Language, "site.cnt.language")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.embeddable": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.Embeddable, ptrutil.ToPtr(int8(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.prod.id": []string{"site.cnt.prod.id"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.Producer.ID, "site.cnt.prod.id")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.prod.name": []string{"site.cnt.prod.name"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.Producer.Name, "site.cnt.prod.name")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.prod.cat": []string{"site,cnt,prod,cat"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.Producer.Cat, []string{"site", "cnt", "prod", "cat"})
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"site.cnt.prod.domain": []string{"site.cnt.prod.domain"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Site.Content.Producer.Domain, "site.cnt.prod.domain")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.id": []string{"app.id"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.ID, "app.id")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.name": []string{"app.name"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Name, "app.name")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.bundle": []string{"app.bundle"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Bundle, "app.bundle")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.domain": []string{"app.domain"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Domain, "app.domain")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.storeurl": []string{"app.storeurl"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.StoreURL, "app.storeurl")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.ver": []string{"app.ver"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Ver, "app.ver")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.paid": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Paid, ptrutil.ToPtr(int8(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cat": []string{"app,cat"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Cat, []string{"app", "cat"})
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.sectioncat": []string{"app,section,cat"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.SectionCat, []string{"app", "section", "cat"})
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.pagecat": []string{"app,page,cat"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.PageCat, []string{"app", "page", "cat"})
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.privacypolicy": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.PrivacyPolicy, ptrutil.ToPtr(int8(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.keywords": []string{"app.keywords"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Keywords, "app.keywords")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.pub.id": []string{"app.pub.id"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Publisher.ID, "app.pub.id")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.pub.name": []string{"app.pub.name"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Publisher.Name, "app.pub.name")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.pub.cat": []string{"app,pub,cat"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Publisher.Cat, []string{"app", "pub", "cat"})
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.pub.domain": []string{"app.pub.domain"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Publisher.Domain, "app.pub.domain")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.id": []string{"app.cnt.id"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.ID, "app.cnt.id")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.episode": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.Episode, int64(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.title": []string{"app.cnt.title"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.Title, "app.cnt.title")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.series": []string{"app.cnt.series"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.Series, "app.cnt.series")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.season": []string{"app.cnt.season"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.Season, "app.cnt.season")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.artist": []string{"app.cnt.artist"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.Artist, "app.cnt.artist")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.genre": []string{"app.cnt.genre"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.Genre, "app.cnt.genre")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.album": []string{"app.cnt.album"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.Album, "app.cnt.album")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.isrc": []string{"app.cnt.isrc"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.ISRC, "app.cnt.isrc")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.url": []string{"app.cnt.url"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.URL, "app.cnt.url")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.cat": []string{"app,cnt,cat"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.Cat, []string{"app", "cnt", "cat"})
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.prodq": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.ProdQ, ptrutil.ToPtr(adcom1.ProductionQuality(int8(1))))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.videoquality": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.VideoQuality, ptrutil.ToPtr(adcom1.ProductionQuality(int8(1))))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.context": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.Context, adcom1.ContentContext(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.contentrating": []string{"app.cnt.contentrating"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.ContentRating, "app.cnt.contentrating")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.userrating": []string{"app.cnt.userrating"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.UserRating, "app.cnt.userrating")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.qagmediarating": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.QAGMediaRating, adcom1.MediaRating(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.keywords": []string{"app.cnt.keywords"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.Keywords, "app.cnt.keywords")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.livestream": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.LiveStream, ptrutil.ToPtr(int8(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.sourcerelationship": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.SourceRelationship, ptrutil.ToPtr(int8(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.len": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.Len, int64(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.language": []string{"app.cnt.language"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.Language, "app.cnt.language")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.embeddable": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.Embeddable, ptrutil.ToPtr(int8(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.prod.id": []string{"app.cnt.prod.id"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.Producer.ID, "app.cnt.prod.id")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.prod.name": []string{"app.cnt.prod.name"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.Producer.Name, "app.cnt.prod.name")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.prod.cat": []string{"app,cnt,prod,cat"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.Producer.Cat, []string{"app", "cnt", "prod", "cat"})
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"app.cnt.prod.domain": []string{"app.cnt.prod.domain"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.App.Content.Producer.Domain, "app.cnt.prod.domain")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.mimes": []string{"imp,vid,mimes"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.MIMEs, []string{"imp", "vid", "mimes"})
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.minduration": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.MinDuration, int64(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.maxduration": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.MaxDuration, int64(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.protocols": []string{"1,2"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.Protocols, []adcom1.MediaCreativeSubtype{1, 2})
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.w": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.W, ptrutil.ToPtr(int64(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.h": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.H, ptrutil.ToPtr(int64(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.startdelay": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.StartDelay, ptrutil.ToPtr(adcom1.StartDelay(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.placement": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.Placement, adcom1.VideoPlacementSubtype(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.plcmt": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.Plcmt, adcom1.VideoPlcmtSubtype(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.linearity": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.Linearity, adcom1.LinearityMode(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.skip": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.Skip, ptrutil.ToPtr(int8(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.skipmin": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.SkipMin, int64(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.sequence": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.Sequence, int8(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.battr": []string{"1,2"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.BAttr, []adcom1.CreativeAttribute{1, 2})
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.maxextended": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.MaxExtended, int64(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.minbitrate": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.MinBitRate, int64(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.maxbitrate": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.MaxBitRate, int64(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.boxingallowed": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.BoxingAllowed, ptrutil.ToPtr(int8(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.playbackmethod": []string{"1,2"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.PlaybackMethod, []adcom1.PlaybackMethod{1, 2})
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.delivery": []string{"1,2"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.Delivery, []adcom1.DeliveryMethod{1, 2})
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.pos": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.Pos, ptrutil.ToPtr(adcom1.PlacementPosition(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.api": []string{"1,2"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.API, []adcom1.APIFramework{1, 2})
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.companiontype": []string{"1,2"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.CompanionType, []adcom1.CompanionType{1, 2})
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.vid.skipafter": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Video.SkipAfter, int64(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"regs.coppa": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Regs.COPPA, int8(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{
	// 	"imp.bidfloor":    []string{"a"},
	// 	"imp.bidfloorcur": []string{"a"},
	// }, http.Header{})
	// parser = NewOpenRTB(request)
	// _, err = parser.ParseORTBRequest(GetORTBParserMap())
	// assert.Error(t, err)

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.ext.bidder": []string{""}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Ext, json.RawMessage(nil))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.ext.bidder": []string{"{"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.Error(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Ext, json.RawMessage(nil))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"imp.ext.bidder": []string{"{}"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Imp[0].Ext, json.RawMessage(`{"bidder":{}}`))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.ua": []string{"dev.ua"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Device.UA, "dev.ua")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.ip": []string{"dev.ip"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Device.IP, "dev.ip")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.ipv6": []string{"dev.ipv6"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Device.IPv6, "dev.ipv6")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.dnt": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Device.DNT, ptrutil.ToPtr(int8(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.lmt": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Device.Lmt, ptrutil.ToPtr(int8(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.devicetype": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Device.DeviceType, adcom1.DeviceType(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.make": []string{"dev.make"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Device.Make, "dev.make")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.model": []string{"dev.model"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Device.Model, "dev.model")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.os": []string{"dev.os"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Device.OS, "dev.os")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.osv": []string{"dev.osv"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Device.OSV, "dev.osv")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.hwv": []string{"dev.hwv"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Device.HWV, "dev.hwv")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.h": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Device.H, int64(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.w": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Device.W, int64(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.ppi": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Device.PPI, int64(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.pxratio": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Device.PxRatio, float64(1))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.js": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Device.JS, ptrutil.ToPtr(int8(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.geofetch": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Device.GeoFetch, ptrutil.ToPtr(int8(1)))
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.flashver": []string{"dev.flashver"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Device.FlashVer, "dev.flashver")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.language": []string{"dev.language"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	assert.Equal(t, ortb.Device.Language, "dev.language")
	// }

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.ext.atts": []string{"invalid_value"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// _, err = parser.ParseORTBRequest(GetORTBParserMap())
	// assert.Equal(t, err.Error(), "[parsing error key:dev.ext.atts msg:strconv.ParseFloat: parsing \"invalid_value\": invalid syntax]", "dev.ext.atts error does not match")

	// request = GetHTTPTestRequest("GET", "/ortb/vast", url.Values{"dev.ext.atts": []string{"1"}}, http.Header{})
	// parser = NewOpenRTB(request)
	// ortb, err = parser.ParseORTBRequest(GetORTBParserMap())
	// if assert.NoError(t, err) {
	// 	val, _ := jsonparser.GetFloat(ortb.Device.Ext, ORTBExtATTS)
	// 	assert.Equal(t, 1, int(val))
	// }
}

// func TestORTBRequestExtPrebidTransparencyContent(t *testing.T) {
// 	type fields struct {
// 		request *http.Request
// 		values  URLValues
// 		ortb    *openrtb2.BidRequest
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		wantErr bool
// 	}{
// 		{
// 			name: "Invalid value for content object",
// 			fields: fields{
// 				values: URLValues{
// 					Values: map[string][]string{
// 						"req.ext.prebid.transparency.content": {"abc"},
// 					},
// 				},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "Valid value for content object",
// 			fields: fields{
// 				values: URLValues{
// 					Values: map[string][]string{
// 						"req.ext.prebid.transparency.content": {`{"pubmatic":{"include":false,"keys":["title"]}}`},
// 					},
// 				},
// 				ortb: &openrtb2.BidRequest{},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "Valid value for content object - empty",
// 			fields: fields{
// 				values: URLValues{
// 					Values: map[string][]string{
// 						"req.ext.prebid.transparency.content": {`{}`},
// 					},
// 				},
// 				ortb: &openrtb2.BidRequest{},
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			o := &OpenRTB{
// 				request: tt.fields.request,
// 				values:  tt.fields.values,
// 				ortb:    tt.fields.ortb,
// 			}
// 			if err := o.ORTBRequestExtPrebidTransparencyContent(); (err != nil) != tt.wantErr {
// 				t.Errorf("OpenRTB.ORTBRequestExtPrebidTransparencyContent() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestORTBExtPrebidFloorsEnforceFloorDeals(t *testing.T) {
// 	type fields struct {
// 		request *http.Request
// 		values  URLValues
// 		ortb    *openrtb2.BidRequest
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		wantErr bool
// 	}{
// 		{
// 			name: "Add enforcement data in ortb floor extension",
// 			fields: fields{
// 				request: nil,
// 				values: URLValues{
// 					Values: url.Values{
// 						"req.ext.prebid.floors.enforcement": []string{"%7B%22enforcepbs%22%3A%20true%2C%22floordeals%22%3A%20true%7D"},
// 					},
// 				},
// 				ortb: func() *openrtb2.BidRequest {
// 					r := openrtb2.BidRequest{
// 						ID: "123",
// 					}
// 					return &r
// 				}(),
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			o := &OpenRTB{
// 				request: tt.fields.request,
// 				values:  tt.fields.values,
// 				ortb:    tt.fields.ortb,
// 			}
// 			if err := o.ORTBExtPrebidFloorsEnforceFloorDeals(); (err != nil) != tt.wantErr {
// 				t.Errorf("OpenRTB.ORTBExtPrebidFloorsEnforceFloorDeals() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			assert.Equal(t, string(o.ortb.Ext), "{\"prebid\":{\"floors\":{\"enforcement\":{\"enforcepbs\":true,\"floordeals\":true}}}}", "enforcement object is not updated properly")
// 		})
// 	}
// }

// func TestORTBImpBidFloor(t *testing.T) {
// 	type fields struct {
// 		request *http.Request
// 		values  URLValues
// 		ortb    *openrtb2.BidRequest
// 	}
// 	tests := []struct {
// 		name      string
// 		fields    fields
// 		wantErr   bool
// 		wantFloor float64
// 	}{
// 		{
// 			name: "valid bidfloor value present, but currency not available in request",
// 			fields: fields{
// 				request: nil,
// 				values: URLValues{
// 					Values: url.Values{
// 						"imp.bidfloor": []string{"20"},
// 					},
// 				},
// 				ortb: &openrtb2.BidRequest{
// 					Imp: []openrtb2.Imp{
// 						{},
// 					},
// 				},
// 			},
// 			wantErr:   false,
// 			wantFloor: 20,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			o := &OpenRTB{
// 				request: tt.fields.request,
// 				values:  tt.fields.values,
// 				ortb:    tt.fields.ortb,
// 			}
// 			if err := o.ORTBImpBidFloor(); (err != nil) != tt.wantErr {
// 				t.Errorf("OpenRTB.ORTBImpBidFloor() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			assert.Equal(t, tt.wantFloor, o.ortb.Imp[0].BidFloor, "Bid Floor value does not match")
// 		})
// 	}
// }

// func TestORTBImpBidFloorCur(t *testing.T) {
// 	type fields struct {
// 		request *http.Request
// 		values  URLValues
// 		ortb    *openrtb2.BidRequest
// 	}
// 	tests := []struct {
// 		name              string
// 		fields            fields
// 		wantErr           bool
// 		wantFloorCurrency string
// 	}{
// 		{
// 			name: "valid bidfloor currency value present, but floor value not available in request",
// 			fields: fields{
// 				request: nil,
// 				values: URLValues{
// 					Values: url.Values{
// 						"imp.bidfloorcur": []string{"USD"},
// 					},
// 				},
// 				ortb: &openrtb2.BidRequest{
// 					Imp: []openrtb2.Imp{
// 						{},
// 					},
// 				},
// 			},
// 			wantErr:           false,
// 			wantFloorCurrency: "",
// 		},
// 		{
// 			name: "valid bidfloor currency and bidfloor value present",
// 			fields: fields{
// 				request: nil,
// 				values: URLValues{
// 					Values: url.Values{
// 						"imp.bidfloorcur": []string{"INR"},
// 						"imp.bidfloor":    []string{"20"},
// 					},
// 				},
// 				ortb: &openrtb2.BidRequest{
// 					Imp: []openrtb2.Imp{
// 						{},
// 					},
// 				},
// 			},
// 			wantErr:           false,
// 			wantFloorCurrency: "INR",
// 		},
// 		{
// 			name: "when floor value is zero, floorval and floor currency will be discarded",
// 			fields: fields{
// 				request: nil,
// 				values: URLValues{
// 					Values: url.Values{
// 						"imp.bidfloorcur": []string{"INR"},
// 						"imp.bidfloor":    []string{"0"},
// 					},
// 				},
// 				ortb: &openrtb2.BidRequest{
// 					Imp: []openrtb2.Imp{
// 						{},
// 					},
// 				},
// 			},
// 			wantErr:           false,
// 			wantFloorCurrency: "",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			o := &OpenRTB{
// 				request: tt.fields.request,
// 				values:  tt.fields.values,
// 				ortb:    tt.fields.ortb,
// 			}
// 			if err := o.ORTBImpBidFloorCur(); (err != nil) != tt.wantErr {
// 				t.Errorf("OpenRTB.ORTBImpBidFloorCur() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			assert.Equal(t, tt.wantFloorCurrency, o.ortb.Imp[0].BidFloorCur, "Currency value does not match")
// 		})
// 	}
// }

// func TestOpenRTBORTBImpExtPrebidFloorMin(t *testing.T) {
// 	type fields struct {
// 		Parser  Parser
// 		request *http.Request
// 		values  URLValues
// 		ortb    *openrtb2.BidRequest
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		wantErr bool
// 		want    json.RawMessage
// 	}{
// 		{
// 			name: "Floor Min present in imp.ext.prebid",
// 			fields: fields{
// 				Parser: nil,
// 				request: func() *http.Request {
// 					r := httptest.NewRequest("GET", "http://localhost:8001/video/openrtb?imp.vid.maxbitrate=2000&imp.vid.boxingallowed=1&imp.secure=0&req.ext.wrapper.ssauction=0&req.ext.wrapper.sumry_disable=0&req.ext.wrapper.clientconfig=1&req.at=1&app.name=OpenWrapperSample&imp.ext.bidder=%7B%22appnexus%22%3A%7B%22keywords%22%3A%5B%5D%2C%22dealtier%22%3A%7B%22prefix%22%3A%22apnx%22%2C%22mindealtier%22%3A4%7D%7D%2C%22pubmatic%22%3A%7B%22keywords%22%3A%5B%7B%22key%22%3A%22dctr%22%2C%22value%22%3A%5B%22abBucket%3D4%7CadType%3Dpage%22%5D%7D%2C%7B%22key%22%3A%22pmZoneID%22%2C%22value%22%3A%5B%22Zone1%22%2C%22Zone2%22%5D%7D%5D%2C%22dealtier%22%3A%7B%22prefix%22%3A%22pubdeal%22%2C%22mindealtier%22%3A5%7D%7D%7D&src.tid=edc7717c-ca43-4ad6-b2a1-354bd8b10f78&imp.tagid=%2F15671365%2FMG_VideoAdUnit&app.ver=1.0&imp.vid.delivery=2&req.cur=USD&req.ext.wrapper.versionid=2&app.storeurl=https%3A%2F%2Fitunes.apple.com%2Fus%2Fapp%2Fpubmatic-sdk-app%2Fid1175273098%3Fvideobid%3D10&app.pub.id=5890&app.bundle=com.pubmatic.openbid.app&imp.vid.placement=5&imp.vid.mimes=video%2F3gpp%2Cvideo%2Fmp4%2Cvideo%2Fwebm&req.id=1559039248176&owLoggerDebug=1&imp.vid.protocols=2%2C3%2C5%2C6%2C7%2C8&imp.id=28635736ddc2bb&imp.vid.pos=7&req.ext.wrapper.profileid=13573&imp.vid.companiontype=1%2C2%2C3&imp.vid.startdelay=0&imp.vid.linearity=1&imp.vid.playbackmethod=1&debug=1&imp.ext.prebid=%7B%22floors%22%3A%7B%22floormin%22%3A17%2C%22floormincur%22%3A%22USD%22%7D%7D", nil)
// 					return r
// 				}(),
// 				values: URLValues{
// 					Values: url.Values{
// 						"imp.ext.prebid": []string{"{\"floors\":{\"floormin\":17,\"floormincur\":\"USD\"}}"},
// 					},
// 				},
// 				ortb: &openrtb2.BidRequest{
// 					Imp: []openrtb2.Imp{
// 						{},
// 					},
// 				},
// 			},
// 			wantErr: false,
// 			want:    json.RawMessage(`{"prebid":{"floors":{"floormin":17,"floormincur":"USD"}}}`),
// 		},
// 		{
// 			name: "Floor Min not present in imp.ext.prebid",
// 			fields: fields{
// 				Parser: nil,
// 				request: func() *http.Request {
// 					r := httptest.NewRequest("GET", "http://localhost:8001/video/openrtb?imp.vid.maxbitrate=2000&imp.vid.boxingallowed=1&imp.secure=0&req.ext.wrapper.ssauction=0&req.ext.wrapper.sumry_disable=0&req.ext.wrapper.clientconfig=1&req.at=1&app.name=OpenWrapperSample&imp.ext.bidder=%7B%22appnexus%22%3A%7B%22keywords%22%3A%5B%5D%2C%22dealtier%22%3A%7B%22prefix%22%3A%22apnx%22%2C%22mindealtier%22%3A4%7D%7D%2C%22pubmatic%22%3A%7B%22keywords%22%3A%5B%7B%22key%22%3A%22dctr%22%2C%22value%22%3A%5B%22abBucket%3D4%7CadType%3Dpage%22%5D%7D%2C%7B%22key%22%3A%22pmZoneID%22%2C%22value%22%3A%5B%22Zone1%22%2C%22Zone2%22%5D%7D%5D%2C%22dealtier%22%3A%7B%22prefix%22%3A%22pubdeal%22%2C%22mindealtier%22%3A5%7D%7D%7D&src.tid=edc7717c-ca43-4ad6-b2a1-354bd8b10f78&imp.tagid=%2F15671365%2FMG_VideoAdUnit&app.ver=1.0&imp.vid.delivery=2&req.cur=USD&req.ext.wrapper.versionid=2&app.storeurl=https%3A%2F%2Fitunes.apple.com%2Fus%2Fapp%2Fpubmatic-sdk-app%2Fid1175273098%3Fvideobid%3D10&app.pub.id=5890&app.bundle=com.pubmatic.openbid.app&imp.vid.placement=5&imp.vid.mimes=video%2F3gpp%2Cvideo%2Fmp4%2Cvideo%2Fwebm&req.id=1559039248176&owLoggerDebug=1&imp.vid.protocols=2%2C3%2C5%2C6%2C7%2C8&imp.id=28635736ddc2bb&imp.vid.pos=7&req.ext.wrapper.profileid=13573&imp.vid.companiontype=1%2C2%2C3&imp.vid.startdelay=0&imp.vid.linearity=1&imp.vid.playbackmethod=1&debug=1&imp.ext.prebid=%7B%22floors%22%3A%7B%22floormin%22%3A17%2C%22floormincur%22%3A%22USD%22%7D%7D", nil)
// 					return r
// 				}(),
// 				values: URLValues{
// 					Values: url.Values{
// 						"imp.ext.prebid": []string{""},
// 					},
// 				},
// 				ortb: &openrtb2.BidRequest{
// 					Imp: []openrtb2.Imp{
// 						{},
// 					},
// 				},
// 			},
// 			wantErr: false,
// 			want:    nil,
// 		},
// 		{
// 			name: "Floor Min present in imp.ext.prebid with invalid json",
// 			fields: fields{
// 				Parser: nil,
// 				request: func() *http.Request {
// 					r := httptest.NewRequest("GET", "http://localhost:8001/video/openrtb?imp.vid.maxbitrate=2000&imp.vid.boxingallowed=1&imp.secure=0&req.ext.wrapper.ssauction=0&req.ext.wrapper.sumry_disable=0&req.ext.wrapper.clientconfig=1&req.at=1&app.name=OpenWrapperSample&imp.ext.bidder=%7B%22appnexus%22%3A%7B%22keywords%22%3A%5B%5D%2C%22dealtier%22%3A%7B%22prefix%22%3A%22apnx%22%2C%22mindealtier%22%3A4%7D%7D%2C%22pubmatic%22%3A%7B%22keywords%22%3A%5B%7B%22key%22%3A%22dctr%22%2C%22value%22%3A%5B%22abBucket%3D4%7CadType%3Dpage%22%5D%7D%2C%7B%22key%22%3A%22pmZoneID%22%2C%22value%22%3A%5B%22Zone1%22%2C%22Zone2%22%5D%7D%5D%2C%22dealtier%22%3A%7B%22prefix%22%3A%22pubdeal%22%2C%22mindealtier%22%3A5%7D%7D%7D&src.tid=edc7717c-ca43-4ad6-b2a1-354bd8b10f78&imp.tagid=%2F15671365%2FMG_VideoAdUnit&app.ver=1.0&imp.vid.delivery=2&req.cur=USD&req.ext.wrapper.versionid=2&app.storeurl=https%3A%2F%2Fitunes.apple.com%2Fus%2Fapp%2Fpubmatic-sdk-app%2Fid1175273098%3Fvideobid%3D10&app.pub.id=5890&app.bundle=com.pubmatic.openbid.app&imp.vid.placement=5&imp.vid.mimes=video%2F3gpp%2Cvideo%2Fmp4%2Cvideo%2Fwebm&req.id=1559039248176&owLoggerDebug=1&imp.vid.protocols=2%2C3%2C5%2C6%2C7%2C8&imp.id=28635736ddc2bb&imp.vid.pos=7&req.ext.wrapper.profileid=13573&imp.vid.companiontype=1%2C2%2C3&imp.vid.startdelay=0&imp.vid.linearity=1&imp.vid.playbackmethod=1&debug=1&imp.ext.prebid=%7B%22floors%22%3A%7B%22floormin%22%3A17%2C%22floormincur%22%3A%22USD%22%7D%7D", nil)
// 					return r
// 				}(),
// 				values: URLValues{
// 					Values: url.Values{
// 						"imp.ext.prebid": []string{"%%"},
// 					},
// 				},
// 				ortb: &openrtb2.BidRequest{
// 					Imp: []openrtb2.Imp{
// 						{},
// 					},
// 				},
// 			},
// 			wantErr: true,
// 			want:    nil,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			o := &OpenRTB{
// 				request: tt.fields.request,
// 				values:  tt.fields.values,
// 				ortb:    tt.fields.ortb,
// 			}
// 			if err := o.ORTBImpExtPrebid(); (err != nil) != tt.wantErr {
// 				t.Errorf("OpenRTB.ORTBImpExtPrebid() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			assert.Equal(t, string(tt.want), string(o.ortb.Imp[0].Ext), "Extension is not formed properly")
// 		})
// 	}
// }

// func TestORTBRegsGpp(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		o        OpenRTB
// 		wantErr  bool
// 		wantRegs *openrtb2.Regs
// 	}{
// 		{
// 			name: "regs.gpp have value, populate in regs.gpp",
// 			o: OpenRTB{
// 				values: URLValues{
// 					Values: url.Values{
// 						ORTBRegsGpp: []string{"GPP-TEST"},
// 					},
// 				},
// 				ortb: &openrtb2.BidRequest{
// 					Regs: nil,
// 				},
// 			},
// 			wantErr: false,
// 			wantRegs: &openrtb2.Regs{
// 				GPP: "GPP-TEST",
// 			},
// 		},
// 		{
// 			name: "reg.gpp have invalid value, do not populate regs.gpp",
// 			o: OpenRTB{
// 				values: URLValues{
// 					Values: url.Values{},
// 				},
// 				ortb: &openrtb2.BidRequest{
// 					Regs: nil,
// 				},
// 			},
// 			wantErr:  false,
// 			wantRegs: nil,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.o.ORTBRegsGpp(); (err != nil) != tt.wantErr {
// 				t.Errorf("ORTBRegsGpp() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			assert.Equal(t, tt.wantRegs, tt.o.ortb.Regs, "Regs does not match")
// 		})
// 	}
// }

// func TestORTBRegsGppSid(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		o        OpenRTB
// 		wantErr  bool
// 		wantRegs openrtb2.Regs
// 	}{
// 		{
// 			name: "reg.gpp_sid have value, populate in regs.gpp",
// 			o: OpenRTB{
// 				values: URLValues{
// 					Values: url.Values{
// 						ORTBRegsGppSid: []string{"3,1"},
// 					},
// 				},
// 				ortb: &openrtb2.BidRequest{
// 					Regs: nil,
// 				},
// 			},
// 			wantErr: false,
// 			wantRegs: openrtb2.Regs{
// 				GPPSID: []int8{3, 1},
// 			},
// 		},
// 		{
// 			name: "reg.gpp_sid have invalid value, do not populate regs.gpp",
// 			o: OpenRTB{
// 				values: URLValues{
// 					Values: url.Values{
// 						ORTBRegsGppSid: []string{"Error"},
// 					},
// 				},
// 				ortb: &openrtb2.BidRequest{
// 					Regs: nil,
// 				},
// 			},
// 			wantErr:  true,
// 			wantRegs: openrtb2.Regs{},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.o.ORTBRegsGppSid(); (err != nil) != tt.wantErr {
// 				t.Errorf("ORTBRegsGppSid() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			if !reflect.DeepEqual(tt.o.ortb.Regs.GPPSID, tt.wantRegs.GPPSID) {
// 				t.Errorf("ORTBRegsGppSid() error = %v, wantErr %v", tt.o.ortb.Regs.GPPSID, tt.wantRegs.GPPSID)
// 			}
// 		})
// 	}
// }

// func TestORTBUserData(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		o       OpenRTB
// 		wantErr bool
// 	}{
// 		{
// 			name: "ORTBUserData is nil",
// 			o: OpenRTB{
// 				values: URLValues{
// 					Values: url.Values{
// 						ORTBDeviceExtSessionID: []string{"anything"},
// 					},
// 				},
// 				ortb: &openrtb2.BidRequest{},
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.o.ORTBUserData(); (err != nil) != tt.wantErr {
// 				t.Errorf("ORTBUserData() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestORTBDeviceExtSessionID(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		o       OpenRTB
// 		wantErr bool
// 	}{
// 		{
// 			name: "ORTBDeviceExtSessionID with nil values",
// 			o: OpenRTB{
// 				values: URLValues{
// 					Values: url.Values{
// 						ORTBDeviceExtSessionID: []string{"anything"},
// 					},
// 				},
// 				ortb: &openrtb2.BidRequest{},
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.o.ORTBDeviceExtSessionID(); (err != nil) != tt.wantErr {
// 				t.Errorf("ORTBDeviceExtSessionID() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestORTBUserExtEIDS(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		o       OpenRTB
// 		wantErr bool
// 	}{
// 		{
// 			name: "ORTBUserExtEIDS with nil values",
// 			o: OpenRTB{
// 				values: URLValues{
// 					Values: url.Values{
// 						ORTBDeviceExtSessionID: []string{"anything"},
// 					},
// 				},
// 				ortb: &openrtb2.BidRequest{},
// 			},
// 			wantErr: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Setup the OpenRTB object
// 			if err := tt.o.ORTBUserExtEIDS(); (err != nil) != tt.wantErr {
// 				t.Errorf("ORTBUserExtEIDS() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestORTBDeviceExtIfaType(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		o       OpenRTB
// 		wantErr bool
// 	}{
// 		{
// 			name: "ORTBDeviceExtIfaType with nil values",
// 			o: OpenRTB{
// 				values: URLValues{
// 					Values: url.Values{
// 						ORTBDeviceExtIfaType: []string{"anything"},
// 					},
// 				},
// 				ortb: &openrtb2.BidRequest{},
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := tt.o.ORTBDeviceExtIfaType(); (err != nil) != tt.wantErr {
// 				t.Errorf("ORTBDeviceExtIfaType() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestOpenRTB_ORTBUserExtSessionDuration(t *testing.T) {
// 	type fields struct {
// 		values URLValues
// 	}
// 	tests := []struct {
// 		name       string
// 		fields     fields
// 		user       *openrtb2.User
// 		wantResult *openrtb2.User
// 		wantErr    error
// 	}{
// 		{
// 			name: "Nil_User_and_Ext",
// 			fields: fields{
// 				values: URLValues{
// 					Values: url.Values{
// 						ORTBUserExtSessionDuration: []string{"3600"},
// 					},
// 				},
// 			},
// 			wantResult: &openrtb2.User{
// 				Ext: json.RawMessage(`{"sessionduration":3600}`),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name: "Valid_sessionduration",
// 			fields: fields{
// 				values: URLValues{
// 					Values: url.Values{
// 						ORTBUserExtSessionDuration: []string{"3600"},
// 					},
// 				},
// 			},
// 			user: &openrtb2.User{},
// 			wantResult: &openrtb2.User{
// 				Ext: json.RawMessage(`{"sessionduration":3600}`),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name: "Zero_sessionduration",
// 			fields: fields{
// 				values: URLValues{
// 					Values: url.Values{
// 						ORTBUserExtSessionDuration: {"0"},
// 					},
// 				},
// 			},
// 			user: &openrtb2.User{},
// 			wantResult: &openrtb2.User{
// 				Ext: json.RawMessage(`{"sessionduration":0}`),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name: "Negative_sessionduration",
// 			fields: fields{
// 				values: URLValues{
// 					Values: url.Values{
// 						ORTBUserExtSessionDuration: {"-10"},
// 					},
// 				},
// 			},
// 			user:       &openrtb2.User{},
// 			wantResult: &openrtb2.User{Ext: nil},
// 			wantErr:    nil,
// 		},
// 		{
// 			name: "Empty_sessionduration",
// 			fields: fields{
// 				values: URLValues{
// 					Values: url.Values{
// 						ORTBUserExtSessionDuration: {""},
// 					},
// 				},
// 			},
// 			user: &openrtb2.User{},
// 			wantResult: &openrtb2.User{
// 				Ext: nil,
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name: "Missing_sessionduration",
// 			fields: fields{
// 				values: URLValues{
// 					Values: url.Values{},
// 				},
// 			},
// 			user: &openrtb2.User{},
// 			wantResult: &openrtb2.User{
// 				Ext: nil,
// 			},
// 			wantErr: nil,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			o := &OpenRTB{
// 				values: tt.fields.values,
// 				ortb:   &openrtb2.BidRequest{ID: "request-ID", User: tt.user},
// 			}
// 			err := o.ORTBUserExtSessionDuration()
// 			assert.Equal(t, tt.wantErr, err)
// 			assert.Equal(t, tt.wantResult, o.ortb.User)
// 		})
// 	}
// }

// func TestOpenRTB_ORTBUserExtImpDepth(t *testing.T) {
// 	type fields struct {
// 		values URLValues
// 	}
// 	tests := []struct {
// 		name       string
// 		fields     fields
// 		user       *openrtb2.User
// 		wantResult *openrtb2.User
// 		wantErr    error
// 	}{
// 		{
// 			name: "Nil_User_and_Ext",
// 			fields: fields{
// 				values: URLValues{
// 					Values: url.Values{
// 						ORTBUserExtImpDepth: []string{"2"},
// 					},
// 				},
// 			},
// 			wantResult: &openrtb2.User{
// 				Ext: json.RawMessage(`{"impdepth":2}`),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name: "Valid_impdepth",
// 			fields: fields{
// 				values: URLValues{
// 					Values: url.Values{
// 						ORTBUserExtImpDepth: []string{"2"},
// 					},
// 				},
// 			},
// 			user: &openrtb2.User{},
// 			wantResult: &openrtb2.User{
// 				Ext: json.RawMessage(`{"impdepth":2}`),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name: "Zero_impdepth",
// 			fields: fields{
// 				values: URLValues{
// 					Values: url.Values{
// 						ORTBUserExtImpDepth: {"0"},
// 					},
// 				},
// 			},
// 			user: &openrtb2.User{},
// 			wantResult: &openrtb2.User{
// 				Ext: json.RawMessage(`{"impdepth":0}`),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name: "Negative_impdepth",
// 			fields: fields{
// 				values: URLValues{
// 					Values: url.Values{
// 						ORTBUserExtImpDepth: {"-10"},
// 					},
// 				},
// 			},
// 			user:       &openrtb2.User{},
// 			wantResult: &openrtb2.User{Ext: nil},
// 			wantErr:    nil,
// 		},
// 		{
// 			name: "Empty_impdepth",
// 			fields: fields{
// 				values: URLValues{
// 					Values: url.Values{
// 						ORTBUserExtImpDepth: {""},
// 					},
// 				},
// 			},
// 			user: &openrtb2.User{},
// 			wantResult: &openrtb2.User{
// 				Ext: nil,
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name: "Missing_impdepth",
// 			fields: fields{
// 				values: URLValues{},
// 			},
// 			user: &openrtb2.User{},
// 			wantResult: &openrtb2.User{
// 				Ext: nil,
// 			},
// 			wantErr: nil,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			o := &OpenRTB{
// 				values: tt.fields.values,
// 				ortb:   &openrtb2.BidRequest{ID: "request-ID", User: tt.user},
// 			}
// 			err := o.ORTBUserExtImpDepth()
// 			assert.Equal(t, tt.wantErr, err)
// 			assert.Equal(t, tt.wantResult, o.ortb.User)
// 		})
// 	}
// }
