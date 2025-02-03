package openrtb_ext

import (
	"encoding/json"
	"regexp"

	jsoniter "github.com/json-iterator/go"
)

var videoRegex *regexp.Regexp

func init() {
	videoRegex, _ = regexp.Compile(`<VAST\s+`)
}

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

func IsNative(adm string) bool {
	var temp map[string]json.RawMessage

	if err := jsonCompatible.UnmarshalFromString(adm, &temp); err == nil {
		if _, exists := temp["native"]; exists {
			return true
		}
		if _, exists := temp["link"]; exists {
			return true
		}
		if _, exists := temp["assets"]; exists {
			return true
		}
	}
	return false
}

func IsVideo(adm string) bool {
	return videoRegex.MatchString(adm)
}
