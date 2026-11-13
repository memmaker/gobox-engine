package engine

import (
	"os"
	"testing"

	"gobox/ecl"
)

// Every long text entry (names/entries.json) is a text of its book, and
// the scripts name it (Journal(Name)).
func TestEntriesInBooks(t *testing.T) {
	e := testEngine(t)
	data := os.DirFS("../data/por/sources/assets/scripts")
	es, err := ecl.ReadEntries(data)
	if err != nil || len(es) == 0 {
		t.Fatalf("entries: %v (%d)", err, len(es))
	}
	n, err := ecl.ReadNames(data, 0, false, ecl.V1)
	if err != nil {
		t.Fatal(err)
	}
	for _, en := range es {
		if b := e.book(en.Book); b == nil || !b.Has(en.Title) {
			t.Errorf("entry %d %s: no %q in %s", en.ID, en.Name, en.Title, en.Book)
		}
		if nm := ecl.ValueName(n.Values, "Entry", n.Set, int(en.ID)); nm != en.Name {
			t.Errorf("entry %d: the scripts name it %q, want %s", en.ID, nm, en.Name)
		}
	}
	sc, err := ecl.Parse("version 1 2\non load {\n    Journal(TavernTale3)\n    exit\n}\non move { exit }\non step { exit }\non camp { exit }\non rest { exit }\n", n)
	if err != nil {
		t.Fatal(err)
	}
	if in := sc.Code[0]; in.Op != ecl.OpJournal || len(in.Args) != 1 || in.Args[0].Type != ecl.ArgWord || in.Args[0].N != 1003 { // a word: a byte operand reads as its low byte
		t.Errorf("Journal(TavernTale3) is %+v, want the pager on entry 1003", in)
	}
}
