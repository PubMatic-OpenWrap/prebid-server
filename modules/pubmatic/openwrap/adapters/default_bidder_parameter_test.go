package adapters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func getExpectedOpenWrapParameterMappings() map[string]map[string]*ParameterMapping {
	mapping := make(map[string]map[string]*ParameterMapping)
	dmxMap := make(map[string]*ParameterMapping)
	dmxMap["tagid"] = &ParameterMapping{
		KeyName: "dmxid",
	}
	mapping["dmx"] = dmxMap

	vrtcalMap := make(map[string]*ParameterMapping)
	vrtcalMap["just_an_unused_vrtcal_param"] = &ParameterMapping{
		KeyName:      "dummyParam",
		DefaultValue: "1",
	}
	mapping["vrtcal"] = vrtcalMap

	gridMap := make(map[string]*ParameterMapping)
	gridMap["uid"] = &ParameterMapping{
		Required: true,
	}
	mapping["grid"] = gridMap

	adkernelMap := make(map[string]*ParameterMapping)
	adkernelMap["zoneId"] = &ParameterMapping{
		Datatype: "integer",
	}
	mapping["adkernel"] = adkernelMap

	return mapping
}

func TestGetType(t *testing.T) {
	type args struct {
		param BidderParameter
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Array of strings",
			args: args{
				param: BidderParameter{
					Type: "array",
					Items: ArrayItemsType{
						Type: "string",
					},
				},
			},
			want: "[]string",
		},
		{
			name: "Array of integers",
			args: args{
				param: BidderParameter{
					Type: "array",
					Items: ArrayItemsType{
						Type: "integer",
					},
				},
			},
			want: "[]integer",
		},
		{
			name: "Array of numbers",
			args: args{
				param: BidderParameter{
					Type: "array",
					Items: ArrayItemsType{
						Type: "number",
					},
				},
			},
			want: "[]number",
		},
		{
			name: "String from array of options",
			args: args{
				param: BidderParameter{
					Type: []string{"integer", "string"},
				},
			},
			want: "string",
		},
		{
			name: "First item from array of options",
			args: args{
				param: BidderParameter{
					Type: []string{"integer", "number"},
				},
			},
			want: "integer",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getType(tt.args.param); got != tt.want {
				t.Errorf("getType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseBidderParams(t *testing.T) {
	parseBidderParams("../../static/bidder-params")
	assert.Equal(t, 161, len(adapterParams), "Length of expected entries should match")
	// calculate this number using X-Y
	// where X is calculated using command - `ls -l | wc -l` (substract 1 from result)
	// Y is calculated using command `grep -EinR 'oneof|not|anyof|dependenc' static/bidder-params | grep -v "description" | grep -oE './.*.json'  | uniq | wc -l`
}

func TestParseBidderSchemaDefinitions(t *testing.T) {
	schemaDefinitions, _ := parseBidderSchemaDefinitions("../../../../static/bidder-params")
	assert.Equal(t, 196, len(schemaDefinitions), "Length of expected entries should match")
	// calculate this number using command - `ls -l | wc -l` (substract 1 from result)
}

func TestParseOpenWrapParameterMappings(t *testing.T) {
	tests := []struct {
		name string
		want map[string]map[string]*ParameterMapping
	}{
		{
			name: "Verify mappings are correctly parsed",
			want: getExpectedOpenWrapParameterMappings(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOpenWrapParameterMappings()
			assert.Equal(t, tt.want, got)
		})
	}
}
