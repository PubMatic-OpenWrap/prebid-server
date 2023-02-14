package macros

import (
	"bytes"
	"strings"
)

const (
	macroPrefix          = `##` //macro prefix can not be empty
	macroSuffix          = `##` //macro suffix can not be empty
	macroEscapeSuffix    = `_ESC`
	macroPrefixLen       = len(macroPrefix)
	macroSuffixLen       = len(macroSuffix)
	macroEscapeSuffixLen = len(macroEscapeSuffix)
	startIndex           = -1
)

func Replace(eventURL string, macroValues map[string]string) (string, error) {
	var out bytes.Buffer
	pos, start, end, size := 0, 0, 0, len(eventURL)

	for pos < size {
		//find macro prefix index
		if start = strings.Index(eventURL[pos:], macroPrefix); start == startIndex {
			//[prefix_not_found] append remaining string to response
			out.WriteString(eventURL[pos:])

			//macro prefix not found
			break
		}

		//prefix index w.r.t original string
		start = start + pos

		//append non macro prefix content
		out.WriteString(eventURL[pos:start])

		if (end - macroSuffixLen) <= (start + macroPrefixLen) {
			//string contains {{TEXT_{{MACRO}} -> it should replace it with{{TEXT_MACROVALUE
			//find macro suffix index
			if end = strings.Index(eventURL[start+macroPrefixLen:], macroSuffix); -1 == end {
				//[suffix_not_found] append remaining string to response
				out.WriteString(eventURL[start:])

				// We Found First %% and Not Found Second %% But We are in between of string
				break
			}

			end = start + macroPrefixLen + end + macroSuffixLen
		}

		//get actual macro key by removing macroPrefix and macroSuffix from key itself
		key := eventURL[start+macroPrefixLen : end-macroSuffixLen]

		//process macro
		// value, found := mp.processKey(key)
		value, found := macroValues[key]
		if found {
			out.WriteString(value)
			pos = end
		} else {
			out.WriteByte(macroPrefix[0])
			pos = start + 1
		}
		//glog.Infof("\nSearch[%d] <start,end,key>: [%d,%d,%s]", count, start, end, key)
	}
	// glog.V(3).Infof("[MACRO]:in:[%s] replaced:[%s]", in, )
	return out.String(), nil
}
