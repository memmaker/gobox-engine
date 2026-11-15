package engine

import (
	"testing"

	"gobox/game"
)

// TestSotsbOpeningLoadsArea: sotsb's opening (area 17) ends in NEW ECL 16;
// the game goes on into area 16's start events (Priam) before any input,
// as the original's GameLoop runs NewEclDone after the start script. Before,
// the party waited in Exploring with no map loaded (GEO 0, no walls).
func TestSotsbOpeningLoadsArea(t *testing.T) {
	t.Setenv(advEnvSOTSB, "../data/sotsb")
	e := advStarted(t, advEnvSOTSB)
	for _, a := range []game.Action{game.QuickStart{}, game.Begin{}} {
		if _, err := e.doNow(a); err != nil {
			t.Fatal(err)
		}
	}
	for range 4 { // the opening's four "press Enter" pages
		ps := e.State()
		if _, ok := ps.(game.Exploring); ok {
			t.Fatalf("exploring during the opening, GEO %d", int(e.world.Geo))
		}
		if _, err := e.doNow(game.Reply{N: 0}); err != nil {
			t.Fatal(err)
		}
		e.Advance()
	}
	if _, ok := e.State().(game.Exploring); ok {
		t.Fatalf("exploring after the opening, GEO %d: area 16's start events did not run", int(e.world.Geo))
	}
	if g := int(e.world.Geo); g != 16 {
		t.Fatalf("GEO %d after the opening, want 16", g)
	}
}
