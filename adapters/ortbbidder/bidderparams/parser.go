package bidderparams

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/prebid/prebid-server/v2/util/jsonutil"
)

const (
	propertiesKey = "properties"
	locationKey   = "location"
)

// LoadBidderConfig creates a bidderConfig from JSON files specified in dirPath directory.
func LoadBidderConfig(requestParamsDirPath, responseParamsDirPath string, isBidderAllowed func(string) bool) (*BidderConfig, error) {
	bidderConfigMap := NewBidderConfig()

	err := handleParams(requestParamsDirPath, isBidderAllowed, bidderConfigMap.SetRequestParams)
	if err != nil {
		return nil, fmt.Errorf("error handling request params: %w", err)
	}

	err = handleParams(responseParamsDirPath, isBidderAllowed, bidderConfigMap.SetResponseParams)
	if err != nil {
		return nil, fmt.Errorf("error handling response params: %w", err)
	}

	return bidderConfigMap, nil
}

func handleParams(dirPath string, isBidderAllowed func(string) bool, setParams func(string, map[string]BidderParamMapper)) error {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("error:[%s] dirPath:[%s]", err.Error(), dirPath)
	}
	for _, file := range files {
		bidderName, ok := strings.CutSuffix(file.Name(), ".json")
		if !ok {
			return fmt.Errorf("error:[invalid_json_file_name] filename:[%s]", file.Name())
		}
		if !isBidderAllowed(bidderName) {
			continue
		}
		paramsConfig, err := readFile(dirPath, file.Name())
		if err != nil {
			return fmt.Errorf("error:[fail_to_read_file] dir:[%s] filename:[%s] err:[%s]", dirPath, file.Name(), err.Error())
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
	err = jsonutil.UnmarshalValid(content, &contentMap)
	return contentMap, err
}

// prepareParams parse the paramsConfig and returns the requestParams
func prepareParams(bidderName string, paramsConfig map[string]any) (map[string]BidderParamMapper, error) {
	paramsProperties, found := paramsConfig[propertiesKey]
	if !found {
		return nil, nil
	}
	paramsMap, ok := paramsProperties.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("error:[invalid_json_file_content_malformed_properties] bidderName:[%s]", bidderName)
	}
	params := make(map[string]BidderParamMapper, len(paramsMap))
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
		params[paramName] = BidderParamMapper{
			Location: locationStr,
		}
	}
	return params, nil
}
