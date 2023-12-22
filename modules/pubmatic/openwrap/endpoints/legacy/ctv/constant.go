package ctv

const (
	//ArraySeparator get api array separator
	ArraySeparator = ","
	//Ext get api ext parameter
	Ext = ".ext."
	//ExtLen get api ext parameter length
	ExtLen = len(Ext)
)

const (
	// USD denotes currency USD
	USD = "USD"
)

const (
	BIDDER_KEY = "bidder"
	PrebidKey  = "prebid"
)

const (
	// BidRequest level parameters

	//ORTBBidRequestID get api parameter req.id
	ORTBBidRequestID = "req.id"
	//ORTBBidRequestTest get api parameter req.test
	ORTBBidRequestTest = "req.test"
	//ORTBBidRequestAt get api parameter req.at
	ORTBBidRequestAt = "req.at"
	//ORTBBidRequestTmax get api parameter req.tmax
	ORTBBidRequestTmax = "req.tmax"
	//ORTBBidRequestWseat get api parameter req.wseat
	ORTBBidRequestWseat = "req.wseat"
	//ORTBBidRequestWlang get api parameter req.wlang
	ORTBBidRequestWlang = "req.wlang"
	//ORTBBidRequestBseat get api parameter req.bseat
	ORTBBidRequestBseat = "req.bseat"
	//ORTBBidRequestAllImps get api parameter req.allimps
	ORTBBidRequestAllImps = "req.allimps"
	//ORTBBidRequestCur get api parameter req.cur
	ORTBBidRequestCur = "req.cur"
	//ORTBBidRequestBcat get api parameter req.bcat
	ORTBBidRequestBcat = "req.bcat"
	//ORTBBidRequestBadv get api parameter req.badv
	ORTBBidRequestBadv = "req.badv"
	//ORTBBidRequestBapp get api parameter req.bapp
	ORTBBidRequestBapp = "req.bapp"

	// Source level parameters

	//ORTBSourceFD get api parameter src.fd
	ORTBSourceFD = "src.fd"
	//ORTBSourceTID get api parameter src.tid
	ORTBSourceTID = "src.tid"
	//ORTBSourcePChain get api parameter src.pchain
	ORTBSourcePChain = "src.pchain"
	//ORTBSourcesChain get api parameter src.pchain
	ORTBSourceSChain = "src.schain"

	// Regs level parameters

	//ORTBRegsCoppa get api parameter regs.coppa
	ORTBRegsCoppa = "regs.coppa"
	// ORTB get api parameter for gpp and gpp_sid in regs
	ORTBRegsGpp    = "regs.gpp"
	ORTBRegsGppSid = "regs.gpp_sid"

	// Imp level parameters

	//ORTBImpID get api parameter imp.id
	ORTBImpID = "imp.id"
	//ORTBImpDisplayManager get api parameter imp.displaymanager
	ORTBImpDisplayManager = "imp.displaymanager"
	//ORTBImpDisplayManagerVer get api parameter imp.displaymanagerver
	ORTBImpDisplayManagerVer = "imp.displaymanagerver"
	//ORTBImpInstl get api parameter imp.instl
	ORTBImpInstl = "imp.instl"
	//ORTBImpTagID get api parameter imp.tagid
	ORTBImpTagID = "imp.tagid"
	//ORTBImpBidFloor get api parameter imp.bidfloor
	ORTBImpBidFloor = "imp.bidfloor"
	//ORTBImpBidFloorCur get api parameter imp.bidfloorcur
	ORTBImpBidFloorCur = "imp.bidfloorcur"
	//ORTBImpClickBrowser get api parameter imp.clickbrowser
	ORTBImpClickBrowser = "imp.clickbrowser"
	//ORTBImpSecure get api parameter imp.secure
	ORTBImpSecure = "imp.secure"
	//ORTBImpIframeBuster get api parameter imp.iframebuster
	ORTBImpIframeBuster = "imp.iframebuster"
	//ORTBImpExp get api parameter imp.exp
	ORTBImpExp = "imp.exp"
	//ORTBImpPmp get api parameter imp.pmp
	ORTBImpPmp = "imp.pmp"
	//ORTBImpExtBidder get api parameter imp.ext
	ORTBImpExtBidder = "imp.ext.bidder"
	//ORTBImpExtPrebid get api parameter imp.ext.prebid
	ORTBImpExtPrebid = "imp.ext.prebid"

	// Metric level parameters

	//ORTBMetricType   = "type" //// get api parameter  = "ty

	//ORTBMetricValue  = "value" //// get api parameter = "val

	//ORTBMetricVendor = "vendor" //// get api parameter  "vend

	// Banner level parameters

	//ORTBBannerSize     = "banner.bsize" //// get api parameter    = "banner.bsi

	//ORTBBannerWMax     = "banner.wmax" //// get api parameter    = "banner.wm

	//ORTBBannerHMax     = "banner.hmax" //// get api parameter    = "banner.hm

	//ORTBBannerWMin     = "banner.wmin" //// get api parameter    = "banner.wm

	//ORTBBannerHMin     = "banner.hmin" //// get api parameter    = "banner.hm

	//ORTBBannerBType    = "banner.btype" //// get api parameter   = "banner.bty

	//ORTBBannerBAttr    = "banner.battr" //// get api parameter   = "banner.bat

	//ORTBBannerPos      = "banner.pos" //// get api parameter     = "banner.p

	//ORTBBannerMimes    = "banner.mimes" //// get api parameter   = "banner.mim

	//ORTBBannerTopFrame = "banner.topframe" //// get api parameter  "banner.topfra

	//ORTBBannerExpdir   = "banner.expdir" //// get api parameter  = "banner.expd

	//ORTBBannerAPI      = "banner.api" //// get api parameter     = "banner.a

	//ORTBBannerID       = "banner.id" //// get api parameter      = "banner.

	//ORTBBannerVcm      = "banner.vcm" //// get api parameter     = "banner.v

	// Native level parameters

	//ORTBNativeRequest = "native.request" //// get api parameter  "native.reque

	//ORTBNativeVer     = "native.ver" //// get api parameter    = "native.v

	//ORTBNativeAPI     = "native.api" //// get api parameter    = "native.a

	//ORTBNativeBAttr   = "native.battr" //// get api parameter  = "native.bat

	// Video level parameters

	//ORTBImpVideoMimes get api parameter imp.vid.mimes
	ORTBImpVideoMimes = "imp.vid.mimes"
	//ORTBImpVideoMinDuration get api parameter imp.vid.minduration
	ORTBImpVideoMinDuration = "imp.vid.minduration"
	//ORTBImpVideoMaxDuration get api parameter imp.vid.maxduration
	ORTBImpVideoMaxDuration = "imp.vid.maxduration"
	//ORTBImpVideoProtocols get api parameter imp.vid.protocols
	ORTBImpVideoProtocols = "imp.vid.protocols"
	//ORTBImpVideoPlayerWidth get api parameter imp.vid.w
	ORTBImpVideoPlayerWidth = "imp.vid.w"
	//ORTBImpVideoPlayerHeight get api parameter imp.vid.h
	ORTBImpVideoPlayerHeight = "imp.vid.h"
	//ORTBImpVideoStartDelay get api parameter imp.vid.startdelay
	ORTBImpVideoStartDelay = "imp.vid.startdelay"
	//ORTBImpVideoPlacement get api parameter imp.vid.placement
	ORTBImpVideoPlacement = "imp.vid.placement"
	// ORTBImpVideoPlacement get api parameter imp.vid.plcmt
	ORTBImpVideoPlcmt = "imp.vid.plcmt"
	//ORTBImpVideoLinearity get api parameter imp.vid.linearity
	ORTBImpVideoLinearity = "imp.vid.linearity"
	//ORTBImpVideoSkip get api parameter imp.vid.skip
	ORTBImpVideoSkip = "imp.vid.skip"
	//ORTBImpVideoSkipMin get api parameter imp.vid.skipmin
	ORTBImpVideoSkipMin = "imp.vid.skipmin"
	//ORTBImpVideoSkipAfter get api parameter imp.vid.skipafter
	ORTBImpVideoSkipAfter = "imp.vid.skipafter"
	//ORTBImpVideoSequence get api parameter imp.vid.sequence
	ORTBImpVideoSequence = "imp.vid.sequence"
	//ORTBImpVideoBAttr get api parameter imp.vid.battr
	ORTBImpVideoBAttr = "imp.vid.battr"
	//ORTBImpVideoMaxExtended get api parameter imp.vid.maxextended
	ORTBImpVideoMaxExtended = "imp.vid.maxextended"
	//ORTBImpVideoMinBitrate get api parameter imp.vid.minbitrate
	ORTBImpVideoMinBitrate = "imp.vid.minbitrate"
	//ORTBImpVideoMaxBitrate get api parameter imp.vid.maxbitrate
	ORTBImpVideoMaxBitrate = "imp.vid.maxbitrate"
	//ORTBImpVideoBoxingAllowed get api parameter imp.vid.boxingallowed
	ORTBImpVideoBoxingAllowed = "imp.vid.boxingallowed"
	//ORTBImpVideoPlaybackMethod get api parameter imp.vid.playbackmethod
	ORTBImpVideoPlaybackMethod = "imp.vid.playbackmethod"
	//ORTBImpVideoDelivery get api parameter imp.vid.delivery
	ORTBImpVideoDelivery = "imp.vid.delivery"
	//ORTBImpVideoPos get api parameter imp.vid.pos
	ORTBImpVideoPos = "imp.vid.pos"
	//ORTBImpVideoAPI get api parameter imp.vid.api
	ORTBImpVideoAPI = "imp.vid.api"
	//ORTBImpVideoCompanionType get api parameter imp.vid.companiontype
	ORTBImpVideoCompanionType = "imp.vid.companiontype"

	// Site level parameters

	//ORTBSiteID get api parameter site.id
	ORTBSiteID = "site.id"
	//ORTBSiteName get api parameter site.name
	ORTBSiteName = "site.name"
	//ORTBSiteDomain get api parameter site.domain
	ORTBSiteDomain = "site.domain"
	//ORTBSitePage get api parameter site.page
	ORTBSitePage = "site.page"
	//ORTBSiteRef get api parameter site.ref
	ORTBSiteRef = "site.ref"
	//ORTBSiteSearch get api parameter site.search
	ORTBSiteSearch = "site.search"
	//ORTBSiteMobile get api parameter site.mobile
	ORTBSiteMobile = "site.mobile"
	//ORTBSiteCat get api parameter site.cat
	ORTBSiteCat = "site.cat"
	//ORTBSiteSectionCat get api parameter site.sectioncat
	ORTBSiteSectionCat = "site.sectioncat"
	//ORTBSitePageCat get api parameter site.pagecat
	ORTBSitePageCat = "site.pagecat"
	//ORTBSitePrivacyPolicy get api parameter site.privacypolicy
	ORTBSitePrivacyPolicy = "site.privacypolicy"
	//ORTBSiteKeywords get api parameter site.keywords
	ORTBSiteKeywords = "site.keywords"

	// App level parameters

	//ORTBAppID get api parameter app.id
	ORTBAppID = "app.id"
	//ORTBAppName get api parameter app.name
	ORTBAppName = "app.name"
	//ORTBAppBundle get api parameter app.bundle
	ORTBAppBundle = "app.bundle"
	//ORTBAppDomain get api parameter app.domain
	ORTBAppDomain = "app.domain"
	//ORTBAppStoreURL get api parameter app.storeurl
	ORTBAppStoreURL = "app.storeurl"
	//ORTBAppVer get api parameter app.ver
	ORTBAppVer = "app.ver"
	//ORTBAppPaid get api parameter app.paid
	ORTBAppPaid = "app.paid"
	//ORTBAppCat get api parameter app.cat
	ORTBAppCat = "app.cat"
	//ORTBAppSectionCat get api parameter app.sectioncat
	ORTBAppSectionCat = "app.sectioncat"
	//ORTBAppPageCat get api parameter app.pagecat
	ORTBAppPageCat = "app.pagecat"
	//ORTBAppPrivacyPolicy get api parameter app.privacypolicy
	ORTBAppPrivacyPolicy = "app.privacypolicy"
	//ORTBAppKeywords get api parameter app.keywords
	ORTBAppKeywords = "app.keywords"

	// Site.Publisher level parameters

	//ORTBSitePublisherID get api parameter site.pub.id
	ORTBSitePublisherID = "site.pub.id"
	//ORTBSitePublisherName get api parameter site.pub.name
	ORTBSitePublisherName = "site.pub.name"
	//ORTBSitePublisherCat get api parameter site.pub.cat
	ORTBSitePublisherCat = "site.pub.cat"
	//ORTBSitePublisherDomain get api parameter site.pub.domain
	ORTBSitePublisherDomain = "site.pub.domain"

	// Site.Content level parameters

	//ORTBSiteContentID get api parameter site.cnt.id
	ORTBSiteContentID = "site.cnt.id"
	//ORTBSiteContentEpisode get api parameter site.cnt.episode
	ORTBSiteContentEpisode = "site.cnt.episode"
	//ORTBSiteContentTitle get api parameter site.cnt.title
	ORTBSiteContentTitle = "site.cnt.title"
	//ORTBSiteContentSeries get api parameter site.cnt.series
	ORTBSiteContentSeries = "site.cnt.series"
	//ORTBSiteContentSeason get api parameter site.cnt.season
	ORTBSiteContentSeason = "site.cnt.season"
	//ORTBSiteContentArtist get api parameter site.cnt.artist
	ORTBSiteContentArtist = "site.cnt.artist"
	//ORTBSiteContentGenre get api parameter site.cnt.genre
	ORTBSiteContentGenre = "site.cnt.genre"
	//ORTBSiteContentAlbum get api parameter site.cnt.album
	ORTBSiteContentAlbum = "site.cnt.album"
	//ORTBSiteContentIsRc get api parameter site.cnt.isrc
	ORTBSiteContentIsRc = "site.cnt.isrc"
	//ORTBSiteContentURL get api parameter site.cnt.url
	ORTBSiteContentURL = "site.cnt.url"
	//ORTBSiteContentCat get api parameter site.cnt.cat
	ORTBSiteContentCat = "site.cnt.cat"
	//ORTBSiteContentProdQ get api parameter site.cnt.prodq
	ORTBSiteContentProdQ = "site.cnt.prodq"
	//ORTBSiteContentVideoQuality get api parameter site.cnt.videoquality
	ORTBSiteContentVideoQuality = "site.cnt.videoquality"
	//ORTBSiteContentContext get api parameter site.cnt.context
	ORTBSiteContentContext = "site.cnt.context"
	//ORTBSiteContentContentRating get api parameter site.cnt.contentrating
	ORTBSiteContentContentRating = "site.cnt.contentrating"
	//ORTBSiteContentUserRating get api parameter site.cnt.userrating
	ORTBSiteContentUserRating = "site.cnt.userrating"
	//ORTBSiteContentQaGmeDiarating get api parameter site.cnt.qagmediarating
	ORTBSiteContentQaGmeDiarating = "site.cnt.qagmediarating"
	//ORTBSiteContentKeywords get api parameter site.cnt.keywords
	ORTBSiteContentKeywords = "site.cnt.keywords"
	//ORTBSiteContentLiveStream get api parameter site.cnt.livestream
	ORTBSiteContentLiveStream = "site.cnt.livestream"
	//ORTBSiteContentSourceRelationship get api parameter site.cnt.sourcerelationship
	ORTBSiteContentSourceRelationship = "site.cnt.sourcerelationship"
	//ORTBSiteContentLen get api parameter site.cnt.len
	ORTBSiteContentLen = "site.cnt.len"
	//ORTBSiteContentLanguage get api parameter site.cnt.language
	ORTBSiteContentLanguage = "site.cnt.language"
	//ORTBSiteContentEmbeddable get api parameter site.cnt.embeddable
	ORTBSiteContentEmbeddable = "site.cnt.embeddable"

	// Site.Producer level parameters

	//ORTBSiteContentProducerID get api parameter site.cnt.prod.id
	ORTBSiteContentProducerID = "site.cnt.prod.id"
	//ORTBSiteContentProducerName get api parameter site.cnt.prod.name
	ORTBSiteContentProducerName = "site.cnt.prod.name"
	//ORTBSiteContentProducerCat get api parameter site.cnt.prod.cat
	ORTBSiteContentProducerCat = "site.cnt.prod.cat"
	//ORTBSiteContentProducerDomain get api parameter site.cnt.prod.domain
	ORTBSiteContentProducerDomain = "site.cnt.prod.domain"

	// App.Publisher level parameters

	//ORTBAppPublisherID get api parameter app.pub.id
	ORTBAppPublisherID = "app.pub.id"
	//ORTBAppPublisherName get api parameter app.pub.name
	ORTBAppPublisherName = "app.pub.name"
	//ORTBAppPublisherCat get api parameter app.pub.cat
	ORTBAppPublisherCat = "app.pub.cat"
	//ORTBAppPublisherDomain get api parameter app.pub.domain
	ORTBAppPublisherDomain = "app.pub.domain"

	// App.Content level parameters

	//ORTBAppContentID get api parameter app.cnt.id
	ORTBAppContentID = "app.cnt.id"
	//ORTBAppContentEpisode get api parameter app.cnt.episode
	ORTBAppContentEpisode = "app.cnt.episode"
	//ORTBAppContentTitle get api parameter app.cnt.title
	ORTBAppContentTitle = "app.cnt.title"
	//ORTBAppContentSeries get api parameter app.cnt.series
	ORTBAppContentSeries = "app.cnt.series"
	//ORTBAppContentSeason get api parameter app.cnt.season
	ORTBAppContentSeason = "app.cnt.season"
	//ORTBAppContentArtist get api parameter app.cnt.artist
	ORTBAppContentArtist = "app.cnt.artist"
	//ORTBAppContentGenre get api parameter app.cnt.genre
	ORTBAppContentGenre = "app.cnt.genre"
	//ORTBAppContentAlbum get api parameter app.cnt.album
	ORTBAppContentAlbum = "app.cnt.album"
	//ORTBAppContentIsRc get api parameter app.cnt.isrc
	ORTBAppContentIsRc = "app.cnt.isrc"
	//ORTBAppContentURL get api parameter app.cnt.url
	ORTBAppContentURL = "app.cnt.url"
	//ORTBAppContentCat get api parameter app.cnt.cat
	ORTBAppContentCat = "app.cnt.cat"
	//ORTBAppContentProdQ get api parameter app.cnt.prodq
	ORTBAppContentProdQ = "app.cnt.prodq"
	//ORTBAppContentVideoQuality get api parameter app.cnt.videoquality
	ORTBAppContentVideoQuality = "app.cnt.videoquality"
	//ORTBAppContentContext get api parameter app.cnt.context
	ORTBAppContentContext = "app.cnt.context"
	//ORTBAppContentContentRating get api parameter app.cnt.contentrating
	ORTBAppContentContentRating = "app.cnt.contentrating"
	//ORTBAppContentUserRating get api parameter app.cnt.userrating
	ORTBAppContentUserRating = "app.cnt.userrating"
	//ORTBAppContentQaGmeDiarating get api parameter app.cnt.qagmediarating
	ORTBAppContentQaGmeDiarating = "app.cnt.qagmediarating"
	//ORTBAppContentKeywords get api parameter app.cnt.keywords
	ORTBAppContentKeywords = "app.cnt.keywords"
	//ORTBAppContentLiveStream get api parameter app.cnt.livestream
	ORTBAppContentLiveStream = "app.cnt.livestream"
	//ORTBAppContentSourceRelationship get api parameter app.cnt.sourcerelationship
	ORTBAppContentSourceRelationship = "app.cnt.sourcerelationship"
	//ORTBAppContentLen get api parameter app.cnt.len
	ORTBAppContentLen = "app.cnt.len"
	//ORTBAppContentLanguage get api parameter app.cnt.language
	ORTBAppContentLanguage = "app.cnt.language"
	//ORTBAppContentEmbeddable get api parameter app.cnt.embeddable
	ORTBAppContentEmbeddable = "app.cnt.embeddable"

	// App.Producer level parameters

	//ORTBAppContentProducerID get api parameter app.cnt.prod.id
	ORTBAppContentProducerID = "app.cnt.prod.id"
	//ORTBAppContentProducerName get api parameter app.cnt.prod.name
	ORTBAppContentProducerName = "app.cnt.prod.name"
	//ORTBAppContentProducerCat get api parameter app.cnt.prod.cat
	ORTBAppContentProducerCat = "app.cnt.prod.cat"
	//ORTBAppContentProducerDomain get api parameter app.cnt.prod.domain
	ORTBAppContentProducerDomain = "app.cnt.prod.domain"

	// Device level parameters

	//ORTBDeviceUserAgent get api parameter dev.ua
	ORTBDeviceUserAgent = "dev.ua"
	//ORTBDeviceDnt get api parameter dev.dnt
	ORTBDeviceDnt = "dev.dnt"
	//ORTBDeviceLmt get api parameter dev.lmt
	ORTBDeviceLmt = "dev.lmt"
	//ORTBDeviceIP get api parameter dev.ip
	ORTBDeviceIP = "dev.ip"
	//ORTBDeviceIpv6 get api parameter dev.ipv6
	ORTBDeviceIpv6 = "dev.ipv6"
	//ORTBDeviceDeviceType get api parameter dev.devicetype
	ORTBDeviceDeviceType = "dev.devicetype"
	//ORTBDeviceMake get api parameter dev.make
	ORTBDeviceMake = "dev.make"
	//ORTBDeviceModel get api parameter dev.model
	ORTBDeviceModel = "dev.model"
	//ORTBDeviceOs get api parameter dev.os
	ORTBDeviceOs = "dev.os"
	//ORTBDeviceOsv get api parameter dev.osv
	ORTBDeviceOsv = "dev.osv"
	//ORTBDeviceHwv get api parameter dev.hwv
	ORTBDeviceHwv = "dev.hwv"
	//ORTBDeviceWidth get api parameter dev.w
	ORTBDeviceWidth = "dev.w"
	//ORTBDeviceHeight get api parameter dev.h
	ORTBDeviceHeight = "dev.h"
	//ORTBDevicePpi get api parameter dev.ppi
	ORTBDevicePpi = "dev.ppi"
	//ORTBDevicePxRatio get api parameter dev.pxratio
	ORTBDevicePxRatio = "dev.pxratio"
	//ORTBDeviceJS get api parameter dev.js
	ORTBDeviceJS = "dev.js"
	//ORTBDeviceGeoFetch get api parameter dev.geofetch
	ORTBDeviceGeoFetch = "dev.geofetch"
	//ORTBDeviceFlashVer get api parameter dev.flashver
	ORTBDeviceFlashVer = "dev.flashver"
	//ORTBDeviceLanguage get api parameter dev.language
	ORTBDeviceLanguage = "dev.language"
	//ORTBDeviceCarrier get api parameter dev.carrier
	ORTBDeviceCarrier = "dev.carrier"
	//ORTBDeviceMccmnc get api parameter dev.mccmnc
	ORTBDeviceMccmnc = "dev.mccmnc"
	//ORTBDeviceConnectionType get api parameter dev.connectiontype
	ORTBDeviceConnectionType = "dev.connectiontype"
	//ORTBDeviceIfa get api parameter dev.ifa
	ORTBDeviceIfa = "dev.ifa"
	//ORTBDeviceDidSha1 get api parameter dev.didsha1
	ORTBDeviceDidSha1 = "dev.didsha1"
	//ORTBDeviceDidMd5 get api parameter dev.didmd5
	ORTBDeviceDidMd5 = "dev.didmd5"
	//ORTBDeviceDpidSha1 get api parameter dev.dpidsha1
	ORTBDeviceDpidSha1 = "dev.dpidsha1"
	//ORTBDeviceDpidMd5 get api parameter dev.dpidmd5
	ORTBDeviceDpidMd5 = "dev.dpidmd5"
	//ORTBDeviceMacSha1 get api parameter dev.macsha1
	ORTBDeviceMacSha1 = "dev.macsha1"
	//ORTBDeviceMacMd5 get api parameter dev.macmd5
	ORTBDeviceMacMd5 = "dev.macmd5"

	// Device.Geo level parameters

	//ORTBDeviceGeoLat get api parameter dev.geo.lat
	ORTBDeviceGeoLat = "dev.geo.lat"
	//ORTBDeviceGeoLon get api parameter dev.geo.lon
	ORTBDeviceGeoLon = "dev.geo.lon"
	//ORTBDeviceGeoType get api parameter dev.geo.type
	ORTBDeviceGeoType = "dev.geo.type"
	//ORTBDeviceGeoAccuracy get api parameter dev.geo.accuracy
	ORTBDeviceGeoAccuracy = "dev.geo.accuracy"
	//ORTBDeviceGeoLastFix get api parameter dev.geo.lastfix
	ORTBDeviceGeoLastFix = "dev.geo.lastfix"
	//ORTBDeviceGeoIPService get api parameter dev.geo.ipservice
	ORTBDeviceGeoIPService = "dev.geo.ipservice"
	//ORTBDeviceGeoCountry get api parameter dev.geo.country
	ORTBDeviceGeoCountry = "dev.geo.country"
	//ORTBDeviceGeoRegion get api parameter dev.geo.region
	ORTBDeviceGeoRegion = "dev.geo.region"
	//ORTBDeviceGeoRegionFips104 get api parameter dev.geo.regionfips104
	ORTBDeviceGeoRegionFips104 = "dev.geo.regionfips104"
	//ORTBDeviceGeoMetro get api parameter dev.geo.metro
	ORTBDeviceGeoMetro = "dev.geo.metro"
	//ORTBDeviceGeoCity get api parameter dev.geo.city
	ORTBDeviceGeoCity = "dev.geo.city"
	//ORTBDeviceGeoZip get api parameter dev.geo.zip
	ORTBDeviceGeoZip = "dev.geo.zip"
	//ORTBDeviceGeoUtcOffset get api parameter dev.geo.utcoffset
	ORTBDeviceGeoUtcOffset = "dev.geo.utcoffset"

	// User level parameters

	//ORTBUserID get api parameter user.id
	ORTBUserID = "user.id"
	//ORTBUserBuyerUID get api parameter user.buyeruid
	ORTBUserBuyerUID = "user.buyeruid"
	//ORTBUserYob get api parameter user.yob
	ORTBUserYob = "user.yob"
	//ORTBUserGender get api parameter user.gender
	ORTBUserGender = "user.gender"
	//ORTBUserKeywords get api parameter user.keywords
	ORTBUserKeywords = "user.keywords"
	//ORTBUserCustomData get api parameter user.customdata
	ORTBUserCustomData = "user.customdata"

	// User.Geo level parameters

	//ORTBUserGeoLat get api parameter user.geo.lat
	ORTBUserGeoLat = "user.geo.lat"
	//ORTBUserGeoLon get api parameter user.geo.lon
	ORTBUserGeoLon = "user.geo.lon"
	//ORTBUserGeoType get api parameter user.geo.type
	ORTBUserGeoType = "user.geo.type"
	//ORTBUserGeoAccuracy get api parameter user.geo.accuracy
	ORTBUserGeoAccuracy = "user.geo.accuracy"
	//ORTBUserGeoLastFix get api parameter user.geo.lastfix
	ORTBUserGeoLastFix = "user.geo.lastfix"
	//ORTBUserGeoIPService get api parameter user.geo.ipservice
	ORTBUserGeoIPService = "user.geo.ipservice"
	//ORTBUserGeoCountry get api parameter user.geo.country
	ORTBUserGeoCountry = "user.geo.country"
	//ORTBUserGeoRegion get api parameter user.geo.region
	ORTBUserGeoRegion = "user.geo.region"
	//ORTBUserGeoRegionFips104 get api parameter user.geo.regionfips104
	ORTBUserGeoRegionFips104 = "user.geo.regionfips104"
	//ORTBUserGeoMetro get api parameter user.geo.metro
	ORTBUserGeoMetro = "user.geo.metro"
	//ORTBUserGeoCity get api parameter user.geo.city
	ORTBUserGeoCity = "user.geo.city"
	//ORTBUserGeoZip get api parameter user.geo.zip
	ORTBUserGeoZip = "user.geo.zip"
	//ORTBUserGeoUtcOffset get api parameter user.geo.utcoffset
	ORTBUserGeoUtcOffset = "user.geo.utcoffset"

	// User.Ext.Consent level parameters

	//ORTBUserExtConsent get api parameter user.ext.consent
	ORTBUserExtConsent = "user.ext.consent"

	//ORTBUserExtEIDS get api parameter user.ext.eids
	ORTBUserExtEIDS = "user.ext.eids"

	//ORTBUserData get api parameter user.data
	ORTBUserData = "user.data"

	// Regs.Ext.GDPR level parameters

	//ORTBRegsExtGdpr get api parameter regs.ext.gdpr
	ORTBRegsExtGdpr = "regs.ext.gdpr"
	//ORTBRegsExtUSPrivacy get api parameter regs.ext.us_privacy
	ORTBRegsExtUSPrivacy = "regs.ext.us_privacy"

	// VideoExtension level parameters

	//ORTBImpVideoExtOffset get api parameter imp.vid.ext.offset
	ORTBImpVideoExtOffset = "imp.vid.ext.offset"
	//ORTBImpVideoExtAdPodMinAds get api parameter imp.vid.ext.adpod.minads
	ORTBImpVideoExtAdPodMinAds = "imp.vid.ext.adpod.minads"
	//ORTBImpVideoExtAdPodMaxAds get api parameter imp.vid.ext.adpod.maxads
	ORTBImpVideoExtAdPodMaxAds = "imp.vid.ext.adpod.maxads"
	//ORTBImpVideoExtAdPodMinDuration get api parameter imp.vid.ext.adpod.adminduration
	ORTBImpVideoExtAdPodMinDuration = "imp.vid.ext.adpod.adminduration"
	//ORTBImpVideoExtAdPodMaxDuration get api parameter imp.vid.ext.adpod.admaxduration
	ORTBImpVideoExtAdPodMaxDuration = "imp.vid.ext.adpod.admaxduration"
	//ORTBImpVideoExtAdPodAdvertiserExclusionPercent get api parameter imp.vid.ext.adpod.excladv
	ORTBImpVideoExtAdPodAdvertiserExclusionPercent = "imp.vid.ext.adpod.excladv"
	//ORTBImpVideoExtAdPodIABCategoryExclusionPercent get api parameter imp.vid.ext.adpod.excliabcat
	ORTBImpVideoExtAdPodIABCategoryExclusionPercent = "imp.vid.ext.adpod.excliabcat"

	// ReqWrapperExtension level parameters

	//ORTBProfileID get api parameter req.ext.wrapper.profileid
	ORTBProfileID = "req.ext.wrapper.profileid"
	//ORTBVersionID get api parameter req.ext.wrapper.versionid
	ORTBVersionID = "req.ext.wrapper.versionid"
	//ORTBSSAuctionFlag get api parameter req.ext.wrapper.ssauction
	ORTBSSAuctionFlag = "req.ext.wrapper.ssauction"
	//ORTBSumryDisableFlag get api parameter req.ext.wrapper.sumry_disable
	ORTBSumryDisableFlag = "req.ext.wrapper.sumry_disable"
	//ORTBClientConfigFlag get api parameter req.ext.wrapper.clientconfig
	ORTBClientConfigFlag = "req.ext.wrapper.clientconfig"
	//ORTBSupportDeals get api parameter req.ext.wrapper.supportdeals
	ORTBSupportDeals = "req.ext.wrapper.supportdeals"
	//ORTBIncludeBrandCategory get api parameter req.ext.wrapper.includebrandcategory
	ORTBIncludeBrandCategory = "req.ext.wrapper.includebrandcategory"

	// ReqAdPodExt level parameters

	//ORTBRequestExtAdPodMinAds get api parameter req.ext.adpod.minads
	ORTBRequestExtAdPodMinAds = "req.ext.adpod.minads"
	//ORTBRequestExtAdPodMaxAds get api parameter req.ext.adpod.maxads
	ORTBRequestExtAdPodMaxAds = "req.ext.adpod.maxads"
	//ORTBRequestExtAdPodMinDuration get api parameter req.ext.adpod.adminduration
	ORTBRequestExtAdPodMinDuration = "req.ext.adpod.adminduration"
	//ORTBRequestExtAdPodMaxDuration get api parameter req.ext.adpod.admaxduration
	ORTBRequestExtAdPodMaxDuration = "req.ext.adpod.admaxduration"
	//ORTBRequestExtAdPodAdvertiserExclusionPercent get api parameter req.ext.adpod.excladv
	ORTBRequestExtAdPodAdvertiserExclusionPercent = "req.ext.adpod.excladv"
	//ORTBRequestExtAdPodIABCategoryExclusionPercent get api parameter req.ext.adpod.excliabcat
	ORTBRequestExtAdPodIABCategoryExclusionPercent = "req.ext.adpod.excliabcat"
	//ORTBRequestExtAdPodCrossPodAdvertiserExclusionPercent get api parameter req.ext.adpod.crosspodexcladv
	ORTBRequestExtAdPodCrossPodAdvertiserExclusionPercent = "req.ext.adpod.crosspodexcladv"
	//ORTBRequestExtAdPodCrossPodIABCategoryExclusionPercent get api parameter req.ext.adpod.crosspodexcliabcat
	ORTBRequestExtAdPodCrossPodIABCategoryExclusionPercent = "req.ext.adpod.crosspodexcliabcat"
	//ORTBRequestExtAdPodIABCategoryExclusionWindow get api parameter req.ext.adpod.excliabcatwindow
	ORTBRequestExtAdPodIABCategoryExclusionWindow = "req.ext.adpod.excliabcatwindow"
	//ORTBRequestExtAdPodAdvertiserExclusionWindow get api parameter req.ext.adpod.excladvwindow
	ORTBRequestExtAdPodAdvertiserExclusionWindow = "req.ext.adpod.excladvwindow"

	// ORTB Extension Objects */ //// get api parameter xtension Objelevel parameters

	//ORTBBidRequestExt get api parameter req.ext
	ORTBBidRequestExt = "req.ext"
	//ORTBSourceExt get api parameter src.ext
	ORTBSourceExt = "src.ext"
	//ORTBRegsExt get api parameter regs.ext
	ORTBRegsExt = "regs.ext"
	//ORTBImpExt get api parameter imp.ext
	ORTBImpExt = "imp.ext"
	//ORTBImpVideoExt get api parameter imp.vid.ext
	ORTBImpVideoExt = "imp.vid.ext"
	//ORTBSiteExt get api parameter site.ext
	ORTBSiteExt = "site.ext"
	//ORTBAppExt get api parameter app.ext
	ORTBAppExt = "app.ext"
	//ORTBSitePublisherExt get api parameter site.pub.ext
	ORTBSitePublisherExt = "site.pub.ext"
	//ORTBSiteContentExt get api parameter site.cnt.ext
	ORTBSiteContentExt = "site.cnt.ext"
	//ORTBSiteContentProducerExt get api parameter site.cnt.prod.ext
	ORTBSiteContentProducerExt = "site.cnt.prod.ext"
	//ORTBAppPublisherExt get api parameter app.pub.ext
	ORTBAppPublisherExt = "app.pub.ext"
	//ORTBAppContentExt get api parameter app.cnt.ext
	ORTBAppContentExt = "app.cnt.ext"
	//ORTBAppContentProducerExt get api parameter app.cnt.prod.ext
	ORTBAppContentProducerExt = "app.cnt.prod.ext"
	//ORTBDeviceExt get api parameter dev.ext
	ORTBDeviceExt = "dev.ext"
	//ORTBDeviceGeoExt get api parameter dev.geo.ext
	ORTBDeviceGeoExt = "dev.geo.ext"
	//ORTBUserExt get api parameter user.ext
	ORTBUserExt = "user.ext"
	//ORTBUserGeoExt get api parameter user.geo.ext
	ORTBUserGeoExt = "user.geo.ext"

	// ORTB Extension Standard Keys */ //// get api parameter xtension Standard Klevel parameters

	//ORTBExtWrapper get api parameter wrapper
	ORTBExtWrapper = "wrapper"
	//ORTBExtProfileId get api parameter profileid
	ORTBExtProfileId = "profileid"
	ORTBExtSsai      = "ssai"
	//ORTBExtVersionId get api parameter versionid
	ORTBExtVersionId = "versionid"
	//ORTBExtSSAuctionFlag get api parameter ssauction
	ORTBExtSSAuctionFlag = "ssauction"
	//ORTBExtSumryDisableFlag get api parameter sumry_disable
	ORTBExtSumryDisableFlag = "sumry_disable"
	//ORTBExtClientConfigFlag get api parameter clientconfig
	ORTBExtClientConfigFlag = "clientconfig"
	//ORTBExtSupportDeals get api parameter supportdeals
	ORTBExtSupportDeals = "supportdeals"
	//ORTBExtIncludeBrandCategory get api parameter includebrandcategory
	ORTBExtIncludeBrandCategory = "includebrandcategory"
	// ORTBSSAI get the api parameter req.ext.wrapper.ssai
	ORTBSSAI = "req.ext.wrapper.ssai"
	//ORTBExtGDPR get api parameter gdpr
	ORTBExtGDPR = "gdpr"
	//ORTBExtUSPrivacy get api parameter us_privacy
	ORTBExtUSPrivacy = "us_privacy"
	//ORTBExtConsent get api parameter consent
	ORTBExtConsent = "consent"
	//ORTBExtAdPod get api parameter adpod
	ORTBExtAdPod = "adpod"
	//ORTBExtAdPodOffset get api parameter offset
	ORTBExtAdPodOffset = "offset"
	//ORTBExtAdPodMinAds get api parameter minads
	ORTBExtAdPodMinAds = "minads"
	//ORTBExtAdPodMaxAds get api parameter maxads
	ORTBExtAdPodMaxAds = "maxads"
	//ORTBExtAdPodMinDuration get api parameter adminduration
	ORTBExtAdPodMinDuration = "adminduration"
	//ORTBExtAdPodMaxDuration get api parameter admaxduration
	ORTBExtAdPodMaxDuration = "admaxduration"
	//ORTBExtAdPodAdvertiserExclusionPercent get api parameter excladv
	ORTBExtAdPodAdvertiserExclusionPercent = "excladv"
	//ORTBExtAdPodIABCategoryExclusionPercent get api parameter excliabcat
	ORTBExtAdPodIABCategoryExclusionPercent = "excliabcat"
	//ORTBExtAdPodCrossPodAdvertiserExclusionPercent get api parameter crosspodexcladv
	ORTBExtAdPodCrossPodAdvertiserExclusionPercent = "crosspodexcladv"
	//ORTBExtAdPodCrossPodIABCategoryExclusionPercent get api parameter crosspodexcliabcat
	ORTBExtAdPodCrossPodIABCategoryExclusionPercent = "crosspodexcliabcat"
	//ORTBExtAdPodIABCategoryExclusionWindow get api parameter excliabcatwindow
	ORTBExtAdPodIABCategoryExclusionWindow = "excliabcatwindow"
	//ORTBExtAdPodAdvertiserExclusionWindow get api parameter excladvwindow
	ORTBExtAdPodAdvertiserExclusionWindow = "excladvwindow"

	//ORBTExtDeviceIfaType get api parameter
	ORTBExtIfaType       = "ifa_type"
	ORTBDeviceExtIfaType = "dev.ext.ifa_type"

	//ORTBExtSessionID parameter
	ORTBExtSessionID       = "session_id"
	ORTBDeviceExtSessionID = "dev.ext.session_id"

	//ORTBExtEIDS parameter
	ORTBExtEIDS = "eids"

	//ORTBRequestExtPrebidTransparencyContent get api parameter req.ext.prebid.transparency.content
	ORTBRequestExtPrebidTransparencyContent = "req.ext.prebid.transparency.content"
	//ORTBExtPrebid get api parameter prebid
	ORTBExtPrebid = "prebid"
	//ORTBExtPrebidTransparency get api parameter transparency
	ORTBExtPrebidTransparency = "transparency"
	//ORTBExtPrebidTransparencyContent get api parameter content
	ORTBExtPrebidTransparencyContent = "content"

	ORTBExtPrebidFloors            = "floors"
	ORTBExtFloorEnforcement        = "enforcement"
	ORTBExtPrebidFloorsEnforcement = "req.ext.prebid.floors.enforcement"

	Debug = "debug"
)

const (
	ErrJSONMarshalFailed    = `error:[json_marshal_failed] object:[%s] message:[%s]`
	ErrJSONUnmarshalFailed  = `error:[json_unmarshal_failed] object:[%s] message:[%s] payload:[%s]`
	ErrTypeCastFailed       = `error:[type_cast_failed] key:[%s] type:[%s] value:[%v]`
	ErrHTTPNewRequestFailed = `error:[setup_new_request_failed] method:[%s] endpoint:[%s] message:[%s]`
)
