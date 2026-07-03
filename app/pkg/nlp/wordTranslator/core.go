package wordTranslator

type TranslateResult struct {
	IsSimpleWord    bool
	IsValid         bool
	// RecognizedWord is true when Google returned a bilingual dictionary entry
	// (part-of-speech block) for the source word. Google only does this for
	// real, correctly-spelled words it knows, so it is a reliable "this is not
	// a typo" signal — but only in directions where the dictionary is
	// populated (notably de as the source). Its confidence score is not
	// reliable for this: it stays high (~0.98) even for misspellings.
	RecognizedWord  bool
	Transcription   string
	SourceWord      string
	SourceLang      string
	TranslationLang string
	Infinitive      string
	PartOfSpeech    string
	Article         string
	Translations    []string
	Examples        []string
	Conjugation     []string
}

type TranslationStrategy interface {
	PostProcess(*TranslateResult, []interface{})
}

var strategyRegistry = map[string]TranslationStrategy{}

func RegisterStrategy(lang string, strat TranslationStrategy) {
	strategyRegistry[lang] = strat
}

func getStrategy(lang string) TranslationStrategy {
	if s, ok := strategyRegistry[lang]; ok {
		return s
	}
	return &defaultStrategy{}
}
