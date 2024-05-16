package ortbbidder

import (
	"encoding/json"
	"fmt"
	"os"
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

// bidderProperty contains property details like location
type bidderProperty struct {
	location []string
}

// bidderConfig contains mappings of bidder with its request-properties and response-properties.
type bidderConfig struct {
	requestProperties  map[string]bidderProperty
	responseProperties map[string]bidderProperty
}

// biddersConfigMap contains map of bidderName to its bidderConfig
type biddersConfigMap struct {
	biddersConfig map[string]*bidderConfig
}

// init initialises the biddersConfigMap
func (bcm *biddersConfigMap) init() {
	bcm.biddersConfig = make(map[string]*bidderConfig)
}

// getBidderRequestProperties returns bidder specific request-properties
func (bcm *biddersConfigMap) getBidderRequestProperties(bidderName string) (map[string]bidderProperty, bool) {
	bidderConfig, found := bcm.biddersConfig[bidderName]
	if !found {
		return nil, false
	}
	return bidderConfig.requestProperties, true
}

// setBidderRequestProperties sets the bidder specific request-properties
func (bcm *biddersConfigMap) setBidderRequestProperties(bidderName string, properties map[string]bidderProperty) {
	if _, found := bcm.biddersConfig[bidderName]; !found {
		bcm.biddersConfig[bidderName] = &bidderConfig{}
	}
	bcm.biddersConfig[bidderName].requestProperties = properties
}

// global instance of setBidderRequestProperties
var g_biddersConfigMap *biddersConfigMap

// InitBiddersConfigMap initializes a g_biddersConfigMap instance using files in a given directory.
func InitBiddersConfigMap(dirPath string) (err error) {
	g_biddersConfigMap, err = prepareBiddersConfigMap(dirPath)
	return err
}

// prepareBiddersConfigMap creates a biddersConfigMap from JSON files in the specified directory.
func prepareBiddersConfigMap(dirPath string) (*biddersConfigMap, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error:[%s] dirPath:[%s]", err.Error(), dirPath)
	}
	var biddersConfigMap biddersConfigMap
	biddersConfigMap.init()
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
		requestProperties, err := prepareBidderRequestProperties(bidderName, fileContents)
		if err != nil {
			return nil, err
		}
		biddersConfigMap.setBidderRequestProperties(bidderName, requestProperties)
	}
	return &biddersConfigMap, nil
}

// prepareBidderRequestProperties prepares the request properties based on file-content passed as map[string]any
func prepareBidderRequestProperties(bidderName string, propertiesMap map[string]any) (map[string]bidderProperty, error) {
	properties, found := propertiesMap[properties]
	if !found {
		return nil, nil
	}
	propertiesTyped, ok := properties.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("error:[invalid_json_file_content_malformed_properties] bidderName:[%s]", bidderName)
	}
	requestProperties := make(map[string]bidderProperty, len(propertiesTyped))
	for propertyName, propertyValue := range propertiesTyped {
		property, ok := propertyValue.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("error:[invalid_json_file_content] bidder:[%s] bidderParam:[%s]", bidderName, propertyName)
		}
		location, found := property[location]
		if !found {
			continue // if location is absent then bidder-param will remain at its default location.
		}
		locationStr, ok := location.(string)
		if !ok {
			return nil, fmt.Errorf("error:[incorrect_location_in_bidderparam] bidder:[%s] bidderParam:[%s]", bidderName, propertyName)
		}
		requestProperties[propertyName] = bidderProperty{
			location: strings.Split(locationStr, "."),
		}
	}
	return requestProperties, nil
}

// mapBidderParamsInRequest updates the requestBody based on the bidder-params mapping details.
func mapBidderParamsInRequest(requestBody []byte, bidderParamDetails map[string]bidderProperty) ([]byte, error) {
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
