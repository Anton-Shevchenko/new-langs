package wordTranslator

import (
	"regexp"
	"strings"

	"langs/pkg/nlp/wiktionary_de"
)

var germanArticleRe = regexp.MustCompile(`(?i)^(der|die|das)\s+(.+)$`)

func SplitGermanArticle(s string) (article, rest string, ok bool) {
	s = strings.TrimSpace(s)
	if m := germanArticleRe.FindStringSubmatch(s); len(m) == 3 {
		return strings.ToLower(m[1]), m[2], true
	}
	return "", s, false
}

func HasGermanArticle(s string) bool {
	_, _, ok := SplitGermanArticle(s)
	return ok
}

// AttachGermanArticle prefixes a German noun with der/die/das when Wiktionary
// knows its gender. Words that already have an article are left as-is.
// article is empty when nothing was (or could be) attached.
func AttachGermanArticle(word string) (result, article string) {
	word = strings.TrimSpace(word)
	if word == "" {
		return word, ""
	}
	if art, rest, ok := SplitGermanArticle(word); ok {
		return art + " " + rest, art
	}

	art, err := wiktionary_de.Article(capitalizeFirst(word))
	if err != nil || art == "" {
		return word, ""
	}
	return art + " " + capitalizeFirst(word), art
}

// PrefixFirstGermanTranslation puts the article on the first option only.
// Remaining options stay bare so they can be enriched at save time.
func PrefixFirstGermanTranslation(tr *TranslateResult) {
	if tr == nil || tr.TranslationLang != "de" || len(tr.Translations) == 0 {
		return
	}
	if tr.Article == "" {
		if _, art := AttachGermanArticle(tr.Translations[0]); art != "" {
			tr.Article = art
		}
	}
	if tr.Article == "" || HasGermanArticle(tr.Translations[0]) {
		return
	}
	tr.Translations[0] = tr.Article + " " + tr.Translations[0]
}
