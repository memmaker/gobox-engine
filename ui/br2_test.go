package ui

import (
	"testing"

	"gobox/engine"
	"gobox/game"
)

// BR2 on the presenter, as BR1 (br1_test.go, whose helpers these use): the folder of `gobox import br2`, the
// shared engine, no window.

// TestBR2Explore: the title, the party menu, a member added, Begin and the opening's pages (block 19's briefing, the
// treasure screen) to the dock: each state is drawn by the presenter with its question seen.
func TestBR2Explore(t *testing.T) {
	f := brFolder(t, "br2")
	g, eng := br1Front(t, f)
	br1Roster(t, f, eng)
	br1Begin(t, g, eng)
	br1Check(t, g, eng, "opening", nil)
	for range 80 {
		if _, ok := eng.State().(game.Exploring); ok {
			return
		}
		br2Press(g, eng)
		if g.P.stuck() {
			t.Fatalf("opening: stuck on %T asking %s", eng.State(), g.P.asksID())
		}
		if _, menu := eng.State().(game.Assembling); !menu && g.P.unseen() {
			t.Fatalf("opening: %T asks %s, not drawn", eng.State(), g.P.asksID())
		}
	}
	t.Fatalf("no exploring after the opening: %T", eng.State())
}

// TestBR2Fight: a fight (the dev scene -fight 1) and its first turns, each frame drawn and each question seen.
func TestBR2Fight(t *testing.T) {
	f := brFolder(t, "br2")
	g, eng := br1Front(t, f)
	br1Roster(t, f, eng)
	eng.DevMonster = 46 // area17's ambush (data/br2/scripts/area17.txt:172)
	eng.DevAction = eng.DevScene("combat")
	br1Begin(t, g, eng)
	for range 80 { // the opening's pages and treasure first, then the dev scene's fight
		if _, ok := eng.State().(game.Fighting); ok {
			break
		}
		br2Press(g, eng)
	}
	br1Check(t, g, eng, "fight", game.Fighting{})
	br1FightPictures(t, eng)
	for range 10 {
		g.Key(keyEnter)
		br1Check(t, g, eng, "fight turn", nil)
	}
}

// br2Press answers the page or the treasure screen up: Enter continues a page; Esc on the treasure screen is Leave (the
// first Esc takes the announcement away).
func br2Press(g *Front, eng *engine.Engine) {
	if _, loot := eng.State().(game.Looting); loot {
		g.Key(keyEsc)
	} else {
		g.Key(keyEnter)
	}
	br1Frames(g, 6)
}
