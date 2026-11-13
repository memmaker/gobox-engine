package present

import (
	"image"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// A shot with a sprite plays as its flight: the sprite drawn square by
// square (in thirds) from the shooter to the target, over the moment.
func TestMissileFlies(t *testing.T) {
	b := &play.Combat{Units: []game.Unit{{ID: 1, At: image.Pt(27, 14)}}}
	es := []play.Report{{Kind: game.EvMissile, Unit: 1, To: image.Pt(27, 12), Pose: 1, Ms: 300, Sprite: []game.Picture{{File: "SPRIT", Block: 1}}}}
	out := played(es, true, func(play.Report) (*play.Combat, []game.Member) { return b, nil }, 0)
	if len(out) != 7 || out[0].Kind != game.EvMissile {
		t.Fatalf("%d frames", len(out))
	}
	first, last := out[0].battle.Effects, out[6].battle.Effects
	if len(first) != 1 || first[0].Y != 14 || last[0].Y != 12 || len(out[6].battle.EffectPix) != 1 {
		t.Fatalf("flight %v .. %v", first, last)
	}
}

// TestMissileStepMs: the game's StepMs paces each third of a square,
// whatever the moment's Ms.
func TestMissileStepMs(t *testing.T) {
	b := &play.Combat{Units: []game.Unit{{ID: 1, At: image.Pt(27, 14)}}}
	es := []play.Report{{Kind: game.EvMissile, Unit: 1, To: image.Pt(27, 12), Pose: 1, Ms: 300, StepMs: 20, Sprite: []game.Picture{{File: "SPRIT", Block: 1}}}}
	for _, f := range played(es, true, func(play.Report) (*play.Combat, []game.Member) { return b, nil }, 0) {
		if f.Ms != 20 {
			t.Fatalf("frame %d ms, want 20", f.Ms)
		}
	}
}

// TestMissileNoMoment: a shot with a sprite but no time or pose of its
// own (POR's) still flies, over flightMs.
func TestMissileNoMoment(t *testing.T) {
	b := &play.Combat{Units: []game.Unit{{ID: 1, At: image.Pt(27, 14)}}}
	es := []play.Report{{Kind: game.EvMissile, Unit: 1, To: image.Pt(27, 12), Sprite: []game.Picture{{File: "SPRIT", Block: 1}}}}
	out := played(es, true, func(play.Report) (*play.Combat, []game.Member) { return b, nil }, 0)
	if len(out) != 7 || out[0].battle == nil || len(out[0].battle.Effects) != 1 {
		t.Fatalf("%d frames", len(out))
	}
}
