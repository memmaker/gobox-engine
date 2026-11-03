package dax

import "testing"

func TestSiblings(t *testing.T) {
	got := siblings("MON3CHA")
	if len(got) != 7 || got[0] != "MON1CHA" || got[2] != "MON4CHA" {
		t.Errorf("MON3CHA: %v", got)
	}
	if got := siblings("8X8D3.DAX"); got[0] != "8X8D1.DAX" {
		t.Errorf("8X8D3.DAX: %v", got)
	}
	if siblings("TITLE") != nil {
		t.Error("TITLE has no disk")
	}
}

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
