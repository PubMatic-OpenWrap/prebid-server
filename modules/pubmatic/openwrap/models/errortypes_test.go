package models

import (
	"fmt"
	"testing"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
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
			name: `DBError`,
			args: args{
				err: models.NewError(DBErrorType, "Error from the DB"),
			},
			want: want{
				errorMessage: `Error from the DB`,
				code:         DBErrorType,
			},
		},
		{
			name: `NormalError `,
			args: args{
				err: fmt.Errorf("Normal Error"),
			},
			want: want{
				errorMessage: `Normal Error`,
				code:         UnknownErrorType,
			},
		},
		{
			name: `AdUnitUnmarshal`,
			args: args{
				err: models.NewError(AdUnitUnmarshalErrorType, "Error in adUnitConfig Unmarshal"),
			},
			want: want{
				errorMessage: `Error in adUnitConfig Unmarshal`,
				code:         AdUnitUnmarshalErrorType,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want.errorMessage, tt.args.err.Error())
			code := GetErrorCode(tt.args.err)
			assert.Equal(t, tt.want.code, code)
		})
	}
}
