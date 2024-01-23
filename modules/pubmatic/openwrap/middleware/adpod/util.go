package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	validator "github.com/asaskevich/govalidator"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

var (
	middlewareLocation = []string{"prebid", "modules", "errors", "pubmatic.openwrap", "pubmatic.openwrap.middleware"}
	errorLocation      = []string{"prebid", "modules", "errors", "pubmatic.openwrap"}
)

func getAndValidateRedirectURL(r *http.Request) (string, string, CustomError) {
	params := r.URL.Query()
	debug := params.Get(models.Debug)
	if len(debug) == 0 {
		debug = "0"
	}

	format := strings.ToLower(strings.TrimSpace(params.Get(models.ResponseFormatKey)))
	if format != "" {
		if format != models.ResponseFormatJSON && format != models.ResponseFormatRedirect {
			return "", debug, NewError(634, "Invalid response format, must be 'json' or 'redirect'")
		}
	}

	owRedirectURL := params.Get(models.OWRedirectURLKey)
	if len(owRedirectURL) > 0 {
		owRedirectURL = strings.TrimSpace(owRedirectURL)
		if format == models.ResponseFormatRedirect && !isValidURL(owRedirectURL) {
			return "", debug, NewError(633, "Invalid redirect URL")
		}
	}

	return owRedirectURL, debug, nil
}

func isValidURL(urlVal string) bool {
	if !(strings.HasPrefix(urlVal, "http://") || strings.HasPrefix(urlVal, "https://")) {
		return false
	}
	return validator.IsRequestURL(urlVal) && validator.IsURL(urlVal)
}

func addErrorInExtension(errMsg string, ext json.RawMessage, debug string) json.RawMessage {
	if debug != "1" {
		return ext
	}

	var responseExt map[string]interface{}
	if ext != nil {
		err := json.Unmarshal(ext, &responseExt)
		if err != nil {
			return ext
		}
	}

	if responseExt == nil {
		responseExt = map[string]interface{}{}
	}

	prebidExt, ok := responseExt[middlewareLocation[0]].(map[string]interface{})
	if !ok {
		prebidExt = map[string]interface{}{}
	}

	module, ok := prebidExt[middlewareLocation[1]].(map[string]interface{})
	if !ok {
		module = map[string]interface{}{}
	}

	errors, ok := module[middlewareLocation[2]].(map[string]interface{})
	if !ok {
		errors = map[string]interface{}{}
	}

	pubOW, ok := errors[middlewareLocation[3]].(map[string]interface{})
	if !ok {
		pubOW = map[string]interface{}{}
	}

	pubOW[middlewareLocation[4]] = []string{errMsg}
	errors[middlewareLocation[3]] = pubOW
	module[middlewareLocation[2]] = errors
	prebidExt[middlewareLocation[1]] = module
	responseExt[middlewareLocation[0]] = prebidExt

	data, err := json.Marshal(responseExt)
	if err != nil {
		return ext
	}

	return data
}
