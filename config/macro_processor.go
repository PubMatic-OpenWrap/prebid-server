package config

import (
	"sync"

	"github.com/prebid/prebid-server/macros"
)

// macroProcessor a global instance which can replace the macros
// with vaules
var macroProcessor macros.IProcessor
var once sync.Once

func GetMacroProcessor() macros.IProcessor {
	once.Do(func() {
		delimiter := "##"

		macroProcessor, _ = macros.NewProcessor(macros.STRING_INDEX_CACHED, macros.Config{
			Delimiter: delimiter,
			Templates: []string{},
		})
	})
	return macroProcessor
}
