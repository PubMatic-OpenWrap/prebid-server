package ctv

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"git.pubmatic.com/PubMatic/go-common/util"
)

// JSONType New Type Defined for JSON Object
type JSONType byte

const (
	//JSONObject will refer to Object type
	JSONObject JSONType = iota

	//JSONInt will refer to Int type value
	JSONInt

	//JSONDouble will refer to Double type value
	JSONDouble

	//JSONString will refer to String type value
	JSONString

	//JSONObjectArray will refer to ObjectArray type value
	JSONObjectArray

	//JSONIntArray will refer to IntArray type value
	JSONIntArray

	//JSONDoubleArray will refer to DoubleArray type value
	JSONDoubleArray

	//JSONStringArray will refer to StringArray type value
	JSONStringArray
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
}

// JSONNode alias for Generic datatype of json object represented by map
type JSONNode = map[string]interface{}

const (
	parsingErrorFormat = `parsing error key:%v msg:%v`
)

// URLValues Will Parse HTTP Request and return key in specific type
type URLValues struct {
	url.Values
}

// GetInt Read Key from Request and Parse to Int Type
func (values *URLValues) GetInt(key string) (int, bool, error) {
	v := values.Get(key)
	if len(v) == 0 {
		return 0, false, nil
	}

	value, err := strconv.Atoi(v)
	if err != nil {
		return 0, true, fmt.Errorf(parsingErrorFormat, key, err.Error())
	}

	return value, true, nil
}

// GetBoolean Read Key from Request and Parse to Int Type
func (values *URLValues) GetBoolean(key string) (bool, bool, error) {
	v := values.Get(key)
	if len(v) == 0 {
		return false, false, nil
	}
	lowerV := strings.ToLower(v)
	switch lowerV {
	case "true":
		return true, true, nil
	case "false":
		return false, true, nil
	default:
		return false, true, fmt.Errorf(parsingErrorFormat, key, fmt.Sprintf(` '%s' is not a bool`, v))
	}
}

// GetFloat32 Read Key from Request and Parse to Float Type
func (values *URLValues) GetFloat32(key string) (float32, bool, error) {
	v := values.Get(key)
	if len(v) == 0 {
		return 0, false, nil
	}

	f, err := strconv.ParseFloat(v, 32)
	if nil != err {
		return float32(f), true, fmt.Errorf(parsingErrorFormat, key, err.Error())
	}

	return float32(f), true, nil
}

// GetFloat64 Read Key from Request and Parse to Float64 Type
func (values *URLValues) GetFloat64(key string) (float64, bool, error) {
	v := values.Get(key)
	if len(v) == 0 {
		return 0, false, nil
	}

	f, err := strconv.ParseFloat(v, 64)
	if nil != err {
		return f, true, fmt.Errorf(parsingErrorFormat, key, err.Error())
	}

	return f, true, nil
}

// GetString Read Key from Request and Parse to String Type
func (values *URLValues) GetString(key string) (string, bool) {
	v := values.Get(key)
	if len(v) == 0 {
		return v, false
	}
	return v, true
}

// GetString Read Key from Request and Parse to String Type
func (values *URLValues) GetStringPtr(key string) *string {
	if v := values.Get(key); len(v) > 0 {
		return &v
	}
	return nil
}

// GetIntArray Read Key from Request and Parse to IntArray Type
func (values *URLValues) GetIntArray(key string, sep string) ([]int, error) {
	if v := values.Get(key); len(v) > 0 {
		//Parse Value
		array := strings.Split(v, sep)
		retvalue := make([]int, len(array))
		j := 0
		for i := 0; i < len(array); i++ {
			intv, err := strconv.Atoi(array[i])
			if nil != err {
				return nil, fmt.Errorf(parsingErrorFormat, key, err.Error())
			}
			retvalue[j] = intv
			j++
		}
		return retvalue[:j], nil
	}
	return nil, nil
}

// GetInt8Array Read Key from Request and Parse to Int8Array Type
func (values *URLValues) GetInt8Array(key string, sep string) ([]int8, error) {
	if v := values.Get(key); len(v) > 0 {
		//Parse Value
		array := strings.Split(v, sep)
		retvalue := make([]int8, len(array))
		j := 0
		for i := 0; i < len(array); i++ {
			intv, err := strconv.Atoi(array[i])
			if nil != err {
				return nil, fmt.Errorf(parsingErrorFormat, key, err.Error())
			}
			retvalue[j] = int8(intv)
			j++
		}
		return retvalue[:j], nil
	}
	return nil, nil
}

// GetFloat32Array Read Key from Request and Parse to Float32Array Type
func (values *URLValues) GetFloat32Array(key string, sep string) ([]float32, error) {
	if v := values.Get(key); len(v) > 0 {
		//Parse Value
		array := strings.Split(v, sep)
		retvalue := make([]float32, len(array))
		j := 0
		for i := 0; i < len(array); i++ {
			s, err := strconv.ParseFloat(array[i], 32)
			if nil != err {
				return nil, fmt.Errorf(parsingErrorFormat, key, err.Error())
			}
			retvalue[j] = float32(s)
			j++
		}
		return retvalue[:j], nil
	}
	return nil, nil
}

// GetFloat64Array Read Key from Request and Parse to Float64Array Type
func (values *URLValues) GetFloat64Array(key string, sep string) ([]float64, error) {
	if v := values.Get(key); len(v) > 0 {
		//Parse Value
		array := strings.Split(v, sep)
		retvalue := make([]float64, len(array))
		j := 0
		for i := 0; i < len(array); i++ {
			s, err := strconv.ParseFloat(array[i], 64)
			if nil != err {
				return nil, fmt.Errorf(parsingErrorFormat, key, err.Error())
			}
			retvalue[j] = s
			j++
		}
		return retvalue[:j], nil
	}
	return nil, nil
}

// GetStringArray Read Key from Request and Parse to StringArray Type
func (values *URLValues) GetStringArray(key string, sep string) []string {
	if v := values.Get(key); len(v) > 0 {
		//Parse Value
		return strings.Split(v, sep)[:]
	}
	return nil
}

/*
SetValue function will recursively create nested object and set value
node: current JSONNode object
child: nested keys to create in node object (a.b.c)
value: value assigned to last key of child
example:

	child = a.b.c; value = 123  ==> {"a": {"b" : {"c":123}}}
*/
func SetValue(node JSONNode, child string, value *string) {
	if value == nil || len(child) == 0 {
		return
	}

	isLeaf := true
	keyStr := child
	index := strings.IndexByte(child, '.')

	if index != -1 {
		keyStr = child[0:index]
		isLeaf = false
	} else {
		index = len(child) - 1
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
		childNode, ok := node[key.Name]
		if !ok {
			newNode := make(JSONNode)
			node[key.Name] = newNode
			SetValue(newNode, child[index+1:], value)
		} else {
			node, ok := childNode.(JSONNode)
			if ok {
				SetValue(node, child[index+1:], value)
			}
		}

	case JSONString:
		node[key.Name] = value

	case JSONInt:
		node[key.Name] = GetInt(value)

	case JSONObjectArray:
		childNode, ok := node[key.Name]
		if !ok {
			newNode := []JSONNode{}
			node[key.Name] = newNode
			SetValue(newNode[0], child[index+1:], value)
		} else {
			node, ok := childNode.([]JSONNode)
			if ok {
				SetValue(node[0], child[index+1:], value)
			}
		}

	case JSONIntArray:
		node[key.Name] = GetIntArray(value, ArraySeparator)

	case JSONStringArray:
		node[key.Name] = GetStringArray(value, ArraySeparator)

	case JSONDouble:
		node[key.Name] = GetFloat64(value)

	case JSONDoubleArray:
		node[key.Name] = GetFloat64Array(value, ArraySeparator)
	}
}

// GetIntArray Read Key from Request and Parse to IntArray Type
func GetIntArray(v *string, sep string) []int {
	if v != nil {
		//Parse Value
		array := strings.Split(*v, sep)
		retvalue := make([]int, len(array))
		j := 0
		for i := 0; i < len(array); i++ {
			if intv, err := strconv.Atoi(array[i]); err == nil {
				retvalue[j] = intv
				j++
			}
		}
		return retvalue[:j]
	}
	return nil
}

// GetStringArray Read Key from Request and Parse to StringArray Type
func GetStringArray(v *string, sep string) []string {
	if v != nil {
		//Parse Value
		return strings.Split(*v, sep)[:]
	}
	return nil
}

// GetFloat64 Read Key from Request and Parse to Float64 Type
func GetFloat64(v *string) *float64 {
	if v != nil {
		//Parse Value
		if retvalue, err := strconv.ParseFloat(*v, 64); err == nil {
			return &retvalue
		}
	}
	return nil
}

// GetFloat64Array Read Key from Request and Parse to Float64Array Type
func GetFloat64Array(v *string, sep string) []float64 {
	if v != nil {
		//Parse Value
		array := strings.Split(*v, sep)
		retvalue := make([]float64, len(array))
		j := 0
		for i := 0; i < len(array); i++ {
			if s, err := strconv.ParseFloat(array[i], 64); err == nil {
				retvalue[j] = s
				j++
			}
		}
		return retvalue[:j]
	}
	return nil
}

// GetInt Read Key from Request and Parse to Int Type
func GetInt(v *string) *int {
	if v != nil {
		//Parse Value
		if retvalue, err := strconv.Atoi(*v); err == nil {
			return &retvalue
		}
	}
	return nil
}

// GetQueryParams Read the Key and  Parse to Map
func (values *URLValues) GetQueryParams(key string) (map[string]interface{}, error) {
	if v := values.Get(key); len(v) > 0 {
		pairs := strings.Split(v, "&")
		queryParams := make(map[string]interface{})

		for _, pair := range pairs {
			keyValue := strings.SplitN(pair, "=", 2)
			key := keyValue[0]
			if len(keyValue) == 2 {
				value := keyValue[1]
				var jsonValue interface{}
				if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
					queryParams[key] = jsonValue
				} else {
					queryParams[key] = value
				}
			} else {
				return nil, errors.New("error while parsing the query param")
			}
		}
		return queryParams, nil
	}
	return nil, nil
}

// GetJSON Read Key and Parsed it map
func (values *URLValues) GetJSON(key string) (map[string]interface{}, error) {
	if v := values.Get(key); len(v) > 0 {
		parsedData := make(map[string]interface{})
		if err := json.Unmarshal([]byte(v), &parsedData); err != nil {
			return nil, err
		}
		return parsedData, nil
	}
	return nil, nil
}

// GetBoolToInt extracts and parses a boolean value associated with the given key
func (values *URLValues) GetBoolToInt(key string) (*int, error) {
	if v := values.Get(key); len(v) > 0 {
		intVal, err := strconv.Atoi(v)
		if err != nil {
			switch strings.ToLower(fmt.Sprintf("%v", v)) {
			case "true":
				return util.GetIntPtr(1), nil
			case "false":
				return util.GetIntPtr(0), nil
			default:
				return nil, fmt.Errorf(parsingErrorFormat, key, fmt.Sprintf(` '%s' is not a bool`, v))
			}
		}
		return util.GetIntPtr(intVal), nil
	}
	return nil, nil
}
