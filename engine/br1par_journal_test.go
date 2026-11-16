package engine

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// TestBR1JournalEntries ports rules TestLogBook's data checks: BR1's Log Book has the 71 entries the game names by number
// (the numbers 1..71, each once), and every picture a text names is in the journal's folder.
func TestBR1JournalEntries(t *testing.T) {
	dir := filepath.Join("..", "importer", "br1", "journal")
	b, err := os.ReadFile(filepath.Join(dir, "JOURNAL.TXT"))
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]int{}
	for _, m := range regexp.MustCompile(`(?m)^\[Log Entry (\d+)\]$`).FindAllStringSubmatch(string(b), -1) {
		seen[m[1]]++
	}
	if len(seen) != 71 {
		t.Errorf("%d entries, want 71", len(seen))
	}
	for n, c := range seen {
		if c != 1 {
			t.Errorf("entry %s appears %d times", n, c)
		}
	}
	for _, m := range regexp.MustCompile(`(?m)^<(.+\.PNG)>$`).FindAllStringSubmatch(string(b), -1) {
		if _, err := os.Stat(filepath.Join(dir, m[1])); err != nil {
			t.Errorf("picture %s: %v", m[1], err)
		}
	}
}
