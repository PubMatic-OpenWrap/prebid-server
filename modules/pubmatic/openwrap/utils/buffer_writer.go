package utils

import (
	"bytes"
	"net/http"
)

type CustomWriter struct {
	Response *bytes.Buffer
	Headers  http.Header
	Code     int
}

func (cw *CustomWriter) Write(data []byte) (int, error) {
	if data == nil {
		return 0, nil
	}

	if cw.Response == nil {
		cw.Response = new(bytes.Buffer)
	}

	return cw.Response.Write(data)
}

func (cw *CustomWriter) Header() http.Header {
	if cw.Headers == nil {
		cw.Headers = make(http.Header)
	}
	return cw.Headers
}

func (cw *CustomWriter) WriteHeader(statusCode int) {
	cw.Code = statusCode
}
