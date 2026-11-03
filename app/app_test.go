package app

import (
	"slices"
	"testing"
)

func TestParseKeys(t *testing.T) {
	got := ParseKeys(`a\r\M\C\P\w\x\`, map[byte]Key{'r': 13})
	want := []Key{'a', 13, Right, Left, Next, Wait, '\\', 'x', '\\'}
	if !slices.Equal(got, want) {
		t.Fatalf("%x, want %x", got, want)
	}
}
