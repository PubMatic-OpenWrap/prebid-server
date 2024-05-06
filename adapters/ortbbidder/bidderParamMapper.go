package ortbbidder

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	properties = "properties"
	dataType   = "type"
	location   = "location"
)

// JSONNode alias for Generic datatype of json object represented by map
type JSONNode = map[string]interface{}

// Mapper struct holds mappings for bidder parameters and bid responses.
type Mapper struct {
	bidderParamMapper bidderParamMapper
	// bidResponseMapper bidderParamMapper // TODO
}

// bidderParamMapper maps bidder-names to their bidder-params and its details like location, type etc
type bidderParamMapper map[string]map[string]paramDetails

type paramDetails struct {
	location string
}

var mapper *Mapper
var once sync.Once
var mapperErr error

// InitMapper initializes a Mapper instance using files in a given directory.
func InitMapper(dirPath string) (*Mapper, error) {
	once.Do(func() {
		mapper, mapperErr = prepareMapperFromFiles(dirPath)
	})
	return mapper, mapperErr
}

// setBidderParam adds or updates a bidder parameter in the mapper for given bidderName.
func (b bidderParamMapper) setBidderParam(bidderName string, paramName string, paramValue paramDetails) {
	params, ok := b[bidderName]
	if !ok {
		params = make(map[string]paramDetails)
	}
	params[paramName] = paramValue
	b[bidderName] = params
}

// prepareMapperFromFiles creates a Mapper from JSON files in the specified directory.
func prepareMapperFromFiles(dirPath string) (*Mapper, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error:[%s] dirPath:[%s]", err.Error(), dirPath)
	}

	mapper := &Mapper{bidderParamMapper: make(bidderParamMapper)}
	for _, file := range files {
		bidderName, ok := strings.CutSuffix(file.Name(), ".json")
		if !ok {
			return nil, fmt.Errorf("error:[invalid_json_file_name] filename:[%s]", file.Name())
		}
		if !isORTBBidder(bidderName) {
			continue
		}
		filePath := filepath.Join(dirPath, file.Name())
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		var fileContentsMap map[string]interface{}
		err = json.Unmarshal(fileBytes, &fileContentsMap)
		if err != nil {
			return nil, err
		}
		mapper.bidderParamMapper, err = updateBidderParamsMapper(mapper.bidderParamMapper, fileContentsMap, bidderName)
		if err != nil {
			return nil, err
		}
		// add code to build bidResponseMapper below
	}
	return mapper, nil
}

// updateBidderParamsMapper adds the details like location based on the file content.
func updateBidderParamsMapper(mapper bidderParamMapper, fileContentsMap JSONNode, bidderName string) (bidderParamMapper, error) {
	properties, found := fileContentsMap[properties]
	if !found {
		return mapper, nil
	}
	propertiesMap, ok := properties.(JSONNode)
	if !ok {
		return mapper, fmt.Errorf("error:[invalid_json_file_content_malformed_properties] bidderName:[%s]", bidderName)
	}

	for bidderParamName, bidderParamProperty := range propertiesMap {
		property, ok := bidderParamProperty.(JSONNode)
		if !ok {
			return mapper, fmt.Errorf("error:[invalid_json_file_content] bidder:[%s] bidderParam:[%s]", bidderName, bidderParamName)
		}
		location, found := property[location]
		if !found {
			continue // if location is absent then mapper will set it to default location.
		}
		locationStr, ok := location.(string)
		if !ok {
			return mapper, fmt.Errorf("error:[incorrect_location_in_bidderparam] bidder:[%s] bidderParam:[%s]", bidderName, bidderParamName)
		}
		if !strings.HasPrefix(locationStr, "req.") {
			return mapper, fmt.Errorf("error:[incorrect_location_in_bidderparam] bidder:[%s] bidderParam:[%s]", bidderName, bidderParamName)
		}
		mapper.setBidderParam(bidderName, bidderParamName, paramDetails{location: locationStr})
	}
	return mapper, nil
}

/*
setValueAtLocation updates or creates a value in a JSONNode based on a specified location.
The location is a string that specifies a path through the node hierarchy,
separated by dots ('.'). The value can be any type, and the function will
create intermediate nodes as necessary if they do not exist.

Arguments:
- node: The root JSONNode where the value will be set.
- location: A dot-separated string indicating the path to set the value.
- value: The value to set at the specified location. Can be of any type.

Example:
  - location = imp.ext.adunitid; value = 123  ==> {"imp": {"ext" : {"adunitid":123}}}
*/
func setValueAtLocation(node JSONNode, location string, value any) bool {
	if value == nil || len(location) == 0 {
		return false
	}

	parts := strings.Split(location, ".")
	lastPartIndex := len(parts) - 1
	currentNode := node

	for i, part := range parts {
		if len(part) == 0 { // If location part is empty string
			return false
		}
		if i == lastPartIndex { // If it's the last part, set the value
			currentNode[part] = value
			break
		}
		// Not the last part, navigate deeper
		if nextNode, ok := currentNode[part]; ok {
			// Ensure the next node is a JSONNode
			if nextNodeTyped, ok := nextNode.(JSONNode); ok {
				currentNode = nextNodeTyped
			} else {
				return false // Existing node is not a JSONNode, cannot navigate deeper
			}
		} else {
			// Key does not exist, create a new node
			newNode := make(JSONNode)
			currentNode[part] = newNode
			currentNode = newNode
		}
	}

	return true
}

func mapBidderParamsInRequest(requestBody []byte, mapper map[string]paramDetails) ([]byte, error) {
	if len(mapper) == 0 {
		// mapper would be empty if oRTB bidder does not contain any bidder-params
		return requestBody, nil
	}
	requestBodyMap := JSONNode{}
	err := json.Unmarshal(requestBody, &requestBodyMap)
	if err != nil {
		return nil, err
	}
	impList, ok := requestBodyMap["imp"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("error:[invalid_imp_found_in_requestbody], imp:[%v]", requestBodyMap["imp"])
	}
	updatedRequestBody := false
	for ind, eachImp := range impList {
		requestBodyMap["imp"] = eachImp
		imp, ok := eachImp.(JSONNode)
		if !ok {
			return nil, fmt.Errorf("error:[invalid_imp_found_in_implist], imp:[%v]", requestBodyMap["imp"])
		}
		ext, ok := imp["ext"].(JSONNode)
		if !ok {
			continue
		}
		bidderParams, ok := ext["bidder"].(JSONNode)
		if !ok {
			continue
		}
		for paramName, paramValue := range bidderParams {
			details, ok := mapper[paramName]
			if !ok {
				continue
			}
			location, found := strings.CutPrefix(details.location, "req.")
			if !found {
				return nil, fmt.Errorf("error:[invalid_bidder_param_location] param:[%s] location:[%s]", paramName, details.location)
			}
			// TODO: handle app/site
			// TODO: delete request level bidder-param
			if setValueAtLocation(requestBodyMap, location, paramValue) {
				delete(bidderParams, paramName)
				updatedRequestBody = true
			}
		}
		impList[ind] = requestBodyMap["imp"]
	}
	requestBodyMap["imp"] = impList
	if updatedRequestBody {
		requestBody, err = json.Marshal(requestBodyMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request %s", err.Error())
		}
	}
	return requestBody, nil
}
