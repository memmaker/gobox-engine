package classic

import (
	"image"
	"testing"

	"gobox/front"
	"gobox/game"
)

// A shot with a sprite plays as its flight: the sprite drawn square by
// square (in thirds) from the shooter to the target, over the moment.
func TestMissileFlies(t *testing.T) {
	b := &front.Battle{Units: []front.Unit{{ID: 1, X: 27, Y: 14}}}
	es := []game.Event{{Kind: game.EvMissile, Unit: 1, To: image.Pt(27, 12), Pose: 1, Ms: 300, Sprite: []game.Picture{{File: "SPRIT", Block: 1}}}}
	out := combatEvents(es, front.View{Battle: b}, func(game.Event) *front.Battle { return b }, 0)
	if len(out) != 7 || out[0].Kind != front.EvMissile {
		t.Fatalf("%d frames", len(out))
	}
	first, last := out[0].Battle.Effects, out[6].Battle.Effects
	if len(first) != 1 || first[0].Y != 14 || last[0].Y != 12 || len(out[6].Battle.EffectPix) != 1 || len(b.Effects) != 0 {
		t.Fatalf("flight %v .. %v", first, last)
	}
}
