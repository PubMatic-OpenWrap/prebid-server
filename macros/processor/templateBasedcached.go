package processor

import (
	"bytes"
	"fmt"
	"regexp"
	"sync"
	"text/template"

	"github.com/prebid/prebid-server/config"
)

const (
	templateName   = "macro_replace"
	templateOption = "missingkey=zero"
)

type templateWrapper struct {
	template *template.Template
	keys     []string
}

// templateBasedCache implements macro processor interface with text/template caching approach
// new template will be cached for each event url per request.
type templateBasedCached struct {
	templates map[string]*templateWrapper
	cfg       config.MacroProcessorConfig
	sync.RWMutex
}

func (processor *templateBasedCached) Replace(url string, macroProvider Provider) (string, error) {
	tmplt := processor.getTemplates(url)
	if tmplt == nil {
		return "", fmt.Errorf("failed to add template for url: %s", url)
	}
	return resolveMacros(tmplt.template, macroProvider.GetAllMacros(tmplt.keys))
}

func (processor *templateBasedCached) getTemplates(url string) *templateWrapper {
	var (
		tmplate *templateWrapper
		ok      bool
	)
	processor.RLock()
	tmplate, ok = processor.templates[url]
	processor.RUnlock()

	if !ok {
		processor.Lock()

		delimiter := processor.cfg.Delimiter
		tmpl := template.New(templateName)
		tmpl.Option(templateOption)
		tmpl.Delims(delimiter, delimiter)
		// collect all macros based on delimiters
		regex := fmt.Sprintf("%s(.*?)%s", delimiter, delimiter)
		re := regexp.MustCompile(regex)
		subStringMatches := re.FindAllStringSubmatch(url, -1)

		keys := make([]string, len(subStringMatches))
		for indx, value := range subStringMatches {
			keys[indx] = value[1]
		}
		replacedStr := re.ReplaceAllString(url, delimiter+".$1"+delimiter)
		tmpl, err := tmpl.Parse(replacedStr)
		if err != nil {
			return nil
		}
		tmplWrapper := &templateWrapper{
			template: tmpl,
			keys:     keys,
		}
		tmplate = tmplWrapper
		processor.templates[url] = tmplate
		processor.Unlock()
	}

	return tmplate
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
