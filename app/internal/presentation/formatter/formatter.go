package formatter

import (
	"fmt"

	"langs/pkg/nlp/localizer_lib"
)

func FormatWordMessage(sourceWord, translation string) string {
	if sourceWord == "" {
		return "<i>" + localizer_lib.T("no_source_word") + "</i>"
	}

	return fmt.Sprintf("<i>%s</i>\n<b>%s</b> - %s", localizer_lib.T("word_added"), sourceWord, translation)
}
