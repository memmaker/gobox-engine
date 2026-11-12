package pool

import (
	"gobox/ecl/script"
	"os"
	"strings"
	"testing"
)

// Rolf's history of Phlan (extra): after the tour, Rolf waits in New Phlan
// at 2,2 (area 0, RolfEvent) and tells the Adventurer's Journal's "A History
// of Phlan and the Moonsea Reaches" by Jeff Grubb, a chapter at a time
// (journal/HISTORY.TXT; entries History1..6, Journal), voiced a paragraph at
// a time.

// TestPhlanHistory: Rolf's labels parse, every chapter is there, its title
// fits the page, and every paragraph fits one page under the title (the
// pager keeps a voiced paragraph whole).
func TestPhlanHistory(t *testing.T) {
	e := testEngine(t)
	names, err := script.ReadNames(os.DirFS("../../por-go/data/assets/scripts"), 0, true, script.V1)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile("../../por-go/data/assets/scripts/area0.txt")
	s, err := script.Parse(string(b), names)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range []string{"RolfEvent", "RolfHistory", "PhlanHistory1", "PhlanHistory6"} {
		if _, ok := s.Labels[l]; !ok {
			t.Error("area 0 has no label", l)
		}
	}
	const rows = 25 - 2 - 1 - jrTop - 2 // jrRows at 40x25, less the title and its blank line
	es, err := script.ReadEntries(os.DirFS("../../por-go/data/assets/scripts"))
	if err != nil {
		t.Fatal(err)
	}
	for _, en := range es {
		if en.Book != "HISTORY.TXT" {
			continue
		}
		c := en.Title
		lines := e.journalLayout("HISTORY.TXT", []string{c}, true)
		if len(lines) < 3 || len(c) > jrWidth {
			t.Fatalf("%q: missing or its title is too long", c)
		}
		run := 0
		for _, l := range lines[2:] {
			if run++; l == "" {
				run = 0
			} else if run > rows {
				t.Errorf("%q: paragraph longer than a page at %q", c, strings.TrimLeft(l, "\x03"))
			}
		}
	}
}
