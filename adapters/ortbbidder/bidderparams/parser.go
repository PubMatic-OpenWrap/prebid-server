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
// func LoadBidderConfig(requestParamsDirPath, responseParamsDirPath string, isBidderAllowed func(string) bool) (*BidderConfig, error) {
// 	bidderConfigMap := &BidderConfig{bidderConfigMap: make(map[string]*config)}
// 	files, err := os.ReadDir(requestParamsDirPath)
// 	if err != nil {
// 		return nil, fmt.Errorf("error:[%s] dirPath:[%s]", err.Error(), requestParamsDirPath)
// 	}

// 	for _, file := range files {
// 		bidderName, ok := strings.CutSuffix(file.Name(), ".json")
// 		if !ok {
// 			return nil, fmt.Errorf("error:[invalid_json_file_name] filename:[%s]", file.Name())
// 		}
// 		if !isBidderAllowed(bidderName) {
// 			continue
// 		}
// 		requestParamsConfig, err := readFile(requestParamsDirPath, file.Name())
// 		if err != nil {
// 			return nil, fmt.Errorf("error:[fail_to_read_file] dir:[%s] filename:[%s] err:[%s]", requestParamsDirPath, file.Name(), err.Error())
// 		}
// 		requestParams, err := prepareRequestParams(bidderName, requestParamsConfig)
// 		if err != nil {
// 			return nil, err
// 		}
// 		bidderConfigMap.setRequestParams(bidderName, requestParams)
// 	}

// 	files, err = os.ReadDir(responseParamsDirPath)
// 	if err != nil {
// 		return nil, fmt.Errorf("error:[%s] dirPath:[%s]", err.Error(), responseParamsDirPath)
// 	}
// 	for _, file := range files {
// 		bidderName, ok := strings.CutSuffix(file.Name(), ".json")
// 		if !ok {
// 			return nil, fmt.Errorf("error:[invalid_json_file_name] filename:[%s]", file.Name())
// 		}
// 		if !isBidderAllowed(bidderName) {
// 			continue
// 		}
// 		responseParamsConfig, err := readFile(responseParamsDirPath, file.Name())
// 		if err != nil {
// 			return nil, fmt.Errorf("error:[fail_to_read_file] dir:[%s] filename:[%s] err:[%s]", responseParamsDirPath, file.Name(), err.Error())
// 		}
// 		requestParams, err := prepareRequestParams(bidderName, responseParamsConfig)
// 		if err != nil {
// 			return nil, err
// 		}
// 		bidderConfigMap.setResponseParams(bidderName, requestParams)
// 	}

// 	return bidderConfigMap, nil
// }

// LoadBidderConfig creates a bidderConfig from JSON files specified in dirPath directory.
func LoadBidderConfig(requestParamsDirPath, responseParamsDirPath string, isBidderAllowed func(string) bool) (*BidderConfig, error) {
	bidderConfigMap := &BidderConfig{bidderConfigMap: make(map[string]*config)}

	err := processParams(requestParamsDirPath, bidderConfigMap.setRequestParams, isBidderAllowed)
	if err != nil {
		return nil, err
	}

	err = processParams(responseParamsDirPath, bidderConfigMap.setResponseParams, isBidderAllowed)
	if err != nil {
		return nil, err
	}

	return bidderConfigMap, nil
}

func processParams(paramsDirPath string, setParams func(string, map[string]BidderParamMapper), isBidderAllowed func(string) bool) error {
	files, err := os.ReadDir(paramsDirPath)
	if err != nil {
		return fmt.Errorf("error:[%s] dirPath:[%s]", err.Error(), paramsDirPath)
	}

	for _, file := range files {
		bidderName, ok := strings.CutSuffix(file.Name(), ".json")
		if !ok {
			return fmt.Errorf("error:[invalid_json_file_name] filename:[%s]", file.Name())
		}
		if !isBidderAllowed(bidderName) {
			continue
		}
		paramsConfig, err := readFile(paramsDirPath, file.Name())
		if err != nil {
			return fmt.Errorf("error:[fail_to_read_file] dir:[%s] filename:[%s] err:[%s]", paramsDirPath, file.Name(), err.Error())
		}
		params, err := prepareParams(bidderName, paramsConfig)
		if err != nil {
			return err
		}
		setParams(bidderName, params)
	}

	return nil
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

// prepareParams parse the requestParamsConfig and returns the requestParams
func prepareParams(bidderName string, paramsConfig map[string]any) (map[string]BidderParamMapper, error) {
	params, found := paramsConfig[propertiesKey]
	if !found {
		return nil, nil
	}
	paramsMap, ok := params.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("error:[invalid_json_file_content_malformed_properties] bidderName:[%s]", bidderName)
	}
	paramsCfg := make(map[string]BidderParamMapper, len(paramsMap))
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
		paramsCfg[paramName] = BidderParamMapper{
			location: strings.Split(locationStr, "."),
			path:     locationStr,
		}
	}
	return paramsCfg, nil
}
