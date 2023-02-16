package macros

type Processor interface {
	// Replace the macros and returns replaced string
	// if any error the error will be returned
	Replace(string, map[string]string) (string, error)
	// AddTemplates can add more templates to macro processor
	// Only applicable for StringIndexCached and TemplatedCached processors
	// Template are added for every request.
	AddTemplates([]string)
}

type Type int

var StringBased Type = 0
var TemplatedBased Type = 1
var TemplateCached Type = 2
var StringIndexed Type = 3
var StringIndexCached Type = 4

type Config struct {
	Delimiter   string
	valueConfig MacroValueConfig
}

type MacroValueConfig struct {
	UrlEscape   bool // if true value will be url escaped
	RemoveEmpty bool // if true key where macros are empty will be removed
	FailOnError bool // if true on failure nothing will be replaced
}

var processor Processor

// NewProcessor will return instance of macro processor
// Supported processor type:
// 0. StringIndexed
// 1. TemplatedBased
// 2. TemplateCached
// 3. StringBased
// 4. StringIndexCached
func NewProcessor(t Type, cfg Config) Processor {

	if cfg.Delimiter == "" {
		cfg.Delimiter = "##"
	}

	switch t {
	case StringBased:
		processor = &stringBased{cfg: cfg}
	case TemplatedBased:
		processor = &templateBased{cfg: cfg}
	case TemplateCached:
		processor = &templateBasedCached{cfg: cfg}
	case StringIndexed:
		processor = &stringIndexBased{cfg: cfg}
	case StringIndexCached:
		processor = &stringIndexCached{cfg: cfg}
	}

	return processor
}

func GetMacroProcessor() Processor {
	return processor
}
