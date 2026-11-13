package present

import (
	"testing"

	"gobox/front"
)

// Tab (pad Y) toggles the minimap in the presenter's Host options and sends
// the game nothing (it is no game action).
func TestTabTogglesMinimap(t *testing.T) {
	oldHost := Host
	h := DefaultOptions
	Host = &h
	defer func() { Host = oldHost }()
	var p Presenter
	for _, hidden := range []bool{true, false} {
		p.keyAct(keyTab, front.View{})
		if Host.MapOff != hidden {
			t.Fatalf("MapOff %v after Tab, want %v", Host.MapOff, hidden)
		}
	}
	if len(p.inv.acts) != 0 {
		t.Fatalf("Tab sent %v to the game, want none", p.inv.acts)
	}
}
