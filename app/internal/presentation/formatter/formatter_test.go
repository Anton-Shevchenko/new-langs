package formatter

import (
	"strings"
	"testing"
)

func TestFormatWordMessageWithArticle(t *testing.T) {
	got := FormatWordMessageWithArticle("квартира", "die Wohnung", "die")
	if !strings.Contains(got, "квартира") || !strings.Contains(got, "die Wohnung") {
		t.Fatalf("pair missing: %s", got)
	}
	if !strings.Contains(got, "die") {
		t.Fatalf("article mention missing: %s", got)
	}
}

func TestFormatWordMessageOmitsArticleLine(t *testing.T) {
	got := FormatWordMessage("дім", "house")
	if strings.Contains(strings.ToLower(got), "article") || strings.Contains(got, "артикль") {
		t.Fatalf("custom/plain save should not mention article: %s", got)
	}
}
