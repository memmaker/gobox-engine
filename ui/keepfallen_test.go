package ui

import (
	"testing"

	"gobox/game"
)

// A tick where 3 foes are hit and fall: each figure is drawn until its own
// fall plays, then it is gone; the fall showing (the queue's head) shows it
// down (a last blow's fall left the figure standing on its body).
func TestKeepFallenUntilDied(t *testing.T) {
	var p Presenter
	before := &game.Combat{Units: []game.Unit{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}}}
	p.keepFallen(before) // drawn before the tick
	after := &game.Combat{Units: []game.Unit{{ID: 1}}}
	var q []frame
	for id := 2; id <= 4; id++ {
		q = append(q, frame{Moment: Moment{Report: game.Report{Kind: game.EvMiss, Unit: 1, Target: id}}}, frame{Moment: Moment{Report: game.Report{Kind: game.EvDown, Unit: id}}})
	}
	p.replay.q = q
	for len(p.replay.q) > 0 {
		left := 0
		for _, e := range p.replay.q[1:] {
			if e.Kind == game.EvDown {
				left++
			}
		}
		if n := len(p.keepFallen(after).Units); n != 1+left {
			t.Fatalf("%d figures with %d falls to play, want %d", n, left, 1+left)
		}
		p.replay.q = p.replay.q[1:]
	}
	if n := len(p.keepFallen(after).Units); n != 1 {
		t.Fatalf("%d figures after the replay", n)
	}
}
