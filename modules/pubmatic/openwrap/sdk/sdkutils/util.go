package sdkutils

import "github.com/buger/jsonparser"

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
