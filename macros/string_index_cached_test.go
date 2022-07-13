package macros

import "testing"

func TestAddTemplatesMutex(t *testing.T) {

	p := StringIndexCached{
		templates: make(map[string]strMetaTemplate),
	}
	for i := 1; i <= 2; i++ {
		go func() {
			p.AddTemplates("template1")
		}()
	}
}
