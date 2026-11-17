package engine

import (
	"slices"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// TestAimRefusedSaysWhy: a square an aimed spell does not take gets the
// rules' word why from Preview (Stinking Cloud, range 3, at a foe nine
// squares off: "Out of range"); a square it takes gets none.
func TestAimRefusedSaysWhy(t *testing.T) {
	e := devFirstTurn(t, "castfight", false)
	for _, x := range commonOf(e.build()).Allowed.Aimed {
		c, ok := x.Action.(game.Cast)
		if !ok || e.spellID(c.Spell) != "stinking-cloud" || len(x.Squares) == 0 {
			continue
		}
		c.Ats = x.Squares[:1]
		if w := e.Preview(c).Why; w != "" {
			t.Fatalf("an offered square: %q", w)
		}
		for _, u := range e.fightState().Units {
			if slices.Contains(x.Squares, u.At) {
				continue
			}
			c.Ats = []rules.BoardPos{u.At}
			if e.Preview(c).Why == "" {
				t.Fatalf("%v not offered, no word why", u.At)
			}
			return
		}
		t.Skip("every unit in reach")
	}
	t.Fatal("no Stinking Cloud aim")
}
