package engine

import (
	"testing"
)

// TestPodSymbolSetFrames: pod's scripts writing the symbol set (tileSet)
// pick the frame set only; its files stay set 1 (a write switched the
// engine to GEO2, ECL2, which pod has not).
func TestPodSymbolSetFrames(t *testing.T) {
	e, err := NewPool("../data/pod", "../data/pod/sources")
	if err != nil {
		t.Skip(err)
	}
	e.set = e.game().GameSet
	f, ok := e.newEngineFields()["SymbolSet"]
	if !ok {
		t.Fatal("no SymbolSet field")
	}
	const frames = 2
	f.set(0, frames)
	if v, _ := f.get(0); v != frames || e.set != e.game().GameSet || int(e.world.SymbolSet) != frames {
		t.Errorf("SymbolSet %d: read %d, set %d, word %d", frames, v, e.set, int(e.world.SymbolSet))
	}
}
