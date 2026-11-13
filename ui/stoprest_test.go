package ui

import (
	"testing"

	"gobox/game"
)

// TestStopRestListed: a game listing game.StopRest in camp hears the stop;
// one that does not, does not.
func TestStopRestListed(t *testing.T) {
	for _, listed := range []bool{true, false} {
		var al game.Allowed
		if listed {
			al.Plain = []game.Action{game.StopRest{}}
		}
		g := &fake{st: game.Camping{Common: game.Common{Allowed: al}}}
		r := newRules(g, game.Data{})
		r.send(game.StopRest{})
		if got := len(g.sent) == 1; got != listed {
			t.Fatalf("listed %v: sent %v", listed, g.sent)
		}
	}
}
