package ui

import (
	"testing"

	"gobox/game"
	"gobox/rules"
)

// While an area spell is aimed, the battle previews its radius (AimArea is radius+1); before and after, none.
func TestAimAreaPreview(t *testing.T) {
	st := game.Fighting{Common: game.Common{Party: []game.Member{{Spells: []game.SpellRef{{ID: 7, Name: "Fireball", Range: 9}}}},
		Allowed: game.Allowed{Plain: []game.Action{game.EndTurn{}}, Aimed: []game.Aimed{{Action: game.Cast{Spell: 7}, Squares: []rules.BoardPos{{X: 3, Y: 3}}, Area: 2, Targets: 1}}}},
		Field: game.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, Member: 0}}}}
	g := NewGame(&fake{st: st}, game.Data{})
	if b := liveBoard(g); b == nil || b.AimArea != 0 {
		t.Fatalf("before aiming: %+v", b)
	}
	g.Answer(Action{Kind: ActCast, N: 7})
	if b := liveBoard(g); b == nil || b.AimArea != 3 {
		t.Fatalf("aiming: AimArea %v, want 3", b.AimArea)
	}
	g.Answer(Action{Kind: ActClose})
	if b := liveBoard(g); b.AimArea != 0 {
		t.Fatalf("after Esc: AimArea %d", b.AimArea)
	}
}
