package resolver

import "fmt"

// Severity represents the severity level of a error.
type Severity int

const (
	SeverityFatal Severity = iota
	SeverityWarning
	SeverityDebug
)

// Coder is used to indicate the severity of an error.
type Coder interface {
	// Severity returns the severity level of the error.
	Severity() Severity
}

// DefaultValueError is used to flag that a default value was found.
type DefaultValueError struct {
	Message string
}

// Error returns the error message.
func (err *DefaultValueError) Error() string {
	return err.Message
}

// Severity returns the severity level of the error.
func (err *DefaultValueError) Severity() Severity {
	return SeverityDebug
}

// ValidationFailedError is used to flag that the value validation failed.
type ValidationFailedError struct {
	Message string // Message contains the error message.
}

// Error returns the error message for ValidationFailedError.
func (err *ValidationFailedError) Error() string {
	return err.Message
}

// Severity returns the severity level for ValidationFailedError.
func (err *ValidationFailedError) Severity() Severity {
	return SeverityWarning
}

// isWarning returns true if an error is labeled with a Severity of SeverityWarning.
func isWarning(err error) bool {
	s, ok := err.(Coder)
	return ok && s.Severity() == SeverityWarning
}

// ContainsWarning checks if the error list contains a warning.
func ContainsWarning(errors []error) bool {
	for _, err := range errors {
		if isWarning(err) {
			return true
		}
	}

	return false
}

// NewDefaultValueError creates a new DefaultValueError with a formatted message.
func NewDefaultValueError(message string, args ...any) error {
	return &DefaultValueError{
		Message: fmt.Sprintf(message, args...),
	}
}

// NewValidationFailedError creates a new ValidationFailedError with a formatted message.
func NewValidationFailedError(message string, args ...any) error {
	return &ValidationFailedError{
		Message: fmt.Sprintf(message, args...),
	}
}
