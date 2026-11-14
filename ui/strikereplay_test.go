package ui

import (
	"slices"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// TestStrikeReplayed: a blow's lunge plays on its replayed moment (the
// moment's board has the next Strike Seq), so the live board after the
// replay, already at that Seq, does not lunge again with the next member up.
func TestStrikeReplayed(t *testing.T) {
	g := NewGame(&fake{st: game.Exploring{}}, game.Data{})
	g.ended = &game.Combat{Strike: game.Strike{Seq: 4}, Units: []game.Unit{{ID: 1, Active: true}, {ID: 2}}}
	b, _ := g.stepBoard(game.Report{Kind: game.EvAttack, Unit: 1, Target: 2, Hit: true, Pose: 1, Ms: 100})
	if want := (game.Strike{Seq: 5, From: 1, To: 2, Hit: true}); b.Strike != want {
		t.Errorf("attack moment's Strike %+v, want %+v", b.Strike, want)
	}
	if !b.Units[0].Active {
		t.Error("the attacker is no longer the active unit on its own blow")
	}
}

// TestStepMovesSquares: a replayed step takes the unit's Squares with it.
// The figure stands on its Squares: left behind, it stood still through the
// walk's moments and the blow's, and walked only on the live board after.
func TestStepMovesSquares(t *testing.T) {
	g := NewGame(&fake{st: game.Exploring{}}, game.Data{})
	before := []rules.BoardPos{{X: 5, Y: 5}, {X: 6, Y: 5}}
	g.ended = &game.Combat{Units: []game.Unit{{ID: 1, At: before[0], Squares: before}}}
	b, _ := g.stepBoard(game.Report{Kind: game.EvMove, Unit: 1, To: rules.BoardPos{X: 4, Y: 6}})
	if want := []rules.BoardPos{{X: 4, Y: 6}, {X: 5, Y: 6}}; !slices.Equal(b.Units[0].Squares, want) {
		t.Errorf("squares after the step %v, want %v", b.Units[0].Squares, want)
	}
	if before[0] != (rules.BoardPos{X: 5, Y: 5}) {
		t.Error("the step moved the squares of the board before it")
	}
}

// TestBlowShowsAtOnce: a blow's damage lowers the unit's hit points on its
// own moment, and a killing blow's fall (no body: a monster) takes the unit
// off the board on its own moment, held, right after the blow.
func TestBlowShowsAtOnce(t *testing.T) {
	g := NewGame(&fake{st: game.Exploring{}}, game.Data{})
	g.ended = &game.Combat{Units: []game.Unit{{ID: 1, HP: 10, MaxHP: 10}, {ID: 2, HP: 4, MaxHP: 8}}}
	ms := g.moments([]game.Report{
		{Kind: game.EvAttack, Unit: 1, Target: 2, Hit: true, Pose: 1, Ms: 100},
		{Kind: game.EvDamage, Unit: 2, Target: 2, N: 3},
		{Kind: game.EvDamage, Unit: 2, Target: 2, N: 3},
		{Kind: game.EvDown, Unit: 2, Target: 2},
	}, true, g.stepBoard, 500)
	hp := func(m Moment) int {
		if m.Board == nil {
			t.Fatalf("moment %v not held", m.Kind)
		}
		for _, u := range m.Board.Units {
			if u.ID == 2 && !u.Gone {
				return u.HP
			}
		}
		return -1 // off the board
	}
	if got := []int{hp(ms[0]), hp(ms[1]), hp(ms[2]), hp(ms[3])}; !slices.Equal(got, []int{4, 1, 0, -1}) {
		t.Errorf("unit 2's hit points over the moments %v, want [4 1 0 -1]", got)
	}
}
