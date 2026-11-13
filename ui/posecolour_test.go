package ui

import (
	"testing"

	"gobox/game"
)

// A posed timed event holds the fight with its unit in the pose.
func TestPosedEvent(t *testing.T) {
	var got game.Report
	posed := func(e game.Report) (*game.Combat, []game.Member) { got = e; return &game.Combat{}, nil }
	out := played([]game.Report{{Kind: game.EvAttack, Unit: 3, Ms: 100, Pose: 1}, {Kind: game.EvAttack, Unit: 4}}, true, posed, 0)
	if len(out) != 2 || out[0].Ms != 100 || boardOfFrame(out[0]) == nil || got.Unit != 3 || boardOfFrame(out[1]) != nil {
		t.Fatalf("posed events: %+v", out)
	}
}

func TestNameColour(t *testing.T) {
	for _, c := range []struct {
		m    game.Member
		want int
	}{{game.Member{Okay: true}, 11}, {game.Member{}, 12}, {game.Member{Okay: true, Foe: true}, 14},
		{game.Member{Okay: true, Exp: 50, NextExp: 40}, 13}} {
		if g := nameColour(c.m); g != c.want {
			t.Errorf("%+v: %d, want %d", c.m, g, c.want)
		}
	}
}
