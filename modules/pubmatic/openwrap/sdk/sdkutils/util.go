package sdkutils

import (
	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func CopyPath(source []byte, target []byte, path ...string) ([]byte, error) {
	if source == nil {
		return target, nil
	}

	value, dataType, _, err := jsonparser.Get(source, path...)
	if value == nil || err != nil {
		return target, err
	}

	// Initialize target if nil
	if target == nil {
		target = []byte(`{}`)
	}

	// Early return for null values
	if dataType == jsonparser.Null {
		return jsonparser.Set(target, nil, path...)
	}

	// Handle empty values based on data type
	switch dataType {
	case jsonparser.String:
		// Only skip if it's an empty string
		if len(value) == 0 {
			return target, nil
		}
		// Quote the string value
		quotedValue := []byte(`"` + string(value) + `"`)
		return jsonparser.Set(target, quotedValue, path...)

	case jsonparser.Number:
		// Preserve numeric value
		return jsonparser.Set(target, value, path...)

	case jsonparser.Boolean:
		// Preserve boolean value
		return jsonparser.Set(target, value, path...)

	case jsonparser.Array, jsonparser.Object:
		// Skip empty arrays/objects
		if len(value) <= 2 { // "[]" or "{}" are 2 chars
			return target, nil
		}
		// Preserve the complex value as is
		return jsonparser.Set(target, value, path...)

	default:
		// For unknown types, copy as is
		return jsonparser.Set(target, value, path...)
	}
}

func IsSdkIntegration(endpoint string) bool {
	return endpoint == models.EndpointAppLovinMax || endpoint == models.EndpointUnityLevelPlay || endpoint == models.EndpointGoogleSDK
}

func AddSize300x600ForInterstitialBanner(imp *openrtb2.Imp) {
	if imp.Banner == nil {
		return
	}
	var is320x480SizeFound, is300x600SizeFound bool
	if imp.Banner.W != nil && imp.Banner.H != nil {
		if *imp.Banner.W == 320 && *imp.Banner.H == 480 {
			is320x480SizeFound = true
		}
		if *imp.Banner.W == 300 && *imp.Banner.H == 600 {
			is300x600SizeFound = true
		}
	}
	for _, size := range imp.Banner.Format {
		if size.W == 320 && size.H == 480 {
			is320x480SizeFound = true
		}
		if size.W == 300 && size.H == 600 {
			is300x600SizeFound = true
		}
	}
	if is320x480SizeFound && !is300x600SizeFound {
		imp.Banner.Format = append(imp.Banner.Format, openrtb2.Format{W: 300, H: 600})
	}
}
