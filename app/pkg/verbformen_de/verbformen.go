package verbformen_de

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type LookupResult struct {
	Word        string   `json:"word"`
	Article     string   `json:"article,omitempty"`
	IsVerb      bool     `json:"is_verb"`
	Conjugation []string `json:"conjugation,omitempty"`
}

func Lookup(raw string) (*LookupResult, error) {
	word := strings.ToLower(raw)
	html, doc, err := fetchHTML(word)
	if err != nil {
		return nil, fmt.Errorf("fetch error: %w", err)
	}
	res := &LookupResult{Word: raw}
	if art, ok := extractArticle(html); ok {
		res.Article = art
		res.IsVerb = false
	} else {
		res.IsVerb = true
		conj, err := extractConjugation(doc, word)
		if err != nil {
			return nil, fmt.Errorf("conjugation error: %w", err)
		}
		res.Conjugation = conj
	}
	return res, nil
}

func (r *LookupResult) JSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

func fetchHTML(word string) (string, *goquery.Document, error) {
	url := fmt.Sprintf("https://www.verbformen.com/?w=%s", word)
	resp, err := http.Get(url)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("HTTP %d on %s", resp.StatusCode, url)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}
	html := string(data)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", nil, err
	}
	return html, doc, nil
}

func extractArticle(html string) (string, bool) {
	re := regexp.MustCompile(`(?i)article\s*"(?P<art>der|die|das)"`)
	if m := re.FindStringSubmatch(html); m != nil {
		return strings.ToLower(m[1]), true
	}
	return "", false
}

func extractConjugation(doc *goquery.Document, word string) ([]string, error) {
	heading := fmt.Sprintf("%s conjugation", word)
	var lines []string
	found := false
	doc.Find("h3").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if strings.TrimSpace(strings.ToLower(s.Text())) != heading {
			return true
		}
		found = true
		for sib := s.Next(); sib.Length() > 0; sib = sib.Next() {
			if goquery.NodeName(sib) == "h3" {
				break
			}
			text := strings.TrimSpace(sib.Text())
			if text != "" {
				lines = append(lines, text)
			}
		}
		return false
	})
	if !found {
		return nil, fmt.Errorf("conjugation for %q not found", word)
	}
	return lines, nil
}
