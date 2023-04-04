package processor

type Replacer interface {
	// Replace the macros and returns replaced string
	// if any error the error will be returned
	Replace(url string, macroProvider Provider) (string, error)
}

var processor Replacer

// NewReplacer will return instance of macro processor
func NewReplacer() Replacer {

	return &stringBasedProcessor{
		templates: make(map[string]urlMetaTemplate),
	}
}
