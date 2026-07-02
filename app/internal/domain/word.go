package model

type Word struct {
	IsCustom        bool
	IsSimpleWord    bool
	Rate            int8   `json:"rate"`
	ID              int64  `json:"id" gorm:"primaryKey"`
	ChatId          int64  `json:"chat_id"`
	Value           string `json:"value"`
	ValueLang       string `json:"value_lang"`
	Translation     string `json:"translation"`
	TranslationLang string `json:"translation_lang"`
}

type WordOption struct {
	WordID          int64
	Word            string
	Translation     string
	WordLang        string
	TranslationLang string
}

type TranslationPair struct {
	Source []string `json:"source"`
}

// LangPair represents a distinct, direction-independent pair of languages the
// user has saved words for (e.g. de/uk), together with the number of words.
type LangPair struct {
	Lang1 string
	Lang2 string
	Count int
}

// Key returns a stable identifier for the pair, e.g. "de_uk".
func (p LangPair) Key() string {
	return p.Lang1 + "_" + p.Lang2
}
