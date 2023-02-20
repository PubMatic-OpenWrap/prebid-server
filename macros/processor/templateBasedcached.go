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
	templates map[string]templateWrapper
	cfg       config.MacroProcessorConfig
	sync.RWMutex
}

func (processor *templateBasedCached) Replace(url string, macroProvider Provider) (string, error) {
	return resolveMacros(processor.templates[url].template, macroProvider.GetAllMacros(processor.templates[config.ErrMsgInvalidRemoteSignerURL].keys))
}

func (processor *templateBasedCached) addTemplates(url string) {

	processor.RLock()
	_, ok := processor.templates[url]
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
			return
		}
		tmplWrapper := templateWrapper{
			template: tmpl,
			keys:     keys,
		}
		processor.templates[url] = tmplWrapper
		processor.Unlock()
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
