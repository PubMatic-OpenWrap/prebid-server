package ortbbidder

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	properties = "properties"
	dataType   = "type"
	location   = "location"
	impKey     = "imp"
	extKey     = "ext"
	bidderKey  = "bidder"
	reqExtPath = "req."
	appsiteKey = "appsite"
	siteKey    = "site"
	appKey     = "app"
)

// mapper struct holds mappings for bidder parameters and bid responses.
type mapper struct {
	bidderParamMapper bidderParamMapper
	// bidResponseMapper bidderParamMapper // TODO
}

// bidderParamMapper maps bidder-names to their bidder-params and its location
type bidderParamMapper map[string]map[string]paramDetails

// paramDetails contains details like bidder-param locations
type paramDetails struct {
	location []string
}

// global instance of Mapper
var g_mapper *mapper

// InitMapper initializes a mapper instance using files in a given directory.
func InitMapper(dirPath string) (err error) {
	g_mapper, err = prepareMapperFromFiles(dirPath)
	return err
}

// prepareMapperFromFiles creates a Mapper from JSON files in the specified directory.
func prepareMapperFromFiles(dirPath string) (*mapper, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error:[%s] dirPath:[%s]", err.Error(), dirPath)
	}

	mapper := &mapper{bidderParamMapper: make(bidderParamMapper)}
	for _, file := range files {
		bidderName, ok := strings.CutSuffix(file.Name(), ".json")
		if !ok {
			return nil, fmt.Errorf("error:[invalid_json_file_name] filename:[%s]", file.Name())
		}
		if !isORTBBidder(bidderName) {
			continue
		}
		fileContents, err := readFile(dirPath, file.Name())
		if err != nil {
			return nil, fmt.Errorf("error:[fail_to_read_file] dir:[%s] filename:[%s] err:[%s]", dirPath, file.Name(), err.Error())
		}
		err = mapper.bidderParamMapper.setBidderParamsDetails(bidderName, fileContents)
		if err != nil {
			return nil, err
		}
		// add code to build bidResponseMapper below
	}
	return mapper, nil
}

// readFile reads the file from directory and unmarshals it into the map[string]any
func readFile(dirPath, file string) (map[string]any, error) {
	filePath := filepath.Join(dirPath, file)
	fileContents, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var fileContentsNode map[string]any
	err = json.Unmarshal(fileContents, &fileContentsNode)
	return fileContentsNode, err
}

// setBidderParamsDetails sets the bidder-param details in bidderParamMapper based on file-content passed as map[string]any
func (bpm bidderParamMapper) setBidderParamsDetails(bidderName string, params map[string]any) error {
	properties, found := params[properties]
	if !found {
		return nil
	}
	propertiesMap, ok := properties.(map[string]any)
	if !ok {
		return fmt.Errorf("error:[invalid_json_file_content_malformed_properties] bidderName:[%s]", bidderName)
	}
	paramsDetails := make(map[string]paramDetails, len(propertiesMap))
	for bidderParamName, bidderParamProperty := range propertiesMap {
		property, ok := bidderParamProperty.(map[string]any)
		if !ok {
			return fmt.Errorf("error:[invalid_json_file_content] bidder:[%s] bidderParam:[%s]", bidderName, bidderParamName)
		}
		location, found := property[location]
		if !found {
			continue // if location is absent then bidder-param will remain at its default location.
		}
		locationStr, ok := location.(string)
		if !ok {
			return fmt.Errorf("error:[incorrect_location_in_bidderparam] bidder:[%s] bidderParam:[%s]", bidderName, bidderParamName)
		}
		paramsDetails[bidderParamName] = paramDetails{
			location: strings.Split(locationStr, "."),
		}
	}
	bpm[bidderName] = paramsDetails
	return nil
}

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
		// not the last part, navigate deeper
		nextNode, found := currentNode[loc]
		if !found {
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

// mapBidderParamsInRequest updates the requestBody based on the bidder-params mapping details.
func mapBidderParamsInRequest(requestBody []byte, bidderParamDetails map[string]paramDetails) ([]byte, error) {
	if len(bidderParamDetails) == 0 {
		return requestBody, nil // mapper would be empty if oRTB bidder does not contain any bidder-params
	}
	requestBodyMap := map[string]any{}
	err := json.Unmarshal(requestBody, &requestBodyMap)
	if err != nil {
		return nil, err
	}
	impList, ok := requestBodyMap[impKey].([]any)
	if !ok {
		return nil, fmt.Errorf("error:[invalid_imp_found_in_requestbody], imp:[%v]", requestBodyMap[impKey])
	}
	updatedRequestBody := false
	for ind, eachImp := range impList {
		requestBodyMap[impKey] = eachImp
		imp, ok := eachImp.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("error:[invalid_imp_found_in_implist], imp:[%v]", requestBodyMap[impKey])
		}
		ext, ok := imp[extKey].(map[string]any)
		if !ok {
			continue
		}
		bidderParams, ok := ext[bidderKey].(map[string]any)
		if !ok {
			continue
		}
		for paramName, paramValue := range bidderParams {
			details, ok := bidderParamDetails[paramName]
			if !ok {
				continue
			}
			details = applyConditionalMapping(requestBodyMap, details)
			// set the value in the requestBody according to the mapping details and remove the parameter if successful.
			if setValue(requestBodyMap, details.location, paramValue) {
				delete(bidderParams, paramName)
				updatedRequestBody = true
			}
			// TODO - what if failed to set the bidder-param
		}
		impList[ind] = requestBodyMap[impKey]
	}
	// update the impression list in the requestBody
	requestBodyMap[impKey] = impList
	// if the requestBody was modified, marshal it back to JSON.
	if updatedRequestBody {
		requestBody, err = json.Marshal(requestBodyMap)
		if err != nil {
			return nil, fmt.Errorf("error:[fail_to_update_request_body] msg:[%s]", err.Error())
		}
	}
	return requestBody, nil
}

// applyConditionalMapping applies the custom rules and updates the location of bidder-param
func applyConditionalMapping(requestBodyMap map[string]any, details paramDetails) paramDetails {
	if len(details.location) == 0 || len(requestBodyMap) == 0 {
		return details
	}
	// if location is "appsite" and if request contains "app" object then set location to "app" else set location to "site"
	// example - if req.site is present and location is {"appsite","publisher","id"} then update location to {"site","publisher","id"}
	if details.location[0] == appsiteKey {
		details.location[0] = siteKey
		if _, found := requestBodyMap[appKey]; found {
			details.location[0] = appKey
		}
	}
	return details
}
