package classic

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestStopRestListed: a game listing play.StopRest in camp hears the stop;
// one that does not, does not.
func TestStopRestListed(t *testing.T) {
	for _, listed := range []bool{true, false} {
		var al play.Allowed
		if listed {
			al.Plain = []play.Action{play.StopRest{}}
		}
		g := &fake{st: play.Camping{Common: play.Common{Allowed: al}}}
		r := newRules(g, game.Data{})
		r.send(play.StopRest{})
		if got := len(g.sent) == 1; got != listed {
			t.Fatalf("listed %v: sent %v", listed, g.sent)
		}
	}
}
