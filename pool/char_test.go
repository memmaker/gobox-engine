package pool

import (
	"bytes"
	"encoding/binary"
	"gobox/char"
	"gobox/dax"
	"rules1e"
	"testing"
	"unsafe"
)

func TestCharRecordLayout(t *testing.T) {
	var r CharRecord
	if n := binary.Size(&r); n != charRecSize {
		t.Fatalf("record is %#x bytes, want %#x", n, charRecSize)
	}
	// every field is bytes or little-endian words, so Go lays the struct out
	// without padding exactly where the DOS record has it
	for _, f := range []struct {
		name string
		off  uintptr
		want uintptr
	}{
		{"MaxHP", unsafe.Offsetof(r.MaxHP), 0x32},
		{"Control", unsafe.Offsetof(r.Control), 0x84},
		{"Levels", unsafe.Offsetof(r.Levels), 0x96},
		{"Exp", unsafe.Offsetof(r.Exp), 0xAC},
		{"ItemCount", unsafe.Offsetof(r.ItemCount), 0xC7},
		{"Hands", unsafe.Offsetof(r.Hands), 0x100},
		{"Status", unsafe.Offsetof(r.Status), 0x10C},
		{"HP", unsafe.Offsetof(r.HP), 0x11B},
	} {
		if f.off != f.want {
			t.Errorf("%s at %#x, want %#x", f.name, f.off, f.want)
		}
	}
	b := make([]byte, charRecSize)
	for i := range b {
		b[i] = byte(i * 7)
	}
	for _, o := range []int{0x7F, 0xC8, 0xCC, 0x104} { // pointers are dropped
		clear(b[o : o+4])
	}
	clear(b[0xCC:0x100])
	clear(b[0x104:0x10C])
	b[0] = 5 // the name: its letters are kept, the bytes past them are not (sheetOf)
	clear(b[6:0x10])
	var c rules1e.Character
	c.Sheet = char.SheetOf(b)
	if !bytes.Equal(char.SheetBytes(c.Sheet), b) {
		t.Fatal("record does not round-trip")
	}
}

func TestItemAffectRecords(t *testing.T) {
	if n := binary.Size(rules1e.Item{}); n != rules1e.ItemRecSize {
		t.Fatalf("item is %#x bytes, want %#x", n, rules1e.ItemRecSize)
	}
	if n := binary.Size(rules1e.AffectRecord{}); n != rules1e.AffectRecSize {
		t.Fatalf("affect is %d bytes, want %d", n, rules1e.AffectRecSize)
	}
	b := make([]byte, 2*rules1e.ItemRecSize)
	for i := range b {
		b[i] = byte(i)
	}
	clear(b[0x2A:0x2E]) // next pointers are dropped
	clear(b[rules1e.ItemRecSize+0x2A : rules1e.ItemRecSize+0x2E])
	items := char.ReadRecords[rules1e.Item](b, rules1e.ItemRecSize)
	if len(items) != 2 || items[1].Price != 0x7A79 || !bytes.Equal(char.WriteRecords(items), b) {
		t.Fatal("items do not round-trip")
	}
}

// Amiga icon pixels follow the DOS icon colours: the DOS defaults leave
// them as drawn, a changed part moves both its Amiga slots.
func TestRecolourAmigaIcon(t *testing.T) {
	seen := map[byte]bool{}
	for _, v := range amigaIconColour {
		seen[v] = true
	}
	if len(seen) != 16 {
		t.Fatal("amigaIconColour is not a permutation")
	}
	im := &dax.Image{Data: []byte{48 + 9, 48 + 1, 48 + 6, 3}, Mask: 16}
	var to [16]byte
	for i := range to {
		to[i] = byte(i)
	}
	recolourIcon(im, to)
	if !bytes.Equal(im.Data, []byte{48 + 9, 48 + 1, 48 + 6, 3}) {
		t.Fatalf("defaults changed the icon: %v", im.Data)
	}
	to[1], to[9] = 4, 12 // body: red
	recolourIcon(im, to)
	if !bytes.Equal(im.Data, []byte{48 + 6, 48 + 7, 48 + 6, 3}) {
		t.Fatalf("got %v", im.Data)
	}
	// a chained change (1 -> 3 -> 5) chains on DOS pixels only
	im = &dax.Image{Data: []byte{1, 48 + 9}, Mask: 16}
	for i := range to {
		to[i] = byte(i)
	}
	to[1], to[3] = 3, 5
	recolourIcon(im, to)
	if !bytes.Equal(im.Data, []byte{5, 48 + amigaIconColour[3]}) {
		t.Fatalf("chain: got %v", im.Data)
	}
}
