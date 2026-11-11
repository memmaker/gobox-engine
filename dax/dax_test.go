package dax

import "testing"

func TestBlockCopy(t *testing.T) {
	f := &File{cache: map[int][]byte{1: {1, 2}}}
	b := f.Block(1)
	b[0] = 9
	if f.BlockShared(1)[0] != 1 {
		t.Fatal("Block aliased the cache")
	}
}

func TestBlockEmptyNotNil(t *testing.T) {
	f := &File{cache: map[int][]byte{1: {}}}
	if f.Block(1) == nil {
		t.Fatal("an empty block reads as missing")
	}
	if f.Block(2) != nil {
		t.Fatal("a missing block reads as present")
	}
}

// TestParseNotPicture: a block whose header claims more pixels than it
// holds (8X8D1 201, the font, reads as 0x0101 x 0x0101 x 255) is no
// picture and allocates nothing past the block.
func TestParseNotPicture(t *testing.T) {
	b := make([]byte, 40)
	for i := range b {
		b[i] = 0xFF
	}
	if ParseImage(b, false, 0) != nil {
		t.Error("ParseImage")
	}
	if fs := ParseSprite(b, false, 0, false); len(fs) != 0 {
		t.Error("ParseSprite", len(fs))
	}
}
