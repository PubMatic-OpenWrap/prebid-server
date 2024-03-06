package models

import "github.com/prebid/openrtb/v19/adcom1"

func GetAPIFramework(api []int) []adcom1.APIFramework {
	if api == nil {
		return nil
	}
	adComAPIs := make([]adcom1.APIFramework, len(api))

	for index, value := range api {
		adComAPIs[index] = adcom1.APIFramework(value)
	}

	return adComAPIs
}

func GetPlaybackMethod(playbackMethods []int) []adcom1.PlaybackMethod {
	if playbackMethods == nil {
		return nil
	}
	methods := make([]adcom1.PlaybackMethod, len(playbackMethods))

	for index, value := range playbackMethods {
		methods[index] = adcom1.PlaybackMethod(value)
	}

	return methods
}

func GetDeliveryMethod(deliveryMethods []int) []adcom1.DeliveryMethod {
	if deliveryMethods == nil {
		return nil
	}
	methods := make([]adcom1.DeliveryMethod, len(deliveryMethods))

	for index, value := range deliveryMethods {
		methods[index] = adcom1.DeliveryMethod(value)
	}

	return methods
}

func GetCompanionType(companionTypes []int) []adcom1.CompanionType {
	if companionTypes == nil {
		return nil
	}
	adcomCompanionTypes := make([]adcom1.CompanionType, len(companionTypes))

	for index, value := range companionTypes {
		adcomCompanionTypes[index] = adcom1.CompanionType(value)
	}

	return adcomCompanionTypes
}

func GetCreativeAttributes(creativeAttributes []int) []adcom1.CreativeAttribute {
	if creativeAttributes == nil {
		return nil
	}
	adcomCreatives := make([]adcom1.CreativeAttribute, len(creativeAttributes))

	for index, value := range creativeAttributes {
		adcomCreatives[index] = adcom1.CreativeAttribute(value)
	}

	return adcomCreatives
}

func GetProtocol(protocols []int) []adcom1.MediaCreativeSubtype {
	if protocols == nil {
		return nil
	}
	adComProtocols := make([]adcom1.MediaCreativeSubtype, len(protocols))

	for index, value := range protocols {
		adComProtocols[index] = adcom1.MediaCreativeSubtype(value)
	}

	return adComProtocols
}

// BannerAdType
// Types of ads that can be accepted by the exchange unless restricted by publisher site settings.
type BannerAdType int8

const (
	BannerAdTypeXHTMLTextAd   BannerAdType = 1 // XHTML Text Ad (usually mobile)
	BannerAdTypeXHTMLBannerAd BannerAdType = 2 // XHTML Banner Ad. (usually mobile)
	BannerAdTypeJavaScriptAd  BannerAdType = 3 // JavaScript Ad; must be valid XHTML (i.e., Script Tags Included)
	BannerAdTypeIframe        BannerAdType = 4 // iframe
)

func GetBannderAdType(adTypes []int) []BannerAdType {
	if adTypes == nil {
		return nil
	}
	bannerAdTypes := make([]BannerAdType, len(adTypes))

	for index, value := range adTypes {
		bannerAdTypes[index] = BannerAdType(value)
	}

	return bannerAdTypes
}

func GetExpandableDirection(expdirs []int) []adcom1.ExpandableDirection {
	if expdirs == nil {
		return nil
	}
	adComExDir := make([]adcom1.ExpandableDirection, len(expdirs))

	for index, value := range expdirs {
		adComExDir[index] = adcom1.ExpandableDirection(value)
	}

	return adComExDir
}

func GetConnectionType(connectionType []int) []adcom1.ConnectionType {
	if connectionType == nil {
		return nil
	}
	adComExDir := make([]adcom1.ConnectionType, len(connectionType))
	for index, value := range connectionType {
		adComExDir[index] = adcom1.ConnectionType(value)
	}

	return adComExDir
}
