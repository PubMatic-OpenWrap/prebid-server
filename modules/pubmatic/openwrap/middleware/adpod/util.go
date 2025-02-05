package middleware

import (
	"encoding/json"
)

var (
	middlewareLocation = []string{"prebid", "modules", "errors", "pubmatic.openwrap", "pubmatic.openwrap.middleware"}
)

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
