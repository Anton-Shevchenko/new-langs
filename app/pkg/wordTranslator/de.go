package wordTranslator

import "langs/pkg/verbformen_de"

type germanStrategy struct {
	defaultStrategy
}

func init() {
	RegisterStrategy("de", &germanStrategy{})
}

func (s *germanStrategy) PostProcess(tr *TranslateResult, raw []interface{}) {
	s.process(tr, raw)
	lookup, err := verbformen_de.Lookup(tr.Infinitive)

	if err != nil || lookup == nil {
		return
	}

	if lookup.Article != "" {
		tr.Article = lookup.Article
	}

	if len(lookup.Conjugation) > 0 {
		tr.Conjugation = lookup.Conjugation
	}
}
