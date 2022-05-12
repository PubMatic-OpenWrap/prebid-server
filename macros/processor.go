package macros

import "errors"

type Processor struct {
	IProcessor
	Cfg Config
}

type IProcessor interface {
	// Replace the macros and returns replaced string
	// if any error the error will be returned
	Replace(string, map[string]string) (string, error)
}

type Type int

var STRING_BASED Type = 0
var TEMPLATE_BASED Type = 1

// following types are temporary kept for benchmarking go template for macro replacement
var TEMPLATE_BASED_INIT_ALWAYS Type = 2
var VAST_BIDDER_MACRO_PROCESSOR Type = 3
var STRING_INDEX_CACHED Type = 4

type Config struct {
	delimiter   string
	valueConfig MacroValueConfig
	templates   []string // Required by TEMPLATE_BASED processors
}

type MacroValueConfig struct {
	UrlEscape   bool // if true value will be url escaped
	RemoveEmpty bool // if true key where macros are empty will be removed
	FailOnError bool // if true on failure nothing will be replaced
}

func NewProcessor(t Type, config Config) (IProcessor, error) {

	if config.delimiter == "" {
		config.delimiter = "##"
	}

	switch t {
	case STRING_BASED:
		p := StringBased{}
		p.Cfg = config
		return &p, nil

	case TEMPLATE_BASED:
		p := TemplateBased{}
		if nil == config.templates || len(config.templates) == 0 {
			// return nil, errors.New("Missing templates")
			panic("Missing config.templates")
		}
		p.Cfg = config
		p.init0(p.Cfg.templates)
		return &p, nil

	case TEMPLATE_BASED_INIT_ALWAYS:
		p := TemplateBasedInitAlways{}
		p.Cfg = config
		return &p, nil

	case VAST_BIDDER_MACRO_PROCESSOR:
		p := VastBidderBased{}
		p.Cfg = config
		return &p, nil

	case STRING_INDEX_CACHED:
		p := StringIndexCached{}
		p.Cfg = config
		p.initTemplate()
		return &p, nil

	}

	return nil, errors.New("Invalid Processor Type")
}
