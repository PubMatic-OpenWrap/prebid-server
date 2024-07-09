package adapters

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/prebid/prebid-server/v2/util/ptrutil"
)

type prepareBidParamJSONDefaultArgs struct {
	adapterName  string
	slotMappings map[string]interface{}
	width        *int
	height       *int
}

func TestPrepareBidParamJSONDefault(t *testing.T) {
	//Skip this test for adapters having entry in OpenWrap parameter mappings JSON file
	adaptersToSkip := make(map[string]bool)
	for bidderName := range parseOpenWrapParameterMappings() {
		adaptersToSkip[bidderName] = true
	}

	var tests []struct {
		name string
		args prepareBidParamJSONDefaultArgs
		want string
	}
	for adapterName, adapterParams := range adapterParams {
		if adaptersToSkip[adapterName] {
			continue
		}

		skipRequiredParam := false
		tests = append(tests, struct {
			name string
			args prepareBidParamJSONDefaultArgs
			want string
		}{name: getTestName(adapterName, skipRequiredParam), args: getTestArgs(adapterName, adapterParams, skipRequiredParam),
			want: getExpectedJSON(adapterParams, skipRequiredParam)})

		skipRequiredParam = true
		tests = append(tests, struct {
			name string
			args prepareBidParamJSONDefaultArgs
			want string
		}{name: getTestName(adapterName, skipRequiredParam), args: getTestArgs(adapterName, adapterParams, skipRequiredParam),
			want: getExpectedJSON(adapterParams, skipRequiredParam)})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := BidderParameters{
				AdapterName: tt.args.adapterName,
				FieldMap:    tt.args.slotMappings,
				Width: func() *int64 {
					if tt.args.width != nil {
						return ptrutil.ToPtr(int64(*tt.args.width))
					}
					return nil
				}(),
				Height: func() *int64 {
					if tt.args.width != nil {
						return ptrutil.ToPtr((int64(*tt.args.height)))
					}
					return nil
				}(),
			}
			got, _ := prepareBidParamJSONDefault(params)
			AssertJSON(t, json.RawMessage(tt.want), got)
		})
	}
}

func getTestName(adapterName string, skipRequiredParam bool) string {
	if skipRequiredParam {
		return fmt.Sprintf("Test for adapter %s with required param missing", adapterName)
	}
	return fmt.Sprintf("Test for adapter %s", adapterName)
}

func getTestArgs(adapterName string, params map[string]*ParameterMapping, skipRequiredParam bool) prepareBidParamJSONDefaultArgs {
	return prepareBidParamJSONDefaultArgs{
		adapterName:  adapterName,
		slotMappings: getDummySlotMappings(params, skipRequiredParam),
		width:        nil,
		height:       nil,
	}

}

func getExpectedJSON(params map[string]*ParameterMapping, skipRequiredParam bool) string {
	allParamsOptional := true
	for _, mapping := range params {
		if mapping.Required {
			allParamsOptional = false
		}
	}
	if !allParamsOptional && skipRequiredParam {
		return ""
	}
	return GetJSON(getExpectedResponseSlotMappings(params, skipRequiredParam))
}

func getExpectedResponseSlotMappings(params map[string]*ParameterMapping, skipRequiredParam bool) map[string]interface{} {
	targetMap := make(map[string]interface{})
	for _, mapping := range params {
		if mapping.Required && skipRequiredParam {
			continue
		}
		targetMap[mapping.KeyName] = getExpectedResponseValue(mapping.Datatype)
		if mapping.KeyName == "rateLimit" {
			delete(targetMap, mapping.KeyName)
		}
	}
	return targetMap
}

func getExpectedResponseValue(datatype string) interface{} {
	switch datatype {
	case "string":
		return "dummyString"
	case "number":
		return 0.10
	case "integer":
		return 1
	case "boolean":
		return true
	case "[]string":
		return []string{"dummyValue1", "dummyValue2"}
	case "[]integer":
		return []int{1, 2, 3}
	case "[]number":
		return []float64{1.1, 2.2}
	default:
		return "defaultDummyString"
	}
}

func getDummySlotMappings(params map[string]*ParameterMapping, skipRequiredParam bool) map[string]interface{} {
	targetMap := make(map[string]interface{})
	for _, mapping := range params {
		if mapping.Required && skipRequiredParam {
			continue
		}
		targetMap[mapping.KeyName] = getDummyValue(mapping.Datatype)
	}
	return targetMap
}

func getDummyValue(datatype string) interface{} {
	switch datatype {
	case "string":
		return "dummyString"
	case "number":
		return 0.10011
	case "integer":
		return 1
	case "boolean":
		return true
	case "[]string":
		return []string{"dummyValue1", "dummyValue2"}
	case "[]integer":
		return []int{1, 2, 3}
	case "[]number":
		return []float64{1.1, 2.2}
	default:
		return "defaultDummyString"
	}
}

func Test_getDataType(t *testing.T) {
	type args struct {
		paramType string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "TEST paramType []number",
			args: args{paramType: "[]number"},
			want: 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDataType(tt.args.paramType); got != tt.want {
				t.Errorf("getDataType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_addBidParam(t *testing.T) {
	type args struct {
		bidParams map[string]interface{}
		name      string
		paramType string
		value     interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{

		{
			name:    "Empty string input for paramType",
			args:    args{bidParams: map[string]interface{}{"test": "test1"}, name: "test", paramType: "string", value: ""},
			wantErr: true,
		},
		{
			name:    "Invalid float input for paramType",
			args:    args{bidParams: map[string]interface{}{"test": "test1"}, name: "test", paramType: "number", value: "abc"},
			wantErr: true,
		},
		{
			name:    "Invalid boolean input for paramType",
			args:    args{bidParams: map[string]interface{}{"test": "test1"}, name: "test", paramType: "boolean", value: "abc"},
			wantErr: true,
		},
		{
			name:    "Invalid array of int input for paramType",
			args:    args{bidParams: map[string]interface{}{"test": "test1"}, name: "test", paramType: "[]integer", value: true},
			wantErr: true,
		},
		{
			name:    "array of float input for paramType",
			args:    args{bidParams: map[string]interface{}{"test": "test1"}, name: "test", paramType: "[]number", value: []float64{11.6}},
			wantErr: false,
		},
		{
			name:    "Invalid array of float input for paramType",
			args:    args{bidParams: map[string]interface{}{"test": "test1"}, name: "test", paramType: "[]number", value: "11.6"},
			wantErr: true,
		},
		{
			name:    "Invalid array of float input for paramType1",
			args:    args{bidParams: map[string]interface{}{"test": "test1"}, name: "test", paramType: "[]number", value: []int64{11}},
			wantErr: true,
		},
		{
			name:    "Invalid array of float input for paramType2",
			args:    args{bidParams: map[string]interface{}{"test": "test1"}, name: "test", paramType: "[]number", value: []interface{}{11}},
			wantErr: true,
		},
		{
			name:    "valid array of float input for paramType",
			args:    args{bidParams: map[string]interface{}{"test": "test1"}, name: "test", paramType: "[]number", value: []interface{}{11.5}},
			wantErr: false,
		},
		{
			name:    "Invalid array of string input for paramType2",
			args:    args{bidParams: map[string]interface{}{"test": "test1"}, name: "test", paramType: "[]string", value: []interface{}{nil}},
			wantErr: true,
		},
		{
			name:    "valid array of string input for paramType",
			args:    args{bidParams: map[string]interface{}{"test": "test1"}, name: "test", paramType: "[]string", value: []interface{}{"test"}},
			wantErr: false,
		},
		{
			name:    "valid array of string input for paramType",
			args:    args{bidParams: map[string]interface{}{"test": "test1"}, name: "test", paramType: "[]string", value: 5},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := addBidParam(tt.args.bidParams, tt.args.name, tt.args.paramType, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("addBidParam() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
