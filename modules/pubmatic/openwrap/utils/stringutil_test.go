package utils

import (
	"reflect"
	"testing"
)

func TestGetIntArrayFromString(t *testing.T) {
	type args struct {
		str      string
		separtor string
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "Get String array from string",
			args: args{
				str:      "1,2,3",
				separtor: ",",
			},
			want: []int{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetIntArrayFromString(tt.args.str, tt.args.separtor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetIntArrayFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
