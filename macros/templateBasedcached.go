package macros

import (
	"bytes"
	"fmt"
	"regexp"
	"text/template"
)

// templateBasedCache implements macro processor interface with text/template caching approach
// new template will be cached for each event url per request.
type templateBasedCached struct {
	templates map[string]*template.Template
	cfg       Config
}

func (processor *templateBasedCached) Replace(str string, macroValues map[string]string) (string, error) {
	return resolveMacros(processor.templates[str], macroValues)
}

func (processor *templateBasedCached) AddTemplates(templates []string) {

	delimiter := processor.cfg.Delimiter
	for _, url := range templates {
		if _, ok := processor.templates[url]; ok {
			continue
		}
		tmpl := template.New(templateName)
		tmpl.Option(templateOption)
		tmpl.Delims(delimiter, delimiter)
		// collect all macros based on delimiters
		regex := fmt.Sprintf("%s(.*?)%s", delimiter, delimiter)
		re := regexp.MustCompile(regex)
		replacedStr := re.ReplaceAllString(url, delimiter+".$1"+delimiter)
		tmpl, err := tmpl.Parse(replacedStr)
		if err != nil {
			return
		}
		processor.templates[url] = tmpl
	}
}

// ResolveMacros resolves macros in the given template with the provided params
func resolveMacros(aTemplate *template.Template, params interface{}) (string, error) {
	strBuf := bytes.Buffer{}

	err := aTemplate.Execute(&strBuf, params)
	if err != nil {
		return "", err
	}
	res := strBuf.String()
	return res, nil
}
