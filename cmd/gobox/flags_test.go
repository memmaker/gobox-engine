package main

import (
	"flag"
	"gobox/ui"
	"slices"
	"testing"
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
