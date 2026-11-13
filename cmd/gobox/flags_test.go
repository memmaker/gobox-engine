package main

import (
	"flag"
	"gobox/ui"
	"slices"
	"testing"
	"time"
)

func TestFlags(t *testing.T) {
	read := captureFlags(map[byte]ui.Key{'r': 13})
	for k, v := range map[string]string{"shot": "a.png", "keys": `x\r`, "pointer": "1,2;3,4", "size": "640x400", "slot": "CAPTURE", "freeze": "true"} {
		flag.Set(k, v)
	}
	c := read()
	if c.Shot != "a.png" || len(c.Keys) != 2 || c.Keys[1] != 13 || len(c.Pointer) != 2 || c.Pointer[1] != [2]int{3, 4} ||
		c.Window != [2]int{640, 400} || c.Slot != "CAPTURE" || !c.Freeze || !c.GPU.Offscreen {
		t.Fatalf("%+v", c)
	}
}

func TestParseKeys(t *testing.T) {
	got := parseKeys(`a\r\M\C\P\w\x\`, map[byte]ui.Key{'r': 13})
	want := []ui.Key{'a', 13, ui.ScriptRight, ui.ScriptLeft, ui.ScriptNext, ui.ScriptWait, '\\', 'x', '\\'}
	if !slices.Equal(got, want) {
		t.Fatalf("%x, want %x", got, want)
	}
}

// TestResolveSeed: -seed >= 0 wins, a scripted run without it plays 1, else the clock.
func TestResolveSeed(t *testing.T) {
	now := time.Unix(0, 42)
	if resolveSeed(7, true, now) != 7 || resolveSeed(0, false, now) != 0 || resolveSeed(-1, true, now) != 1 || resolveSeed(-1, false, now) != 42 {
		t.Fatal("resolveSeed")
	}
}
