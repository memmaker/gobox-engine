package ui

import (
	"image"
	"testing"

	"gobox/game"
)

// A shot with a sprite plays as its flight: the sprite drawn square by
// square (in thirds) from the shooter to the target, over the moment.
func TestMissileFlies(t *testing.T) {
	b := &game.Combat{Units: []game.Unit{{ID: 1, At: image.Pt(27, 14)}}}
	es := []game.Report{{Kind: game.EvMissile, Unit: 1, To: image.Pt(27, 12), Pose: 1, Ms: 300, Sprite: []game.Picture{{File: "SPRIT", Block: 1}}}}
	out := played(es, true, func(game.Report) (*game.Combat, []game.Member) { return b, nil }, 0)
	if len(out) != 7 || out[0].Kind != game.EvMissile {
		t.Fatalf("%d frames", len(out))
	}
	first, last := boardOfFrame(out[0]).Effects, boardOfFrame(out[6]).Effects
	if len(first) != 1 || first[0].Y != 14 || last[0].Y != 12 {
		t.Fatalf("flight %v .. %v", first, last)
	}
}

// TestMissileStepMs: the game's StepMs paces each third of a square,
// whatever the moment's Ms.
func TestMissileStepMs(t *testing.T) {
	b := &game.Combat{Units: []game.Unit{{ID: 1, At: image.Pt(27, 14)}}}
	es := []game.Report{{Kind: game.EvMissile, Unit: 1, To: image.Pt(27, 12), Pose: 1, Ms: 300, StepMs: 20, Sprite: []game.Picture{{File: "SPRIT", Block: 1}}}}
	for _, f := range played(es, true, func(game.Report) (*game.Combat, []game.Member) { return b, nil }, 0) {
		if f.Ms != 20 {
			t.Fatalf("frame %d ms, want 20", f.Ms)
		}
	}
}

// TestMissileNoMoment: a shot with a sprite but no time or pose of its
// own (POR's) still flies, over flightMs.
func TestMissileNoMoment(t *testing.T) {
	b := &game.Combat{Units: []game.Unit{{ID: 1, At: image.Pt(27, 14)}}}
	es := []game.Report{{Kind: game.EvMissile, Unit: 1, To: image.Pt(27, 12), Sprite: []game.Picture{{File: "SPRIT", Block: 1}}}}
	out := played(es, true, func(game.Report) (*game.Combat, []game.Member) { return b, nil }, 0)
	if len(out) != 7 || boardOfFrame(out[0]) == nil || len(boardOfFrame(out[0]).Effects) != 1 {
		t.Fatalf("%d frames", len(out))
	}
}
