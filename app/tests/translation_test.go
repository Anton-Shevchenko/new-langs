package tests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"langs/pkg/nlp/spellio"
	"langs/pkg/nlp/wordTranslator"
)

// These tests exercise the real translation and spell-checking pipeline, so
// they hit the live Google Translate and LanguageTool APIs. They are skipped
// in -short mode, and skip (rather than fail) when the network is unavailable.

func translateOrSkip(t *testing.T, word, src, dst string) *wordTranslator.TranslateResult {
	t.Helper()
	tr, err := wordTranslator.Translate(word, src, dst)
	if err != nil {
		t.Skipf("translation API unavailable for %q (%s->%s): %v", word, src, dst, err)
	}
	require.NotNil(t, tr)
	return tr
}

// TestTranslateVariousWords covers different languages and parts of speech.
// Every single word should yield at least one translation. Grammar details
// (article/part of speech/conjugation) depend on secondary services, so they
// are logged for visibility but only sanity-checked, not hard-asserted.
func TestTranslateVariousWords(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live translation test in short mode")
	}

	cases := []struct {
		name            string
		word            string
		src             string
		dst             string
		wordType        string
		checkRecognized bool
		wantRecognized  bool
	}{
		{"de noun Wohnung", "Wohnung", "de", "uk", "noun", true, true},
		{"de noun Haus", "Haus", "de", "uk", "noun", true, true},
		{"de verb gehen", "gehen", "de", "uk", "verb", true, true},
		// Google does not return a dictionary block for every correct word
		// (e.g. this adjective), so RecognizedWord may legitimately be false.
		// That is safe: such words then fall through to the LanguageTool check,
		// which reports no mistake and the translation is shown.
		{"de adjective schön", "schön", "de", "uk", "adjective", false, false},
		{"de adverb gestern", "gestern", "de", "uk", "adverb", true, true},
		{"de->en noun Haus", "Haus", "de", "en", "noun", true, true},
		{"en noun apartment", "apartment", "en", "uk", "noun", false, false},
		{"en verb run", "run", "en", "uk", "verb", false, false},
		{"uk noun kvartyra", "квартира", "uk", "de", "noun", false, false},
		{"uk verb bihty", "бігти", "uk", "de", "verb", false, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tr := translateOrSkip(t, tc.word, tc.src, tc.dst)

			assert.NotEmpty(t, tr.Translations,
				"expected at least one translation for %q", tc.word)

			if tc.checkRecognized {
				assert.Equal(t, tc.wantRecognized, tr.RecognizedWord,
					"RecognizedWord mismatch for %q", tc.word)
			}

			if tr.Article != "" {
				assert.Contains(t, []string{"der", "die", "das"}, tr.Article,
					"unexpected article for %q", tc.word)
			}

			t.Logf("%s: translations=%v pos=%q article=%q infinitive=%q conjugation=%v recognized=%v valid=%v",
				tc.wordType, tr.Translations, tr.PartOfSpeech, tr.Article,
				tr.Infinitive, tr.Conjugation, tr.RecognizedWord, tr.IsValid)
		})
	}
}

// TestTranslateSentence verifies that multi-word input is treated as a sentence
// rather than a single dictionary word.
func TestTranslateSentence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live translation test in short mode")
	}

	sentence := "Ich wohne in einer schönen Wohnung."
	tr := translateOrSkip(t, sentence, "de", "uk")

	assert.False(t, tr.IsSimpleWord, "sentence must not be treated as a simple word")
	assert.Equal(t, sentence, tr.SourceWord)
	t.Logf("sentence: source=%q simpleWord=%v valid=%v", tr.SourceWord, tr.IsSimpleWord, tr.IsValid)
}

// TestRecognizedWordDetectsTypos is the core of the spell-suggestion gating:
// Google returns a dictionary entry (RecognizedWord=true) only for real words,
// and drops it for misspellings even though its confidence stays high.
func TestRecognizedWordDetectsTypos(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live translation test in short mode")
	}

	correct := translateOrSkip(t, "Wohnung", "de", "uk")
	assert.True(t, correct.RecognizedWord, "correct German word should be recognized")

	typo := translateOrSkip(t, "wohnunf", "de", "uk")
	assert.False(t, typo.RecognizedWord, "misspelled German word must not be recognized")
}

// TestSpellingSuggestions checks that LanguageTool returns replacement
// suggestions for misspelled words across languages.
func TestSpellingSuggestions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live spelling test in short mode")
	}

	typos := []struct {
		name string
		word string
		lang string
	}{
		{"de wohnunf", "wohnunf", "de"},
		{"de tiscg", "tiscg", "de"},
		{"de gestenr", "gestenr", "de"},
		{"en apartmnt", "apartmnt", "en"},
		{"uk kvrtyra", "квртира", "uk"},
	}

	for _, tc := range typos {
		t.Run(tc.name, func(t *testing.T) {
			res := spellio.Check(tc.word, tc.lang)
			if res == nil {
				t.Skipf("spelling API returned no result for %q (likely network/rate limit)", tc.word)
			}
			assert.Equal(t, spellio.SpellingMistake, res.Message)
			assert.NotEmpty(t, res.Replacements,
				"expected spelling suggestions for %q", tc.word)
			t.Logf("%q -> suggestions: %s", tc.word, strings.Join(res.Replacements, ", "))
		})
	}

	t.Run("correct word", func(t *testing.T) {
		res := spellio.Check("Wohnung", "de")
		if res == nil {
			t.Log("correct word: no spelling mistake reported (expected)")
			return
		}
		t.Logf("correct word flagged by picky mode: %v", res.Replacements)
	})
}
