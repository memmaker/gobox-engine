package app

import (
	"flag"
	"testing"
)

func TestFlags(t *testing.T) {
	read := Flags(map[byte]Key{'r': 13})
	for k, v := range map[string]string{"shot": "a.png", "keys": `x\r`, "pointer": "1,2;3,4", "size": "640x400", "slot": "CAPTURE", "freeze": "true"} {
		flag.Set(k, v)
	}
	c := read()
	if c.Shot != "a.png" || len(c.Keys) != 2 || c.Keys[1] != 13 || len(c.Pointer) != 2 || c.Pointer[1] != [2]int{3, 4} ||
		c.Window != [2]int{640, 400} || c.Slot != "CAPTURE" || !c.Freeze || !c.GPU.Offscreen {
		t.Fatalf("%+v", c)
	}
}
