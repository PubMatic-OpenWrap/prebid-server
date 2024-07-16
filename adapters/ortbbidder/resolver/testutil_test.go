package resolver

import (
	"reflect"
	"testing"

	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

func TestValidateStructFields(t *testing.T) {
	type args struct {
		expectedFields map[string]reflect.Type
		structType     reflect.Type
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "mismatch between count of fields",
			args: args{
				expectedFields: map[string]reflect.Type{},
				structType:     reflect.TypeOf(openrtb_ext.ExtBidPrebidVideo{}),
			},
			wantErr: true,
		},
		{
			name: "found unexpected field",
			args: args{
				expectedFields: map[string]reflect.Type{
					"Duration_1": reflect.TypeOf(0.0),
					"field2":     reflect.TypeOf(""),
					"field3":     reflect.TypeOf(""),
				},
				structType: reflect.TypeOf(openrtb_ext.ExtBidPrebidVideo{
					Duration: 0,
				}),
			},
			wantErr: true,
		},
		{
			name: "found field with incorrect data type",
			args: args{
				expectedFields: map[string]reflect.Type{
					"Duration":        reflect.TypeOf(0.0),
					"PrimaryCategory": reflect.TypeOf(""),
					"VASTTagID":       reflect.TypeOf(""),
				},
				structType: reflect.TypeOf(openrtb_ext.ExtBidPrebidVideo{
					Duration: 0,
				}),
			},
			wantErr: true,
		},
		{
			name: "found valid fields",
			args: args{
				expectedFields: map[string]reflect.Type{
					"Duration":        reflect.TypeOf(0),
					"PrimaryCategory": reflect.TypeOf(""),
					"VASTTagID":       reflect.TypeOf(""),
				},
				structType: reflect.TypeOf(openrtb_ext.ExtBidPrebidVideo{
					Duration: 0,
				}),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateStructFields(tt.args.expectedFields, tt.args.structType); (err != nil) != tt.wantErr {
				t.Errorf("ValidateStructFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
