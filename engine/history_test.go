package engine

import (
	"gobox/ecl"
	"gobox/ui"
	"image/color"
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
	names, err := ecl.ReadNames(os.DirFS("../data/por/scripts"), 0, true, ecl.V1)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile("../data/por/scripts/area0.txt")
	s, err := ecl.Parse(string(b), names)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range []string{"RolfEvent", "RolfHistory", "PhlanHistory1", "PhlanHistory6"} {
		if _, ok := s.Labels[l]; !ok {
			t.Error("area 0 has no label", l)
		}
	}
	es, err := ecl.ReadEntries(os.DirFS("../data/por/scripts"))
	if err != nil {
		t.Fatal(err)
	}
	for _, en := range es {
		if en.Book != "HISTORY.TXT" {
			continue
		}
		c := en.Title
		texts := e.journalReadTexts("HISTORY.TXT", []string{c}, true)
		if len(texts) != 1 || len(texts[0].Paras) < 2 || len(c) > 0x24 {
			t.Fatalf("%q: missing or its title is too long", c)
		}
		for n := 0; ; n++ { // a page with no blank line holds a paragraph longer than it
			pg, _ := ui.JournalPage(c, texts, n, 0, 40, 25, [16]color.RGBA{}, nil)
			blank := false
			for _, l := range pg.Lines {
				blank = blank || len(l) == 0 || strings.TrimSpace(l[0].Text) == ""
			}
			if !blank {
				t.Errorf("%q: paragraph longer than a page on page %d", c, n)
			}
			if n >= pg.Pages-1 {
				break
			}
		}
	}
}
