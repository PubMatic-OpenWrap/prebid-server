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

// Mapper struct holds mappings for bidder parameters and bid responses.
type Mapper struct {
	bidderParamMapper bidderParamMapper
	bidResponseMapper map[string]map[string]paramDetails
}

// bidderParamMapper maps bidder-names to their bidder-params and its details like location, type etc
type bidderParamMapper map[string]map[string]paramDetails

type paramDetails struct {
	dataType string
	location string
}

var mapper *Mapper
var once sync.Once
var err error

// NewMapper initializes a Mapper instance using files in a given directory.
func NewMapper(dirPath string) (*Mapper, error) {
	once.Do(func() {
		mapper, err = prepareMapperFromFiles(dirPath)
	})
	return mapper, err
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
		return nil, err
	}

	mapper := &Mapper{
		bidderParamMapper: make(map[string]map[string]paramDetails),
	}
	for _, file := range files {
		bidderName, ok := strings.CutSuffix(file.Name(), ".json")
		if !ok {
			return nil, fmt.Errorf("Err:[invalid_json_file_name] fileName:[%s]", file.Name())
		}
		if !isORTBBidder(bidderName) {
			continue
		}
		filePath := filepath.Join(dirPath, file.Name())
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		var fileBytesMap map[string]interface{}
		err = json.Unmarshal(fileBytes, &fileBytesMap)
		if err != nil {
			return nil, err
		}
		mapper.bidderParamMapper, err = updateBidderParamsMapper(mapper.bidderParamMapper, fileBytesMap, bidderName)
		if err != nil {
			return nil, err
		}
		// add code to build bidResponseMapper below
	}
	return mapper, nil
}

// updateBidderParamsMapper updates the bidder parameter mapping based on file content.
func updateBidderParamsMapper(mapper bidderParamMapper, fileBytesMap map[string]interface{}, bidderName string) (bidderParamMapper, error) {
	if fileBytesMap == nil {
		return mapper, nil
	}
	properties, found := fileBytesMap[properties]
	if !found {
		return mapper, nil
	}
	propertiesMap, ok := properties.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Err:[invalid_json_file_content_malformed_properties] bidderName:[%s]", bidderName)
	}

	for bidderParamName, bidderParamProperty := range propertiesMap {
		property, ok := bidderParamProperty.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Err:[invalid_json_file_content] bidder:[%s]", bidderName)
		}
		dataType, ok := property[dataType].(string)
		if !ok {
			return nil, fmt.Errorf("Err:[missing_type_in_bidderparam] bidder:[%s] bidderParam:[%s]", bidderName, bidderParamName)
		}
		location, found := property[location]
		if !found {
			continue
		}
		locationStr, ok := location.(string)
		if !ok {
			return nil, fmt.Errorf("Err:[incorrect_location_in_bidderparam] bidder:[%s] bidderParam:[%s]", bidderName, bidderParamName)
		}
		mapper.setBidderParam(bidderName, bidderParamName, paramDetails{dataType: dataType, location: locationStr})
	}
	return mapper, nil
}

// JSONType New Type Defined for JSON Object
type JSONType byte

const (
	//JSONObject will refer to Object type
	JSONObject JSONType = iota

	//JSONObjectArray will refer to ObjectArray type value
	JSONObjectArray

	//JSONString will refer to String type value
	JSONString
)

// Key Defines Special Object Type and Respective Name Mapping
type Key struct {
	Type JSONType
	Name string
}

// KeyMap is set of standard key map with their datatype which can be used to generate JSON object
var KeyMap map[string]*Key = map[string]*Key{
	//Standard Keys, No Need to Declare String and Object Parameters
	//"div":           &Key{Type: JSONString, Name: "div"},
	// "imp": &Key{Type: JSONObjectArray, Name: "imp"},
}

// JSONNode alias for Generic datatype of json object represented by map
type JSONNode = map[string]interface{}

// KeyMap["imp"] = &Key{Type: JSONObjectArray, Name: "imp"}

/*
SetValue function will recursively create nested object and set value
node: current JSONNode object
location: nested keys to create in node object (a.b.c)
value: value assigned to last key of location
example:

	location = a.b.c; value = 123  ==> {"a": {"b" : {"c":123}}}
*/
func SetValue(node JSONNode, location string, value any) bool {
	if value == nil || len(location) == 0 {
		return false
	}

	isLeaf := true
	keyStr := location
	index := strings.IndexByte(location, '.')

	if index != -1 {
		keyStr = location[0:index]
		isLeaf = false
	} else {
		index = len(location) - 1
	}

	key, ok := KeyMap[keyStr]
	if !ok {
		if isLeaf {
			key = &Key{Type: JSONString, Name: keyStr}
		} else {
			key = &Key{Type: JSONObject, Name: keyStr}
		}
	}

	switch key.Type {
	case JSONObject:
		locationNode, ok := node[key.Name]
		if !ok {
			newNode := make(JSONNode)
			node[key.Name] = newNode
			SetValue(newNode, location[index+1:], value)
		} else {
			node, ok := locationNode.(JSONNode)
			if ok {
				SetValue(node, location[index+1:], value)
			}
		}

	case JSONObjectArray:
		locationNode, ok := node[key.Name]
		if !ok {
			newNode := []JSONNode{JSONNode{}}
			node[key.Name] = newNode
			SetValue(newNode[0], location[index+1:], value)
		} else {
			nodes, ok := locationNode.([]interface{})
			if ok {
				for _, node := range nodes {
					SetValue(node.(JSONNode), location[index+1:], value)
				}
			}
		}

	case JSONString:
		node[key.Name] = value

	default:
		return false
	}
	return true
}
