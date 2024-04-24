package models

// Defines numeric codes for OW specific Errors
const (
	UnknownErrorType = 999
	NorErrorType     = iota
	DBErrorType
	AdUnitUnmarshalErrorType
)

// ErrorCode new type defined for wrapper errors
type ErrorCode = int

// IError Interface for Custom Errors
type IError interface {
	error
	Code() ErrorCode
}

// Error Structure for Custom Errors
type Error struct {
	code    ErrorCode
	message string
}

// NewError New Object
func NewError(code ErrorCode, message string) *Error {
	return &Error{code: code, message: message}
}

// Code Returns Error Code
func (e *Error) Code() ErrorCode {
	return e.code
}

// Error Returns Error Message
func (e *Error) Error() string {
	return e.message
}

// GetErrorCode returns the error code, or UnknownErrorCode if unavailable.
func GetErrorCode(err error) int {
	if e, ok := err.(IError); ok {
		return e.Code()
	}
	return UnknownErrorType
}
