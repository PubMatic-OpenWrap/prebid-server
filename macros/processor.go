package macros

import "errors"

type Processor struct {
	IProcessor
	Cfg Config
}

type IProcessor interface {
	// Replace the macros and returns replaced string
	// if any error the error will be returned
	Replace(str string) (string, error)
}

type Type int

var STRING_BASED Type = 0
var TEMPLATE_BASED Type = 1

type Config struct {
	delimiter   string
	valueConfig MacroValueConfig
	macroValues map[string]string
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
		p.init0()
		return &p, nil
	}
	return nil, errors.New("Invalid Processor Type")
}
