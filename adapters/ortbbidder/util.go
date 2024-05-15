package ortbbidder

/*
setValue updates or creates a value in a node based on a specified location.
The location is a string that specifies a path through the node hierarchy,
separated by dots ('.'). The value can be any type, and the function will
create intermediate nodes as necessary if they do not exist.

Arguments:
- locations: slice of strings indicating the path to set the value.
- value: The value to set at the specified location. Can be of any type.

Example:
  - location = imp.ext.adunitid; value = 123  ==> {"imp": {"ext" : {"adunitid":123}}}
*/
func setValue(node map[string]any, locations []string, value any) bool {
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

// getNode applies custom rules and returns the appropriate node based on key
func getNode(nodes map[string]any, key string) any {
	switch key {
	case appsiteKey:
		// if location is "appsite" and if request contains "app" object then set location to "app" else set location to "site"
		// example - if req.site is present and location is {"appsite","publisher","id"} then update location to {"site","publisher","id"}
		if value, ok := nodes[siteKey]; ok {
			return value
		}
		return nodes[appKey]
	}
	return nodes[key]
}
