package engine

import (
	"testing"

	"gobox/game"
)

// areaAimUp is castfight with Stinking Cloud's aim up (its Cast sent
// without a square, as the bar's): the aim's Cast and the Leave beside it.
func areaAimUp(t *testing.T) (*Engine, game.Aimed) {
	e := devFirstTurn(t, "castfight", false)
	var cloud game.Aimed
	for _, x := range commonOf(e.build()).Allowed.Aimed {
		if c, ok := x.Action.(game.Cast); ok && x.Area > 0 {
			cloud = x
			if _, err := e.doNow(game.Cast{Member: -1, Spell: c.Spell}); err != nil {
				t.Fatal("Cast:", err)
			}
		}
	}
	if cloud.Action == nil {
		t.Fatalf("no area spell offered: %+v", commonOf(e.build()).Allowed.Aimed)
	}
	al := commonOf(e.build()).Allowed
	if len(al.Aimed) != 1 {
		t.Fatalf("aim up: aimed %+v, want the cast's", al.Aimed)
	}
	return e, al.Aimed[0]
}

// TestAreaAimTaken: every square the area spell's aim offers is in its
// range: the last (the farthest) casts it.
func TestAreaAimTaken(t *testing.T) {
	e, x := areaAimUp(t)
	c := x.Action.(game.Cast)
	c.Ats = x.Squares[len(x.Squares)-1:]
	if _, err := e.doNow(c); err != nil {
		t.Fatal("aimed Cast:", err)
	}
	if e.cs.aim != nil {
		t.Fatalf("the aim is still up after %+v: why %q, ask %+v", c, e.snap.aimWhy, commonOf(e.build()).Ask)
	}
}

// TestAreaAimLeft: Leave gives the area spell's aim up (the bar's Cast: no menu to go back to).
func TestAreaAimLeft(t *testing.T) {
	e, _ := areaAimUp(t)
	if _, err := e.doNow(game.Leave{}); err != nil {
		t.Fatal("Leave:", err)
	}
	if e.cs.aim != nil {
		t.Fatal("the aim is still up after Leave")
	}
}
