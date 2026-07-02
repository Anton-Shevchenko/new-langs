package wordTranslator

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/robertkrimen/otto"
	"io"
	"net/http"
	"strings"
)

func Translate(source, sourceLang, targetLang string) (*TranslateResult, error) {
	encoded, err := encodeURI(source)
	if err != nil {
		return nil, err
	}
	url := buildTranslateURL(encoded, sourceLang, targetLang)
	body, err := fetchTranslationData(url)
	if err != nil {
		return nil, err
	}
	raw, err := parseResponse(body)
	if err != nil {
		return nil, err
	}
	tr := &TranslateResult{
		SourceWord:      source,
		SourceLang:      sourceLang,
		TranslationLang: targetLang,
	}

	getStrategy(sourceLang).PostProcess(tr, raw)

	// When translating into German from another language, the German-specific
	// strategy is not selected (it keys off the source language), so enrich the
	// first German translation with its article here.
	if targetLang == "de" && sourceLang != "de" {
		enrichGerman(tr)
	}

	return tr, nil
}

func encodeURI(s string) (string, error) {
	vm := otto.New()
	if err := vm.Set("sourceText", s); err != nil {
		return "", errors.New("error setting js variable")
	}
	if _, err := vm.Run("eUri = encodeURI(sourceText);"); err != nil {
		return "", errors.New("error executing JavaScript")
	}
	val, err := vm.Get("eUri")
	if err != nil {
		return "", errors.New("error getting variable from js")
	}
	return val.ToString()
}

func buildTranslateURL(encodedSource, sourceLang, targetLang string) string {
	return fmt.Sprintf(
		"https://translate.googleapis.com/translate_a/single?client=gtx&dt=t&dt=ex&dt=md&sl=%s&tl=%s&dt=t&dt=rm&dt=at&q=%s",
		sourceLang, targetLang, encodedSource,
	)
}

// detectPartOfSpeech resolves the part of speech of a word by querying Google's
// bilingual dictionary (dt=bd), which returns POS labels ("noun", "verb",
// "adjective", ...). It is used when the main translation response does not
// carry a part of speech (e.g. uk->de). Returns "" if unknown.
func detectPartOfSpeech(word, lang string) string {
	if word == "" {
		return ""
	}

	// The dictionary block is populated for translations into English; pick a
	// different target when the source is already English.
	target := "en"
	if lang == "en" {
		target = "de"
	}

	encoded, err := encodeURI(word)
	if err != nil {
		return ""
	}

	url := fmt.Sprintf(
		"https://translate.googleapis.com/translate_a/single?client=gtx&sl=%s&tl=%s&dt=t&dt=bd&q=%s",
		lang, target, encoded,
	)
	body, err := fetchTranslationData(url)
	if err != nil {
		return ""
	}
	raw, err := parseResponse(body)
	if err != nil {
		return ""
	}
	return parsePartOfSpeech(raw)
}

var knownPartsOfSpeech = map[string]bool{
	"noun": true, "verb": true, "adjective": true, "adverb": true,
	"pronoun": true, "preposition": true, "conjunction": true,
	"interjection": true, "numeral": true, "article": true,
	"determiner": true, "abbreviation": true, "particle": true,
	"exclamation": true, "prefix": true, "suffix": true,
}

// parsePartOfSpeech extracts the first POS label from a Google dictionary
// response. Dictionary groups look like ["noun", ["term", ...], ...].
func parsePartOfSpeech(raw []interface{}) string {
	for _, el := range raw {
		group, ok := el.([]interface{})
		if !ok || len(group) == 0 {
			continue
		}
		entry, ok := group[0].([]interface{})
		if !ok || len(entry) < 2 {
			continue
		}
		label, ok := entry[0].(string)
		if !ok {
			continue
		}
		if _, ok := entry[1].([]interface{}); !ok {
			continue
		}
		if knownPartsOfSpeech[label] {
			return label
		}
	}
	return ""
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
	if strings.Contains(string(body), "<title>Error 400") {
		return nil, errors.New("error 400 (Bad Request)")
	}
	return body, nil
}

func parseResponse(body []byte) ([]interface{}, error) {
	var result []interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.New("error unmarshaling data")
	}
	if len(result) == 0 {
		return nil, errors.New("no translated data")
	}
	return result, nil
}
