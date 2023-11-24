package middleware

import "github.com/prebid/openrtb/v19/openrtb3"

type CustomError interface {
	error
	Code() int
}

type OWError struct {
	code    int
	message string
}

// NewError New Object
func NewError(code int, message string) CustomError {
	return &OWError{code: code, message: message}
}

// Code Returns Error Code
func (e *OWError) Code() int {
	return e.code
}

// Error Returns Error Message
func (e *OWError) Error() string {
	return e.message
}

func GetNoBidReasonCode(code int) *openrtb3.NoBidReason {
	nbr := openrtb3.NoBidReason(code)
	return &nbr
}
