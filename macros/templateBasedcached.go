package macros

import (
	"bytes"
	"fmt"
	"regexp"
	"text/template"
)

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
		tmpl := template.New("macro_replace")
		tmpl.Option("missingkey=zero")
		tmpl.Delims(delimiter, delimiter)
		// collect all macros based on delimiters
		regex := fmt.Sprintf("%s(.*?)%s", delimiter, delimiter)
		re := regexp.MustCompile(regex)
		// Example
		// http://tracker.com?macro_1=##PBS_EVENTTYPE##&macro_2=##PBS_GDPRCONSENT##&custom=##PBS_MACRO_profileid##&custom=##shri##
		// ##(.*?)##
		// Group 0 => ##PBS_EVENTTYPE##, ##PBS_GDPRCONSENT#
		// Group 1 => PBS_EVENTTYPE, PBS_GDPRCONSENT
		// We are using #1. because we want '.' as a prefix
		replacedStr := re.ReplaceAllString(url, "##.$1##")
		tmpl, err := tmpl.Parse(replacedStr)
		if err != nil {
			panic(err)
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
