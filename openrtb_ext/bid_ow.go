package openrtb_ext

import (
	"encoding/json"
	"regexp"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

var VideoRegex *regexp.Regexp

func init() {
	VideoRegex, _ = regexp.Compile(`<VAST\s*`)
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
	trimmedAdm := strings.TrimSpace(adm)
	if strings.HasSuffix(trimmedAdm, "</VAST>") {
		return Video
	}
	if IsNative(adm) {
		return Native
	}

	return Banner
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
