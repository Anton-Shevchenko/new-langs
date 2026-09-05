package spellio

import "testing"

func TestIsCapitalizationOnly(t *testing.T) {
	if !isCapitalizationOnly("auto", []string{"Auto", "Judo", "auch"}) {
		t.Fatal("auto → Auto must be treated as capitalization, not a typo")
	}
	if !isCapitalizationOnly("Haus", []string{"haus"}) {
		t.Fatal("same word with different case")
	}
	if isCapitalizationOnly("wohnunf", []string{"Wohnung", "wohnen"}) {
		t.Fatal("real typo must not be ignored")
	}
	if isCapitalizationOnly("auto", nil) {
		t.Fatal("empty replacements")
	}
}
