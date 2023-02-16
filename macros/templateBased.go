package macros

import (
	"fmt"
	"regexp"
	"text/template"
)

const (
	templateName   = "macro_replace"
	templateOption = "missingkey=zero"
)

// templateBased implements macro processor with text/template approach
type templateBased struct {
	cfg Config
}

func (processor *templateBased) Replace(url string, macroValues map[string]string) (string, error) {
	tmpl := template.New(templateName)
	tmpl.Option(templateOption)
	tmpl.Delims(processor.cfg.Delimiter, processor.cfg.Delimiter)
	// collect all macros based on delimiters
	regex := fmt.Sprintf("%s(.*?)%s", processor.cfg.Delimiter, processor.cfg.Delimiter)
	re := regexp.MustCompile(regex)
	replacedStr := re.ReplaceAllString(url, processor.cfg.Delimiter+".$1"+processor.cfg.Delimiter)
	tmpl, err := tmpl.Parse(replacedStr)
	if err != nil {
		return "", err
	}

	return resolveMacros(tmpl, macroValues)
}

func (processor *templateBased) AddTemplates([]string) {}
