package config

type StringIndexCacheProcessorConfig struct {
	Enabled bool
}
type StringIndexProcessorConfig struct {
	Enabled bool
}
type TemplateCacheProcessorConfig struct {
	Enabled bool
}
type MacroProcessorConfig struct {
	TemplateCacheProcessorConfig    TemplateCacheProcessorConfig
	StringIndexProcessorConfig      StringIndexProcessorConfig
	StringIndexCacheProcessorConfig StringIndexCacheProcessorConfig
	Delimiter                       string
}
