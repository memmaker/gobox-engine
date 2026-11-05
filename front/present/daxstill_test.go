package present

import "testing"

// An animation block (PIC*: u8 count, a 21-byte frame header, packed pixels) decodes as its frame, not as nothing.
func TestDaxStillSprite(t *testing.T) {
	b := []byte{1, 0, 0, 0, 0, 2, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	b = append(b, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0)
	im := daxStill("PIC1", b, 0)
	if im == nil || im.Width != 1 || im.Height != 2 || im.Frame(0)[0] != 1 {
		t.Fatalf("%+v", im)
	}
}
