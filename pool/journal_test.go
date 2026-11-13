package pool

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"

	"gobox/game/play"
)

// TestJournalSitesNamed: no imported script prints a reference to the
// printed journal ("ENTRY 26", "ENTRIES 23 AND 14", "PROCLAMATION LIX"): the
// import's special cases name each through Journal (bug class: a journal
// text opened by matching printed text). Each Journal names an entry
// (Parse resolves the Entry values).
func TestJournalSitesNamed(t *testing.T) {
	ref := regexp.MustCompile(`\b(?:ENTRY\s*\d|ENTRIES\s*\d|PROCLAMATIONS?\s*(?:[IVXLC]+\b|"))`)
	files, err := filepath.Glob("../data/por/sources/assets/scripts/area*.txt")
	if err != nil || len(files) == 0 {
		t.Skip("no imported scripts", err)
	}
	journals := 0
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for i, l := range strings.Split(string(b), "\n") {
			if (strings.Contains(l, "Text(") || strings.Contains(l, "Print(")) && ref.MatchString(l) {
				t.Errorf("%s:%d prints a journal reference: %s", f, i+1, strings.TrimSpace(l))
			}
			if strings.Contains(l, "Journal(") {
				journals++
			}
		}
	}
	if journals < 23+50 {
		t.Errorf("%d Journal commands", journals)
	}
}

// TestShowEntryTogether: one script's Journal commands are read together,
// in one pager, and each journal text is noted once.
func TestShowEntryTogether(t *testing.T) {
	e := quickStarted(t)
	e.showEntry(e.entryNamed("JournalEntry23"))
	e.showEntry(e.entryNamed("JournalEntry14"))
	e.showEntry(e.entryNamed("JournalEntry23"))
	r := e.ext.read
	if r == nil || !slices.Contains(r.Lines, "\x00Journal Entry 23") || !slices.Contains(r.Lines, "\x00Journal Entry 14") {
		t.Fatalf("read %+v", r)
	}
	n := 0
	for _, no := range e.ext.notes {
		if no.Title == "Journal Entry 23" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("Journal Entry 23 noted %d times", n)
	}
}

// TestClueBook: the clue book's chapters are unique and its pictures are there.
func TestClueBook(t *testing.T) {
	e := testEngine(t)
	seen := map[string]bool{}
	for _, c := range e.clueChapters() {
		if seen[c] {
			t.Errorf("chapter %q twice", c)
		}
		seen[c] = true
	}
	for _, paras := range oldParse(t, "CLUEBOOK.TXT") {
		for _, p := range paras {
			if n := len(p); n > 2 && p[0] == '<' && p[n-1] == '>' && e.journalPicture(p[1:n-1]) == nil {
				t.Errorf("picture %s missing", p)
			}
		}
	}
	if len(seen) < 30 || e.journalPicture("CLUE04A.PNG") == nil {
		t.Errorf("%d chapters", len(seen))
	}
}

// oldParse is por's journal format as it read it before gobox did: each
// title's lines.
func oldParse(t *testing.T, name string) map[string][]string {
	b, err := os.ReadFile("../games/por/importpor/journal/" + name)
	if err != nil {
		t.Fatal(err)
	}
	j := map[string][]string{}
	var title string
	for _, l := range strings.Split(strings.ReplaceAll(string(b), "\r", ""), "\n") {
		l = strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(l, "#"):
		case strings.HasPrefix(l, "[") && strings.HasSuffix(l, "]"):
			title = l[1 : len(l)-1]
			j[title] = nil
		case title != "":
			j[title] = append(j[title], l)
		}
	}
	return j
}

// TestJournalNotes: the notes' texts (journalTexts, through gobox's Book)
// are each journal text's paragraphs, joined and without pictures, as por
// made them from the lines itself.
func TestJournalNotes(t *testing.T) {
	e := testEngine(t)
	old := oldParse(t, "JOURNAL.TXT")
	got := e.journalTexts()
	if len(got) != len(old) || len(got) < 90 {
		t.Fatalf("%d texts, want %d", len(got), len(old))
	}
	for title, ps := range old {
		var paras []string
		for _, p := range ps {
			switch {
			case strings.HasPrefix(p, "<") && strings.HasSuffix(p, ">"):
			case p == "" || len(paras) == 0:
				paras = append(paras, p)
			default:
				paras[len(paras)-1] = strings.TrimSpace(paras[len(paras)-1] + " " + p)
			}
		}
		paras = slices.DeleteFunc(paras, func(p string) bool { return p == "" })
		if !reflect.DeepEqual(got[title], paras) {
			t.Errorf("%s: %q, want %q", title, got[title], paras)
		}
	}
}

// A Read is only published: a decision made while one is up is never
// refused, and clears it; each new text counts Seq up (BUG-CLASSES 2).
func TestReadPublishOnly(t *testing.T) {
	e := testEngine(t)
	seed := uint32(7)
	e.FixedSeed = &seed
	e.Explore = true
	e.Init()
	e.Start(e.Play)
	handout := func() bool { sc, ok := e.State().(play.Scene); return ok && sc.Kind == play.SceneHandout }
	e.showJournal(journalBox, []string{"one"}) // Seq is gobox's now (each Scene a new one): not checked here
	if !handout() {
		t.Fatalf("first Read: %T, want a handout Scene", e.State())
	}
	if _, err := e.Do(play.Leave{}); err != nil {
		t.Fatalf("Leave under a Read: %v", err)
	}
	if handout() {
		t.Fatal("Leave left the Read up")
	}
	e.showJournal(journalBox, []string{"two"})
	if !handout() {
		t.Fatalf("second Read: %T, want a handout Scene", e.State())
	}
}

// TestOwnJournalBooks: the books are the game's own, from its data folder:
// a folder with another game's book gives that text, and one without POR's
// books has none of POR's.
func TestOwnJournalBooks(t *testing.T) {
	adv := t.TempDir()
	dir := filepath.Join(adv, JournalDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "JOURNAL.TXT"), []byte("[Own Entry]\nthe game's own words\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	e := &Engine{adv: adv}
	if got := e.journalTexts()["Own Entry"]; !slices.Equal(got, []string{"the game's own words"}) {
		t.Errorf("own text %q", got)
	}
	if _, ok := e.journalTexts()["Journal Entry 26"]; ok {
		t.Error("POR's text in another game's folder")
	}
	if e.book("CLUEBOOK.TXT") != nil || len(e.clueChapters()) != 0 || e.journalPicture("CLUE04A.PNG") != nil {
		t.Error("POR's clue book in another game's folder")
	}
}
