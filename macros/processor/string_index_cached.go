package processor

import (
	"bytes"
	"strings"
	"sync"

	"github.com/prebid/prebid-server/config"
)

type stringIndexCachedProcessor struct {
	cfg       config.MacroProcessorConfig
	templates map[string]strMetaTemplate
	sync.RWMutex
}

type strMetaTemplate struct {
	indices     []int
	macroLength []int
}

func constructTemplate(str string, delim string) strMetaTemplate {
	si := 0
	tmplt := strMetaTemplate{
		// sIndexMacrosMap: make(map[int]string),
		indices:     []int{},
		macroLength: []int{},
	}
	for {
		si = si + strings.Index(str[si:], delim)
		if si == -1 {
			break
		}
		msi := si + len(delim)
		ei := strings.Index(str[msi:], delim) // ending Delimiter
		if ei == -1 {
			break
		}
		ei = ei + msi // offset adjustment (Delimiter inclusive)
		mei := ei     // just for readiability
		// cache macro and its start index
		// tmplt.sIndexMacrosMap[si] = str[msi:mei]
		tmplt.indices = append(tmplt.indices, si)
		tmplt.macroLength = append(tmplt.macroLength, mei)
		si = ei + 1
		if si >= len(str) {
			break
		}
	}
	return tmplt
}

func (processor *stringIndexCachedProcessor) Replace(url string, macroProvider Provider) (string, error) {
	processor.addTemplate(url)
	tmplt := processor.templates[url]
	var result bytes.Buffer
	// iterate over macros startindex list to get position where value should be put
	// http://tracker.com?macro_1=##PBS_EVENTTYPE##&macro_2=##PBS_GDPRCONSENT##&custom=##PBS_MACRO_profileid##&custom=##shri##
	s := 0
	delimLen := len(processor.cfg.Delimiter)
	for i, index := range tmplt.indices {
		// macro := tmplt.sIndexMacrosMap[index]
		macro := url[index+delimLen : tmplt.macroLength[i]]
		// copy prev part
		result.WriteString(url[s:index])
		value := macroProvider.GetMacro(macro)
		if value != "" {
			result.WriteString(value)
		}
		s = index + len(macro) + len(processor.cfg.Delimiter) + len(processor.cfg.Delimiter)
	}
	result.WriteString(url[s:])
	return result.String(), nil
}

func (processor *stringIndexCachedProcessor) addTemplate(url string) {
	processor.RLock()
	_, ok := processor.templates[url]
	processor.RUnlock()

	if !ok {
		processor.Lock()
		processor.templates[url] = constructTemplate(url, processor.cfg.Delimiter)
		processor.Unlock()
	}
}
