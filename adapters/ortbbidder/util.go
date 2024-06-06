package ortbbidder

import (
	"strconv"

	"github.com/prebid/prebid-server/v2/errortypes"
)

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
func setValue(node map[string]any, location []string, value any) bool {
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

// newBadInputError returns the error of type bad-input
func newBadInputError(message string) error {
	return &errortypes.BadInput{
		Message: message,
	}
}
