package amigaimport

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestFonts: each game's Amiga text font decodes to FontFile's 256 glyphs
// with the digits, capitals and the signs between drawn and the space blank.
func TestFonts(t *testing.T) {
	for g, s := range map[string]Spec{
		"por":   {Disks: []string{"poolgame", "POOLDATA"}, Font: "fonts/poolrad/8"},
		"cotab": {Disks: []string{"CurseA", "CurseB"}, Font: "DISKA/CURSE.FON"},
		"sotsb": {Disks: []string{"Secret 1", "Secret 2"}, Font: "DISK1/SECRET.FON"},
		"pod":   {Disks: []string{"POD_Disk1_v.1.00", "POD_Disk2_v.1.00", "POD_Disk3_v.1.00"}, Font: "DISK1/pod.fon"},
	} {
		b, err := os.ReadFile(filepath.Join(importGame(t, s), FontFile))
		if err != nil {
			t.Fatalf("%s: %v", g, err)
		}
		if len(b) != fontGlyphs*glyphRows {
			t.Fatalf("%s: %d bytes", g, len(b))
		}
		glyph := func(c byte) []byte { return b[int(c)*glyphRows:][:glyphRows] }
		blank := make([]byte, glyphRows)
		if !bytes.Equal(glyph(' '), blank) {
			t.Errorf("%s: space drawn", g)
		}
		for c := byte('0'); c <= 'Z'; c++ { // por's font draws no $ % & (the space's bits)
			if bytes.Equal(glyph(c), blank) {
				t.Errorf("%s: %q blank", g, c)
			}
		}
		if bytes.Equal(glyph('A'), glyph('B')) {
			t.Errorf("%s: A and B alike", g)
		}
	}
}
