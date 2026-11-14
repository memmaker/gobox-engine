package ui

import (
	"testing"

	"gobox/game"
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
