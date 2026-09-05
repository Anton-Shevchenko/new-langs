package whatsapp

import (
	"html"
	"regexp"
	"strings"
)

var (
	htmlBoldRe      = regexp.MustCompile(`(?is)</?(b|strong)>`)
	htmlItalicRe    = regexp.MustCompile(`(?is)</?(i|em)>`)
	htmlStrikeRe    = regexp.MustCompile(`(?is)</?(s|strike|del)>`)
	htmlCodeRe      = regexp.MustCompile(`(?is)</?(code|pre)>`)
	htmlBreakRe     = regexp.MustCompile(`(?is)<br\s*/?>`)
	htmlAnyTagRe    = regexp.MustCompile(`(?s)<[^>]*>`)
	htmlUnderlineRe = regexp.MustCompile(`(?is)</?u>`)
)

// htmlToWhatsApp converts the subset of Telegram HTML used by the bot into
// WhatsApp text formatting (*bold*, _italic_, ~strike~, ```code```).
func htmlToWhatsApp(s string) string {
	s = htmlBreakRe.ReplaceAllString(s, "\n")
	s = htmlBoldRe.ReplaceAllString(s, "*")
	s = htmlItalicRe.ReplaceAllString(s, "_")
	s = htmlStrikeRe.ReplaceAllString(s, "~")
	s = htmlCodeRe.ReplaceAllString(s, "```")
	s = htmlUnderlineRe.ReplaceAllString(s, "")
	s = htmlAnyTagRe.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	return strings.TrimSpace(s)
}
