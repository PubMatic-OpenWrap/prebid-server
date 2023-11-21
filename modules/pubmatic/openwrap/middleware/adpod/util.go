package middleware

import (
	"net/http"
	"strings"

	validator "github.com/asaskevich/govalidator"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func getAndValidateRedirectURL(r *http.Request) (string, string, CustomError) {
	params := r.URL.Query()
	debug := params.Get(models.Debug)

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
