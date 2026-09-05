package wordTranslator

import "testing"

func TestPrefixFirstGermanTranslation(t *testing.T) {
	tr := &TranslateResult{
		TranslationLang: "de",
		Article:         "das",
		Translations:    []string{"Auto", "Wagen", "Kfz"},
	}
	PrefixFirstGermanTranslation(tr)

	if tr.Translations[0] != "das Auto" {
		t.Fatalf("first option: %q", tr.Translations[0])
	}
	if tr.Translations[1] != "Wagen" || tr.Translations[2] != "Kfz" {
		t.Fatalf("other options must stay bare: %v", tr.Translations[1:])
	}

	PrefixFirstGermanTranslation(tr)
	if tr.Translations[0] != "das Auto" {
		t.Fatalf("must not double-prefix: %q", tr.Translations[0])
	}
}

func TestPrefixFirstGermanTranslationSkipsNonGerman(t *testing.T) {
	tr := &TranslateResult{
		TranslationLang: "uk",
		Article:         "das",
		Translations:    []string{"авто"},
	}
	PrefixFirstGermanTranslation(tr)
	if tr.Translations[0] != "авто" {
		t.Fatalf("uk option changed: %q", tr.Translations[0])
	}
}

func TestAttachGermanArticleAlreadyPrefixed(t *testing.T) {
	got, art := AttachGermanArticle("das Auto")
	if got != "das Auto" || art != "das" {
		t.Fatalf("got %q %q", got, art)
	}
}

func TestAttachGermanArticleNoun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live wiktionary lookup")
	}
	got, art := AttachGermanArticle("Auto")
	if art == "" {
		t.Skip("wiktionary unavailable")
	}
	if art != "das" || got != "das Auto" {
		t.Fatalf("got %q article=%q", got, art)
	}
}

func TestSplitGermanArticle(t *testing.T) {
	art, rest, ok := SplitGermanArticle("Der Hund")
	if !ok || art != "der" || rest != "Hund" {
		t.Fatalf("got %q %q %v", art, rest, ok)
	}
	if HasGermanArticle("Hund") {
		t.Fatal("bare noun should not have article")
	}
}
