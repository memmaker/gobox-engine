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

	"gobox/asset"
	"gobox/asset/orig"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: fontexport FILE.DAX out.f8")
		os.Exit(2)
	}
	f, err := orig.Open(os.Args[1])
	if err == nil {
		b := f.Block(asset.FontBlock)
		if len(b) < 512 || len(b)%8 != 0 {
			err = fmt.Errorf("block %d is %d bytes", asset.FontBlock, len(b))
		} else {
			err = os.WriteFile(os.Args[2], asset.F8(b), 0o644)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
