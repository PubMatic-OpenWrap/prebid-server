package macros

import (
	"strings"
)

type StringBased struct {
	Processor
}

func (p *StringBased) Replace(str string, macroValues map[string]string) (string, error) {
	return replaceStringBased(str, p.Cfg.Delimiter, macroValues, p.Cfg.valueConfig)
}

func replaceStringBased(str, Delimiter string, macroValueMap map[string]string, valueConfig MacroValueConfig) (string, error) {
	replacedStr := str
	for macro, value := range macroValueMap {
		// if valueConfig.UrlEscape {
		// 	value = url.QueryEscape(value)
		// }

		// if len(value) == 0 && valueConfig.FailOnError {
		// 	return "", fmt.Errorf("Empty value for Macro '%s'", macro)
		// }
		replacedStr = strings.ReplaceAll(replacedStr, Delimiter+macro+Delimiter, value)

		// replacedStr = strings.ReplaceAll(replacedStr, fmt.Sprintf("%s%s%s", Delimiter, macro, Delimiter), value)
		// if replacedStr == str && valueConfig.FailOnError {
		// 	return "", fmt.Errorf("Empty value for Macro '%s'", macro)
		// }
	}
	return replacedStr, nil
}
