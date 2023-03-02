package processor

import (
	"bytes"
	"strings"
	"sync"

	"github.com/prebid/prebid-server/config"
)

type stringBasedProcessor struct {
	cfg       config.MacroProcessorConfig
	templates map[string]urlMetaTemplate
	sync.RWMutex
}

func newStringBasedProcessor(cfg config.MacroProcessorConfig) *stringBasedProcessor {
	return &stringBasedProcessor{
		cfg:       cfg,
		templates: make(map[string]urlMetaTemplate),
	}
}

type urlMetaTemplate struct {
	indices     []int
	macroLength []int
}

func constructTemplate(url string, delimiter string) urlMetaTemplate {
	currentIndex := 0
	tmplt := urlMetaTemplate{
		indices:     []int{},
		macroLength: []int{},
	}
	delimiterLen := len(delimiter)
	for {
		currentIndex = currentIndex + strings.Index(url[currentIndex:], delimiter)
		if currentIndex == -1 {
			break
		}
		middleIndex := currentIndex + delimiterLen
		endingIndex := strings.Index(url[middleIndex:], delimiter) // ending Delimiter
		if endingIndex == -1 {
			break
		}
		endingIndex = endingIndex + middleIndex // offset adjustment (Delimiter inclusive)
		macroLength := endingIndex              // just for readiability
		tmplt.indices = append(tmplt.indices, currentIndex)
		tmplt.macroLength = append(tmplt.macroLength, macroLength)
		currentIndex = endingIndex + 1
		if currentIndex >= len(url) {
			break
		}
	}
	return tmplt
}

func (processor *stringBasedProcessor) Replace(url string, macroProvider Provider) (string, error) {
	tmplt := processor.getTemplate(url)

	var result bytes.Buffer
	// iterate over macros startindex list to get position where value should be put
	// http://tracker.com?macro_1=##PBS_EVENTTYPE##&macro_2=##PBS_GDPRCONSENT##&custom=##PBS_MACRO_profileid##&custom=##shri##
	currentIndex := 0
	delimLen := len(processor.cfg.Delimiter)
	for i, index := range tmplt.indices {
		macro := url[index+delimLen : tmplt.macroLength[i]]
		// copy prev part
		result.WriteString(url[currentIndex:index])
		value := macroProvider.GetMacro(macro)
		if value != "" {
			result.WriteString(value)
		}
		currentIndex = index + len(macro) + 2*delimLen
	}
	result.WriteString(url[currentIndex:])
	return result.String(), nil
}

func (processor *stringBasedProcessor) getTemplate(url string) urlMetaTemplate {
	var (
		template urlMetaTemplate
		ok       bool
	)
	processor.RLock()
	template, ok = processor.templates[url]
	processor.RUnlock()

	if !ok {
		processor.Lock()
		template = constructTemplate(url, processor.cfg.Delimiter)
		processor.templates[url] = template
		processor.Unlock()
	}
	return template
}

// Test cases:
// 1) Verify NewProcessor returns the instance of stringBasedProcessor if ProcessorType  = 1 in config
// 2) Verify NewProcessor returns the instance of stringBasedProcessor if ProcessorType  = 2 in config
// 3) Verify NewProcessor returns the instance of emptyProcessor if invalid/ ProcessorType = 0 is in config
// 4) Verify NewProcessor instance has delimiter = "##" if default config is used
// 5) Verify NewProcessor instance has delimiter same as delimiter added in config

// 6) Verify NewProvider returns macroProvider instance with all the request level and custom macros set
// 7) Verify customs macro values are truncated if values are beyond 100 chars
// 8) Verify SetContext resets the  existing bid and impression level macros and adds the new the bid and impression level macros.
// 9) Verify GetMacro returns the value of the given macro key when key is present
// 10) Verify GetAllMacros returns the ""(empty) value when the macros value is not present in request
// 11) Verify GetAllMacros returns the value of the all macros. Value is ""(empty if macro value not present in request)

// 12) Verfiy for StringBased Approach, all the macro value are replaced in tracker url. For macros, for which 
//value are not present in request/bid. Emtpy value will be used to replace macro.
 
// 13) Verfiy for TemplateBased Approach, all the macro value are replaced in tracker url. For macros, for which 
//value are not present in request/bid. Emtpy value will be used to replace macro.

// 14) Verify for TemplateBased Approach, if template creation fails, Replace function should return error.
// 15) Verify for TemplateBased Approach, if template execution fails, Replace function returns error.
// 16) Verify for StringBased approach, if no macros is present in url, Replace function should return url as it is.
// 16) Verify for TemplateBased approach, if no macros is present in url, Replace function should return url as it is.
