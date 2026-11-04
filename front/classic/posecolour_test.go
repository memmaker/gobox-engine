package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

// A posed timed event holds the fight with its unit in the pose.
func TestPosedEvent(t *testing.T) {
	var got game.Event
	posed := func(e game.Event) *front.Battle { got = e; return &front.Battle{} }
	out := combatEvents([]game.Event{{Kind: game.EvAttack, Unit: 3, Ms: 100, Pose: 1}, {Kind: game.EvAttack, Unit: 4}}, front.View{}, posed)
	if len(out) != 2 || out[0].Ms != 100 || out[0].Battle == nil || got.Unit != 3 || out[1].Battle != nil {
		t.Fatalf("posed events: %+v", out)
	}
}

func TestNameColour(t *testing.T) {
	for _, c := range []struct {
		m    game.Member
		want int
	}{{game.Member{Okay: true}, 11}, {game.Member{}, 12}, {game.Member{Okay: true, Foe: true}, 14},
		{game.Member{Okay: true, Exp: 50, NextExp: 40}, 13}, {game.Member{Colour: 5}, 5}} {
		if g := nameColour(c.m); g != c.want {
			t.Errorf("%+v: %d, want %d", c.m, g, c.want)
		}
	}
}
