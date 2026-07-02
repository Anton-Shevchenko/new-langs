// Package wiktionary_de resolves the grammatical gender (der/die/das) of a
// German noun using the official de.wiktionary.org MediaWiki API instead of
// scraping HTML pages. The API returns page wikitext that contains a
// "Genus=" field (m/f/n), which maps to der/die/das.
package wiktionary_de

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const apiEndpoint = "https://de.wiktionary.org/w/api.php"

// Wikimedia requires a descriptive User-Agent, otherwise requests may be blocked.
const userAgent = "lang_bot/1.0 (https://github.com/lang_bot; article lookup)"

var httpClient = &http.Client{Timeout: 10 * time.Second}

// genusRe matches lines like "|Genus=m" or "|Genus 1=f".
var genusRe = regexp.MustCompile(`(?mi)^\s*\|\s*Genus(?:\s*\d+)?\s*=\s*(m|f|n)\b`)

var genusToArticle = map[string]string{
	"m": "der",
	"f": "die",
	"n": "das",
}

// Result holds the resolved articles for a word. A noun can have more than one
// valid gender (e.g. "die/der Butter"), so all of them are returned in order.
type Result struct {
	Word     string   `json:"word"`
	Article  string   `json:"article"`  // primary (first) article
	Articles []string `json:"articles"` // all valid articles, deduplicated
}

type apiResponse struct {
	Query struct {
		Pages []struct {
			Missing   bool `json:"missing"`
			Revisions []struct {
				Slots struct {
					Main struct {
						Content string `json:"content"`
					} `json:"main"`
				} `json:"slots"`
			} `json:"revisions"`
		} `json:"pages"`
	} `json:"query"`
}

// Article returns the primary article (der/die/das) for the given German noun.
// It returns an empty string with a nil error if no gender could be resolved.
func Article(word string) (string, error) {
	res, err := Lookup(word)
	if err != nil {
		return "", err
	}
	if res == nil {
		return "", nil
	}
	return res.Article, nil
}

// Lookup fetches the wikitext for the exact word (page titles on
// de.wiktionary.org are case-sensitive) and extracts its gender(s). Only nouns
// carry a gender, and German nouns are always capitalized, so callers should
// pass the word with its correct capitalization.
func Lookup(raw string) (*Result, error) {
	word := strings.TrimSpace(raw)
	if word == "" {
		return nil, fmt.Errorf("empty word")
	}

	content, err := fetchWikitext(word)
	if err != nil {
		return nil, err
	}

	articles := extractArticles(content)
	if len(articles) == 0 {
		return nil, nil
	}

	return &Result{
		Word:     raw,
		Article:  articles[0],
		Articles: articles,
	}, nil
}

func fetchWikitext(word string) (string, error) {
	params := url.Values{}
	params.Set("action", "query")
	params.Set("format", "json")
	params.Set("formatversion", "2")
	params.Set("prop", "revisions")
	params.Set("rvprop", "content")
	params.Set("rvslots", "main")
	params.Set("titles", word)
	params.Set("redirects", "1")

	req, err := http.NewRequest(http.MethodGet, apiEndpoint+"?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("wiktionary HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var parsed apiResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Query.Pages) == 0 {
		return "", nil
	}

	page := parsed.Query.Pages[0]
	if page.Missing || len(page.Revisions) == 0 {
		return "", nil
	}
	return page.Revisions[0].Slots.Main.Content, nil
}

func extractArticles(content string) []string {
	matches := genusRe.FindAllStringSubmatch(content, -1)
	seen := map[string]bool{}
	var articles []string
	for _, m := range matches {
		art, ok := genusToArticle[strings.ToLower(m[1])]
		if !ok || seen[art] {
			continue
		}
		seen[art] = true
		articles = append(articles, art)
	}
	return articles
}
