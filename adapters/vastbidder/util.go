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
	"time"

	"github.com/prebid/prebid-server/v3/adapters"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

func arrayToString(len int, separator string, cb func(i int) string) string {
	if len == 0 {
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

func normalizeJSON(obj map[string]interface{}) map[string]string {
	out := map[string]string{}
	normalizeObject("", out, obj)
	return out
}

var generateRandomID = func() string {
	return strconv.FormatInt(rand.Int63(), intBase)
}

func getJSONString(kvmap any) string {

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

// extractDataFromMap help to get value from nested  map
func getValueFromMap(lookUpOrder []string, m map[string]any) any {
	if len(lookUpOrder) == 0 {
		return ""
	}

	for _, key := range lookUpOrder {
		value, keyExists := m[key]
		if !keyExists {
			return ""
		}
		if nestedMap, isMap := value.(map[string]any); isMap {
			m = nestedMap
		} else {
			return value
		}
	}
	return m
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
			v := fmt.Sprintf("%v", value)
			decodedString, err := url.QueryUnescape(v)
			if err == nil {
				v = decodedString
			}
			values.Add(key, v)
		}
	}
	return values.Encode()
}

// parseVASTVersion convert vast version string to float64
func parseVASTVersion(version string) (float64, error) {
	if version == "" {
		//return default value
		return 2.0, nil
	}
	value, err := strconv.ParseFloat(version, 64)
	if err != nil || value < 0 {
		return 0, errInvalidVASTVersion
	}
	return value, nil
}

// parseDuration extracts the duration of the bid from input creative of Linear type.
// The lookup may vary from vast version provided in the input
// returns duration in seconds or error if failed to obtained the duration.
// If multple Linear tags are present, onlyfirst one will be used
//
// It will lookup for duration only in case of creative type is Linear.
// If creative type other than Linear then this function will return error
// For Linear Creative it will lookup for Duration attribute.Duration value will be in hh:mm:ss.mmm format as per VAST specifications
// If Duration attribute not present this will return error
//
// # After extracing the duration it will convert it into seconds
//
// The ad server uses the <Duration> element to denote
// the intended playback duration for the video or audio component of the ad.
// Time value may be in the format HH:MM:SS.mmm where .mmm indicates milliseconds.
// Providing milliseconds is optional.
//
// Reference
// 1.https://iabtechlab.com/wp-content/uploads/2019/06/VAST_4.2_final_june26.pdf
// 2.https://iabtechlab.com/wp-content/uploads/2018/11/VAST4.1-final-Nov-8-2018.pdf
// 3.https://iabtechlab.com/wp-content/uploads/2016/05/VAST4.0_Updated_April_2016.pdf
// 4.https://iabtechlab.com/wp-content/uploads/2016/04/VASTv3_0.pdf
func parseDuration(duration string) (int, error) {
	// check if milliseconds is provided
	match := durationRegExp.FindStringSubmatch(duration)
	if nil == match {
		return 0, errInvalidVideoDuration
	}

	repl := "${1}h${2}m${3}s"
	ms := match[5]
	if len(ms) > 0 {
		repl += "${5}ms"
	}

	duration = durationRegExp.ReplaceAllString(duration, repl)
	dur, err := time.ParseDuration(duration)
	if err != nil {
		return 0, err
	}

	return int(dur.Seconds()), nil
}
