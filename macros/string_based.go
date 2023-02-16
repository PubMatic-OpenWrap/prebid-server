package macros

import (
	"strings"
)

// stringBased implements macro processor interface with string replacement approach.
type stringBased struct {
	cfg Config
}

func (processor *stringBased) Replace(str string, macroValues map[string]string) (string, error) {

	delimiter := processor.cfg.Delimiter
	replacedStr := str
	for macro, value := range macroValues {
		replacedStr = strings.ReplaceAll(replacedStr, delimiter+macro+delimiter, value)
	}
	return replacedStr, nil
}

func (processor *stringBased) AddTemplates([]string) {

}
