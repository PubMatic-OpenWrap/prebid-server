package util

import (
	"strconv"
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

const (
	impKey     = "imp"
	extKey     = "ext"
	bidderKey  = "bidder"
	appsiteKey = "appsite"
	siteKey    = "site"
	appKey     = "app"
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
func SetValue(node map[string]any, locations []string, value any) bool {
	if value == nil || len(locations) == 0 {
		return false
	}

	lastNodeIndex := len(locations) - 1
	currentNode := node

	for index, loc := range locations {
		if len(loc) == 0 { // if location part is empty string
			return false
		}
		if index == lastNodeIndex { // if it's the last part in location, set the value
			currentNode[loc] = value
			break
		}
		nextNode := getNode(currentNode, loc)
		// not the last part, navigate deeper
		if nextNode == nil {
			// loc does not exist, set currentNode to a new node
			newNode := make(map[string]any)
			currentNode[loc] = newNode
			currentNode = newNode
			continue
		}
		// loc exists, set currentNode to nextNode
		nextNodeTyped, ok := nextNode.(map[string]any)
		if !ok {
			return false
		}
		currentNode = nextNodeTyped
	}
	return true
}

// getNode retrieves the value for a given key from a map with special handling for the "appsite" key
func getNode(nodes map[string]any, key string) any {
	switch key {
	case appsiteKey:
		// if key is "appsite" and if nodes contains "site" object then return nodes["site"] else return nodes["app"]
		if value, ok := nodes[siteKey]; ok {
			return value
		}
		return nodes[appKey]
	}
	return nodes[key]
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

func setValueAtLocation(node map[string]interface{}, path string, value interface{}) bool {
	location := strings.Split(path, ".")
	var (
		ok   bool
		next interface{} = node
	)
	lastIndex := len(location) - 1
	for i, loc := range location {
		switch nxt := next.(type) {
		case map[string]interface{}:
			if i == lastIndex {
				nxt[loc] = value
				return true
			}
			next, ok = nxt[loc]
			if !ok {
				return false
			}
		case []interface{}:
			index, err := strconv.Atoi(loc)
			if err != nil || index < 0 || index >= len(nxt) {
				return false
			}
			if i == lastIndex {
				nxt[index] = value
				return true
			}
			next = nxt[index]
		default:
			return false
		}
	}
	return false
}

func GetPath(path string, array []int) string {
	parts := strings.Split(path, ".")
	j := 0
	for i, part := range parts {
		if part == "#" {
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
