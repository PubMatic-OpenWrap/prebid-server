package macros

type TemplateBasedInitAlways struct {
	TemplateBased
}

func (p *TemplateBasedInitAlways) Replace(str string) (string, error) {
	p.init0([]string{str})
	return replaceTemplateBased(p.templates[str], p.Cfg.macroValues)
}
