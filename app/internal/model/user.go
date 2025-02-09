package model

const CustomTranslationScenario = "cts"
const WritingExamScenario = "wes"

type User struct {
	TestInterval  uint16    `json:"test_interval,omitempty" gorm:"default:60"`
	TargetRate    uint16    `json:"target_rate,omitempty" gorm:"default:5"`
	ChatId        int64     `json:"chat_id" gorm:"primaryKey"`
	BookId        int64     `json:"book_id"`
	InterfaceLang string    `json:"interface_lang,omitempty"`
	NativeLang    string    `json:"native_lang,omitempty"`
	TargetLang    string    `json:"target_lang,omitempty"`
	StateData     StateData `gorm:"embedded"`
}

type StateData struct {
	Scenario string `json:"scenario,omitempty"`
	Value    string `json:"value,omitempty"`
}

func (u *User) GetUserLangs() []string {
	return []string{u.NativeLang, u.TargetLang}
}

func (u *User) IsAwaitingInput() bool {
	return u.StateData.Scenario != ""
}

func (u *User) GetOppositeLang(lang string) string {
	if u.NativeLang == lang {
		return u.TargetLang
	}
	return u.NativeLang
}
