package openrtb_ext

import (
	"encoding/json"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

var jsonCompatible = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	//constant for adformat
	Banner = "banner"
	Video  = "video"
	Native = "native"
)

func GetCreativeTypeFromCreative(adm string) string {
	if adm == "" {
		return ""
	}
	if IsVideo(adm) {
		return Video
	}
	if IsNative(adm) {
		return Native
	}

	return Banner
}

func IsVideo(adm string) bool {
	return strings.HasSuffix(strings.TrimSpace(adm), "</VAST>")
}

func IsNative(adm string) bool {
	var temp map[string]json.RawMessage

	if err := jsonCompatible.UnmarshalFromString(adm, &temp); err != nil {
		return false
	}

	for _, tag := range []string{"native", "link", "assets"} {
		if _, exists := temp[tag]; exists {
			return true
		}
	}

	return false
}
