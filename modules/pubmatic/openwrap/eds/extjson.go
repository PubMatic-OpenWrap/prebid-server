package eds

import (
	"github.com/buger/jsonparser"
)

// nestedObject returns the raw JSON value at key when it is a non-empty object.
func nestedObject(ext []byte, key string) []byte {
	if len(ext) == 0 {
		return nil
	}

	value, dataType, _, err := jsonparser.Get(ext, key)
	if err != nil || dataType != jsonparser.Object || len(value) <= 2 {
		return nil
	}

	return value
}

// mergeExtObjects merges top-level keys from overlay into base without full (de)serialization.
// When overlayWins is false, existing keys in base are preserved (gap-fill).
// Invalid or empty base objects are replaced by overlay, matching prior mergeExtJSON behavior.
func mergeExtObjects(base, overlay []byte, overlayWins bool) []byte {
	if len(overlay) == 0 {
		return base
	}
	if len(base) == 0 || isEmptyJSONObject(base) || !isJSONObject(base) {
		return overlay
	}

	result := base
	var mergeErr error
	_ = jsonparser.ObjectEach(overlay, func(key []byte, value []byte, _ jsonparser.ValueType, _ int) error {
		if !overlayWins {
			if _, _, _, err := jsonparser.Get(result, string(key)); err == nil {
				return nil
			}
		}

		var err error
		result, err = jsonparser.Set(result, value, string(key))
		if err != nil {
			mergeErr = err
		}
		return nil
	})

	if mergeErr != nil {
		if overlayWins {
			return overlay
		}
		return base
	}

	return result
}

func deleteExtKeysFromObject(ext []byte, keysSource []byte) []byte {
	if len(ext) == 0 || len(keysSource) == 0 {
		return nilIfEmptyExt(ext)
	}

	result := ext
	_ = jsonparser.ObjectEach(keysSource, func(key []byte, _ []byte, _ jsonparser.ValueType, _ int) error {
		result = jsonparser.Delete(result, string(key))
		return nil
	})

	return nilIfEmptyExt(result)
}

func isJSONObject(ext []byte) bool {
	return jsonparser.ObjectEach(ext, func(_ []byte, _ []byte, _ jsonparser.ValueType, _ int) error {
		return nil
	}) == nil
}

func isEmptyJSONObject(ext []byte) bool {
	if len(ext) == 0 {
		return true
	}

	empty := true
	_ = jsonparser.ObjectEach(ext, func(_ []byte, _ []byte, _ jsonparser.ValueType, _ int) error {
		empty = false
		return nil
	})

	return empty
}

func nilIfEmptyExt(ext []byte) []byte {
	if isEmptyJSONObject(ext) {
		return nil
	}
	return ext
}
