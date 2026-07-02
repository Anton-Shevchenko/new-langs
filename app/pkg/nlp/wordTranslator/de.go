package wordTranslator

import "langs/pkg/nlp/verbformen_de"

type germanStrategy struct {
	defaultStrategy
}

func init() {
	RegisterStrategy("de", &germanStrategy{})
}

func (s *germanStrategy) PostProcess(tr *TranslateResult, raw []interface{}) {
	s.process(tr, raw)

	word := tr.Infinitive
	if word == "" {
		word = tr.SourceWord
	}

	lookup, err := verbformen_de.Lookup(word)
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
