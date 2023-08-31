package middleware

type CustomError interface {
	error
	Code() int
}

type OwError struct {
	code    int
	message string
}

// NewError New Object
func NewError(code int, message string) CustomError {
	return &OwError{code: code, message: message}
}

// Code Returns Error Code
func (e *OwError) Code() int {
	return e.code
}

// Error Returns Error Message
func (e *OwError) Error() string {
	return e.message
}
