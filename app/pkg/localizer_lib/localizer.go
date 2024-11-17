package localizer_lib

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var Localaizer *i18n.Localizer
var bundle *i18n.Bundle
var translationsPath = "internal/translations/"

func init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile(fmt.Sprintf("%s%s", translationsPath, "active.en.toml"))

	Localaizer = i18n.NewLocalizer(bundle, "en")
}

func T(messageID string) string {
	msg, err := Localaizer.Localize(&i18n.LocalizeConfig{MessageID: messageID})
	if err != nil {
		return "translation missing"
	}
	return msg
}

func LoadLang(lang string) {
	bundle.MustLoadMessageFile(fmt.Sprintf("%s%s", translationsPath, fmt.Sprintf("active.%s.toml", lang)))
	Localaizer = i18n.NewLocalizer(bundle, lang)
}
