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
// noun. Verbs are detected via part of speech; when the case of the word is
// reliable (the German side comes from the translator, not raw user input) an
// uppercase first letter is also required, since German nouns are always
// capitalized. This prevents matching a noun homograph of a verb such as the
// noun "Essen" for the verb "essen".
func enrichGerman(tr *TranslateResult) {
	word := germanWord(tr)
	if word == "" {
		return
	}

	articleWord, tryArticle := nounLookupWord(tr, word)

	if tryArticle {
		if art, err := wiktionary_de.Article(articleWord); err == nil && art != "" {
			tr.Article = art
		}
	}

	lookup, err := verbformen_de.Lookup(word)
	if err != nil || lookup == nil {
		return
	}

	if tryArticle && tr.Article == "" && lookup.Article != "" {
		tr.Article = lookup.Article
	}

	if len(lookup.Conjugation) > 0 {
		tr.Conjugation = lookup.Conjugation
	}
}

// nounLookupWord returns the word to use for the article lookup and whether an
// article should be looked up at all.
func nounLookupWord(tr *TranslateResult, word string) (string, bool) {
	if tr.PartOfSpeech == "verb" {
		return "", false
	}

	// German source: user input is lower-cased by the handlers, so the case is
	// not reliable. Capitalize for the (case-sensitive) noun lookup and rely on
	// the part-of-speech check above to filter out verbs.
	if tr.SourceLang == "de" {
		return capitalizeFirst(word), true
	}

	// German target: the word comes from the translator, which preserves case.
	// A lowercase word is therefore not a noun and must not get an article.
	if tr.TranslationLang == "de" && startsUpper(word) {
		return word, true
	}

	return "", false
}

func startsUpper(word string) bool {
	for _, r := range word {
		return unicode.IsUpper(r)
	}
	return false
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
