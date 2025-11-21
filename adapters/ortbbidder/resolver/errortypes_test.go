package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationFailedError(t *testing.T) {
	t.Run("validationFailedError", func(t *testing.T) {
		err := validationFailedError{Message: "any validation message"}
		assert.Equal(t, "any validation message", err.Error())
		assert.Equal(t, severityWarning, err.Severity())
	})
}

func TestDefaultValueError(t *testing.T) {
	t.Run("defaultValueError", func(t *testing.T) {
		err := defaultValueError{Message: "any validation message"}
		assert.Equal(t, "any validation message", err.Error())
		assert.Equal(t, severityIgnore, err.Severity())
	})
}

func TestIsWarning(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "input err  is of severity warning",
			args: args{
				err: NewValidationFailedError("error"),
			},
			want: true,
		},
		{
			name: "input err is of severity ignore",
			args: args{
				err: NewDefaultValueError("error"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsWarning(tt.args.err); got != tt.want {
				t.Errorf("IsWarning() = %v, want %v", got, tt.want)
			}
		})
	}
}
