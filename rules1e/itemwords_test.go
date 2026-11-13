package rules1e

import (
	"encoding/json"
	"os"
	"testing"
)

// TestGameItemWords: an item's name is made of its game's own words
// (rules.json itemWords), not POR's: sotsb's word 36 is not POR's.
func TestGameItemWords(t *testing.T) {
	for _, g := range []string{"cotab", "sotsb", "pod"} {
		path := "../data/" + g + "/data/rules.json"
		b, err := os.ReadFile(path)
		if err != nil {
			t.Skip("no imported " + g)
		}
		var raw struct {
			Words []string `json:"itemWords"`
		}
		if err := json.Unmarshal(b, &raw); err != nil {
			t.Fatal(err)
		}
		l, err := ReadLayout("../data/" + g + "/data/" + LayoutFile)
		if err != nil {
			t.Fatal(err)
		}
		tb, err := POR.LoadRules(path, l)
		if err != nil {
			t.Fatalf("%s: %v", g, err)
		}
		r := &RuleState{Tables: tb}
		for n, w := range raw.Words {
			if got := r.itemWord(WordID(n)); got != w {
				t.Fatalf("%s word %d: %q, want %q", g, n, got, w)
			}
		}
	}
}
