package model

type User struct {
	ChatId        int64  `json:"chat_id" gorm:"primaryKey"`
	BookId        int64  `json:"book_id"`
	InterfaceLang string `json:"interface_lang,omitempty"`
	NativeLang    string `json:"native_lang,omitempty"`
	TargetLang    string `json:"target_lang,omitempty"`
	WaitingFor    string `json:"waiting_for,omitempty"`
	TestInterval  uint16 `json:"test_interval,omitempty" gorm:"default:60"`
}

func (u *User) GetUserLangs() []string {
	return []string{u.NativeLang, u.TargetLang}
}

func (u *User) IsAwaitingInput() bool {
	return u.WaitingFor != ""
}

func (u *User) GetOppositeLang(lang string) string {
	if u.NativeLang == lang {
		return u.TargetLang
	}
	return u.NativeLang
}
