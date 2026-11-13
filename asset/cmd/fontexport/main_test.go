package main

import (
	"testing"

	"gobox/asset"
	"gobox/asset/orig"
)

// glyph is the games' decoder (cotab cmd/cotab/combat.go glyph, sotsb
// engine/fonts.go glyph with upcase): upper(ch) mod 64 into the block.
func glyph(font []byte, ch byte) (g [8]byte) {
	if ch >= 'a' && ch <= 'z' {
		ch -= 0x20
	}
	if i := int(ch%0x40) * 8; i+8 <= len(font) {
		copy(g[:], font[i:])
	}
	return g
}

func TestPool(t *testing.T) {
	for _, c := range []struct{ dir, dax, name string }{
		{"/Users/felix/Projects/cotab-re/gog", "8X8D1", "Gold_Box_8x8D1"},
		{"/Users/felix/Projects/sotsb-re/gog", "8X8D5", "Gold_Box_8x8D1"},
	} {
		l, err := orig.NewLoader(c.dir)
		if err != nil {
			t.Skip(err)
		}
		blk := l.Block(c.dax, block) // what the game gets
		pool := asset.Get(c.name)
		if len(blk) != 1416 || len(pool) != 2048 {
			t.Fatal(c.name, len(blk), len(pool))
		}
		if string(pool) != string(toF8(blk)) {
			t.Fatal(c.name, "pool file differs from the exporter")
		}
		for ch := 0x20; ch < 0x80; ch++ {
			if glyph(blk, byte(ch)) != [8]byte(pool[ch*8:]) {
				t.Fatalf("%s: glyph %#x", c.name, ch)
			}
		}
		for k := 64; k < len(blk)/8; k++ { // the symbol tiles
			if string(blk[k*8:k*8+8]) != string(pool[(0x80+k-64)*8:][:8]) {
				t.Fatalf("%s: tile %d", c.name, k)
			}
		}
	}
}
