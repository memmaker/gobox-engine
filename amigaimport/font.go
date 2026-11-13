package amigaimport

import (
	"errors"
	"fmt"
)

// FontFile is the Amiga text font as Import writes it: 256 glyphs of 8
// rows, bit 7 leftmost (asset.F8's form), each character at its ASCII slot.
const FontFile = "FONT.F8"

const (
	glyphRows  = 8
	fontGlyphs = 256
	firstChar  = 0x20 // a .FON's first glyph: 96 of them, ' '..0x7F
	fonGlyphs  = 0x60
	hunkHeader = 0x3f3 // an AmigaOS load file's first word
)

// font is an Amiga text font in FontFile's form. The later games' .FON
// (Curse, Silver Blades, Darkness) is 96 glyphs of 8 rows from ' '; Pool
// of Radiance's fonts/poolrad/8 an AmigaOS disk font (a load file holding
// a DiskFontHeader and its TextFont, 8 rows high, at most 8 wide).
func font(b []byte) ([]byte, error) {
	out := make([]byte, fontGlyphs*glyphRows)
	if len(b) == fonGlyphs*glyphRows {
		copy(out[firstChar*glyphRows:], b)
		return out, nil
	}
	if len(b) < 0x20 || be32(b, 0) != hunkHeader {
		return nil, fmt.Errorf("font: %d bytes, neither a .FON nor a disk font", len(b))
	}
	// one code hunk: header words 0x3f3, 0 (no names), 1 hunk, first 0,
	// last 0, its size, then 0x3e9 and the size again; the data at 0x20,
	// where pointers are offsets
	h := b[0x20:]
	const (
		textFont = 0x3a // dfh_TF: after the 4-byte moveq/rts, Node, FileID, Revision, Segment, Name[32]
		ySize    = textFont + 20
		loChar   = textFont + 32
		charData = textFont + 34
	)
	if len(h) < charData+14 || be16(h, 0x12) != 0x0f80 || be16(h, ySize) != glyphRows {
		return nil, errors.New("font: not an 8-row disk font")
	}
	lo, hi := int(h[loChar]), int(h[loChar+1])
	data, modulo, loc := be32(h, charData), be16(h, charData+4), be32(h, charData+6)
	for c := lo; c <= hi; c++ {
		at, w := be16(h, loc+4*(c-lo)), be16(h, loc+4*(c-lo)+2)
		for r := range glyphRows {
			var g byte
			for x := range min(w, 8) {
				bit := at + x
				if h[data+r*modulo+bit/8]&(0x80>>(bit%8)) != 0 {
					g |= 0x80 >> x
				}
			}
			out[c*glyphRows+r] = g
		}
	}
	return out, nil
}
