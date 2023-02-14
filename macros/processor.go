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
)

// Replace replaces event macros in vast event url
func Replace(eventURL string, macroValues map[string]string) string {
	var out bytes.Buffer
	pos, start, end, size := 0, 0, 0, len(eventURL)

	for pos < size {

		if start = strings.Index(eventURL[pos:], macroPrefix); start == -1 {
			out.WriteString(eventURL[pos:])
			break
		}

		start = start + pos
		out.WriteString(eventURL[pos:start])

		if (end - macroSuffixLen) <= (start + macroPrefixLen) {
			if end = strings.Index(eventURL[start+macroPrefixLen:], macroSuffix); end == -1 {
				out.WriteString(eventURL[start:])
				break
			}

			end = start + macroPrefixLen + end + macroSuffixLen
		}

		key := eventURL[start+macroPrefixLen : end-macroSuffixLen]

		value, found := macroValues[key]
		if found {
			out.WriteString(value)
			pos = end
		} else {
			out.WriteByte(macroPrefix[0])
			pos = start + 1
		}
	}
	return out.String()
}
