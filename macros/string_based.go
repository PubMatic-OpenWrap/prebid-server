package macros

import (
	"strings"
)

type stringBased struct {
	cfg Config
}

func (processor *stringBased) Replace(str string, macroValues map[string]string) (string, error) {
	return replaceStringBased(str, processor.cfg.Delimiter, macroValues)
}

func replaceStringBased(str, delimiter string, macroValueMap map[string]string) (string, error) {
	replacedStr := str
	for macro, value := range macroValueMap {
		replacedStr = strings.ReplaceAll(replacedStr, delimiter+macro+delimiter, value)
	}
	return replacedStr, nil
}

func (processor *stringBased) AddTemplates([]string) {

}
