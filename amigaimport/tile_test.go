package amigaimport

import "testing"

// TestBlankSparse: a sparse frame with no colour bytes (B 0; Curse's PIC
// 57 and 74, SotSB's 6 and 17) decodes as all clear, not as malformed.
func TestBlankSparse(t *testing.T) {
	const h, w = 88, 11
	d := append([]byte{0, h, 0, 0, 0, 0, w, 3, 0, h, 0, 0}, make([]byte, h)...) // h "next row" codes
	tl := decodeTile(d)
	if tl == nil || len(tl.px) != h*w*8 {
		t.Fatalf("decoded %v", tl)
	}
	for _, c := range tl.px {
		if c != clear {
			t.Fatal("not clear")
		}
	}
}
