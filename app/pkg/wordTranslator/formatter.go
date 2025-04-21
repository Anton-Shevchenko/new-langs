package wordTranslator

import (
	"fmt"
	"regexp"
	"strings"

	"langs/internal/consts"
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
	sb.WriteString("<strong>❓ Unknown word.</strong>\n")
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
	sb.WriteString(fmt.Sprintf("<strong>🔊 Transcription:</strong> %s\n", tr.Transcription))
	if len(tr.Examples) > 0 {
		sb.WriteString("<strong>📖 Examples:</strong>\n")
		for _, example := range tr.Examples {
			sb.WriteString(fmt.Sprintf("  - %s\n", example))
		}
	}

	if tr.Article != "" {
		sb.WriteString(fmt.Sprintf(
			"<strong>🔍 Article: %s</strong>\n",
			tr.Article,
		))
	}
	WriteTranslationString(&sb, tr.TranslationLang)
	return sb.String()
}

func WriteSourceWordString(sb *strings.Builder, sourceWord, sourceLang string) {
	sb.WriteString(fmt.Sprintf(
		"<strong>%s Source Word:</strong> %s\n",
		consts.LangFlags[sourceLang], sourceWord,
	))
}

func WriteTranslationString(sb *strings.Builder, targetLang string) {
	sb.WriteString(fmt.Sprintf(
		"<strong>%s Translations:</strong>\n",
		consts.LangFlags[targetLang],
	))
}

// TODO move
func ParseSourceWordsFromTranslateMsg(input string) string {
	re := regexp.MustCompile(`Source (Sentence|Word):\s+(.*)(\n|$)`)
	match := re.FindStringSubmatch(input)
	if len(match) > 2 {
		return match[2]
	}
	return ""
}

func ParseArticleFromTranslateMsg(input string) string {
	re := regexp.MustCompile(`🔍 Article:\s*([[:alpha:]]+)`)
	m := re.FindStringSubmatch(input)
	fmt.Println("RRRRRRRRRR", m)
	if len(m) == 2 {
		fmt.Println("RRRRRRRRRR2", m)
		return m[1]
	} else {
		fmt.Println("RRRRRRR")
		return ""
	}
}
