package macros

import (
	"strings"
)

type StringBased struct {
	Processor
}

func (p *StringBased) Replace(str string) (string, error) {
	return replaceStringBased(str, p.Cfg.delimiter, p.Cfg.macroValues, p.Cfg.valueConfig)
}

func replaceStringBased(str, delimiter string, macroValueMap map[string]string, valueConfig MacroValueConfig) (string, error) {
	replacedStr := str
	for macro, value := range macroValueMap {
		// if valueConfig.UrlEscape {
		// 	value = url.QueryEscape(value)
		// }

		// if len(value) == 0 && valueConfig.FailOnError {
		// 	return "", fmt.Errorf("Empty value for Macro '%s'", macro)
		// }
		replacedStr = strings.ReplaceAll(replacedStr, delimiter+macro+delimiter, value)

		// replacedStr = strings.ReplaceAll(replacedStr, fmt.Sprintf("%s%s%s", delimiter, macro, delimiter), value)
		// if replacedStr == str && valueConfig.FailOnError {
		// 	return "", fmt.Errorf("Empty value for Macro '%s'", macro)
		// }
	}
	return replacedStr, nil
}
