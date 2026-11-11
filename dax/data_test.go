package dax

import (
	"os"
	"path/filepath"
	"testing"
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
