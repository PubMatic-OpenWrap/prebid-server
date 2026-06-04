package sdkutils

import (
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
)

// DeviceExtSignalKeys lists LTV-related device.ext keys copied from OW SDK signal into the bid request.
// device.ext.ifv is merged via CopyIFV inside MergeDeviceExtFromSignal.
// device.ext.atts is not merged here: only integrations that historically copied it (e.g. AppLovin Max, Unity, APS) call CopyPath separately.
var DeviceExtSignalKeys = []string{
	"boottime",
	"pbtime",
	"diskspace",
	"totaldisk",
	"inputlanguage",
	"charging",
	"batterylevel",
	"totalmem",
	"dnh",
	"ringmute",
	"darkmode",
	"bluetooth",
	"airplane",
	"dnd",
	"headset",
	"screenbright",
}

// AppExtSignalKeys lists app.ext keys copied from OW SDK signal into the bid request.
var AppExtSignalKeys = []string{
	"install_time",
	"first_launch_time",
}

// CopyExtKeys copies the given top-level keys from source JSON into target JSON when present in source.
func CopyExtKeys(source, target []byte, keys ...string) []byte {
	newTarget := target
	if len(keys) > 0 && len(newTarget) == 0 {
		newTarget = []byte(`{}`)
	}

	for _, key := range keys {
		field, dataType, _, err := jsonparser.Get(source, key)
		if err != nil {
			continue
		}

		if dataType == jsonparser.String {
			quotedStr := strconv.Quote(string(field))
			field = []byte(quotedStr)
		}

		newTarget, err = jsonparser.Set(newTarget, field, key)
		if err != nil {
			return target
		}
	}

	if len(newTarget) == 2 {
		return target
	}
	return newTarget
}

// MergeDeviceExtFromSignal copies OW SDK signal device.ext LTV keys into the request and merges ifv.
// Callers that should copy device.ext.atts from the signal (not Google SDK) must use CopyPath separately.
func MergeDeviceExtFromSignal(source, target []byte) []byte {
	target = CopyExtKeys(source, target, DeviceExtSignalKeys...)
	return CopyIFV(source, target)
}

// MergeAppExtFromSignal copies OW SDK signal app.ext params into the request.
func MergeAppExtFromSignal(source, target []byte) []byte {
	return CopyExtKeys(source, target, AppExtSignalKeys...)
}

// MergeImpLTVFieldsFromSignal copies imp.rwdd from the OW SDK signal into the request impression.
// imp.banner.mimes are merged in each integration's modifyBanner (or AppLovin equivalent), not here, to avoid duplicate work.
func MergeImpLTVFieldsFromSignal(dst, src *openrtb2.Imp) {
	if dst == nil || src == nil {
		return
	}

	if src.Rwdd != 0 {
		dst.Rwdd = src.Rwdd
	}
}
