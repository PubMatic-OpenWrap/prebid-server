package tagbidder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/golang/glog"
)

const (
	macroFormatChar      byte   = '%'
	macroFormat          string = `%%`
	macroFormatLen       int    = len(macroFormat)
	macroEOF             int    = -1
	customMacroFormat    string = `%%VAR_%s%%`
	macroEscapeSuffix    string = `_ESC`
	macroEscapeSuffixLen int    = len(macroEscapeSuffix)
)

//MacroProcessor struct to hold openrtb request and cache values
type MacroProcessor struct {
	bidder     IBidderMacro
	mapper     mapper
	macroCache map[string]string
}

//NewMacroProcessor will process macro's of openrtb bid request
func NewMacroProcessor(mapper mapper) *MacroProcessor {
	return &MacroProcessor{
		mapper:     mapper,
		macroCache: make(map[string]string),
	}
}

//SetMacro : Adding Custom Macro Manually
func (mp *MacroProcessor) SetMacro(key, value string) {
	mp.macroCache[key] = value
}

//GetCutsomMacroKey : Returns Custom Macro Keys
func (mp *MacroProcessor) GetCutsomMacroKey(key string) string {
	return fmt.Sprintf(customMacroFormat, key)
}

//Process : Substitute macros in input string
func (mp *MacroProcessor) Process(in string) (response string) {
	var out bytes.Buffer
	pos, start, end, size, nEscaping := 0, 0, 0, len(in), 0
	skip, found := false, false

	for pos < size {
		if skip == false {
			if start = strings.Index(in[pos:], macroFormat); -1 == start {
				out.WriteString(in[pos:])
				// Normal Exit
				//glog.Infof("\n[EXIT=1]")
				break
			}
			start = start + pos
			out.WriteString(in[pos:start])
		}

		if end = strings.Index(in[start+macroFormatLen:], macroFormat); -1 == end {
			out.WriteString(in[start:])
			// We Found First %% and Not Found Second %% But We are in between of string
			//glog.Infof("\n[EXIT=2]")
			break
		}
		end = start + end + (macroFormatLen << 1)

		key := in[start:end]
		//glog.Infof("\nSearch[%d] <start,end,key>: [%d,%d,%s]", count, start, end, key)
		if value, ok := mp.macroCache[key]; ok {
			//Found Key and Value: Replace Macro Value
			//glog.Infof("\n<Start,End,Token,Value> : <%d,%d,%s,%s>", start, end, key, value)
			out.WriteString(value)
			pos = end
			skip = false
		} else {
			found = false
			nEscaping = 0
			tmpKey := key
			for {
				if valueCallback, ok := mp.mapper[tmpKey]; ok {
					// Found Callback Function for Key
					if value := valueCallback.callback(mp.bidder, tmpKey); len(value) > 0 {
						if nEscaping > 0 {
							//Escaping string nEscaping times
							value = escape(value, nEscaping)
						}

						if valueCallback.cached {
							// Get Value and add it in macro list
							mp.macroCache[key] = value
						}

						// Replace it in MACRO
						out.WriteString(value)
						pos = end
						skip = false
						found = true
						break
					}
				} else if strings.HasSuffix(tmpKey, macroEscapeSuffix) {
					//escaping macro found
					tmpKey = tmpKey[0 : len(tmpKey)-macroEscapeSuffixLen]
					nEscaping++
					continue
				}
				break
			}

			if !found {
				if in[start+macroFormatLen] == macroFormatChar {
					// Next Character is % then end = start+1, and write '%' in string
					end = start + 1
				} else {
					// Not Found Key as well as ValueCallback Function
					end = end - macroFormatLen
				}
				out.WriteString(in[start:end])
				pos, start = end, end
				skip = true
			}
		}
	}
	response = out.String()
	glog.V(3).Infof("[MACRO]:in:[%s]\nreplaced:[%s]\n", in, response)

	return
}

//Dump : will print all cached macro and its values
func (mp *MacroProcessor) Dump() {
	if glog.V(3) {
		cacheStr, _ := json.Marshal(mp.macroCache)
		glog.Infof("[MACRO]: Map:[%s]", string(cacheStr))
	}
}

func escape(str string, n int) string {
	for ; n > 0; n-- {
		str = url.QueryEscape(str)
	}
	return str[:]
}
