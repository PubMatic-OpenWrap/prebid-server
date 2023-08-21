package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	validator "github.com/asaskevich/govalidator"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/middleware"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

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

type OWError struct {
	Code    int
	Message string
}

// OperRTB Writer
type AdpodOpenRTBWriter struct {
	W http.ResponseWriter
}

func (aw AdpodOpenRTBWriter) Write(data []byte) (int, error) {
	data = middleware.FormOperRTBResponse(data)
	return aw.W.Write(data)
}

func (aw AdpodOpenRTBWriter) Header() http.Header {
	return aw.W.Header()
}

func (aw AdpodOpenRTBWriter) WriteHeader(statusCode int) {
	aw.W.WriteHeader(statusCode)
}

// VAST writer
type AdpodVastWriter struct {
	W http.ResponseWriter
}

func (aw AdpodVastWriter) Write(data []byte) (int, error) {
	data = middleware.FormVastResponse(data)
	return aw.W.Write(data)
}

func (aw AdpodVastWriter) Header() http.Header {
	return aw.W.Header()
}

func (aw AdpodVastWriter) WriteHeader(statusCode int) {
	aw.W.WriteHeader(statusCode)
}

// JSON Writer
type AdpodJSONWriter struct {
	W           http.ResponseWriter
	RedirectURL string
}

func (aw AdpodJSONWriter) Write(data []byte) (int, error) {
	data = middleware.FormJSONResponse(g_cacheClient, data, aw.RedirectURL)
	return aw.W.Write(data)
}

func (aw AdpodJSONWriter) Header() http.Header {
	return aw.W.Header()
}

func (aw AdpodJSONWriter) WriteHeader(statusCode int) {
	aw.W.WriteHeader(statusCode)
}

func GetAndValidateRedirectURL(r *http.Request) (string, CustomError) {
	params := r.URL.Query()

	format := strings.ToLower(strings.TrimSpace(params.Get(models.ResponseFormatKey)))
	if format != "" {
		if format != models.ResponseFormatJSON && format != models.ResponseFormatRedirect {
			return "", NewError(634, "Invalid response format, must be 'json' or 'redirect'")
		}
	}

	owRedirectURL := params.Get(models.OWRedirectURLKey)
	if len(owRedirectURL) > 0 {
		owRedirectURL = strings.TrimSpace(owRedirectURL)
		if format == models.ResponseFormatRedirect && !IsValidURL(owRedirectURL) {
			return "", NewError(633, "Invalid redirect URL")
		}
	}

	return owRedirectURL, nil
}

func IsValidURL(urlVal string) bool {
	if !(strings.HasPrefix(urlVal, "http://") || strings.HasPrefix(urlVal, "https://")) {
		return false
	}
	return validator.IsRequestURL(urlVal) && validator.IsURL(urlVal)
}

func writeErrorResponse(w http.ResponseWriter, code int, err CustomError) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	errResponse := GetErrorResponse(err)
	fmt.Fprintln(w, errResponse)
}

func GetErrorResponse(err CustomError) []byte {
	if err == nil {
		return nil
	}

	response, _ := json.Marshal(map[string]interface{}{
		"ErrorCode": err.Code(),
		"Error":     err.Error(),
	})
	return response
}
