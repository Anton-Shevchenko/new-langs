package language_detector

import (
	"errors"
	"github.com/pemistahl/lingua-go"
	"strings"
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

func Detect(word string, languages []string) (string, error) {
	var fromLanguages []lingua.Language

	for _, lang := range languages {
		fromLanguages = append(fromLanguages, lingua.GetLanguageFromIsoCode639_1(lingua.IsoCode639_1(Map[lang])))
	}

	detector := lingua.NewLanguageDetectorBuilder().FromLanguages(fromLanguages...).Build()

	if language, exists := detector.DetectLanguageOf(word); exists {
		return strings.ToLower(language.IsoCode639_1().String()), nil
	}

	return "", errors.New("language is not detected")
}
