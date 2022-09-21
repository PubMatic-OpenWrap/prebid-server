package config

import (
	"github.com/prebid/prebid-server/macros"
)

// macroProcessor a global instance which can replace the macros
// with vaules
var macroProcessor macros.IProcessor

func GetMacroProcessor() macros.IProcessor {
	return macroProcessor
}

func UpdateMacroProcessor(events Events) {
	// add account specific templates to macroprocessor
	// TODO: Check if multiple publishers have same template URL
	// e.g.
	// pub1 : http://example.com?k=${val} - T1
	// pub2 : http://example.com?k=${val} /
	// TODO: Do we need to introduce account context to maintain
	// templates seperately though they might be same??
	templates := make([]string, 0)
	templates = append(templates, events.DefaultURL)
	for _, vEvent := range events.VASTEvents {
		if vEvent.URLs != nil {
			templates = append(templates, vEvent.URLs...)
		}
	}
	GetMacroProcessor().AddTemplates(templates...)
}
