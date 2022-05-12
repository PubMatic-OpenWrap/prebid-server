package macros

import (
	"bytes"
	"fmt"
	"regexp"
	"text/template"
)

type TemplateBased struct {
	Processor
	templates map[string]*template.Template
}

func (p *TemplateBased) Replace(str string, macroValues map[string]string) (string, error) {
	return replaceTemplateBased(p.templates[str], macroValues)
}

func (p *TemplateBased) init0(templates []string) {
	delimiter := p.Cfg.delimiter
	p.templates = make(map[string]*template.Template, len(p.Cfg.templates))
	for _, str := range templates {
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
		replacedStr := re.ReplaceAllString(str, "##.$1##")
		tmpl, err := tmpl.Parse(replacedStr)
		if err != nil {
			panic(err)
		}
		p.templates[str] = tmpl
	}
}

func replaceTemplateBased(tmpl *template.Template, macroValueMap map[string]string) (string, error) {
	// http://abc.co?key=##mac##

	return resolveMacros(tmpl, macroValueMap)
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
