package ui

import (
	"testing"

	"gobox/game"
	"gobox/rules"
)

// TestMoveMarks: the walk (the reach, the hovered path) is shown only while
// a party member has its turn and may walk: not on a foe's turn, not for a
// member with nowhere to go, and not while what happened before is still
// played (the turn that ended).
func TestMoveMarks(t *testing.T) {
	units := []game.Unit{{ID: 1, Member: 0, At: rules.BoardPos{X: 1, Y: 1}}, {ID: 2, Member: -1, At: rules.BoardPos{X: 5, Y: 3}, Foe: true}}
	moves := []game.Reach{{At: rules.BoardPos{X: 2, Y: 1}}}
	fight := func(turn int, moves []game.Reach) *Presenter {
		g := NewGame(&fake{st: game.Fighting{Common: game.Common{Party: []game.Member{{Name: "A"}}, Allowed: game.Allowed{Plain: []game.Action{game.EndTurn{}}, Moves: moves}},
			Field: game.Battlefield{Size: rules.Size{W: 8, H: 8}, Turn: turn, Units: units}}}, game.Data{})
		p := presenterOf(g)
		p.fightView = g.combat()
		return p
	}
	shown := func(p *Presenter) bool { return p.moveMarks(p.boardOf(p.fightView, p.party(), nil)).CanMove }
	if p := fight(1, moves); !shown(p) {
		t.Error("a member's turn with squares to walk to: no walk shown")
	} else if p.replay.q = []frame{{}}; shown(p) {
		t.Error("the walk is shown while the turn before is still played")
	}
	if shown(fight(2, moves)) {
		t.Error("the walk is shown on a foe's turn")
	}
	if shown(fight(1, nil)) {
		t.Error("the walk is shown for a member with nowhere to go")
	}
}
