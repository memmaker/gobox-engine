package dax

import (
	"os"
	"path/filepath"
	"testing"

	"gobox/asset"
)

type rec struct {
	A byte
	B uint16
}

func TestRecords(t *testing.T) {
	got := Records[rec]([]byte{1, 2, 0, 3, 4, 0, 9})
	if len(got) != 2 || got[0] != (rec{1, 2}) || got[1] != (rec{3, 4}) {
		t.Fatal(got)
	}
	p := filepath.Join(t.TempDir(), "ITEMS")
	os.WriteFile(p, []byte{1, 0, 5, 6, 0}, 0o644)
	tab, err := Table[rec](p, 2)
	if err != nil || tab[0] != (rec{5, 6}) || tab[1] != (rec{}) {
		t.Fatal(tab, err)
	}
}

// TestData: Data reads back what WriteBlocks and WriteJSON wrote; files that
// are not there are none.
func TestData(t *testing.T) {
	dir := t.TempDir()
	if err := asset.WriteBlocks(dir, "GEO2", "test", map[int][]byte{3: {7, 8}}); err != nil {
		t.Fatal(err)
	}
	if err := asset.WriteJSON(dir, "MON2", map[int]Monster[rec, rec, rec]{5: {Record: rec{1, 2}, Items: []rec{{3, 4}}}}); err != nil {
		t.Fatal(err)
	}
	if err := asset.WriteJSON(dir, "ITEM2", map[int][]rec{1: {{9, 9}}}); err != nil {
		t.Fatal(err)
	}
	d := OpenData[rec, rec, rec](dir)
	if b := d.Block("GEO2", 3); len(b) != 2 || b[1] != 8 || d.Block("GEO2", 4) != nil || d.Block("GEO9", 3) != nil {
		t.Error("blocks")
	}
	if m, ok := d.Monster("MON2", 5); !ok || m.Record != (rec{1, 2}) || len(m.Items) != 1 {
		t.Error("monster", m, ok)
	}
	if _, ok := d.Monster("MON3", 5); ok {
		t.Error("monster of a missing file")
	}
	if it := d.Items("ITEM2", 1); len(it) != 1 || it[0] != (rec{9, 9}) {
		t.Error("items", it)
	}
}

// TestMonstersEmptyBlock: an empty CHA block is no monster (no entry), so
// the engine's Monster lookup fails as the original's load did.
func TestMonstersEmptyBlock(t *testing.T) {
	dir := t.TempDir()
	// two index entries (id, offset, raw size 0 = stored raw, stored size): id 1 empty, id 2 three bytes
	idx := []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 3, 0}
	b := append([]byte{byte(len(idx)), 0}, idx...)
	b = append(b, 7, 8, 0)
	if err := os.WriteFile(filepath.Join(dir, "MON0CHA.DAX"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	l, err := NewLoader(dir)
	if err != nil {
		t.Fatal(err)
	}
	m := Monsters[rec, rec, rec](l, "MON0")
	if _, ok := m[1]; ok {
		t.Error("empty block loaded as a monster")
	}
	if r, ok := m[2]; !ok || r.Record != (rec{7, 8}) {
		t.Error("monster 2", r, ok)
	}
}
