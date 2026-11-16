package ui

import (
	"testing"

	"gobox/game"
	"gobox/rules"
)

// A shot with a sprite plays as its flight: the sprite drawn square by
// square (in thirds) from the shooter to the target, over the moment.
func TestMissileFlies(t *testing.T) {
	b := &game.Combat{Units: []game.Unit{{ID: 1, At: rules.BoardPos{X: 27, Y: 14}}}}
	es := []game.Report{{Kind: game.EvMissile, Unit: 1, At: rules.BoardPos{X: 27, Y: 14}, To: rules.BoardPos{X: 27, Y: 12}, Pose: 1, Ms: 300, Sprite: []game.Picture{{File: "SPRIT", Block: 1}}}}
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
	b := &game.Combat{Units: []game.Unit{{ID: 1, At: rules.BoardPos{X: 27, Y: 14}}}}
	es := []game.Report{{Kind: game.EvMissile, Unit: 1, At: rules.BoardPos{X: 27, Y: 14}, To: rules.BoardPos{X: 27, Y: 12}, Pose: 1, Ms: 300, StepMs: 20, Sprite: []game.Picture{{File: "SPRIT", Block: 1}}}}
	for _, f := range played(es, true, func(game.Report) (*game.Combat, []game.Member) { return b, nil }, 0) {
		if f.Ms != 20 {
			t.Fatalf("frame %d ms, want 20", f.Ms)
		}
	}
}

// TestMissileNoMoment: a shot with a sprite but no time or pose of its
// own (POR's) still flies, over flightMs.
func TestMissileNoMoment(t *testing.T) {
	b := &game.Combat{Units: []game.Unit{{ID: 1, At: rules.BoardPos{X: 27, Y: 14}}}}
	es := []game.Report{{Kind: game.EvMissile, Unit: 1, At: rules.BoardPos{X: 27, Y: 14}, To: rules.BoardPos{X: 27, Y: 12}, Sprite: []game.Picture{{File: "SPRIT", Block: 1}}}}
	out := played(es, true, func(game.Report) (*game.Combat, []game.Member) { return b, nil }, 0)
	if len(out) != 7 || boardOfFrame(out[0]) == nil || len(boardOfFrame(out[0]).Effects) != 1 {
		t.Fatalf("%d frames", len(out))
	}
}

// TestMissileFromStart: a shot flies from the game's start square (the
// shooter's), never from a square guessed off the board: one whose shooter
// the moment's board does not hold flew from the map's corner (0,0).
func TestMissileFromStart(t *testing.T) {
	b := &game.Combat{Units: []game.Unit{{ID: 1, At: rules.BoardPos{X: 27, Y: 14}}}}
	es := []game.Report{{Kind: game.EvMissile, Unit: 9, At: rules.BoardPos{X: 20, Y: 10}, To: rules.BoardPos{X: 26, Y: 10}, Sprite: []game.Picture{{File: "SPRIT", Block: 1}}}}
	out := played(es, true, func(game.Report) (*game.Combat, []game.Member) { return b, nil }, 0)
	if f := boardOfFrame(out[0]).Effects; len(f) != 1 || f[0].X != 20 || f[0].Y != 10 {
		t.Fatalf("starts at %v, want 20,10", f)
	}
}
