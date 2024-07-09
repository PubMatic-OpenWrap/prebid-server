package adapters

import (
	"encoding/json"
	"errors"
	"fmt"

	"strconv"
	"strings"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

// Map containing []ParameterMapping for all partners (partner name)
var adapterParams map[string]map[string]*ParameterMapping

func prepareBidParamJSONDefault(params BidderParameters) (json.RawMessage, error) {
	bidderParamMapping, present := adapterParams[params.AdapterName]
	if !present {
		return nil, fmt.Errorf(errInvalidS2SPartnerFormat, params.AdapterName, params.SlotKey)
	}

	bidderParams := make(map[string]interface{})
	for _, mapping := range bidderParamMapping {
		paramValue, present := params.FieldMap[mapping.KeyName]
		if !present && mapping.DefaultValue != nil {
			present = true
			paramValue = mapping.DefaultValue
		}

		if !present && mapping.Required {
			return nil, fmt.Errorf(errDefaultBidderParameterMissingFormat, params.AdapterName, mapping.BidderParamName, mapping.KeyName)
		}

		if present {
			err := addBidParam(bidderParams, mapping.BidderParamName, mapping.Datatype, paramValue)
			if err != nil && mapping.Required {
				return nil, err
			}
			if params.AdapterName == "aidem" && mapping.KeyName == "rateLimit" {
				delete(bidderParams, mapping.KeyName)
			}
		}
	}

	jsonBuf, err := json.Marshal(bidderParams)
	if err != nil {
		return nil, err
	}

	return jsonBuf, nil
}

func addBidParam(bidParams map[string]interface{}, name string, paramType string, value interface{}) error {
	dataType := getDataType(paramType)

	switch dataType {
	case models.DataTypeInteger:
		//DataTypeInteger
		intVal, err := strconv.Atoi(fmt.Sprintf("%v", value))
		if err != nil {
			return err
		}
		bidParams[name] = intVal
	case models.DataTypeFloat:
		//DataTypeFloat
		floatVal, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64)
		if err != nil {
			return err
		}
		bidParams[name] = toFixed(floatVal, FloatValuePrecision)
	case models.DataTypeString:
		//DataTypeString
		val := fmt.Sprintf("%v", value)
		if val == "" {
			return errors.New("value is empty")
		}
		bidParams[name] = fmt.Sprintf("%v", value)
	case models.DataTypeBoolean:
		//DataTypeBoolean
		boolVal, err := strconv.ParseBool(fmt.Sprintf("%v", value))
		if err != nil {
			return err
		}
		bidParams[name] = boolVal
	case models.DataTypeArrayOfIntegers:
		//Array of DataTypeInteger
		switch v := value.(type) {
		case string:
			var arr []int
			err := json.Unmarshal([]byte(value.(string)), &arr)
			if err != nil {
				return err
			}
			bidParams[name] = arr
		case []int:
			bidParams[name] = v
		case []interface{}:
			//Unmarshal's default type for array. Refer https://pkg.go.dev/encoding/json#Unmarshal
			arr := make([]int, 0, len(v))
			for _, elem := range v {
				elemFloat, ok := elem.(float64) //Unmarshal's default type interface values
				if !ok {
					return fmt.Errorf("ErrTypeCastFailed %s float64 %v", name, elem)
				}
				arr = append(arr, int(elemFloat))
			}

			bidParams[name] = arr
		default:
			errMsg := fmt.Sprintf("unknown array type %T!\n", v)
			return errors.New(errMsg)
		}
	case models.DataTypeArrayOfFloats:
		//Array of DataTypeFloat
		switch v := value.(type) {
		case string:
			var arr []float64
			err := json.Unmarshal([]byte(value.(string)), &arr)
			if err != nil {
				return err
			}
			bidParams[name] = arr
		case []float64:
			bidParams[name] = v
		case []interface{}:
			//Unmarshal's default type for array. Refer https://pkg.go.dev/encoding/json#Unmarshal
			arr := make([]float64, 0, len(v))
			for _, elem := range v {
				elemFloat, ok := elem.(float64) //Unmarshal's default type interface values
				if !ok {
					return fmt.Errorf("ErrTypeCastFailed %s float64 %v", name, elem)
				}
				arr = append(arr, elemFloat)
			}

			bidParams[name] = arr
		default:
			errMsg := fmt.Sprintf("unknown array type %T!\n", v)
			return errors.New(errMsg)
		}
	case models.DataTypeArrayOfStrings:
		//Array of DataTypeString
		switch v := value.(type) {
		case string:
			var arr []string
			stringValue := strings.Trim(value.(string), "[]")
			arr = strings.Split(stringValue, ",")
			bidParams[name] = arr
		case []string:
			bidParams[name] = v
		case []interface{}:
			arr := make([]string, 0, len(v))
			for _, elem := range v {
				elemStr, ok := elem.(string)
				if !ok {
					return fmt.Errorf("ErrTypeCastFailed %s float64 %v", name, elem)
				}
				arr = append(arr, elemStr)
			}
			bidParams[name] = arr
		default:
			errMsg := fmt.Sprintf("unknown array type %T!\n", v)
			return errors.New(errMsg)
		}
	default:
		bidParams[name] = fmt.Sprintf("%v", value)
	}

	return nil
}

func getDataType(paramType string) int {
	switch paramType {
	case "string":
		return models.DataTypeString
	case "number":
		return models.DataTypeFloat
	case "integer":
		return models.DataTypeInteger
	case "boolean":
		return models.DataTypeBoolean
	case "[]string":
		return models.DataTypeArrayOfStrings
	case "[]integer":
		return models.DataTypeArrayOfIntegers
	case "[]number":
		return models.DataTypeArrayOfFloats
	default:
		return models.DataTypeUnknown
	}
}
