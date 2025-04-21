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
	article, err := verbformen_de.Lookup(tr.SourceWord)

	if err != nil || article == nil || article.Article == "" {
		return
	}

	tr.Article = article.Article
}
