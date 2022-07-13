package config

import "github.com/prebid/prebid-server/macros"

// macroProcessor a global instance which can replace the macros
// with vaules
var macroProcessor macros.IProcessor

func GetMacroProcessor() macros.IProcessor {
	return macroProcessor
}
