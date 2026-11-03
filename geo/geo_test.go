package geo

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestMap(t *testing.T) {
	b := make([]byte, Size)
	i := 2 + 3<<4 + 4 // x=4,y=3
	b[i] = 0x12
	b[0x100+i] = 0x34
	b[0x200+i] = 0x99
	b[0x300+i] = 0b11100100
	m, err := ParseMap(b)
	if err != nil {
		t.Fatal(err)
	}
	for d, w := range map[int]byte{0: 1, 2: 2, 4: 3, 6: 4} {
		if m.WallAt(4, 3, d) != w || m.DoorAt(4, 3, d) != byte(d/2) {
			t.Fatal("dir", d, m.WallAt(4, 3, d), m.DoorAt(4, 3, d))
		}
	}
	if m.MapCell(4, 3) != 0x99 {
		t.Fatal("cell")
	}
	m.SolidEdges = true
	if m.WallAt(16, 3, 0) != 0 || m.DoorAt(-1, 0, 0) != 0 {
		t.Fatal("solid")
	}
	m.SolidEdges = false
	if m.MapCell(-1, 3) != m.MapCell(15, 3) {
		t.Fatal("wrap")
	}
	if _, err := ParseMap(b[:10]); err == nil {
		t.Fatal("size")
	}
}

func TestWalldef(t *testing.T) {
	b := make([]byte, 2*SetSize+5)
	b[0], b[1], b[SetSize] = 0x2C, 0x2D, 0x50
	w, err := ParseWalldef(b)
	if err != nil || len(w.Sets) != 2 {
		t.Fatal(err, len(w.Sets))
	}
	w.Relocate(0, 45)
	if w.Sets[0][0] != 0x2C || w.Sets[0][1] != 0x2D+45 || w.Sets[1][0] != 0x50 {
		t.Fatal("reloc")
	}
	if !w.Fits(2) || w.Fits(3) || SymbolBlock(7, 2, 2) != 72 || SymbolBlock(7, 1, 1) != 7 {
		t.Fatal("fits")
	}
	if _, err := ParseWalldef(b[:10]); err == nil {
		t.Fatal("short")
	}
}

// Real data: skips when the DOS files are missing.
func TestRealData(t *testing.T) {
	files, _ := filepath.Glob("../../por-go/data/dos/GEO*.DAX")
	if len(files) == 0 {
		t.Skip("no game files")
	}
	n := 0
	for _, f := range files {
		for id, b := range daxBlocks(t, f) {
			m, err := ParseMap(b)
			if err != nil {
				t.Fatalf("%s %d: %v", f, id, err)
			}
			_ = m.MapCell(3, 3)
			n++
		}
	}
	if n == 0 {
		t.Fatal("no blocks")
	}
	for _, f := range mustGlob("../../por-go/data/dos/WALLDEF*.DAX") {
		for id, b := range daxBlocks(t, f) {
			if _, err := ParseWalldef(b); err != nil {
				t.Fatalf("%s %d: %v", f, id, err)
			}
		}
	}
}

func mustGlob(p string) []string { f, _ := filepath.Glob(p); return f }

// daxBlocks is a minimal DAX reader (test only; geo must not import dax).
func daxBlocks(t *testing.T, path string) map[int][]byte {
	b, err := os.ReadFile(path)
	if err != nil {
		t.Skip(err)
	}
	n := int(binary.LittleEndian.Uint16(b))
	data := b[2+n:]
	out := map[int][]byte{}
	for i := 2; i+9 <= 2+n; i += 9 {
		off := int(binary.LittleEndian.Uint32(b[i+1:]))
		raw := int(binary.LittleEndian.Uint16(b[i+5:]))
		src := data[off : off+int(binary.LittleEndian.Uint16(b[i+7:]))]
		var o []byte
		for j := 0; j < len(src) && len(o) < raw; {
			c := int(int8(src[j]))
			if c >= 0 {
				e := min(j+c+2, len(src))
				o = append(o, src[j+1:e]...)
				j = e
			} else {
				for k := 0; k < -c && j+1 < len(src); k++ {
					o = append(o, src[j+1])
				}
				j += 2
			}
		}
		out[int(b[i])] = o[:min(len(o), raw)]
	}
	return out
}
