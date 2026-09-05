package wordTranslator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	browserUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	googleDictionaryURL = "https://translate.googleapis.com/translate_a/single?client=gtx&dt=t&dt=ex&dt=md&sl=%s&tl=%s&dt=t&dt=rm&dt=at&q=%s"
	googleChromeSingleURL = "https://clients5.google.com/translate_a/single?client=dict-chrome-ex&dt=t&dt=ex&dt=md&sl=%s&tl=%s&dt=rm&dt=at&q=%s"
	googleChromeDictURL   = "https://clients5.google.com/translate_a/t?client=dict-chrome-ex&sl=%s&tl=%s&q=%s"
	myMemoryURL           = "https://api.mymemory.translated.net/get?q=%s&langpair=%s|%s"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

func Translate(source, sourceLang, targetLang string) (*TranslateResult, error) {
	encoded := url.QueryEscape(source)
	var errs []string

	// Chrome-dict first: the `single` URL returns the same dictionary JSON as
	// gtx (alternatives, POS, examples) and is less often rate-limited.
	if raw, err := fetchGoogleArray(fmt.Sprintf(googleChromeSingleURL, sourceLang, targetLang, encoded)); err == nil {
		return finishGoogleResult(source, sourceLang, targetLang, raw), nil
	} else {
		errs = append(errs, "chrome-single: "+err.Error())
	}

	if text, err := fetchChromeDict(encoded, sourceLang, targetLang); err == nil && text != "" {
		return simpleResult(source, sourceLang, targetLang, text), nil
	} else if err != nil {
		errs = append(errs, "chrome-t: "+err.Error())
	}

	if text, err := fetchMyMemory(encoded, sourceLang, targetLang); err == nil && text != "" {
		return simpleResult(source, sourceLang, targetLang, text), nil
	} else if err != nil {
		errs = append(errs, "mymemory: "+err.Error())
	}

	if raw, err := fetchGoogleArray(fmt.Sprintf(googleDictionaryURL, sourceLang, targetLang, encoded)); err == nil {
		return finishGoogleResult(source, sourceLang, targetLang, raw), nil
	} else {
		errs = append(errs, "gtx: "+err.Error())
	}

	fmt.Printf("translation providers failed for %q (%s->%s): %s\n", source, sourceLang, targetLang, strings.Join(errs, "; "))
	return nil, fmt.Errorf("all translation providers failed: %s", strings.Join(errs, "; "))
}

func finishGoogleResult(source, sourceLang, targetLang string, raw []interface{}) *TranslateResult {
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
	return tr
}

func simpleResult(source, sourceLang, targetLang, text string) *TranslateResult {
	tr := &TranslateResult{
		SourceWord:      source,
		SourceLang:      sourceLang,
		TranslationLang: targetLang,
		Translations:    []string{text},
		IsValid:         true,
		IsSimpleWord:    len(strings.Fields(source)) == 1,
	}
	if sourceLang == "de" || targetLang == "de" {
		enrichGerman(tr)
	}
	return tr
}

func fetchGoogleArray(rawURL string) ([]interface{}, error) {
	body, err := doGET(rawURL)
	if err != nil {
		return nil, err
	}
	return parseResponse(body)
}

func fetchChromeDict(encoded, sourceLang, targetLang string) (string, error) {
	body, err := doGET(fmt.Sprintf(googleChromeDictURL, sourceLang, targetLang, encoded))
	if err != nil {
		return "", err
	}
	return parseChromeDict(body)
}

func fetchMyMemory(encoded, sourceLang, targetLang string) (string, error) {
	body, err := doGET(fmt.Sprintf(myMemoryURL, encoded, sourceLang, targetLang))
	if err != nil {
		return "", err
	}
	return parseMyMemory(body)
}

func doGET(rawURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", browserUA)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("error reading response body")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	if len(body) == 0 || body[0] == '<' {
		return nil, errors.New("non-JSON response")
	}
	return body, nil
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

	encoded := url.QueryEscape(word)
	reqURL := fmt.Sprintf(
		"https://clients5.google.com/translate_a/single?client=dict-chrome-ex&sl=%s&tl=%s&dt=t&dt=bd&q=%s",
		lang, target, encoded,
	)
	body, err := doGET(reqURL)
	if err != nil {
		reqURL = fmt.Sprintf(
			"https://translate.googleapis.com/translate_a/single?client=gtx&sl=%s&tl=%s&dt=t&dt=bd&q=%s",
			lang, target, encoded,
		)
		body, err = doGET(reqURL)
	}
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

func parseChromeDict(body []byte) (string, error) {
	var raw []interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", err
	}
	text := firstString(raw)
	if text == "" {
		return "", errors.New("empty translation")
	}
	return text, nil
}

type myMemoryResponse struct {
	ResponseData struct {
		TranslatedText string `json:"translatedText"`
	} `json:"responseData"`
	ResponseStatus int `json:"responseStatus"`
}

func parseMyMemory(body []byte) (string, error) {
	var parsed myMemoryResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if parsed.ResponseStatus != 0 && parsed.ResponseStatus != http.StatusOK {
		return "", fmt.Errorf("mymemory status %d", parsed.ResponseStatus)
	}
	text := strings.TrimSpace(parsed.ResponseData.TranslatedText)
	if text == "" || strings.Contains(strings.ToUpper(text), "MYMEMORY") {
		return "", errors.New("mymemory returned no translation")
	}
	return text, nil
}

func firstString(raw []interface{}) string {
	if len(raw) == 0 {
		return ""
	}
	if s, ok := raw[0].(string); ok {
		return strings.TrimSpace(s)
	}
	if nested, ok := raw[0].([]interface{}); ok {
		return firstString(nested)
	}
	return ""
}

