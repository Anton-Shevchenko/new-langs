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

type TranslationPair struct {
	Source []string `json:"source"`
}
