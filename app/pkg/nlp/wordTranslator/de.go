package wordTranslator

import (
	"langs/pkg/nlp/verbformen_de"
	"langs/pkg/nlp/wiktionary_de"
)

type germanStrategy struct {
	defaultStrategy
}

func init() {
	RegisterStrategy("de", &germanStrategy{})
}

func (s *germanStrategy) PostProcess(tr *TranslateResult, raw []interface{}) {
	s.process(tr, raw)
	enrichGerman(tr)
}

// enrichGerman resolves the German article (der/die/das) and, for verbs, the
// conjugation. It works in both directions:
//   - German source word (de -> xx): enrich the source word itself.
//   - German target (xx -> de): enrich the first German translation.
func enrichGerman(tr *TranslateResult) {
	word := germanWord(tr)
	if word == "" {
		return
	}

	// Prefer the official Wiktionary API to resolve the article (der/die/das).
	if art, err := wiktionary_de.Article(word); err == nil && art != "" {
		tr.Article = art
	}

	lookup, err := verbformen_de.Lookup(word)
	if err != nil || lookup == nil {
		return
	}

	// Fall back to verbformen only if Wiktionary did not return an article.
	if tr.Article == "" && lookup.Article != "" {
		tr.Article = lookup.Article
	}

	if len(lookup.Conjugation) > 0 {
		tr.Conjugation = lookup.Conjugation
	}
}

// germanWord picks the German word to look up depending on the direction.
func germanWord(tr *TranslateResult) string {
	if tr.SourceLang == "de" {
		if tr.Infinitive != "" {
			return tr.Infinitive
		}
		return tr.SourceWord
	}

	if tr.TranslationLang == "de" && len(tr.Translations) > 0 {
		return tr.Translations[0]
	}

	return ""
}
