package wordTranslator

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/robertkrimen/otto"
	"io"
	"langs/internal/consts"
	"net/http"
	"regexp"
	"strings"
)

type TranslateResult struct {
	IsSimpleWord    bool
	IsValid         bool
	Transcription   string
	SourceWord      string
	SourceLang      string
	TranslationLang string
	Translations    []string
	Definitions     []string
	Examples        []string
}

func Translate(source, sourceLang, targetLang string) (*TranslateResult, error) {

	encodedSource, err := encodeURI(source)
	if err != nil {
		return nil, err
	}

	url := buildTranslateURL(encodedSource, sourceLang, targetLang)
	body, err := fetchTranslationData(url)
	if err != nil {
		return nil, err
	}

	response, err := parseResponse(body)

	if err != nil {
		return nil, err
	}

	var translateResult TranslateResult
	translateResult.Translations = extractTranslations(response)
	translateResult.SourceWord = source
	translateResult.SourceLang = sourceLang
	translateResult.TranslationLang = targetLang
	fields := strings.Fields(source)
	translateResult.IsSimpleWord = len(fields) == 1
	fmt.Println("LEEEEEEттттттттттт", len(response))
	if (len(response) >= 12 || sourceLang == "uk") && response[0] != nil {
		fmt.Println("LEEEEEEn", len(response), response[2], targetLang, sourceLang)
		firstElement, ok := response[0].([]interface{})

		if ok {
			responseLangBlock, ok := response[8].([]interface{})
			fmt.Println("AAAAA000000", responseLangBlock)
			if ok {
				responseLangArr, ok := responseLangBlock[0].([]interface{})
				fmt.Println("AAAAA1", responseLangArr)
				if !ok {
					fmt.Println("AAAAA2", responseLangArr)
					translateResult.IsValid = false
					return &translateResult, nil
				}
				responseLang := responseLangArr[0]
				fmt.Println("AAAAA3", responseLang, targetLang, sourceLang)
				if responseLang != targetLang && responseLang != sourceLang {
					fmt.Println("AAAAA4", responseLang, targetLang, sourceLang)
					translateResult.IsValid = false
					return &translateResult, nil
				}
			}

			fmt.Println("BBBB", firstElement)

			wordTranscription, ok := firstElement[1].([]interface{})
			if ok && len(wordTranscription) >= 4 {
				translateResult.Transcription = wordTranscription[3].(string)
			}
		} else {
			fmt.Println("AAAAA5")
			translateResult.IsValid = false
			return &translateResult, nil
		}
		translationBlock, ok := firstElement[0].([]interface{})

		if ok && sourceLang == "uk" && len(translationBlock) > 5 {
			translateResult.IsValid = false
			return &translateResult, nil
		}
	} else {
		fmt.Println("AAAAA7")
		translateResult.IsValid = false
		return &translateResult, nil
	}

	translateResult.IsValid = true

	if translateResult.IsSimpleWord && len(response) > 8 {
		translateResult.Definitions = extractDefinitions(response)
		translateResult.Examples = extractExamples(response)
	}

	return &translateResult, nil
}

func encodeURI(s string) (string, error) {
	vm := otto.New()
	err := vm.Set("sourceText", s)
	if err != nil {
		return "", errors.New("error setting js variable")
	}

	_, err = vm.Run("eUri = encodeURI(sourceText);")
	if err != nil {
		return "", errors.New("error executing JavaScript")
	}

	val, err := vm.Get("eUri")
	if err != nil {
		return "", errors.New("error getting variable value from js")
	}

	encodedString, err := val.ToString()
	if err != nil {
		return "", errors.New("error converting js var to string")
	}

	return encodedString, nil
}

func (tr TranslateResult) ToSentenceString() string {
	var sb strings.Builder
	WriteSourceWordString(&sb, tr.SourceWord, tr.SourceLang)
	WriteTranslationString(&sb, tr.TranslationLang)
	return sb.String()

}

func (tr TranslateResult) ToNativeWordString() string {
	var sb strings.Builder
	WriteSourceWordString(&sb, tr.SourceWord, tr.SourceLang)
	WriteTranslationString(&sb, tr.TranslationLang)
	return sb.String()
}

func ParseSourceWordsFromTranslateMsg(input string) string {
	re := regexp.MustCompile(`Source (Sentence|Word):\s+(.*)(\n|$)`)
	match := re.FindStringSubmatch(input)

	if match != nil && len(match) > 2 {
		return match[2]
	} else {
		return ""
	}
}

func (tr TranslateResult) ToString(msgLang string) string {
	if msgLang == tr.SourceLang {
		return tr.ToNativeWordString()
	}
	if !tr.IsSimpleWord {
		return tr.ToSentenceString()
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

	WriteTranslationString(&sb, tr.TranslationLang)
	return sb.String()
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

func buildTranslateURL(encodedSource, sourceLang, targetLang string) string {
	url := fmt.Sprintf(
		"https://translate.googleapis.com/translate_a/single?client=gtx&dt=t&dt=ex&dt=md&sl=%s&tl=%s&dt=t&dt=rm&dt=at&q=%s",
		sourceLang, targetLang, encodedSource,
	)

	fmt.Println(url)
	return url
}

func fetchTranslationData(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.New("error getting translate.googleapis.com")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("error reading response body")
	}

	if strings.Contains(string(body), "<title>Error 400 (Bad Request)") {
		return nil, errors.New("error 400 (Bad Request)")
	}

	return body, nil
}

func parseResponse(body []byte) ([]interface{}, error) {
	var result []interface{}
	err := json.Unmarshal(body, &result)
	if err != nil {
		return nil, errors.New("error unmarshaling data")
	}

	if len(result) == 0 {
		return nil, errors.New("no translated data")
	}

	return result, nil
}

func extractTranslations(response []interface{}) []string {
	var translations []string
	if len(response) > 5 {
		sliceOneResp, ok := response[5].([]interface{})

		if !ok {
			return translations
		}

		for _, sliceOne := range sliceOneResp {
			for i, sliceTwo := range sliceOne.([]interface{}) {
				if i == 2 {
					for _, sliceThree := range sliceTwo.([]interface{}) {
						for _, translate := range sliceThree.([]interface{}) {
							translations = append(translations, strings.ToLower(fmt.Sprintf("%v", translate)))
							break
						}
					}
				}
			}
		}
	}
	return SliceUnique(translations)
}

func SliceUnique(slice []string) []string {
	uniqueMap := make(map[string]struct{}, len(slice))
	uniqueSlice := make([]string, 0, len(slice))

	for _, elem := range slice {
		if _, exists := uniqueMap[elem]; !exists {
			uniqueMap[elem] = struct{}{}
			uniqueSlice = append(uniqueSlice, elem)
		}
	}

	return uniqueSlice
}

func extractDefinitions(response []interface{}) []string {
	var definitions []string
	if len(response) > 12 {
		for _, def := range response[12].([]interface{}) {
			entry := def.([]interface{})
			if len(entry) > 1 {
				definitions = append(definitions, entry[2].(string))
			}
		}
	}
	return definitions
}

func extractExamples(response []interface{}) []string {
	var examples []string
	if len(response) > 13 {
		for _, exArr := range response[13].([]interface{}) {
			examplesSet := exArr.([]interface{})
			for i, ex := range examplesSet {
				if i == 5 {
					break
				}
				example, ok := ex.([]interface{})
				if ok && len(example) > 0 {
					examples = append(examples, example[0].(string))
				}
			}
		}
	}

	return examples
}

func WriteSourceWordString(sb *strings.Builder, sourceWord, sourceLang string) {
	sb.WriteString(
		fmt.Sprintf("<strong>%s Source Word:</strong> %s\n", consts.LangFlags[sourceLang], sourceWord),
	)
}

func WriteTranslationString(sb *strings.Builder, targetLang string) {
	sb.WriteString(
		fmt.Sprintf(
			"<strong>%s Translations:</strong>\n",
			consts.LangFlags[targetLang]),
	)
}
