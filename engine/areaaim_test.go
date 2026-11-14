package engine

import (
	"slices"
	"testing"

	"gobox/game"
	"gobox/rules"
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
	e, cloud := areaAimUp(t)
	caster := e.cs.actor
	if _, err := e.doNow(game.Leave{}); err != nil {
		t.Fatal("Leave:", err)
	}
	if e.cs.aim != nil || e.cs.sa != nil {
		t.Fatal("the aim is still up after Leave")
	}
	if e.cs.turn == nil || e.cs.actor != caster {
		t.Fatal("Leave ended the turn: want its menu again")
	}
	if !slices.ContainsFunc(commonOf(e.build()).Allowed.Aimed, func(x game.Aimed) bool {
		c, ok := x.Action.(game.Cast)
		return ok && c.Spell == cloud.Action.(game.Cast).Spell
	}) {
		t.Fatal("Leave spent the spell: want it offered again")
	}
}

// TestAreaAimGround: an area spell is aimed at an empty square in reach.
func TestAreaAimGround(t *testing.T) {
	e, x := areaAimUp(t)
	b := e.build().(game.Fighting).Field
	for _, at := range x.Squares {
		if slices.ContainsFunc(b.Units, func(u game.Unit) bool { return u.At == at || slices.Contains(u.Squares, at) }) {
			continue
		}
		c := x.Action.(game.Cast)
		c.Ats = []rules.BoardPos{at}
		if _, err := e.doNow(c); err != nil || e.cs.aim != nil {
			t.Fatalf("Cast at the empty %v: %v, aim up %v (why %q)", at, err, e.cs.aim != nil, e.snap.aimWhy)
		}
		return
	}
	t.Fatal("no empty square offered")
}

// TestAreaAimOffered: every square the turn's Cast offers for the area
// spell is taken (the corner ones were "Out of sight": the aim stayed up,
// the cast silently not sent).
func TestAreaAimOffered(t *testing.T) {
	var squares []rules.BoardPos
	for _, x := range commonOf(devFirstTurn(t, "castfight", false).build()).Allowed.Aimed {
		if _, ok := x.Action.(game.Cast); ok && x.Area > 0 {
			squares = x.Squares
		}
	}
	if len(squares) == 0 {
		t.Fatal("no area spell offered")
	}
	for _, at := range squares {
		e := devFirstTurn(t, "castfight", false)
		for _, x := range commonOf(e.build()).Allowed.Aimed {
			if c, ok := x.Action.(game.Cast); ok && x.Area > 0 {
				c.Ats = []rules.BoardPos{at}
				if _, err := e.doNow(c); err != nil || e.cs.aim != nil {
					t.Fatalf("Cast at %v: %v, aim up %v (why %q)", at, err, e.cs.aim != nil, e.snap.aimWhy)
				}
			}
		}
	}
}
