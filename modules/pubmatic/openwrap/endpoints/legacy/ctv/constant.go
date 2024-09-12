package ctv

const (
	//ArraySeparator get api array separator
	ArraySeparator = ","
	//Ext get api ext parameter
	Ext = ".ext."
	//ExtLen get api ext parameter length
	ExtLen = len(Ext)

	// USD denotes currency USD
	USD   = "USD"
	Debug = "debug"

	BIDDER_KEY = "bidder"
	PrebidKey  = "prebid"
)

const (
	// BidRequest level parameters
	ORTBBidRequestID      = "req.id"      //ORTBBidRequestID get api parameter req.id
	ORTBBidRequestTest    = "req.test"    //ORTBBidRequestTest get api parameter req.test
	ORTBBidRequestAt      = "req.at"      //ORTBBidRequestAt get api parameter req.at
	ORTBBidRequestTmax    = "req.tmax"    //ORTBBidRequestTmax get api parameter req.tmax
	ORTBBidRequestWseat   = "req.wseat"   //ORTBBidRequestWseat get api parameter req.wseat
	ORTBBidRequestWlang   = "req.wlang"   //ORTBBidRequestWlang get api parameter req.wlang
	ORTBBidRequestBseat   = "req.bseat"   //ORTBBidRequestBseat get api parameter req.bseat
	ORTBBidRequestAllImps = "req.allimps" //ORTBBidRequestAllImps get api parameter req.allimps
	ORTBBidRequestCur     = "req.cur"     //ORTBBidRequestCur get api parameter req.cur
	ORTBBidRequestBcat    = "req.bcat"    //ORTBBidRequestBcat get api parameter req.bcat
	ORTBBidRequestBadv    = "req.badv"    //ORTBBidRequestBadv get api parameter req.badv
	ORTBBidRequestBapp    = "req.bapp"    //ORTBBidRequestBapp get api parameter req.bapp

	// Source level parameters
	ORTBSourceFD     = "src.fd"     //ORTBSourceFD get api parameter src.fd
	ORTBSourceTID    = "src.tid"    //ORTBSourceTID get api parameter src.tid
	ORTBSourcePChain = "src.pchain" //ORTBSourcePChain get api parameter src.pchain
	ORTBSourceSChain = "src.schain" //ORTBSourceSChain get api parameter src.ext.schain

	// Regs level parameters
	ORTBRegsCoppa  = "regs.coppa"   //ORTBRegsCoppa get api parameter regs.coppa
	ORTBRegsGpp    = "regs.gpp"     // ORTB get api parameter for gpp in regs
	ORTBRegsGppSid = "regs.gpp_sid" // ORTB get api parameter for gpp_sid in regs

	// Imp level parameters
	ORTBImpID                = "imp.id"                //ORTBImpID get api parameter imp.id
	ORTBImpDisplayManager    = "imp.displaymanager"    //ORTBImpDisplayManager get api parameter imp.displaymanager
	ORTBImpDisplayManagerVer = "imp.displaymanagerver" //ORTBImpDisplayManagerVer get api parameter imp.displaymanagerver
	ORTBImpInstl             = "imp.instl"             //ORTBImpInstl get api parameter imp.instl
	ORTBImpTagID             = "imp.tagid"             //ORTBImpTagID get api parameter imp.tagid
	ORTBImpBidFloor          = "imp.bidfloor"          //ORTBImpBidFloor get api parameter imp.bidfloor
	ORTBImpBidFloorCur       = "imp.bidfloorcur"       //ORTBImpBidFloorCur get api parameter imp.bidfloorcur
	ORTBImpClickBrowser      = "imp.clickbrowser"      //ORTBImpClickBrowser get api parameter imp.clickbrowser
	ORTBImpSecure            = "imp.secure"            //ORTBImpSecure get api parameter imp.secure
	ORTBImpIframeBuster      = "imp.iframebuster"      //ORTBImpIframeBuster get api parameter imp.iframebuster
	ORTBImpExp               = "imp.exp"               //ORTBImpExp get api parameter imp.exp
	ORTBImpPmp               = "imp.pmp"               //ORTBImpPmp get api parameter imp.pmp
	ORTBImpExtBidder         = "imp.ext.bidder"        //ORTBImpExtBidder get api parameter imp.ext
	ORTBImpExtPrebid         = "imp.ext.prebid"        //ORTBImpExtPrebid get api parameter imp.ext.prebid

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

	ORTBImpVideoMimes          = "imp.vid.mimes"          //ORTBImpVideoMimes get api parameter imp.vid.mimes
	ORTBImpVideoMinDuration    = "imp.vid.minduration"    //ORTBImpVideoMinDuration get api parameter imp.vid.minduration
	ORTBImpVideoMaxDuration    = "imp.vid.maxduration"    //ORTBImpVideoMaxDuration get api parameter imp.vid.maxduration
	ORTBImpVideoProtocols      = "imp.vid.protocols"      //ORTBImpVideoProtocols get api parameter imp.vid.protocols
	ORTBImpVideoPlayerWidth    = "imp.vid.w"              //ORTBImpVideoPlayerWidth get api parameter imp.vid.w
	ORTBImpVideoPlayerHeight   = "imp.vid.h"              //ORTBImpVideoPlayerHeight get api parameter imp.vid.h
	ORTBImpVideoStartDelay     = "imp.vid.startdelay"     //ORTBImpVideoStartDelay get api parameter imp.vid.startdelay
	ORTBImpVideoPlacement      = "imp.vid.placement"      //ORTBImpVideoPlacement get api parameter imp.vid.placement
	ORTBImpVideoPlcmt          = "imp.vid.plcmt"          //ORTBImpVideoPlacement get api parameter imp.vid.plcmt
	ORTBImpVideoLinearity      = "imp.vid.linearity"      //ORTBImpVideoLinearity get api parameter imp.vid.linearity
	ORTBImpVideoSkip           = "imp.vid.skip"           //ORTBImpVideoSkip get api parameter imp.vid.skip
	ORTBImpVideoSkipMin        = "imp.vid.skipmin"        //ORTBImpVideoSkipMin get api parameter imp.vid.skipmin
	ORTBImpVideoSkipAfter      = "imp.vid.skipafter"      //ORTBImpVideoSkipAfter get api parameter imp.vid.skipafter
	ORTBImpVideoSequence       = "imp.vid.sequence"       //ORTBImpVideoSequence get api parameter imp.vid.sequence
	ORTBImpVideoBAttr          = "imp.vid.battr"          //ORTBImpVideoBAttr get api parameter imp.vid.battr
	ORTBImpVideoMaxExtended    = "imp.vid.maxextended"    //ORTBImpVideoMaxExtended get api parameter imp.vid.maxextended
	ORTBImpVideoMinBitrate     = "imp.vid.minbitrate"     //ORTBImpVideoMinBitrate get api parameter imp.vid.minbitrate
	ORTBImpVideoMaxBitrate     = "imp.vid.maxbitrate"     //ORTBImpVideoMaxBitrate get api parameter imp.vid.maxbitrate
	ORTBImpVideoBoxingAllowed  = "imp.vid.boxingallowed"  //ORTBImpVideoBoxingAllowed get api parameter imp.vid.boxingallowed
	ORTBImpVideoPlaybackMethod = "imp.vid.playbackmethod" //ORTBImpVideoPlaybackMethod get api parameter imp.vid.playbackmethod
	ORTBImpVideoDelivery       = "imp.vid.delivery"       //ORTBImpVideoDelivery get api parameter imp.vid.delivery
	ORTBImpVideoPos            = "imp.vid.pos"            //ORTBImpVideoPos get api parameter imp.vid.pos
	ORTBImpVideoAPI            = "imp.vid.api"            //ORTBImpVideoAPI get api parameter imp.vid.api
	ORTBImpVideoCompanionType  = "imp.vid.companiontype"  //ORTBImpVideoCompanionType get api parameter imp.vid.companiontype

	// Site level parameters
	ORTBSiteID            = "site.id"            //ORTBSiteID get api parameter site.id
	ORTBSiteName          = "site.name"          //ORTBSiteName get api parameter site.name
	ORTBSiteDomain        = "site.domain"        //ORTBSiteDomain get api parameter site.domain
	ORTBSitePage          = "site.page"          //ORTBSitePage get api parameter site.page
	ORTBSiteRef           = "site.ref"           //ORTBSiteRef get api parameter site.ref
	ORTBSiteSearch        = "site.search"        //ORTBSiteSearch get api parameter site.search
	ORTBSiteMobile        = "site.mobile"        //ORTBSiteMobile get api parameter site.mobile
	ORTBSiteCat           = "site.cat"           //ORTBSiteCat get api parameter site.cat
	ORTBSiteSectionCat    = "site.sectioncat"    //ORTBSiteSectionCat get api parameter site.sectioncat
	ORTBSitePageCat       = "site.pagecat"       //ORTBSitePageCat get api parameter site.pagecat
	ORTBSitePrivacyPolicy = "site.privacypolicy" //ORTBSitePrivacyPolicy get api parameter site.privacypolicy
	ORTBSiteKeywords      = "site.keywords"      //ORTBSiteKeywords get api parameter site.keywords

	// App level parameters

	ORTBAppID            = "app.id"            //ORTBAppID get api parameter app.id
	ORTBAppName          = "app.name"          //ORTBAppName get api parameter app.name
	ORTBAppBundle        = "app.bundle"        //ORTBAppBundle get api parameter app.bundle
	ORTBAppDomain        = "app.domain"        //ORTBAppDomain get api parameter app.domain
	ORTBAppStoreURL      = "app.storeurl"      //ORTBAppStoreURL get api parameter app.storeurl
	ORTBAppVer           = "app.ver"           //ORTBAppVer get api parameter app.ver
	ORTBAppPaid          = "app.paid"          //ORTBAppPaid get api parameter app.paid
	ORTBAppCat           = "app.cat"           //ORTBAppCat get api parameter app.cat
	ORTBAppSectionCat    = "app.sectioncat"    //ORTBAppSectionCat get api parameter app.sectioncat
	ORTBAppPageCat       = "app.pagecat"       //ORTBAppPageCat get api parameter app.pagecat
	ORTBAppPrivacyPolicy = "app.privacypolicy" //ORTBAppPrivacyPolicy get api parameter app.privacypolicy
	ORTBAppKeywords      = "app.keywords"      //ORTBAppKeywords get api parameter app.keywords

	// Site.Publisher level parameters

	ORTBSitePublisherID     = "site.pub.id"     //ORTBSitePublisherID get api parameter site.pub.id
	ORTBSitePublisherName   = "site.pub.name"   //ORTBSitePublisherName get api parameter site.pub.name
	ORTBSitePublisherCat    = "site.pub.cat"    //ORTBSitePublisherCat get api parameter site.pub.cat
	ORTBSitePublisherDomain = "site.pub.domain" //ORTBSitePublisherDomain get api parameter site.pub.domain

	// Site.Content level parameters

	ORTBSiteContentID                 = "site.cnt.id"                 //ORTBSiteContentID get api parameter site.cnt.id
	ORTBSiteContentEpisode            = "site.cnt.episode"            //ORTBSiteContentEpisode get api parameter site.cnt.episode
	ORTBSiteContentTitle              = "site.cnt.title"              //ORTBSiteContentTitle get api parameter site.cnt.title
	ORTBSiteContentSeries             = "site.cnt.series"             //ORTBSiteContentSeries get api parameter site.cnt.series
	ORTBSiteContentSeason             = "site.cnt.season"             //ORTBSiteContentSeason get api parameter site.cnt.season
	ORTBSiteContentArtist             = "site.cnt.artist"             //ORTBSiteContentArtist get api parameter site.cnt.artist
	ORTBSiteContentGenre              = "site.cnt.genre"              //ORTBSiteContentGenre get api parameter site.cnt.genre
	ORTBSiteContentAlbum              = "site.cnt.album"              //ORTBSiteContentAlbum get api parameter site.cnt.album
	ORTBSiteContentIsRc               = "site.cnt.isrc"               //ORTBSiteContentIsRc get api parameter site.cnt.isrc
	ORTBSiteContentURL                = "site.cnt.url"                //ORTBSiteContentURL get api parameter site.cnt.url
	ORTBSiteContentCat                = "site.cnt.cat"                //ORTBSiteContentCat get api parameter site.cnt.cat
	ORTBSiteContentProdQ              = "site.cnt.prodq"              //ORTBSiteContentProdQ get api parameter site.cnt.prodq
	ORTBSiteContentVideoQuality       = "site.cnt.videoquality"       //ORTBSiteContentVideoQuality get api parameter site.cnt.videoquality
	ORTBSiteContentContext            = "site.cnt.context"            //ORTBSiteContentContext get api parameter site.cnt.context
	ORTBSiteContentContentRating      = "site.cnt.contentrating"      //ORTBSiteContentContentRating get api parameter site.cnt.contentrating
	ORTBSiteContentUserRating         = "site.cnt.userrating"         //ORTBSiteContentUserRating get api parameter site.cnt.userrating
	ORTBSiteContentQaGmeDiarating     = "site.cnt.qagmediarating"     //ORTBSiteContentQaGmeDiarating get api parameter site.cnt.qagmediarating
	ORTBSiteContentKeywords           = "site.cnt.keywords"           //ORTBSiteContentKeywords get api parameter site.cnt.keywords
	ORTBSiteContentLiveStream         = "site.cnt.livestream"         //ORTBSiteContentLiveStream get api parameter site.cnt.livestream
	ORTBSiteContentSourceRelationship = "site.cnt.sourcerelationship" //ORTBSiteContentSourceRelationship get api parameter site.cnt.sourcerelationship
	ORTBSiteContentLen                = "site.cnt.len"                //ORTBSiteContentLen get api parameter site.cnt.len
	ORTBSiteContentLanguage           = "site.cnt.language"           //ORTBSiteContentLanguage get api parameter site.cnt.language
	ORTBSiteContentEmbeddable         = "site.cnt.embeddable"         //ORTBSiteContentEmbeddable get api parameter site.cnt.embeddable

	// Site.Producer level parameters
	ORTBSiteContentProducerID     = "site.cnt.prod.id"     //ORTBSiteContentProducerID get api parameter site.cnt.prod.id
	ORTBSiteContentProducerName   = "site.cnt.prod.name"   //ORTBSiteContentProducerName get api parameter site.cnt.prod.name
	ORTBSiteContentProducerCat    = "site.cnt.prod.cat"    //ORTBSiteContentProducerCat get api parameter site.cnt.prod.cat
	ORTBSiteContentProducerDomain = "site.cnt.prod.domain" //ORTBSiteContentProducerDomain get api parameter site.cnt.prod.domain

	// App.Publisher level parameters
	ORTBAppPublisherID     = "app.pub.id"     //ORTBAppPublisherID get api parameter app.pub.id
	ORTBAppPublisherName   = "app.pub.name"   //ORTBAppPublisherName get api parameter app.pub.name
	ORTBAppPublisherCat    = "app.pub.cat"    //ORTBAppPublisherCat get api parameter app.pub.cat
	ORTBAppPublisherDomain = "app.pub.domain" //ORTBAppPublisherDomain get api parameter app.pub.domain

	// App.Content level parameters
	ORTBAppContentID                 = "app.cnt.id"                 //ORTBAppContentID get api parameter app.cnt.id
	ORTBAppContentEpisode            = "app.cnt.episode"            //ORTBAppContentEpisode get api parameter app.cnt.episode
	ORTBAppContentTitle              = "app.cnt.title"              //ORTBAppContentTitle get api parameter app.cnt.title
	ORTBAppContentSeries             = "app.cnt.series"             //ORTBAppContentSeries get api parameter app.cnt.series
	ORTBAppContentSeason             = "app.cnt.season"             //ORTBAppContentSeason get api parameter app.cnt.season
	ORTBAppContentArtist             = "app.cnt.artist"             //ORTBAppContentArtist get api parameter app.cnt.artist
	ORTBAppContentGenre              = "app.cnt.genre"              //ORTBAppContentGenre get api parameter app.cnt.genre
	ORTBAppContentAlbum              = "app.cnt.album"              //ORTBAppContentAlbum get api parameter app.cnt.album
	ORTBAppContentIsRc               = "app.cnt.isrc"               //ORTBAppContentIsRc get api parameter app.cnt.isrc
	ORTBAppContentURL                = "app.cnt.url"                //ORTBAppContentURL get api parameter app.cnt.url
	ORTBAppContentCat                = "app.cnt.cat"                //ORTBAppContentCat get api parameter app.cnt.cat
	ORTBAppContentProdQ              = "app.cnt.prodq"              //ORTBAppContentProdQ get api parameter app.cnt.prodq
	ORTBAppContentVideoQuality       = "app.cnt.videoquality"       //ORTBAppContentVideoQuality get api parameter app.cnt.videoquality
	ORTBAppContentContext            = "app.cnt.context"            //ORTBAppContentContext get api parameter app.cnt.context
	ORTBAppContentContentRating      = "app.cnt.contentrating"      //ORTBAppContentContentRating get api parameter app.cnt.contentrating
	ORTBAppContentUserRating         = "app.cnt.userrating"         //ORTBAppContentUserRating get api parameter app.cnt.userrating
	ORTBAppContentQaGmeDiarating     = "app.cnt.qagmediarating"     //ORTBAppContentQaGmeDiarating get api parameter app.cnt.qagmediarating
	ORTBAppContentKeywords           = "app.cnt.keywords"           //ORTBAppContentKeywords get api parameter app.cnt.keywords
	ORTBAppContentLiveStream         = "app.cnt.livestream"         //ORTBAppContentLiveStream get api parameter app.cnt.livestream
	ORTBAppContentSourceRelationship = "app.cnt.sourcerelationship" //ORTBAppContentSourceRelationship get api parameter app.cnt.sourcerelationship
	ORTBAppContentLen                = "app.cnt.len"                //ORTBAppContentLen get api parameter app.cnt.len
	ORTBAppContentLanguage           = "app.cnt.language"           //ORTBAppContentLanguage get api parameter app.cnt.language
	ORTBAppContentEmbeddable         = "app.cnt.embeddable"         //ORTBAppContentEmbeddable get api parameter app.cnt.embeddable

	// App.Producer level parameters
	ORTBAppContentProducerID     = "app.cnt.prod.id"     //ORTBAppContentProducerID get api parameter app.cnt.prod.id
	ORTBAppContentProducerName   = "app.cnt.prod.name"   //ORTBAppContentProducerName get api parameter app.cnt.prod.name
	ORTBAppContentProducerCat    = "app.cnt.prod.cat"    //ORTBAppContentProducerCat get api parameter app.cnt.prod.cat
	ORTBAppContentProducerDomain = "app.cnt.prod.domain" //ORTBAppContentProducerDomain get api parameter app.cnt.prod.domain

	// Device level parameters
	ORTBDeviceUserAgent      = "dev.ua"             //ORTBDeviceUserAgent get api parameter dev.ua
	ORTBDeviceDnt            = "dev.dnt"            //ORTBDeviceDnt get api parameter dev.dnt
	ORTBDeviceLmt            = "dev.lmt"            //ORTBDeviceLmt get api parameter dev.lmt
	ORTBDeviceIP             = "dev.ip"             //ORTBDeviceIP get api parameter dev.ip
	ORTBDeviceIpv6           = "dev.ipv6"           //ORTBDeviceIpv6 get api parameter dev.ipv6
	ORTBDeviceDeviceType     = "dev.devicetype"     //ORTBDeviceDeviceType get api parameter dev.devicetype
	ORTBDeviceMake           = "dev.make"           //ORTBDeviceMake get api parameter dev.make
	ORTBDeviceModel          = "dev.model"          //ORTBDeviceModel get api parameter dev.model
	ORTBDeviceOs             = "dev.os"             //ORTBDeviceOs get api parameter dev.os
	ORTBDeviceOsv            = "dev.osv"            //ORTBDeviceOsv get api parameter dev.osv
	ORTBDeviceHwv            = "dev.hwv"            //ORTBDeviceHwv get api parameter dev.hwv
	ORTBDeviceWidth          = "dev.w"              //ORTBDeviceWidth get api parameter dev.w
	ORTBDeviceHeight         = "dev.h"              //ORTBDeviceHeight get api parameter dev.h
	ORTBDevicePpi            = "dev.ppi"            //ORTBDevicePpi get api parameter dev.ppi
	ORTBDevicePxRatio        = "dev.pxratio"        //ORTBDevicePxRatio get api parameter dev.pxratio
	ORTBDeviceJS             = "dev.js"             //ORTBDeviceJS get api parameter dev.js
	ORTBDeviceGeoFetch       = "dev.geofetch"       //ORTBDeviceGeoFetch get api parameter dev.geofetch
	ORTBDeviceFlashVer       = "dev.flashver"       //ORTBDeviceFlashVer get api parameter dev.flashver
	ORTBDeviceLanguage       = "dev.language"       //ORTBDeviceLanguage get api parameter dev.language
	ORTBDeviceCarrier        = "dev.carrier"        //ORTBDeviceCarrier get api parameter dev.carrier
	ORTBDeviceMccmnc         = "dev.mccmnc"         //ORTBDeviceMccmnc get api parameter dev.mccmnc
	ORTBDeviceConnectionType = "dev.connectiontype" //ORTBDeviceConnectionType get api parameter dev.connectiontype
	ORTBDeviceIfa            = "dev.ifa"            //ORTBDeviceIfa get api parameter dev.ifa
	ORTBDeviceDidSha1        = "dev.didsha1"        //ORTBDeviceDidSha1 get api parameter dev.didsha1
	ORTBDeviceDidMd5         = "dev.didmd5"         //ORTBDeviceDidMd5 get api parameter dev.didmd5
	ORTBDeviceDpidSha1       = "dev.dpidsha1"       //ORTBDeviceDpidSha1 get api parameter dev.dpidsha1
	ORTBDeviceDpidMd5        = "dev.dpidmd5"        //ORTBDeviceDpidMd5 get api parameter dev.dpidmd5
	ORTBDeviceMacSha1        = "dev.macsha1"        //ORTBDeviceMacSha1 get api parameter dev.macsha1
	ORTBDeviceMacMd5         = "dev.macmd5"         //ORTBDeviceMacMd5 get api parameter dev.macmd5

	// Device.Geo level parameters
	ORTBDeviceGeoLat           = "dev.geo.lat"           //ORTBDeviceGeoLat get api parameter dev.geo.lat
	ORTBDeviceGeoLon           = "dev.geo.lon"           //ORTBDeviceGeoLon get api parameter dev.geo.lon
	ORTBDeviceGeoType          = "dev.geo.type"          //ORTBDeviceGeoType get api parameter dev.geo.type
	ORTBDeviceGeoAccuracy      = "dev.geo.accuracy"      //ORTBDeviceGeoAccuracy get api parameter dev.geo.accuracy
	ORTBDeviceGeoLastFix       = "dev.geo.lastfix"       //ORTBDeviceGeoLastFix get api parameter dev.geo.lastfix
	ORTBDeviceGeoIPService     = "dev.geo.ipservice"     //ORTBDeviceGeoIPService get api parameter dev.geo.ipservice
	ORTBDeviceGeoCountry       = "dev.geo.country"       //ORTBDeviceGeoCountry get api parameter dev.geo.country
	ORTBDeviceGeoRegion        = "dev.geo.region"        //ORTBDeviceGeoRegion get api parameter dev.geo.region
	ORTBDeviceGeoRegionFips104 = "dev.geo.regionfips104" //ORTBDeviceGeoRegionFips104 get api parameter dev.geo.regionfips104
	ORTBDeviceGeoMetro         = "dev.geo.metro"         //ORTBDeviceGeoMetro get api parameter dev.geo.metro
	ORTBDeviceGeoCity          = "dev.geo.city"          //ORTBDeviceGeoCity get api parameter dev.geo.city
	ORTBDeviceGeoZip           = "dev.geo.zip"           //ORTBDeviceGeoZip get api parameter dev.geo.zip
	ORTBDeviceGeoUtcOffset     = "dev.geo.utcoffset"     //ORTBDeviceGeoUtcOffset get api parameter dev.geo.utcoffset

	// User level parameters
	ORTBUserID         = "user.id"         //ORTBUserID get api parameter user.id
	ORTBUserBuyerUID   = "user.buyeruid"   //ORTBUserBuyerUID get api parameter user.buyeruid
	ORTBUserYob        = "user.yob"        //ORTBUserYob get api parameter user.yob
	ORTBUserGender     = "user.gender"     //ORTBUserGender get api parameter user.gender
	ORTBUserKeywords   = "user.keywords"   //ORTBUserKeywords get api parameter user.keywords
	ORTBUserCustomData = "user.customdata" //ORTBUserCustomData get api parameter user.customdata

	// User.Geo level parameters
	ORTBUserGeoLat           = "user.geo.lat"           //ORTBUserGeoLat get api parameter user.geo.lat
	ORTBUserGeoLon           = "user.geo.lon"           //ORTBUserGeoLon get api parameter user.geo.lon
	ORTBUserGeoType          = "user.geo.type"          //ORTBUserGeoType get api parameter user.geo.type
	ORTBUserGeoAccuracy      = "user.geo.accuracy"      //ORTBUserGeoAccuracy get api parameter user.geo.accuracy
	ORTBUserGeoLastFix       = "user.geo.lastfix"       //ORTBUserGeoLastFix get api parameter user.geo.lastfix
	ORTBUserGeoIPService     = "user.geo.ipservice"     //ORTBUserGeoIPService get api parameter user.geo.ipservice
	ORTBUserGeoCountry       = "user.geo.country"       //ORTBUserGeoCountry get api parameter user.geo.country
	ORTBUserGeoRegion        = "user.geo.region"        //ORTBUserGeoRegion get api parameter user.geo.region
	ORTBUserGeoRegionFips104 = "user.geo.regionfips104" //ORTBUserGeoRegionFips104 get api parameter user.geo.regionfips104
	ORTBUserGeoMetro         = "user.geo.metro"         //ORTBUserGeoMetro get api parameter user.geo.metro
	ORTBUserGeoCity          = "user.geo.city"          //ORTBUserGeoCity get api parameter user.geo.city
	ORTBUserGeoZip           = "user.geo.zip"           //ORTBUserGeoZip get api parameter user.geo.zip
	ORTBUserGeoUtcOffset     = "user.geo.utcoffset"     //ORTBUserGeoUtcOffset get api parameter user.geo.utcoffset

	// User.Ext.Consent level parameters
	ORTBUserExtConsent = "user.ext.consent" //ORTBUserExtConsent get api parameter user.ext.consent
	ORTBUserExtEIDS    = "user.ext.eids"    //ORTBUserExtEIDS get api parameter user.ext.eids
	ORTBUserData       = "user.data"        //ORTBUserData get api parameter user.data
	ORTBExtEIDS        = "eids"             //ORTBExtEIDS parameter

	// Regs.Ext.GDPR level parameters
	ORTBRegsExtGdpr      = "regs.ext.gdpr"       //ORTBRegsExtGdpr get api parameter regs.ext.gdpr
	ORTBRegsExtUSPrivacy = "regs.ext.us_privacy" //ORTBRegsExtUSPrivacy get api parameter regs.ext.us_privacy
	ORTBExtUSPrivacy     = "us_privacy"          //ORTBExtUSPrivacy get api parameter us_privacy

	// VideoExtension level parameters
	ORTBImpVideoExtOffset                           = "imp.vid.ext.offset"              //ORTBImpVideoExtOffset get api parameter imp.vid.ext.offset
	ORTBImpVideoExtAdPodMinAds                      = "imp.vid.ext.adpod.minads"        //ORTBImpVideoExtAdPodMinAds get api parameter imp.vid.ext.adpod.minads
	ORTBImpVideoExtAdPodMaxAds                      = "imp.vid.ext.adpod.maxads"        //ORTBImpVideoExtAdPodMaxAds get api parameter imp.vid.ext.adpod.maxads
	ORTBImpVideoExtAdPodMinDuration                 = "imp.vid.ext.adpod.adminduration" //ORTBImpVideoExtAdPodMinDuration get api parameter imp.vid.ext.adpod.adminduration
	ORTBImpVideoExtAdPodMaxDuration                 = "imp.vid.ext.adpod.admaxduration" //ORTBImpVideoExtAdPodMaxDuration get api parameter imp.vid.ext.adpod.admaxduration
	ORTBImpVideoExtAdPodAdvertiserExclusionPercent  = "imp.vid.ext.adpod.excladv"       //ORTBImpVideoExtAdPodAdvertiserExclusionPercent get api parameter imp.vid.ext.adpod.excladv
	ORTBImpVideoExtAdPodIABCategoryExclusionPercent = "imp.vid.ext.adpod.excliabcat"    //ORTBImpVideoExtAdPodIABCategoryExclusionPercent get api parameter imp.vid.ext.adpod.excliabcat

	// ReqWrapperExtension level parameters
	ORTBProfileID            = "req.ext.wrapper.profileid"            //ORTBProfileID get api parameter req.ext.wrapper.profileid
	ORTBVersionID            = "req.ext.wrapper.versionid"            //ORTBVersionID get api parameter req.ext.wrapper.versionid
	ORTBSSAuctionFlag        = "req.ext.wrapper.ssauction"            //ORTBSSAuctionFlag get api parameter req.ext.wrapper.ssauction
	ORTBSumryDisableFlag     = "req.ext.wrapper.sumry_disable"        //ORTBSumryDisableFlag get api parameter req.ext.wrapper.sumry_disable
	ORTBClientConfigFlag     = "req.ext.wrapper.clientconfig"         //ORTBClientConfigFlag get api parameter req.ext.wrapper.clientconfig
	ORTBSupportDeals         = "req.ext.wrapper.supportdeals"         //ORTBSupportDeals get api parameter req.ext.wrapper.supportdeals
	ORTBIncludeBrandCategory = "req.ext.wrapper.includebrandcategory" //ORTBIncludeBrandCategory get api parameter req.ext.wrapper.includebrandcategory
	ORTBSSAI                 = "req.ext.wrapper.ssai"                 //ORTBSSAI get the api parameter req.ext.wrapper.ssai
	ORTBKeyValues            = "req.ext.wrapper.kv"                   //ORTBKeyValues get the api parameter req.ext.wrapper.kv
	ORTBKeyValuesMap         = "req.ext.wrapper.kvm"                  //ORTBKeyValuesMap get the api parameter req.ext.wrapper.kvm

	// ReqAdPodExt level parameters
	ORTBRequestExtAdPodMinAds                              = "req.ext.adpod.minads"             //ORTBRequestExtAdPodMinAds get api parameter req.ext.adpod.minads
	ORTBRequestExtAdPodMaxAds                              = "req.ext.adpod.maxads"             //ORTBRequestExtAdPodMaxAds get api parameter req.ext.adpod.maxads
	ORTBRequestExtAdPodMinDuration                         = "req.ext.adpod.adminduration"      //ORTBRequestExtAdPodMinDuration get api parameter req.ext.adpod.adminduration
	ORTBRequestExtAdPodMaxDuration                         = "req.ext.adpod.admaxduration"      //ORTBRequestExtAdPodMaxDuration get api parameter req.ext.adpod.admaxduration
	ORTBRequestExtAdPodAdvertiserExclusionPercent          = "req.ext.adpod.excladv"            //ORTBRequestExtAdPodAdvertiserExclusionPercent get api parameter req.ext.adpod.excladv
	ORTBRequestExtAdPodIABCategoryExclusionPercent         = "req.ext.adpod.excliabcat"         //ORTBRequestExtAdPodIABCategoryExclusionPercent get api parameter req.ext.adpod.excliabcat
	ORTBRequestExtAdPodCrossPodAdvertiserExclusionPercent  = "req.ext.adpod.crosspodexcladv"    //ORTBRequestExtAdPodCrossPodAdvertiserExclusionPercent get api parameter req.ext.adpod.crosspodexcladv
	ORTBRequestExtAdPodCrossPodIABCategoryExclusionPercent = "req.ext.adpod.crosspodexcliabcat" //ORTBRequestExtAdPodCrossPodIABCategoryExclusionPercent get api parameter req.ext.adpod.crosspodexcliabcat
	ORTBRequestExtAdPodIABCategoryExclusionWindow          = "req.ext.adpod.excliabcatwindow"   //ORTBRequestExtAdPodIABCategoryExclusionWindow get api parameter req.ext.adpod.excliabcatwindow
	ORTBRequestExtAdPodAdvertiserExclusionWindow           = "req.ext.adpod.excladvwindow"      //ORTBRequestExtAdPodAdvertiserExclusionWindow get api parameter req.ext.adpod.excladvwindow

	// ORTB Extension Objects */ //// get api parameter xtension Objelevel parameters
	ORTBBidRequestExt          = "req.ext"           //ORTBBidRequestExt get api parameter req.ext
	ORTBSourceExt              = "src.ext"           //ORTBSourceExt get api parameter src.ext
	ORTBRegsExt                = "regs.ext"          //ORTBRegsExt get api parameter regs.ext
	ORTBImpExt                 = "imp.ext"           //ORTBImpExt get api parameter imp.ext
	ORTBImpVideoExt            = "imp.vid.ext"       //ORTBImpVideoExt get api parameter imp.vid.ext
	ORTBSiteExt                = "site.ext"          //ORTBSiteExt get api parameter site.ext
	ORTBAppExt                 = "app.ext"           //ORTBAppExt get api parameter app.ext
	ORTBSitePublisherExt       = "site.pub.ext"      //ORTBSitePublisherExt get api parameter site.pub.ext
	ORTBSiteContentExt         = "site.cnt.ext"      //ORTBSiteContentExt get api parameter site.cnt.ext
	ORTBSiteContentProducerExt = "site.cnt.prod.ext" //ORTBSiteContentProducerExt get api parameter site.cnt.prod.ext
	ORTBAppPublisherExt        = "app.pub.ext"       //ORTBAppPublisherExt get api parameter app.pub.ext
	ORTBAppContentExt          = "app.cnt.ext"       //ORTBAppContentExt get api parameter app.cnt.ext
	ORTBAppContentProducerExt  = "app.cnt.prod.ext"  //ORTBAppContentProducerExt get api parameter app.cnt.prod.ext
	ORTBDeviceExt              = "dev.ext"           //ORTBDeviceExt get api parameter dev.ext
	ORTBDeviceGeoExt           = "dev.geo.ext"       //ORTBDeviceGeoExt get api parameter dev.geo.ext
	ORTBUserExt                = "user.ext"          //ORTBUserExt get api parameter user.ext
	ORTBUserGeoExt             = "user.geo.ext"      //ORTBUserGeoExt get api parameter user.geo.ext
	ORTBUserExtUIDS            = "uids"              //ORTBUserExtUIDs get api parameter user.ext.eids.uids
	ORTBUserExtID              = "id"                //ORTBUserExtID get api parameter user.ext.eids.uids.id

	// ORTB Extension Standard Keys */ //// get api parameter xtension Standard Klevel parameters
	ORTBExtWrapper                                  = "wrapper"              //ORTBExtWrapper get api parameter wrapper
	ORTBExtProfileId                                = "profileid"            //ORTBExtProfileId get api parameter profileid
	ORTBExtSsai                                     = "ssai"                 //ORTBExtSsai get api parameter ssai
	ORTBExtKV                                       = "kv"                   //ORTBExtKV get api parameter kv
	ORTBExtVersionId                                = "versionid"            //ORTBExtVersionId get api parameter versionid
	ORTBExtSSAuctionFlag                            = "ssauction"            //ORTBExtSSAuctionFlag get api parameter ssauction
	ORTBExtSumryDisableFlag                         = "sumry_disable"        //ORTBExtSumryDisableFlag get api parameter sumry_disable
	ORTBExtClientConfigFlag                         = "clientconfig"         //ORTBExtClientConfigFlag get api parameter clientconfig
	ORTBExtSupportDeals                             = "supportdeals"         //ORTBExtSupportDeals get api parameter supportdeals
	ORTBExtIncludeBrandCategory                     = "includebrandcategory" //ORTBExtIncludeBrandCategory get api parameter includebrandcategory
	ORTBExtGDPR                                     = "gdpr"                 //ORTBExtGDPR get api parameter gdpr
	ORTBExtConsent                                  = "consent"              //ORTBExtConsent get api parameter consent
	ORTBExtAdPod                                    = "adpod"                //ORTBExtAdPod get api parameter adpod
	ORTBExtAdPodOffset                              = "offset"               //ORTBExtAdPodOffset get api parameter offset
	ORTBExtAdPodMinAds                              = "minads"               //ORTBExtAdPodMinAds get api parameter minads
	ORTBExtAdPodMaxAds                              = "maxads"               //ORTBExtAdPodMaxAds get api parameter maxads
	ORTBExtAdPodMinDuration                         = "adminduration"        //ORTBExtAdPodMinDuration get api parameter adminduration
	ORTBExtAdPodMaxDuration                         = "admaxduration"        //ORTBExtAdPodMaxDuration get api parameter admaxduration
	ORTBExtAdPodAdvertiserExclusionPercent          = "excladv"              //ORTBExtAdPodAdvertiserExclusionPercent get api parameter excladv
	ORTBExtAdPodIABCategoryExclusionPercent         = "excliabcat"           //ORTBExtAdPodIABCategoryExclusionPercent get api parameter excliabcat
	ORTBExtAdPodCrossPodAdvertiserExclusionPercent  = "crosspodexcladv"      //ORTBExtAdPodCrossPodAdvertiserExclusionPercent get api parameter crosspodexcladv
	ORTBExtAdPodCrossPodIABCategoryExclusionPercent = "crosspodexcliabcat"   //ORTBExtAdPodCrossPodIABCategoryExclusionPercent get api parameter crosspodexcliabcat
	ORTBExtAdPodIABCategoryExclusionWindow          = "excliabcatwindow"     //ORTBExtAdPodIABCategoryExclusionWindow get api parameter excliabcatwindow
	ORTBExtAdPodAdvertiserExclusionWindow           = "excladvwindow"        //ORTBExtAdPodAdvertiserExclusionWindow get api parameter excladvwindow

	//Device Extensions Parameters
	ORTBDeviceExtIfaType   = "dev.ext.ifa_type"   //ORTBDeviceExtIfaType get api parameter ifa_type
	ORTBDeviceExtSessionID = "dev.ext.session_id" //ORTBDeviceExtSessionID get api parameter session_id
	ORTBDeviceExtATTS      = "dev.ext.atts"       //ORTBDeviceExtATTS get api parameter atts
	ORTBExtIfaType         = "ifa_type"           //ORBTExtDeviceIfaType get api parameter
	ORTBExtSessionID       = "session_id"         //ORTBExtSessionID parameter
	ORTBExtATTS            = "atts"               //ORBTExtDeviceATTS get api parameter

	ORTBExtPrebidBidderParams               = "bidderparams"                        //ORTBExtPrebidBidderParams get api parameter bidderparams
	ORTBExtPrebidBidderParamsPubmaticCDS    = "cds"                                 //ORTBExtPrebidBidderParamsPubmaticCDS get api parameter cds
	ORTBExtPrebid                           = "prebid"                              //ORTBExtPrebid get api parameter prebid
	ORTBExtPrebidTransparencyContent        = "content"                             //ORTBExtPrebidTransparencyContent get api parameter content
	ORTBExtPrebidTransparency               = "transparency"                        //ORTBExtPrebidTransparency get api parameter transparency
	ORTBRequestExtPrebidTransparencyContent = "req.ext.prebid.transparency.content" //ORTBRequestExtPrebidTransparencyContent get api parameter req.ext.prebid.transparency.content
	ORTBExtPrebidFloors                     = "floors"                              //ORTBExtPrebidFloors get api parameter for floors
	ORTBExtFloorEnforcement                 = "enforcement"                         //ORTBExtFloorEnforcement get api parameter for enforcement
	ORTBExtPrebidFloorsEnforcement          = "req.ext.prebid.floors.enforcement"   //ORTBExtPrebidFloorsEnforcement get api parameter for enforcement
	ORTBExtPrebidReturnAllBidStatus         = "req.ext.prebid.returnallbidstatus"   //ORTBExtPrebidReturnAllBidStatus get api parameter for returnallbidstatus
	ReturnAllBidStatus                      = "returnallbidstatus"

	ORTBExtAdrule = "adrule" ////ORTBExtAdrule get api parameter adrule
	ORTBVideo     = "video"
	ORTBAdrule    = "req.ext.wrapper.video.adrule"
)

const (
	ErrJSONMarshalFailed      = `error:[json_marshal_failed] object:[%s] message:[%s]`
	ErrJSONUnmarshalFailed    = `error:[json_unmarshal_failed] object:[%s] message:[%s] payload:[%s]`
	ErrTypeCastFailed         = `error:[type_cast_failed] key:[%s] type:[%s] value:[%v]`
	ErrHTTPNewRequestFailed   = `error:[setup_new_request_failed] method:[%s] endpoint:[%s] message:[%s]`
	ErrDeserializationFailed  = `error:[schain_validation_failed] object:[%s] message:[%s] pubid:[%s] payload:[%s]`
	ErrSchainValidationFailed = `error:[schain_validation_failed] object:[%s] message:[%s] pubid:[%s] profileid:[%s] payload:[%s]`
)
