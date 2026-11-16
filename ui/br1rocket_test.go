package ui

import (
	"testing"

	"gobox/game"
)

// TestBR1Rocket: block 34's script flies the rocket (AreaKind 2): the presenter draws the system map and the orbit-map menu (Move)
// with its question seen, then the heading question after Move, then a move, with the ship's square and the map's tiles published.
func TestBR1Rocket(t *testing.T) {
	f := br1Folder(t)
	g, eng := br1Front(t, f)
	br1Roster(t, f, eng)
	eng.DevAction = func() { eng.DevAt(34, 0, 0) }
	br1Begin(t, g, eng)
	flying := func() *game.Question { // the orbit-map menu
		if sc, ok := eng.State().(game.Scene); ok && sc.Where.Mode == "space" && sc.Ask != nil && len(sc.Ask.Replies) == 4 && sc.Ask.Replies[0] == "Move" {
			return sc.Ask
		}
		return nil
	}
	for range 60 {
		if flying() != nil {
			break
		}
		if sc, ok := eng.State().(game.Scene); ok && sc.Where.Mode == "space" && sc.Ask != nil {
			flowKey(g, keyEsc) // the script's landing menu (Salvation Tycho Exit) of the square the ship starts on: Esc is Exit
			continue
		}
		g.Key(keyEnter)
		br1Frames(g, 6)
		if g.P.stuck() {
			t.Fatalf("on the way to the rocket: stuck on %T asking %s", eng.State(), g.P.asksID())
		}
	}
	q := flying()
	if q == nil {
		t.Fatalf("not flying: %T", eng.State())
	}
	br1Check(t, g, eng, "orbit-map menu", game.Scene{})
	c := game.CommonOf(eng.State())
	if len(q.Replies) != 4 || q.Replies[0] != "Move" || c.Automap.Size.W != 22 || len(c.Automap.Terrain) != 22*22 {
		t.Errorf("menu %+v map %v", q, c.Automap.Size)
	}
	if !g.P.areaMapUp() {
		t.Error("the system map is not the screen")
	}
	flowKey(g, 'r') // Rocket: the own ship's record, drawn with no undrawn question, its key leaves
	br1Check(t, g, eng, "ship record", game.Scene{})
	if sc, ok := eng.State().(game.Scene); !ok || sc.Kind != game.SceneScreen || sc.Screen == nil || sc.Screen.ShipSheet == nil {
		t.Fatalf("the ship record: %T", eng.State())
	}
	flowKey(g, keyEnter)
	br1Check(t, g, eng, "menu after the record", game.Scene{})
	flowKey(g, 'm')
	br1Check(t, g, eng, "heading", game.Scene{})
	if sc := eng.State().(game.Scene); sc.Kind != game.SceneHeading || sc.Ask == nil || !sc.Ask.Heading {
		t.Fatalf("heading: %+v", sc.Ask)
	}
	from := c.Where
	flowKey(g, keyUp) // north
	br1Check(t, g, eng, "after a move", game.Scene{})
	if to := game.CommonOf(eng.State()).Where; to.Mode != "space" || to.Y != from.Y-1 || to.X != from.X || to.Facing != 0 {
		t.Errorf("moved from %+v to %+v", from, to)
	}
	flowKey(g, keyEsc) // Done: the menu again
	br1Check(t, g, eng, "menu again", game.Scene{})
	if sc := eng.State().(game.Scene); sc.Ask == nil || sc.Ask.Heading || len(sc.Ask.Replies) != 4 {
		t.Errorf("after Done: %+v", sc.Ask)
	}
}
