package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	type args struct {
		err error
	}
	type want struct {
		errorMessage string
		code         int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: `normal_error`,
			args: args{
				err: fmt.Errorf(`normal_error`),
			},
			want: want{
				errorMessage: `normal_error`,
				code:         UnknownErrorCode,
			},
		},
		{
			name: `DBError`,
			args: args{
				err: &DBError{Message: `DBError_ErrorMessage`},
			},
			want: want{
				errorMessage: `DBError_ErrorMessage`,
				code:         DBErrorCode,
			},
		},
		{
			name: `AdUnitUnmarshal`,
			args: args{
				err: &AdUnitUnmarshalError{Message: `AdUnitUnmarshal_ErrorMessage`},
			},
			want: want{
				errorMessage: `AdUnitUnmarshal_ErrorMessage`,
				code:         AdUnitUnmarshalErrorCode,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want.errorMessage, tt.args.err.Error())
			if code, ok := tt.args.err.(Coder); ok {
				assert.Equal(t, tt.want.code, code.Code())
			}
		})
	}
}
