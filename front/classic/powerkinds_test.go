package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

// game's UsePower kinds are the front's acts' numbers, as classic passes them.
func TestPowerKinds(t *testing.T) {
	for g, f := range map[int]front.ActKind{game.PowerCast: front.ActSpells, game.PowerDisplay: front.ActMagic, game.PowerHeal: front.ActHeal, game.PowerCure: front.ActCure} {
		if g != int(f) {
			t.Errorf("game kind %d, front act %d", g, f)
		}
	}
}
