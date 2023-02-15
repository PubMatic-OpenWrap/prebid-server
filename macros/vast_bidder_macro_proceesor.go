package macros

import (
	"bytes"
	"strings"
)

type VastBidderBased struct {
	Processor
}

const (
	macroPrefix          string = `##` //macro prefix can not be empty
	macroSuffix          string = `##` //macro suffix can not be empty
	macroEscapeSuffix    string = `_ESC`
	macroPrefixLen       int    = len(macroPrefix)
	macroSuffixLen       int    = len(macroSuffix)
	macroEscapeSuffixLen int    = len(macroEscapeSuffix)
)

func (p *VastBidderBased) Replace(in string, macroValues map[string]string) (string, error) {
	var out bytes.Buffer
	pos, start, end, size := 0, 0, 0, len(in)

	for pos < size {
		//find macro prefix index
		if start = strings.Index(in[pos:], macroPrefix); -1 == start {
			//[prefix_not_found] append remaining string to response
			out.WriteString(in[pos:])

			//macro prefix not found
			break
		}

		//prefix index w.r.t original string
		start = start + pos

		//append non macro prefix content
		out.WriteString(in[pos:start])

		if (end - macroSuffixLen) <= (start + macroPrefixLen) {
			//string contains {{TEXT_{{MACRO}} -> it should replace it with{{TEXT_MACROVALUE
			//find macro suffix index
			if end = strings.Index(in[start+macroPrefixLen:], macroSuffix); -1 == end {
				//[suffix_not_found] append remaining string to response
				out.WriteString(in[start:])

				// We Found First %% and Not Found Second %% But We are in between of string
				break
			}

			end = start + macroPrefixLen + end + macroSuffixLen
		}

		//get actual macro key by removing macroPrefix and macroSuffix from key itself
		key := in[start+macroPrefixLen : end-macroSuffixLen]

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
