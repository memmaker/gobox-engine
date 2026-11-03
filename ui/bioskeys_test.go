package ui

import (
	"testing"

	"gobox/gpu"
	"gobox/input"
)

func TestBIOSKey(t *testing.T) {
	for k, want := range map[gpu.Key]uint16{gpu.KeyF1: 0x3B00, gpu.KeyF8: 0x4200, gpu.KeyF9: 0x4300, gpu.KeyF10: 0x4400,
		gpu.KeyKP4: 0x4B00, gpu.KeyPageDown: 0x5100, gpu.KeyKP5: 0x4C00, gpu.KeyDelete: 0x5300, gpu.KeyKPEnter: 13, gpu.KeyTab: 9} {
		if got, ok := BIOSKey(k); !ok || got != want {
			t.Errorf("%v: %#x %v", k, got, ok)
		}
	}
	if _, ok := BIOSKey(gpu.KeyA); ok {
		t.Error("A is typed text, not a BIOS key")
	}
}

func TestMenuNavH(t *testing.T) {
	sel := 0
	keys := []byte("ABC")
	MenuNavH(input.Right, 0x4D00, &sel, keys)
	if sel != 1 {
		t.Fatal(sel)
	}
	MenuNavH(input.Left, 0x4B00, &sel, keys)
	if sel != 0 {
		t.Fatal(sel)
	}
	if _, _, ok := MenuNav(input.Right, 0x4D00, &sel, keys); ok {
		t.Fatal("MenuNav must ignore Right")
	}
}
