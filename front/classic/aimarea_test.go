package classic

import (
	"testing"

	"gobox/front"
	"gobox/game"
)

// aimRules is a fight whose one aimed spell has a burst of radius 2.
type aimRules struct{ seqRules }

func (r *aimRules) State() game.State {
	return game.State{Mode: game.ModeCombat, Combat: &game.Combat{Aims: []game.SpellRef{{ID: 7, Name: "Fireball", Range: 9, Area: 2}}}}
}

// While an area spell is aimed, the battle previews its radius (AimArea is radius+1); before and after, none.
func TestAimAreaPreview(t *testing.T) {
	g := NewGame(&aimRules{}, game.Data{})
	if b := g.View().Battle; b == nil || b.AimArea != 0 {
		t.Fatalf("before aiming: %+v", b)
	}
	g.Answer(front.Action{Kind: front.ActCast, N: 7})
	if b := g.View().Battle; b == nil || b.AimArea != 3 {
		t.Fatalf("aiming: AimArea %v, want 3", b.AimArea)
	}
	g.Answer(front.Action{Kind: front.ActClose})
	if b := g.View().Battle; b.AimArea != 0 {
		t.Fatalf("after Esc: AimArea %d", b.AimArea)
	}
}
