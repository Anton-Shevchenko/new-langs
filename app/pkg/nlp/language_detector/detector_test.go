package language_detector

import (
	"testing"

	"github.com/pemistahl/lingua-go"
)

func TestDetect(t *testing.T) {
	cases := []struct {
		name  string
		word  string
		langs []string
		want  string
	}{
		{"english", "hello", []string{"en", "uk"}, "en"},
		{"ukrainian", "привіт", []string{"en", "uk"}, "uk"},
		{"german", "Wohnung", []string{"de", "uk"}, "de"},
		{"ukrainian among de", "квартира", []string{"de", "uk"}, "uk"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Detect(tc.word, tc.langs)
			if err != nil {
				t.Fatalf("Detect(%q) returned error: %v", tc.word, err)
			}
			if got != tc.want {
				t.Errorf("Detect(%q) = %q, want %q", tc.word, got, tc.want)
			}
		})
	}
}

// TestDetectorCache verifies that a detector is cached per language set and
// that the set is treated as order-independent (the lingua detector type is
// not comparable, so we assert on the cache entries rather than on instances).
func TestDetectorCache(t *testing.T) {
	detectorCache.Delete("en,uk")
	detectorCache.Delete("de,uk")

	getDetector([]string{"en", "uk"})
	if _, ok := detectorCache.Load("en,uk"); !ok {
		t.Fatal("expected detector to be cached under sorted key \"en,uk\"")
	}

	countKeys := func() int {
		n := 0
		detectorCache.Range(func(_, _ any) bool { n++; return true })
		return n
	}
	before := countKeys()

	// Same set in a different order must hit the same cache entry.
	getDetector([]string{"uk", "en"})
	if after := countKeys(); after != before {
		t.Errorf("language order created a new cache entry: %d -> %d", before, after)
	}

	// A different set must create its own entry.
	getDetector([]string{"de", "uk"})
	if _, ok := detectorCache.Load("de,uk"); !ok {
		t.Error("expected a separate cache entry for a different language set")
	}
}

func TestCacheKey(t *testing.T) {
	if got := cacheKey([]string{"uk", "en"}); got != "en,uk" {
		t.Errorf("cacheKey not order-independent: got %q", got)
	}
}

// BenchmarkDetectCached measures Detect with the detector cache in place.
func BenchmarkDetectCached(b *testing.B) {
	langs := []string{"en", "uk"}
	getDetector(langs) // warm the cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Detect("hello", langs)
	}
}

// BenchmarkDetectRebuild measures the old behaviour: rebuilding the lingua
// detector on every call.
func BenchmarkDetectRebuild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		detector := lingua.NewLanguageDetectorBuilder().
			FromLanguages(lingua.English, lingua.Ukrainian).Build()
		_, _ = detector.DetectLanguageOf("hello")
	}
}
