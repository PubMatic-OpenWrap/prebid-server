package eds

import (
	"bytes"
	"encoding/json"

	"github.com/buger/jsonparser"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

var (
	edsBidderDevicePath = []string{models.PubmaticBidderKey, models.ExtEDSKey, models.EDSDeviceKey}
	edsBidderAppPath    = []string{models.PubmaticBidderKey, models.ExtEDSKey, models.EDSAppKey}
	edsBidderPath       = []string{models.PubmaticBidderKey, models.ExtEDSKey}
)

func StripEDSTier1ParamsForBlockedCountry(bidderParams json.RawMessage) json.RawMessage {
	if len(bidderParams) == 0 {
		return bidderParams
	}

	modified := stripEDSParamsAtPath(bidderParams, edsBidderDevicePath, models.DeviceEDSTier1BlockedParams)
	modified = stripEDSParamsAtPath(modified, edsBidderAppPath, models.AppEDSTier1BlockedParams)

	if edsData, _, _, err := jsonparser.Get(modified, edsBidderPath...); err == nil && isEmptyJSONObject(edsData) {
		modified = jsonparser.Delete(modified, edsBidderPath...)
	}

	if isEmptyJSONObject(modified) {
		return nil
	}

	return modified
}

func stripEDSParamsAtPath(ext []byte, path []string, keys []string) []byte {
	if len(ext) == 0 || len(keys) == 0 || len(path) == 0 {
		return ext
	}

	objData, dataType, _, err := jsonparser.Get(ext, path...)
	if err != nil || dataType != jsonparser.Object {
		return ext
	}

	modified := ext
	changed := false
	for _, key := range keys {
		if _, _, _, err := jsonparser.Get(objData, key); err != nil {
			continue
		}
		keyPath := append(append([]string{}, path...), key)
		updated := jsonparser.Delete(modified, keyPath...)
		if !bytes.Equal(updated, modified) {
			modified = updated
			changed = true
		}
	}

	if !changed {
		return ext
	}

	if objData, _, _, err = jsonparser.Get(modified, path...); err == nil && isEmptyJSONObject(objData) {
		modified = jsonparser.Delete(modified, path...)
	}

	return modified
}

func isEmptyJSONObject(data []byte) bool {
	return len(data) == 2 && data[0] == '{' && data[1] == '}'
}
