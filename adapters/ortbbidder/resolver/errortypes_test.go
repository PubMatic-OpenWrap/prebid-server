package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationFailedError(t *testing.T) {
	t.Run("validationFailedError", func(t *testing.T) {
		err := ValidationFailedError{Message: "any validation message"}
		assert.Equal(t, "any validation message", err.Error())
		assert.Equal(t, SeverityWarning, err.Severity())
	})
}

func TestDefaultValueError(t *testing.T) {
	t.Run("defaultValueError", func(t *testing.T) {
		err := DefaultValueError{Message: "any validation message"}
		assert.Equal(t, "any validation message", err.Error())
		assert.Equal(t, SeverityDebug, err.Severity())
	})
}

func TestContainsWarning(t *testing.T) {
	type args struct {
		errors []error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "error list contains warning",
			args: args{
				errors: []error{
					NewDefaultValueError("default value error"),
					NewValidationFailedError("validation failed"),
				},
			},
			want: true,
		},
		{
			name: "error list not contains warning",
			args: args{
				errors: []error{
					NewDefaultValueError("default value error"),
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsWarning(tt.args.errors); got != tt.want {
				t.Errorf("ContainsWarning() = %v, want %v", got, tt.want)
			}
		})
	}
}
