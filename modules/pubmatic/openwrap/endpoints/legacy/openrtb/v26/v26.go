package v26

import "github.com/prebid/openrtb/v20/adcom1"

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
