package main

import (
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Buck Rogers' VGA WALLDEF/8X8D blocks dump as pictures, not as transparent EGA misreads
// (skipped without buckrogers-go's data/dos).
func TestViewdumpVGA(t *testing.T) {
	data := "../../../../buckrogers-go/data/dos"
	if _, err := os.Stat(filepath.Join(data, "WALLDEF1.DAX")); err != nil {
		t.Skip("no Buck Rogers data")
	}
	out := t.TempDir()
	if b, err := exec.Command("go", "run", ".", "-transparent", data, out).CombinedOutput(); err != nil {
		t.Fatal(err, string(b))
	}
	f, err := os.Open(filepath.Join(out, "WALLDEF1-007-1-wall1.png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	m, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	opaque := 0
	b := m.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if _, _, _, a := m.At(x, y).RGBA(); a != 0 {
				opaque++
			}
		}
	}
	if opaque < b.Dx()*b.Dy()/2 {
		t.Fatalf("%d of %d pixels drawn", opaque, b.Dx()*b.Dy())
	}
}
