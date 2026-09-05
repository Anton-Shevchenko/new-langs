package wordTranslator

import (
	"testing"
)

func TestParseResponseRejectsHTML(t *testing.T) {
	_, err := parseResponse([]byte(`<html><title>Sorry...</title></html>`))
	if err == nil {
		t.Fatal("expected error for HTML body")
	}
}

func TestParseResponseAcceptsGoogleArray(t *testing.T) {
	raw, err := parseResponse([]byte(`[[["Будинок","Haus"]],null,"de"]`))
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) == 0 {
		t.Fatal("empty result")
	}
}

func TestParseChromeDict(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		want    string
		wantErr bool
	}{
		{"flat", `["Будинок"]`, "Будинок", false},
		{"nested", `[["house","Haus"]]`, "house", false},
		{"empty", `[]`, "", true},
		{"html", `<html>nope</html>`, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseChromeDict([]byte(tc.body))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("got %q, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseMyMemory(t *testing.T) {
	ok := `{"responseData":{"translatedText":"Будинок"},"responseStatus":200}`
	got, err := parseMyMemory([]byte(ok))
	if err != nil {
		t.Fatal(err)
	}
	if got != "Будинок" {
		t.Fatalf("got %q", got)
	}

	if _, err := parseMyMemory([]byte(`{"responseData":{"translatedText":"MYMEMORY WARNING"},"responseStatus":200}`)); err == nil {
		t.Fatal("expected error for MyMemory warning payload")
	}
}

func TestSimpleResult(t *testing.T) {
	tr := simpleResult("hello", "en", "uk", "привіт")
	if !tr.IsValid || !tr.IsSimpleWord {
		t.Fatalf("flags: valid=%v simple=%v", tr.IsValid, tr.IsSimpleWord)
	}
	if len(tr.Translations) != 1 || tr.Translations[0] != "привіт" {
		t.Fatalf("translations: %v", tr.Translations)
	}
}
