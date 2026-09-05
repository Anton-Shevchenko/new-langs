package formatter

import (
	"fmt"

	"langs/pkg/nlp/localizer_lib"
)

func FormatWordMessage(sourceWord, translation string) string {
	return FormatWordMessageWithArticle(sourceWord, translation, "")
}

func FormatWordMessageWithArticle(sourceWord, translation, article string) string {
	if sourceWord == "" {
		return "<i>" + localizer_lib.T("no_source_word") + "</i>"
	}

	msg := fmt.Sprintf("✅ <i>%s</i>\n\n<b>%s</b> → %s", localizer_lib.T("word_added"), sourceWord, translation)
	if article != "" {
		msg += "\n\n" + fmt.Sprintf(localizer_lib.T("word_added_article"), article)
	}
	return msg
}
