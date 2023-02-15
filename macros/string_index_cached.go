package macros

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"sync"
)

type stringIndexCached struct {
	cfg       Config
	templates map[string]strMetaTemplate
	dup       int
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

func (processor *stringIndexCached) Replace(str string, macroValues map[string]string) (string, error) {
	tmplt := processor.templates[str]
	var result bytes.Buffer
	// iterate over macros startindex list to get position where value should be put
	// http://tracker.com?macro_1=##PBS_EVENTTYPE##&macro_2=##PBS_GDPRCONSENT##&custom=##PBS_MACRO_profileid##&custom=##shri##
	s := 0
	delimLen := len(processor.cfg.Delimiter)
	for i, index := range tmplt.indices {
		// macro := tmplt.sIndexMacrosMap[index]
		macro := str[index+delimLen : tmplt.macroLength[i]]
		// copy prev part
		result.WriteString(str[s:index])
		if value, found := macroValues[macro]; found {
			// replace macro with value
			if processor.cfg.valueConfig.UrlEscape {
				value = url.QueryEscape(value)
			}
			result.WriteString(value)
		}
		s = index + len(macro) + len(processor.cfg.Delimiter) + len(processor.cfg.Delimiter)
	}
	result.WriteString(str[s:])
	return result.String(), nil
}

func (processor *stringIndexCached) AddTemplates(templates []string) {
	processor.dup = 0
	for _, str := range templates {
		processor.RLock()
		_, ok := processor.templates[str]
		processor.RUnlock()

		if !ok {
			processor.Lock()
			processor.templates[str] = constructTemplate(str, processor.cfg.Delimiter)
			fmt.Println("Template constructed")
			processor.Unlock()
		} else {
			processor.dup++
		}
	}
	if processor.dup == len(templates) {
		fmt.Printf("Templates already processed\n")
	}
	fmt.Printf("Macroprocessor initialized %d templates\nDuplicate=%d\n", len(processor.templates), processor.dup)
}
