package wordTranslator

type TranslateResult struct {
	IsSimpleWord    bool
	IsValid         bool
	Transcription   string
	SourceWord      string
	SourceLang      string
	TranslationLang string
	Infinitive      string
	Translations    []string
	Examples        []string
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
