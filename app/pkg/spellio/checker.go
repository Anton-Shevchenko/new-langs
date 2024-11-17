package spellio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const SpellingMistake = "Spelling mistake"

type APIResponse struct {
	Software Software `json:"software"`
	Warnings Warnings `json:"warnings"`
	Language Language `json:"language"`
	Matches  []Match  `json:"matches"`
}

type Software struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	ApiVersion  int    `json:"apiVersion"`
	Premium     bool   `json:"premium"`
	PremiumHint string `json:"premiumHint"`
	Status      string `json:"status"`
}

type Warnings struct {
	IncompleteResults bool `json:"incompleteResults"`
}

type Language struct {
	Name             string           `json:"name"`
	Code             string           `json:"code"`
	DetectedLanguage DetectedLanguage `json:"detectedLanguage"`
}

type DetectedLanguage struct {
	Name       string  `json:"name"`
	Code       string  `json:"code"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
}

type Match struct {
	Message                     string        `json:"message"`
	ShortMessage                string        `json:"shortMessage"`
	Replacements                []Replacement `json:"replacements"`
	Offset                      int           `json:"offset"`
	Length                      int           `json:"length"`
	Context                     Context       `json:"context"`
	Sentence                    string        `json:"sentence"`
	Type                        Type          `json:"type"`
	Rule                        Rule          `json:"rule"`
	IgnoreForIncompleteSentence bool          `json:"ignoreForIncompleteSentence"`
	ContextForSureMatch         int           `json:"contextForSureMatch"`
}

type Replacement struct {
	Value string `json:"value"`
}

type Context struct {
	Text   string `json:"text"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

type Type struct {
	TypeName string `json:"typeName"`
}

type Rule struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	IssueType   string   `json:"issueType"`
	Category    Category `json:"category"`
	IsPremium   bool     `json:"isPremium"`
	Confidence  float64  `json:"confidence"`
}

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Result struct {
	Message      string
	Replacements []string
}

func Check(text string, lang string) *Result {
	if lang == "en" {
		lang = "en-US"
	}
	if lang == "uk" {
		lang = "uk-UA"
	}
	if lang == "de" {
		lang = "de-DE"
	}
	baseURL := "https://api.languagetool.org/v2/check"
	params := url.Values{}
	params.Add("text", text)
	params.Add("level", "picky")
	params.Add("language", lang)

	url := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	reqBody := bytes.NewBuffer(nil)

	resp, err := http.Post(url, "application/json", reqBody)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil
	}

	var apiResponse APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		fmt.Printf("Error unmarshaling response: %v\n", err)
		return nil
	}

	for _, match := range apiResponse.Matches {
		// check
		if len(match.Replacements) > 0 {
			return &Result{
				Message:      SpellingMistake,
				Replacements: getReplacements(match.Replacements),
			}
		}
	}

	return nil
}

func getReplacements(replacements []Replacement) []string {
	var values []string

	for i, item := range replacements {
		if i == 5 {
			break
		}
		values = append(values, item.Value)
	}

	return values
}
