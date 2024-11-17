package formatter

import "fmt"

func FormatWordMessage(sourceWord, translation string) string {
	if sourceWord == "" {
		return "<i>No source word found.</i>"
	}

	return fmt.Sprintf("<i>Word added to your list</i>\n<b>%s</b> - %s", sourceWord, translation)
}
