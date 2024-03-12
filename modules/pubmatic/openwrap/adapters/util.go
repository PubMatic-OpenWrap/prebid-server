package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

const FloatValuePrecision = 2

func getKeywordStringForPartner(impExt *models.ImpExtension, partner string) string {
	if impExt != nil && impExt.Bidder != nil {
		bidder := impExt.Bidder[partner]
		if nil != bidder && len(bidder.KeyWords) > 0 {
			if byts, err := json.Marshal(bidder.KeyWords); err == nil {
				return string(byts)
			}
		}
	}
	return ""
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func trimComma(buf *bytes.Buffer) {
	b := buf.Bytes()
	if len(b) > 0 && b[len(b)-1] == ',' {
		b[len(b)-1] = ' '
	}
}

func isDirectoryExists(location string) bool {
	if _, err := os.Stat(location); err == nil {
		// path to schemaDirectory exists
		return true
	}
	return false
}

// ----------- datatype utilities ----------
func getInt(val interface{}) (int, bool) {
	if val == nil {
		return 0, false
	}

	var result int
	switch v := val.(type) {
	case int:
		result = v
	case string:
		iVal, err := strconv.Atoi(v)
		if err != nil {
			return 0, false
		}
		result = iVal
	case float64:
		result = int(v)
	case float32:
		result = int(v)
	default:
		iVal, err := strconv.Atoi(fmt.Sprint(v))
		if err != nil {
			return 0, false
		}
		result = iVal
	}
	return result, true
}

func getFloat64(val interface{}) (float64, bool) {
	if val == nil {
		return 0, false
	}

	var result float64
	switch v := val.(type) {
	case float64:
		result = v
	case string:
		fVal, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, false
		}
		result = fVal
	case int:
		result = float64(v)
	default:
		fVal, err := strconv.ParseFloat(fmt.Sprint(v), 64)
		if err != nil {
			return 0, false
		}
		result = fVal
	}
	return result, true
}

func getString(val interface{}) (string, bool) {
	if val == nil {
		return "", false
	}

	var result string
	switch v := val.(type) {
	case string:
		result = v
	case int:
		result = strconv.Itoa(v)
	case map[string]interface{}:
		val, err := json.Marshal(v)
		if err != nil {
			return "", false
		}
		result = string(val)
	default:
		result = fmt.Sprint(val)
	}

	return result, true
}

func getBool(val interface{}) (bool, bool) {
	if val == nil {
		return false, false
	}

	var result bool
	switch v := val.(type) {
	case bool:
		result = v
	case string:
		bVal, err := strconv.ParseBool(v)
		if err != nil {
			return false, false
		}
		result = bVal
	default:
		bVal, err := strconv.ParseBool(fmt.Sprint(v))
		if err != nil {
			return false, false
		}
		result = bVal
	}

	return result, true
}

func getIntArray(val interface{}) ([]int, bool) {
	if val == nil {
		return nil, false
	}

	valArray, ok := val.([]interface{})
	if !ok {
		return nil, false
	}

	result := make([]int, 0)
	for _, x := range valArray {
		if val, ok := getInt(x); ok {
			result = append(result, val)
		}
	}

	return result, true
}
