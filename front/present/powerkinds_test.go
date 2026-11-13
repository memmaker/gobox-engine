package present

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// A power picked on the front reaches the game as its PowerKind.
func TestPowerKinds(t *testing.T) {
	for k, want := range map[front.ActKind]game.PowerKind{front.ActSpells: game.PowerCast, front.ActMagic: game.PowerDisplay,
		front.ActHeal: game.PowerHeal, front.ActCure: game.PowerCure} {
		f := &fake{st: play.Camping{Common: allow(play.UsePower{Member: 0, Kind: want})}}
		g := NewGame(f, game.Data{})
		if err := g.UsePower(0, k, 0, 0, true); err != nil {
			t.Fatal(err)
		}
		if len(f.sent) != 1 || f.sent[0].(play.UsePower).Kind != want {
			t.Errorf("%v sent %+v, want kind %v", k, f.sent, want)
		}
	}
}
