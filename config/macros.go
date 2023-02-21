package config

type ProcessorType int

const (
	EmptyProcessor            = 0
	StringIndexCacheProcessor = 1
	TemplateCacheProcessor    = 2
)

type MacroProcessorConfig struct {
	ProcessorType ProcessorType
	Delimiter     string
}
