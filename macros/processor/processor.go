package processor

import (
	"github.com/prebid/prebid-server/config"
)

type Processor interface {
	// Replace the macros and returns replaced string
	// if any error the error will be returned
	Replace(url string, macroProvider Provider) (string, error)
}

var processor Processor

// NewProcessor will return instance of macro processor
func NewProcessor(cfg config.MacroProcessorConfig) Processor {

	if cfg.Delimiter == "" {
		cfg.Delimiter = "##"
	}

	switch cfg.ProcessorType {
	case config.StringIndexCacheProcessor:
		processor = &stringIndexCachedProcessor{cfg: cfg}
	case config.TemplateCacheProcessor:
		processor = &templateBasedCached{cfg: cfg}
	default:
		processor = &emptyProcessor{}
	}

	return processor
}

func GetMacroProcessor() Processor {
	return processor
}
