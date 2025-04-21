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
	fmt.Println("TTTTTTTTTT", tr)
	getStrategy(sourceLang).PostProcess(tr, raw)
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
