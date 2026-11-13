package engine

import (
	"fmt"
	"gobox/rules1e"
	"os"
	"path/filepath"
	"testing"

	"gobox/asset/orig"
	"gobox/char"
)

// maxTurnKind is the last row of the turn table; a higher record would read past it.
const maxTurnKind = 10

// checkRecordInTables recomputes a record, which panics on a read past a table, and
// checks the bounds the vectors cover. Levels here bound rules1e's por-range (-1, fighter 12).
func checkRecordInTables(t *testing.T, where string, b []byte) {
	t.Helper()
	c := &rules1e.Character{}
	c.Sheet = char.SheetOf(b)
	(&Engine{}).state().RecomputeCharacter(c)
	if c.Undead > maxTurnKind {
		t.Errorf("%s %s: undead kind %d", where, c.Name(), c.Undead)
	}
	for k, lvl := range c.Levels {
		if l := int8(lvl); l < -1 || l > 12 {
			t.Errorf("%s %s: class %d level %d outside the vectors", where, c.Name(), k, l)
		}
	}
}

// Every record the data can put in play stays inside the tables looked up by level or
// stat. NPCs join through AddMonsterToParty from these same MON<n>CHA files (no other
// NPC records exist). Saved characters are local files, so they count when present.
func TestMonsterRecordsStayInTables(t *testing.T) {
	files, err := filepath.Glob("../data/por/sources/dos/MON*CHA.DAX")
	if err != nil || len(files) == 0 {
		t.Fatalf("no monster files: %v", err)
	}
	for _, path := range files {
		f, err := orig.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range f.Entries {
			checkRecordInTables(t, fmt.Sprintf("%s %d", path, entry.ID), f.BlockShared(entry.ID))
		}
	}
	saved, err := filepath.Glob("../data/por/sources/save/*.CHA")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range saved {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		checkRecordInTables(t, path, b)
	}
}
