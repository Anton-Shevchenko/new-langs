package wordTranslator

import (
	"fmt"
	"regexp"
	"strings"

	"langs/internal/consts"
	"langs/pkg/nlp/localizer_lib"
)

func (tr TranslateResult) ToSentenceString() string {
	var sb strings.Builder
	WriteSourceWordString(&sb, tr.SourceWord, tr.SourceLang)
	WriteTranslationString(&sb, tr.TranslationLang)
	return sb.String()
}

func (tr TranslateResult) ToNativeWordString() string {
	return tr.ToSentenceString()
}

// cardSeparator is a thin visual divider between the headword and its
// grammatical details, giving the message a clean "card" look.
const cardSeparator = "➖➖➖➖➖➖➖➖"

func (tr TranslateResult) SimpleString() string {
	var sb strings.Builder
	sb.WriteString("<strong>" + localizer_lib.T("fmt_unknown_word") + "</strong>\n")
	WriteSourceWordString(&sb, tr.SourceWord, tr.SourceLang)
	if len(tr.Translations) > 0 {
		sb.WriteString("\n")
		WriteTranslationString(&sb, tr.TranslationLang)
		for _, translation := range tr.Translations {
			sb.WriteString("• " + translation + "\n")
		}
	}
	return sb.String()
}

func (tr TranslateResult) ToString(msgLang string) string {
	if msgLang == tr.SourceLang {
		return tr.ToNativeWordString()
	}
	if !tr.IsSimpleWord {
		return tr.ToSentenceString()
	}
	if !tr.IsValid {
		return tr.SimpleString()
	}

	var sb strings.Builder
	WriteSourceWordString(&sb, tr.SourceWord, tr.SourceLang)

	if grammar := tr.grammarLines(); len(grammar) > 0 {
		sb.WriteString(cardSeparator + "\n")
		for _, line := range grammar {
			sb.WriteString(line + "\n")
		}
	}

	if len(tr.Examples) > 0 {
		sb.WriteString("\n<strong>" + localizer_lib.T("fmt_examples") + "</strong>\n")
		for _, example := range tr.Examples {
			sb.WriteString("• " + example + "\n")
		}
	}

	sb.WriteString("\n")
	WriteTranslationString(&sb, tr.TranslationLang)
	return sb.String()
}

// grammarLines renders the article/infinitive/conjugation details as a group of
// bold-labelled lines. The article label keeps the 🏷 marker so the value can be
// parsed back out of the rendered message regardless of interface language.
func (tr TranslateResult) grammarLines() []string {
	var lines []string

	if tr.Article != "" {
		lines = append(lines, fmt.Sprintf(
			"<strong>%s</strong> %s", localizer_lib.T("fmt_article"), tr.Article,
		))
	}

	if tr.PartOfSpeech == "verb" && tr.Infinitive != "" {
		lines = append(lines, fmt.Sprintf(
			"<strong>%s</strong> %s", localizer_lib.T("fmt_infinitive"), tr.Infinitive,
		))
	}

	if len(tr.Conjugation) > 0 {
		lines = append(lines, fmt.Sprintf(
			"<strong>%s</strong> %s",
			localizer_lib.T("fmt_forms"), strings.Join(tr.Conjugation, " · "),
		))
	}

	return lines
}

func WriteSourceWordString(sb *strings.Builder, sourceWord, sourceLang string) {
	sb.WriteString(fmt.Sprintf(
		"<strong>%s %s</strong> %s\n",
		consts.LangFlags[sourceLang], localizer_lib.T("fmt_source_word"), sourceWord,
	))
}

func WriteTranslationString(sb *strings.Builder, targetLang string) {
	sb.WriteString(fmt.Sprintf(
		"<strong>%s %s</strong>\n",
		consts.LangFlags[targetLang], localizer_lib.T("fmt_translations"),
	))
}

// ParseSourceWordsFromTranslateMsg extracts the source word from the first line
// of a rendered translate message. It is language-independent: the first line
// always has the shape "{flag} {label}: {word}", so everything after the first
// ": " is the word.
func ParseSourceWordsFromTranslateMsg(input string) string {
	firstLine := input
	if idx := strings.IndexByte(input, '\n'); idx != -1 {
		firstLine = input[:idx]
	}
	sep := strings.Index(firstLine, ": ")
	if sep == -1 {
		return ""
	}
	return strings.TrimSpace(firstLine[sep+2:])
}

// articleMsgRe matches the article line regardless of the (localized) label,
// anchoring on the 🏷 marker and the der/die/das value.
var articleMsgRe = regexp.MustCompile(`(?i)🏷[^:\n]*:\s*(der|die|das)\b`)

func ParseArticleFromTranslateMsg(input string) string {
	m := articleMsgRe.FindStringSubmatch(input)
	if len(m) == 2 {
		return strings.ToLower(m[1])
	}
	return ""
}
