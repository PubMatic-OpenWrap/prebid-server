package util

import (
	"strconv"
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

const (
	impKey        = "imp"
	extKey        = "ext"
	bidderKey     = "bidder"
	appsiteKey    = "appsite"
	siteKey       = "site"
	appKey        = "app"
	owOrtbPrefix  = "owortb_"
	locationMacro = "#"
)

/*
setValue updates or creates a value in a node based on a specified location.
The location is a string that specifies a path through the node hierarchy,
separated by dots ('.'). The value can be any type, and the function will
create intermediate nodes as necessary if they do not exist.

Arguments:
- node: the root of the map in which to set the value
- locations: slice of strings indicating the path to set the value.
- value: The value to set at the specified location. Can be of any type.

Example:
  - location = imp.ext.adunitid; value = 123  ==> {"imp": {"ext" : {"adunitid":123}}}
*/
/*
setValue updates or creates a value in a node based on a specified location.
The location is a string that specifies a path through the node hierarchy,
separated by dots ('.'). The value can be any type, and the function will
create intermediate nodes as necessary if they do not exist.

Arguments:
- node: the root of the map in which to set the value
- location: slice of strings indicating the path to set the value.
- value: The value to set at the specified location. Can be of any type.

Example:
  - location = imp.0.ext.adunitid; value = 123  ==> {"imp": {"ext" : {"adunitid":123}}}
*/
func SetValue(node map[string]any, location []string, value any) bool {
	if node == nil || value == nil {
		return false
	}
	var nextNode any = node
	lastNodeIndex := len(location) - 1
	for locIndex, loc := range location {
		if len(loc) == 0 {
			// if location is invalid
			return false
		}
		switch nextNodeTyped := nextNode.(type) {
		case map[string]any:
			if locIndex == lastNodeIndex {
				// set value at last index
				nextNodeTyped[loc] = value
				return true
			}
			nextNode = getNode(nextNodeTyped, loc)
			if nextNode == nil {
				// create a new node if the next node does not exist
				newNode := make(map[string]any)
				nextNodeTyped[loc] = newNode
				nextNode = newNode
			}
		case []any:
			// extract the array nodeIndex from the location to determine where to set the value
			nodeIndex, err := strconv.Atoi(loc)
			if err != nil || nodeIndex < 0 || nodeIndex >= len(nextNodeTyped) {
				return false
			}
			if locIndex == lastNodeIndex {
				nextNodeTyped[nodeIndex] = value
				return true
			}
			nextNode = nextNodeTyped[nodeIndex]
			if nextNode == nil {
				// create a new node if the next node does not exist
				newNode := make(map[string]any)
				nextNodeTyped[nodeIndex] = newNode
				nextNode = newNode
			}
		default:
			return false
		}
	}
	return false
}

// getNode retrieves the value for a given key from a map with special handling for the "appsite", "imp" key
func getNode(requestNode map[string]any, key string) any {
	switch key {
	case appsiteKey:
		// if key is "appsite" and if nodes contains "site" object then return nodes["site"] else return nodes["app"]
		if value, ok := requestNode[siteKey]; ok {
			return value
		}
		return requestNode[appKey]
	}
	return requestNode[key]
}

// getValueFromLocation retrieves a value from a map based on a specified location.
// getValueFromLocation retrieves a value from a map based on a specified location.
func GetValueFromLocation(val interface{}, path string) (interface{}, bool) {
	location := strings.Split(path, ".")
	var (
		ok   bool
		next interface{} = val
	)
	for _, loc := range location {
		switch nxt := next.(type) {
		case map[string]interface{}:
			next, ok = nxt[loc]
			if !ok {
				return nil, false
			}
		case []interface{}:
			index, err := strconv.Atoi(loc)
			if err != nil {
				return nil, false
			}
			if index < 0 || index >= len(nxt) {
				return nil, false
			}
			next = nxt[index]
		default:
			return nil, false
		}
	}
	return next, true
}

func ReplaceLocationMacro(path string, array []int) string {
	parts := strings.Split(path, ".")
	j := 0
	for i, part := range parts {
		if part == locationMacro {
			if j >= len(array) {
				break
			}
			parts[i] = strconv.Itoa(array[j])
			j++
		}
	}
	return strings.Join(parts, ".")
}

// GetMediaType returns the bidType from the MarkupType field
func GetMediaType(mtype openrtb2.MarkupType) openrtb_ext.BidType { // change name
	var bidType openrtb_ext.BidType
	switch mtype {
	case openrtb2.MarkupBanner:
		bidType = openrtb_ext.BidTypeBanner
	case openrtb2.MarkupVideo:
		bidType = openrtb_ext.BidTypeVideo
	case openrtb2.MarkupAudio:
		bidType = openrtb_ext.BidTypeAudio
	case openrtb2.MarkupNative:
		bidType = openrtb_ext.BidTypeNative
	}
	return bidType
}

// IsORTBBidder returns true if the bidder is an oRTB bidder
func IsORTBBidder(bidderName string) bool {
	return strings.HasPrefix(bidderName, "owortb_")
}
