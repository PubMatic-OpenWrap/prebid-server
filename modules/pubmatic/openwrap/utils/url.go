package utils

import (
	"strings"

	validator "github.com/asaskevich/govalidator"
)

func IsValidURL(urlVal string) bool {
	if !(strings.HasPrefix(urlVal, "http://") || strings.HasPrefix(urlVal, "https://")) {
		return false
	}
	return validator.IsRequestURL(urlVal) && validator.IsURL(urlVal)
}
