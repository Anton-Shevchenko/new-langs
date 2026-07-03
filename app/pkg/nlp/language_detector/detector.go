package language_detector

import (
	"errors"
	"sort"
	"strings"
	"sync"

	"github.com/pemistahl/lingua-go"
)

var Map = map[string]int{
	"en": 13,
	"ru": 51,
	"uk": 68,
	"de": 11,
	"fr": 20,
	"nl": 45,
	"es": 58,
}

func getLanguages() []lingua.Language {
	return []lingua.Language{
		lingua.English,
		lingua.French,
		lingua.German,
		lingua.Russian,
		lingua.Ukrainian,
		lingua.Dutch,
		lingua.Spanish,
	}
}

// detectorCache memoizes built detectors by their language set. Building a
// lingua detector loads language models into memory and is by far the most
// expensive step, so we build one per distinct set of languages and reuse it
// across requests.
var detectorCache sync.Map // map[string]lingua.LanguageDetector

func getDetector(languages []string) lingua.LanguageDetector {
	key := cacheKey(languages)
	if cached, ok := detectorCache.Load(key); ok {
		return cached.(lingua.LanguageDetector)
	}

	var fromLanguages []lingua.Language
	for _, lang := range languages {
		fromLanguages = append(fromLanguages, lingua.GetLanguageFromIsoCode639_1(lingua.IsoCode639_1(Map[lang])))
	}
	detector := lingua.NewLanguageDetectorBuilder().FromLanguages(fromLanguages...).Build()

	// LoadOrStore guards against two goroutines building the same detector
	// concurrently: only the first stored instance is kept and returned.
	actual, _ := detectorCache.LoadOrStore(key, detector)
	return actual.(lingua.LanguageDetector)
}

// cacheKey builds a stable, order-independent key for a set of language codes.
func cacheKey(languages []string) string {
	sorted := make([]string, len(languages))
	copy(sorted, languages)
	sort.Strings(sorted)
	return strings.Join(sorted, ",")
}

func Detect(word string, languages []string) (string, error) {
	detector := getDetector(languages)

	if language, exists := detector.DetectLanguageOf(word); exists {
		return strings.ToLower(language.IsoCode639_1().String()), nil
	}

	return "", errors.New("language is not detected")
}
