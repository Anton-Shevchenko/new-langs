package wordTranslator

import (
	"unicode"

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
//
// Only nouns have a gender, so the article is looked up only when the word is a
// noun. Relying on capitalization alone is unreliable (Google may capitalize
// adjectives, e.g. "гарний" -> "Gut", which collides with the noun "das Gut"),
// so the part of speech drives the decision.
func enrichGerman(tr *TranslateResult) {
	word := germanWord(tr)
	if word == "" {
		return
	}

	isNoun := germanConceptPOS(tr) == "noun"

	if isNoun {
		// German nouns are capitalized; user input is lower-cased by the
		// handlers, and Google is inconsistent, so normalize for the
		// case-sensitive Wiktionary lookup.
		if art, err := wiktionary_de.Article(capitalizeFirst(word)); err == nil && art != "" {
			tr.Article = art
		}
	}

	lookup, err := verbformen_de.Lookup(word)
	if err != nil || lookup == nil {
		return
	}

	if isNoun && tr.Article == "" && lookup.Article != "" {
		tr.Article = lookup.Article
	}

	if len(lookup.Conjugation) > 0 {
		tr.Conjugation = lookup.Conjugation
	}
}

// germanConceptPOS returns the part of speech of the concept being translated.
// The main translation response carries it for some directions (e.g. de->xx);
// otherwise it is resolved from the source word via an auxiliary lookup.
func germanConceptPOS(tr *TranslateResult) string {
	if tr.PartOfSpeech != "" {
		return tr.PartOfSpeech
	}
	return detectPartOfSpeech(tr.SourceWord, tr.SourceLang)
}

func capitalizeFirst(word string) string {
	runes := []rune(word)
	if len(runes) == 0 {
		return word
	}
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
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
