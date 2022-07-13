package macros

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"sync"
)

type StringIndexCached struct {
	Processor
	templates map[string]strMetaTemplate
	dup       int
	sync.RWMutex
}

type strMetaTemplate struct {
	// macroSIndexMap  map[string]int
	// sIndexMacrosMap map[int]string
	// macroSIndexList []*string // ordered list of indices (useful for replace method)
	indices     []int
	macroLength []int
}

func (p *StringIndexCached) initTemplate() {
	delim := p.Cfg.delimiter
	p.templates = make(map[string]strMetaTemplate)
	if nil == p.Processor.Cfg.Templates || len(p.Processor.Cfg.Templates) == 0 {
		panic("No input templates")
	}
	for _, str := range p.Processor.Cfg.Templates {
		p.templates[str] = constructTemplate(str, delim)
	}
	fmt.Printf("Macroprocessor initialized %d templates\n", len(p.templates))
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
		ei := strings.Index(str[msi:], delim) // ending delimiter
		if ei == -1 {
			break
		}
		ei = ei + msi // offset adjustment (delimiter inclusive)
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

func (p *StringIndexCached) Replace(str string, macroValues map[string]string) (string, error) {
	tmplt := p.templates[str]
	var result bytes.Buffer
	// iterate over macros startindex list to get position where value should be put
	// http://tracker.com?macro_1=##PBS_EVENTTYPE##&macro_2=##PBS_GDPRCONSENT##&custom=##PBS_MACRO_profileid##&custom=##shri##
	s := 0
	delimLen := len(p.Cfg.delimiter)
	for i, index := range tmplt.indices {
		// macro := tmplt.sIndexMacrosMap[index]
		macro := str[index+delimLen : tmplt.macroLength[i]]
		// copy prev part
		result.WriteString(str[s:index])
		if value, found := macroValues[macro]; found {
			// replace macro with value
			if p.Cfg.valueConfig.UrlEscape {
				value = url.QueryEscape(value)
			}
			result.WriteString(value)
		}
		s = index + len(macro) + len(p.Cfg.delimiter) + len(p.Cfg.delimiter)
	}
	result.WriteString(str[s:])
	return result.String(), nil
}

func (p *StringIndexCached) AddTemplates(templates ...string) {
	p.dup = 0
	for _, str := range templates {
		p.RLock()
		_, ok := p.templates[str]
		p.RUnlock()

		if !ok {
			p.Lock()
			p.templates[str] = constructTemplate(str, p.Cfg.delimiter)
			fmt.Println("Template constructed")
			p.Unlock()
		} else {
			p.dup++
		}
	}
	if p.dup == len(templates) {
		fmt.Printf("Templates already processed\n")
	}
	fmt.Printf("Macroprocessor initialized %d templates\nDuplicate=%d\n", len(p.templates), p.dup)
}
