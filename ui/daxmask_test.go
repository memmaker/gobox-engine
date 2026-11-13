package ui

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"gobox/asset/orig"
)

// Data.DaxMask makes a ref's colour index transparent, wherever it is in the
// picture; no entry keeps it (skipped without buckrogers-go's data/dos). PNG: $DAXMASK_PNG.
func TestDaxMask(t *testing.T) {
	home, _ := os.UserHomeDir()
	l, err := orig.NewLoader(filepath.Join(home, "Projects", "buckrogers-go", "data", "dos"))
	if err != nil || !l.Exists("CHARS.DAX") {
		t.Skip("no buckrogers data")
	}
	useArt(t, l.Dir)
	defer func(m map[string]int) { daxMask = m }(daxMask)
	daxMask = nil
	clear(daxPics)
	pc := Picture{File: "CHARS", Block: 0}
	if picImage(pc, egaInk).(interface{ Opaque() bool }).Opaque() != true {
		t.Fatal("masked without an entry")
	}
	daxMask = map[string]int{"CHARS": 8}
	m := picImage(pc, egaInk)
	want := orig.ParseVGA(l.BlockShared("CHARS", 0), false, 0)
	w, clear8, kept := want.Width*8, 0, 0
	for i, c := range want.Frame(0) {
		_, _, _, a := m.At(i%w, i/w).RGBA()
		switch {
		case c == 8 && a != 0, c != 8 && a == 0:
			t.Fatalf("pixel %d (index %d): alpha %d", i, c, a)
		case c == 8:
			clear8++
		default:
			kept++
		}
	}
	if clear8 == 0 || kept == 0 {
		t.Fatalf("nothing keyed (%d) or kept (%d)", clear8, kept)
	}
	if p := os.Getenv("DAXMASK_PNG"); p != "" {
		f, _ := os.Create(p)
		png.Encode(f, m)
		f.Close()
	}
}
