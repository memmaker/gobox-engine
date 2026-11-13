package main

import (
	"testing"

	"gobox/engine"
	"gobox/ui"
)

// TestArtSwitch: the pool games' one art switch (F12, the Graphics page)
// turns the engine's pictures and the front's portraits and cards alike
// (the front's own art stayed DOS: its portraits never followed F12).
func TestArtSwitch(t *testing.T) {
	s := start{Adventure: "../../data/cotab", Data: "../../data/cotab/sources"}
	s.Capture.Dev.Saves = t.TempDir()
	l, err := playPool(s)
	if err != nil {
		t.Skip(err)
	}
	defer engine.DOSArt.Store(engine.DOSArt.Load())
	for _, dos := range []bool{false, true, false} {
		l.Art.Set(dos)
		if engine.DOSArt.Load() != dos || ui.Amiga() == dos {
			t.Errorf("Set(dos %v): engine DOS %v, front Amiga %v", dos, engine.DOSArt.Load(), ui.Amiga())
		}
	}
}
