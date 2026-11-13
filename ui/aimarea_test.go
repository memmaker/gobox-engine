package ui

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// While an area spell is aimed, the battle previews its radius (AimArea is radius+1); before and after, none.
func TestAimAreaPreview(t *testing.T) {
	st := play.Fighting{Common: play.Common{Party: []game.Member{{Spells: []game.SpellRef{{ID: 7, Name: "Fireball", Range: 9}}}},
		Allowed: play.Allowed{Aimed: []play.Aimed{{Action: play.Cast{Spell: 7}, Area: 2, Targets: 1}}}},
		Field: play.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, Member: 0}}}}
	g := NewGame(&fake{st: st}, game.Data{})
	if b := liveBoard(g); b == nil || b.AimArea != 0 {
		t.Fatalf("before aiming: %+v", b)
	}
	g.Answer(FrontAction{Kind: ActCast, N: 7})
	if b := liveBoard(g); b == nil || b.AimArea != 3 {
		t.Fatalf("aiming: AimArea %v, want 3", b.AimArea)
	}
	g.Answer(FrontAction{Kind: ActClose})
	if b := liveBoard(g); b.AimArea != 0 {
		t.Fatalf("after Esc: AimArea %d", b.AimArea)
	}
}
