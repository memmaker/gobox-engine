// Command fontexport writes a game's DAX text font (block 201, 64 glyphs of
// 8 rows, 1 bit per pixel, bit 7 leftmost, upper case only, then 113 more
// 8-row symbol tiles) as a pool .f8:
//
//	fontexport 8X8D1.DAX out.f8
//
// The game draws ch as glyph upper(ch)%64; the .f8 holds that glyph at every
// ASCII 0x20..0x7F (a-z show the capitals); symbol tile k (from 64) is at
// slot 0x80+k-64; the rest is blank.
package main

import (
	"fmt"
	"os"

	"gobox/dax"
)

const block = 201

// toF8 is the 256-glyph form of the 64-glyph font.
func toF8(b []byte) []byte {
	out := make([]byte, 256*8)
	for ch := 0x20; ch < 0x80; ch++ {
		c := ch
		if c >= 'a' && c <= 'z' {
			c -= 0x20
		}
		i := c % 0x40 * 8
		if i+8 <= len(b) {
			copy(out[ch*8:], b[i:i+8])
		}
	}
	for k := 64; k*8+8 <= len(b) && 0x80+k-64 < 256; k++ {
		copy(out[(0x80+k-64)*8:], b[k*8:k*8+8])
	}
	return out
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: fontexport FILE.DAX out.f8")
		os.Exit(2)
	}
	f, err := dax.Open(os.Args[1])
	if err == nil {
		b := f.Block(block)
		if len(b) < 512 || len(b)%8 != 0 {
			err = fmt.Errorf("block %d is %d bytes", block, len(b))
		} else {
			err = os.WriteFile(os.Args[2], toF8(b), 0o644)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
