package localizer_lib

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"os"
)

var Localaizer *i18n.Localizer
var bundle *i18n.Bundle
var translationsPath = "internal/translations/"

func init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	filePath := fmt.Sprintf("%s%s", translationsPath, "active.en.toml")
	if _, err := os.Stat(filePath); err == nil {
		bundle.MustLoadMessageFile(filePath)
	}

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
	if lang == "" {
		lang = "en"
	}
	filePath := fmt.Sprintf("%s%s", translationsPath, fmt.Sprintf("active.%s.toml", lang))
	if _, err := os.Stat(filePath); err == nil {
		bundle.MustLoadMessageFile(filePath)
	}
	Localaizer = i18n.NewLocalizer(bundle, lang)
}
