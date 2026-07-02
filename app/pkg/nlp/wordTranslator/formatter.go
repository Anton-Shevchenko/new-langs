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

func (tr TranslateResult) SimpleString() string {
	var sb strings.Builder
	sb.WriteString("<strong>" + localizer_lib.T("fmt_unknown_word") + "</strong>\n")
	WriteSourceWordString(&sb, tr.SourceWord, tr.SourceLang)
	WriteTranslationString(&sb, tr.TranslationLang)
	for _, translation := range tr.Translations {
		sb.WriteString(fmt.Sprintf("  - %s\n", translation))
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

	if len(tr.Examples) > 0 {
		sb.WriteString("<strong>" + localizer_lib.T("fmt_examples") + "</strong>\n")
		for _, example := range tr.Examples {
			sb.WriteString(fmt.Sprintf("  - %s\n", example))
		}
	}

	if len(tr.Conjugation) > 0 {
		sb.WriteString("<strong>" + localizer_lib.T("fmt_forms") + "</strong> ")
		sb.WriteString(strings.Join(tr.Conjugation, " * "))
		sb.WriteString("\n")
	}

	if tr.Article != "" {
		sb.WriteString(fmt.Sprintf(
			"<strong>%s %s</strong>\n",
			localizer_lib.T("fmt_article"), tr.Article,
		))
	}

	if tr.PartOfSpeech == "verb" && tr.Infinitive != "" {
		sb.WriteString(fmt.Sprintf(
			"<strong>%s %s</strong>\n",
			localizer_lib.T("fmt_infinitive"), tr.Infinitive,
		))
	}

	WriteTranslationString(&sb, tr.TranslationLang)
	return sb.String()
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
// anchoring on the 🔍 marker and the der/die/das value.
var articleMsgRe = regexp.MustCompile(`(?i)🔍[^:\n]*:\s*(der|die|das)\b`)

func ParseArticleFromTranslateMsg(input string) string {
	m := articleMsgRe.FindStringSubmatch(input)
	if len(m) == 2 {
		return strings.ToLower(m[1])
	}
	return ""
}
