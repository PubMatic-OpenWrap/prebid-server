package processor


import (
	"bytes"
	"strings"

	"github.com/prebid/prebid-server/config"
)

// stringIndexBasedProcessor implements macro processor interface with string indexing approach
type stringIndexBasedProcessor struct {
	cfg config.MacroProcessorConfig
}

const (
	macroEscapeSuffix    string = `_ESC`
	macroEscapeSuffixLen int    = len(macroEscapeSuffix)
)

func (p *stringIndexBasedProcessor) Replace(url string, macroProvider Provider) (string, error) {
	macroPrefix := p.cfg.Delimiter
	macroSuffix := p.cfg.Delimiter
	macroPrefixLen := len(macroPrefix)
	macroSuffixLen := len(macroSuffix)
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

		value := macroProvider.GetMacro(key)
		if value != "" {
			out.WriteString(value)
			currIndex = end
		} else {
			out.WriteByte(macroPrefix[0])
			currIndex = start + 1
		}
	}
	return out.String(), nil
}
