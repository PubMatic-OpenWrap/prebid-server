package bidderparams

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	propertiesKey = "properties"
	locationKey   = "location"
)

// LoadBidderConfig creates a bidderConfig from JSON files specified in dirPath directory.
func LoadBidderConfig(dirPath string, isBidderAllowed func(string) bool) (*BidderConfig, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error:[%s] dirPath:[%s]", err.Error(), dirPath)
	}
	bidderConfigMap := &BidderConfig{bidderConfigMap: make(map[string]*config)}
	for _, file := range files {
		bidderName, ok := strings.CutSuffix(file.Name(), ".json")
		if !ok {
			return nil, fmt.Errorf("error:[invalid_json_file_name] filename:[%s]", file.Name())
		}
		if !isBidderAllowed(bidderName) {
			continue
		}
		requestParamsConfig, err := readFile(dirPath, file.Name())
		if err != nil {
			return nil, fmt.Errorf("error:[fail_to_read_file] dir:[%s] filename:[%s] err:[%s]", dirPath, file.Name(), err.Error())
		}
		requestParams, err := loadRequestParams(bidderName, requestParamsConfig)
		if err != nil {
			return nil, err
		}
		bidderConfigMap.setRequestParams(bidderName, requestParams)
	}
	return bidderConfigMap, nil
}

// readFile reads the file from directory and unmarshals it into the map[string]any
func readFile(dirPath, file string) (map[string]any, error) {
	filePath := filepath.Join(dirPath, file)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var contentMap map[string]any
	err = json.Unmarshal(content, &contentMap)
	return contentMap, err
}

// loadRequestParams parse the requestParamsConfig and returns the requestParams
func loadRequestParams(bidderName string, requestParamsConfig map[string]any) (map[string]BidderParamMapper, error) {
	params, found := requestParamsConfig[propertiesKey]
	if !found {
		return nil, nil
	}
	paramsMap, ok := params.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("error:[invalid_json_file_content_malformed_properties] bidderName:[%s]", bidderName)
	}
	requestParams := make(map[string]BidderParamMapper, len(paramsMap))
	for paramName, paramValue := range paramsMap {
		paramValueMap, ok := paramValue.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("error:[invalid_json_file_content] bidder:[%s] bidderParam:[%s]", bidderName, paramName)
		}
		location, found := paramValueMap[locationKey]
		if !found {
			continue
		}
		locationStr, ok := location.(string)
		if !ok {
			return nil, fmt.Errorf("error:[incorrect_location_in_bidderparam] bidder:[%s] bidderParam:[%s]", bidderName, paramName)
		}
		requestParams[paramName] = BidderParamMapper{
			location: strings.Split(locationStr, "."),
		}
	}
	return requestParams, nil
}
