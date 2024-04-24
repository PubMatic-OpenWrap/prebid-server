package models

// Defines numeric codes for OW specific Errors
const (
	UnknownErrorCode = 999
	NorErrorCode     = iota
	DBErrorCode
	AdUnitUnmarshalErrorCode
)

// DBError is used to in case of DB query gives error.
type DBError struct {
	Message string
}

func (err *DBError) Error() string {
	return err.Message
}

func (err *DBError) Code() int {
	return DBErrorCode
}

// AdUnitUnmarshalError is used to in case of Invalid adUnitConfig is present
type AdUnitUnmarshalError struct {
	Message string
}

func (err *AdUnitUnmarshalError) Error() string {
	return err.Message
}

func (err *AdUnitUnmarshalError) Code() int {
	return AdUnitUnmarshalErrorCode
}

// Coder provides an error or warning code with severity.
type Coder interface {
	Code() int
}

// GetErrorCode returns the error code, or UnknownErrorCode if unavailable.
func GetErrorCode(err error) int {
	if e, ok := err.(Coder); ok {
		return e.Code()
	}
	return UnknownErrorCode
}
