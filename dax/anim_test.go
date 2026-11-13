package dax

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Every view picture block (PIC*, FINAL*, SPRIT*) of every Gold Box port,
// DOS and Amiga, decodes whole: a still picture, an EGA animation
// (ParseSprite's layout, the Amiga import's too) or a VGA one (Buck
// Rogers, PoD), each frame its full size and the block used up.
func TestAnimBlocksDecode(t *testing.T) {
	seen := 0
	for _, data := range []string{"../data/por/sources", "../data/cotab/sources", "../data/sotsb/sources", "../data/pod/sources", "../../buckrogers-go/data", "../../buckrogers2-go/data"} {
		for _, side := range []string{"dos", "amiga"} {
			dir := filepath.Join(data, side)
			ents, _ := os.ReadDir(dir)
			for _, e := range ents {
				n := strings.ToUpper(e.Name())
				if !strings.HasSuffix(n, ".DAX") || strings.Contains(n, "PAL") || !(strings.HasPrefix(n, "PIC") || strings.HasPrefix(n, "FINAL") || strings.HasPrefix(n, "SPRIT")) {
					continue
				}
				f, err := Open(filepath.Join(dir, e.Name()))
				if err != nil {
					t.Fatal(err)
				}
				for _, en := range f.Entries {
					b := f.BlockShared(en.ID)
					if len(b) == 0 {
						continue
					}
					seen++
					if !animWhole(b) {
						t.Errorf("%s %s %s block %d: does not decode", data, side, n, en.ID)
					}
				}
			}
		}
	}
	if seen == 0 {
		t.Skip("no game data")
	}
}

// animWhole: b is a still picture, or an animation every frame of which
// fills its size and that uses the block up.
func animWhole(b []byte) bool {
	if fs, _ := ParseVGAAnim(b); fs != nil {
		return true // ParseVGAAnim checks sizes and the block's end itself
	}
	if len(b) >= 17 {
		h, w, c := int(binary.LittleEndian.Uint16(b)), int(binary.LittleEndian.Uint16(b[2:])), int(b[8])
		sz := h * w * 4
		if wide(b[9:]) {
			sz *= 2
		}
		if h > 0 && w > 0 && 17+c*sz == len(b) {
			return true
		}
	}
	p := 1
	for range int(b[0]) {
		if p+21 > len(b) {
			return false
		}
		sz := int(binary.LittleEndian.Uint16(b[p+4:])) * int(binary.LittleEndian.Uint16(b[p+6:])) * 4
		if wide(b[p+12:]) {
			sz *= 2
		}
		p += 21 + sz
	}
	return b[0] > 0 && p == len(b) && len(ParseSprite(b, false, 0, true)) == int(b[0])
}

// PoD's PIC1 block 44 (the mud man): 8 frames, their delays, and play
// going back to frame 4 after the last (O108:0834's loop).
func TestVGAAnimPod(t *testing.T) {
	f, err := Open("../data/pod/sources/dos/PIC1.DAX")
	if err != nil {
		t.Skip("no PoD data")
	}
	fs, loop := ParseVGAAnim(f.BlockShared(44))
	var d []int
	for _, x := range fs {
		d = append(d, x.Delay)
		if x.Image.Width != 11 || x.Image.Height != 88 || len(x.Image.Palette) != 256 {
			t.Fatalf("%+v", x.Image)
		}
	}
	if loop != 4 || !slices.Equal(d, []int{2, 3, 4, 4, 2, 3, 3, 3}) {
		t.Fatalf("loop %d delays %v", loop, d)
	}
	if slices.Equal(fs[0].Image.Data, fs[5].Image.Data) {
		t.Fatal("frames not reconstructed from the key frame")
	}
}

// unpackRuns: literals (c <= 0x7F: c+1 bytes) and runs (257-c copies).
func TestUnpackRuns(t *testing.T) {
	dst := make([]byte, 6)
	if n := unpackRuns(dst, []byte{1, 7, 8, 0xFE, 9}); n != 5 || !slices.Equal(dst, []byte{7, 8, 9, 9, 9, 0}) {
		t.Fatal(n, dst)
	}
}
