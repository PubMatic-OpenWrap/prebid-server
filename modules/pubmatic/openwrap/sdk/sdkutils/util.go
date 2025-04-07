package sdkutils

import "github.com/buger/jsonparser"

func CopyPath(source []byte, target []byte, path ...string) ([]byte, error) {
	value, dataType, _, err := jsonparser.Get(source, path...)
	if err != nil {
		return nil, err
	}

	switch dataType {
	case jsonparser.Null:
		return target, nil
	case jsonparser.String:
		if len(value) == 0 || string(value) == `""` {
			return target, nil
		}
	case jsonparser.Array, jsonparser.Object:
		if len(value) <= 2 { // "[]" or "{}" are 2 chars
			return target, nil
		}
	}

	return jsonparser.Set(target, value, path...)
}
