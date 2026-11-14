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
