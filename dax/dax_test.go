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
