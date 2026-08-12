package resolver

import "fmt"

// severity represents the severity level of a error.
type severity int

const (
	severityFatal   severity = iota // use this to discard the entire bid response
	severityWarning                 // use this to include errors in responseExt.warnings
	severityIgnore                  // use this to exclude errors from responseExt.warnings
)

// coder is used to indicate the severity of an error.
type coder interface {
	// Severity returns the severity level of the error.
	Severity() severity
}

// defaultValueError is used to flag that a default value was found.
type defaultValueError struct {
	Message string
}

// Error returns the error message.
func (err *defaultValueError) Error() string {
	return err.Message
}

// Severity returns the severity level of the error.
func (err *defaultValueError) Severity() severity {
	return severityIgnore
}

// validationFailedError is used to flag that the value validation failed.
type validationFailedError struct {
	Message string // Message contains the error message.
}

// Error returns the error message for ValidationFailedError.
func (err *validationFailedError) Error() string {
	return err.Message
}

// Severity returns the severity level for ValidationFailedError.
func (err *validationFailedError) Severity() severity {
	return severityWarning
}

// IsWarning returns true if an error is labeled with a Severity of SeverityWarning.
func IsWarning(err error) bool {
	s, ok := err.(coder)
	return ok && s.Severity() == severityWarning
}

// NewDefaultValueError creates a new DefaultValueError with a formatted message.
func NewDefaultValueError(message string, args ...any) error {
	return &defaultValueError{
		Message: fmt.Sprintf(message, args...),
	}
}

// NewValidationFailedError creates a new ValidationFailedError with a formatted message.
func NewValidationFailedError(message string, args ...any) error {
	return &validationFailedError{
		Message: fmt.Sprintf(message, args...),
	}
}
