package macros

import (
	"sort"
	"strings"
)

type StringIndexCached struct {
	Processor
	templates map[string]strMetaTemplate
}

type strMetaTemplate struct {
	macroSIndexMap  map[string]int
	sIndexMacrosMap map[int]string
	// macroSIndexList []*string // ordered list of indices (useful for replace method)
	indices []int
}

func (p *StringIndexCached) initTemplate() {
	delim := p.Cfg.delimiter
	p.templates = make(map[string]strMetaTemplate)
	if nil == p.Cfg.templates || len(p.Cfg.templates) == 0 {
		panic("No input templates")
	}
	for _, str := range p.Cfg.templates {
		si := 0
		tmplt := strMetaTemplate{
			macroSIndexMap:  make(map[string]int),
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
			// fmt.Println(str[msi:mei])
			// cache macro and its start index
			tmplt.macroSIndexMap[str[msi:mei]] = si
			tmplt.sIndexMacrosMap[si] = str[msi:mei]
			tmplt.indices = append(tmplt.indices, si)
			si = ei + 1
			if si >= len(str) {
				break
			}
		}

		// form macroSIndexList  - and ordered list
		sort.Ints(tmplt.indices)
		// create index array with len = max value present at end sorted indices array
		// tmplt.macroSIndexList = make([]*string, tmplt.indices[len(tmplt.indices)-1]+1)
		// for _, index := range tmplt.indices {
		// 	val := tmplt.sIndexMacrosMap[index]
		// 	// tmplt.macroSIndexList[index] = &val
		// }
		// store template for str
		p.templates[str] = tmplt
	}
}
func (p *StringIndexCached) Replace(str string, macroValues map[string]string) (string, error) {
	tmplt := p.templates[str]
	res := ""
	// for macro, value := range macroValues {
	// 	if si, found := tmplt.macroSIndexMap[macro]; found {
	// 		res += str[i : si-1]
	// 		res += value
	// 	}
	// }

	// iterate over macros startindex list to get position where value should be put
	// http://tracker.com?macro_1=##PBS_EVENTTYPE##&macro_2=##PBS_GDPRCONSENT##&custom=##PBS_MACRO_profileid##&custom=##shri##
	s := 0
	for _, index := range tmplt.indices {
		macro := tmplt.sIndexMacrosMap[index]
		// copy prev part
		res += str[s:index]
		if value, found := macroValues[macro]; found {
			// replace macro with value
			res += value
			s = index + len(macro) + len(p.Cfg.delimiter) + len(p.Cfg.delimiter)
		}
	}

	return res, nil
}
