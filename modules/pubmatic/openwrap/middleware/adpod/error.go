package middleware

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
