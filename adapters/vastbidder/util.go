package vastbidder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func ObjectArrayToString(len int, separator string, cb func(i int) string) string {
	if 0 == len {
		return ""
	}

	var out bytes.Buffer
	for i := 0; i < len; i++ {
		if out.Len() > 0 {
			out.WriteString(separator)
		}
		out.WriteString(cb(i))
	}
	return out.String()
}

func readImpExt(impExt json.RawMessage) (*openrtb_ext.ExtImpVASTBidder, error) {
	var bidderExt adapters.ExtImpBidder
	if err := json.Unmarshal(impExt, &bidderExt); err != nil {
		return nil, err
	}

	vastBidderExt := openrtb_ext.ExtImpVASTBidder{}
	if err := json.Unmarshal(bidderExt.Bidder, &vastBidderExt); err != nil {
		return nil, err
	}
	return &vastBidderExt, nil
}

func normalizeObject(prefix string, out map[string]string, obj map[string]interface{}) {
	for k, value := range obj {
		key := k
		if len(prefix) > 0 {
			key = prefix + "." + k
		}

		switch val := value.(type) {
		case string:
			out[key] = val
		case []interface{}: //array
			continue
		case map[string]interface{}: //object
			normalizeObject(key, out, val)
		default: //all int, float
			out[key] = fmt.Sprint(value)
		}
	}
}

func NormalizeJSON(obj map[string]interface{}) map[string]string {
	out := map[string]string{}
	normalizeObject("", out, obj)
	return out
}

var GetRandomID = func() string {
	return strconv.FormatInt(rand.Int63(), intBase)
}

func getJsonString(kvmap any) string {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)

	// Disable HTML escaping for special characters
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(kvmap); err != nil {
		return ""
	}
	return strings.TrimRight(buf.String(), "\n")

}

func isMap(data any) bool {
	return reflect.TypeOf(data).Kind() == reflect.Map
}

// getValueFromKV help to get value from nested  map
func (tag *BidderMacro) getValueFromKV(keyOrder []string) any {
	if tag.KV == nil {
		return ""
	}
	dataMap := tag.KV
	for _, key := range keyOrder {
		value, keyExists := dataMap[key]
		if !keyExists {
			return ""
		}

		if nestedMap, isMap := value.(map[string]any); isMap {
			dataMap = nestedMap
		} else {
			return value
		}
	}
	return dataMap
}

// mapToQuery convert the map data into & seperated string
func mapToQuery(m map[string]any) string {
	values := url.Values{}
	for key, value := range m {
		switch reflect.TypeOf(value).Kind() {
		case reflect.Map:
			mvalue, ok := value.(map[string]any)
			if ok {
				values.Add(key, mapToQuery(mvalue))
			}
		default:
			values.Add(key, fmt.Sprintf("%v", value))
		}
	}
	return values.Encode()
}
