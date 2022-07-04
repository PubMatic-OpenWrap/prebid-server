package macros

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
)

type StringIndexCached struct {
	Processor
	templates map[string]strMetaTemplate
}

type strMetaTemplate struct {
	// macroSIndexMap  map[string]int
	sIndexMacrosMap map[int]string
	// macroSIndexList []*string // ordered list of indices (useful for replace method)
	indices []int
}

func (p *StringIndexCached) initTemplate() {
	delim := p.Cfg.delimiter
	p.templates = make(map[string]strMetaTemplate)
	if nil == p.Processor.Cfg.Templates || len(p.Processor.Cfg.Templates) == 0 {
		panic("No input templates")
	}
	for _, str := range p.Processor.Cfg.Templates {
		si := 0
		tmplt := strMetaTemplate{
			sIndexMacrosMap: make(map[int]string),
			indices:         []int{},
		}
		for true {
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
			tmplt.sIndexMacrosMap[si] = str[msi:mei]
			tmplt.indices = append(tmplt.indices, si)
			si = ei + 1
			if si >= len(str) {
				break
			}
		}
		p.templates[str] = tmplt
	}
	fmt.Printf("Macroprocessor initialized %d templates", len(p.templates))
}
func (p *StringIndexCached) Replace(str string, macroValues map[string]string) (string, error) {
	tmplt := p.templates[str]
	var result bytes.Buffer
	// iterate over macros startindex list to get position where value should be put
	// http://tracker.com?macro_1=##PBS_EVENTTYPE##&macro_2=##PBS_GDPRCONSENT##&custom=##PBS_MACRO_profileid##&custom=##shri##
	s := 0
	for _, index := range tmplt.indices {
		macro := tmplt.sIndexMacrosMap[index]
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
