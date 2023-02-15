package macros

import (
	"fmt"
	"regexp"
	"text/template"
)

type templateBased struct {
	cfg Config
}

func (processor *templateBased) Replace(url string, macroValues map[string]string) (string, error) {
	tmpl := template.New("macro_replace")
	tmpl.Option("missingkey=zero")
	tmpl.Delims(processor.cfg.Delimiter, processor.cfg.Delimiter)
	// collect all macros based on delimiters
	regex := fmt.Sprintf("%s(.*?)%s", processor.cfg.Delimiter, processor.cfg.Delimiter)
	re := regexp.MustCompile(regex)
	replacedStr := re.ReplaceAllString(url, "##.$1##")
	tmpl, err := tmpl.Parse(replacedStr)
	if err != nil {
		panic(err)
	}

	return resolveMacros(tmpl, macroValues)
}

func (processor *templateBased) AddTemplates([]string) {}
