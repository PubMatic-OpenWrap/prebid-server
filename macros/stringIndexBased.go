package macros

import (
	"bytes"
	"strings"
)

// stringIndexBased implements macro processor interface with string indexing approach
type stringIndexBased struct {
	cfg Config
}

const (
	macroPrefix          string = `##` //macro prefix can not be empty
	macroSuffix          string = `##` //macro suffix can not be empty
	macroEscapeSuffix    string = `_ESC`
	macroPrefixLen       int    = len(macroPrefix)
	macroSuffixLen       int    = len(macroSuffix)
	macroEscapeSuffixLen int    = len(macroEscapeSuffix)
)

func (p *stringIndexBased) Replace(url string, macroValues map[string]string) (string, error) {
	var out bytes.Buffer
	currIndex, start, end, size := 0, 0, 0, len(url)

	for currIndex < size {
		//find macro prefix index
		if start = strings.Index(url[currIndex:], macroPrefix); start == -1 {
			//[prefix_not_found] append remaining string to response
			out.WriteString(url[currIndex:])
			//macro prefix not found
			break
		}

		//prefix index w.r.t original string
		start = start + currIndex

		//append non macro prefix content
		out.WriteString(url[currIndex:start])

		if (end - macroSuffixLen) <= (start + macroPrefixLen) {
			//string contains {{TEXT_{{MACRO}} -> it should replace it with{{TEXT_MACROVALUE
			//find macro suffix index
			if end = strings.Index(url[start+macroPrefixLen:], macroSuffix); end == -1 {
				//[suffix_not_found] append remaining string to response
				out.WriteString(url[start:])

				// We Found First %% and Not Found Second %% But We are url between of string
				break
			}

			end = start + macroPrefixLen + end + macroSuffixLen
		}

		//get actual macro key by removing macroPrefix and macroSuffix from key itself
		key := url[start+macroPrefixLen : end-macroSuffixLen]

		//process macro
		// value, found := mp.processKey(key)
		value, found := macroValues[key]
		if found {
			out.WriteString(value)
			currIndex = end
		} else {
			out.WriteByte(macroPrefix[0])
			currIndex = start + 1
		}
	}
	return out.String(), nil
}

func (p *stringIndexBased) AddTemplates([]string) {}
