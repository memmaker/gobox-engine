package classic

import (
	"image"
	"testing"

	"gobox/game"
)

// TestEndedFightBody: the last blow ends the fight (no Combat after it):
// an EvDown's Picture is the body on the unit's square of the boards the
// events after it play on (POD's bodies vanished).
func TestEndedFightBody(t *testing.T) {
	r := &fightRules{}
	g := NewGame(r, game.Data{})
	g.lastCombat = &game.Combat{Size: image.Pt(8, 8), Units: []game.Unit{{ID: 4, At: image.Pt(3, 2), HP: 1}}}
	body := game.Picture{W: 1, H: 1, Pixels: []byte{5}, Clear: -1}
	g.events = []game.Event{{Kind: game.EvDown, Unit: 4, Picture: body, Ms: 1}, {Kind: game.EvMessage, Text: "won", Ms: 100}}
	v := g.Settle()
	if len(v.Events) != 2 || v.Events[1].Battle == nil {
		t.Fatalf("events %+v", v.Events)
	}
	if b := v.Events[1].Battle; b.Ground[2][3].Overlay == 0 {
		t.Fatal("no body on the square after the last blow")
	}
}

// TestEndedFightTileBody: a body named by a Combat.Tiles name (BR2's
// "venuscom_tile_003") is that tile's slot, for a 2x2 unit too (on its At).
func TestEndedFightTileBody(t *testing.T) {
	r := &fightRules{}
	g := NewGame(r, game.Data{})
	g.lastCombat = &game.Combat{Size: image.Pt(8, 8), Tiles: []string{"grass", "venuscom_tile_003"},
		Squares: []game.Ground{{At: image.Pt(3, 2), Tile: 0}},
		Units:   []game.Unit{{ID: 1, At: image.Pt(3, 2), Size: 2, HP: 1}}}
	g.events = []game.Event{{Kind: game.EvDown, Unit: 1, Picture: game.Picture{Name: "venuscom_tile_003"}, Ms: 1}, {Kind: game.EvMessage, Text: "won", Ms: 100}}
	v := g.Settle()
	if len(v.Events) != 2 || v.Events[1].Battle == nil {
		t.Fatalf("events %+v", v.Events)
	}
	if o := v.Events[1].Battle.Ground[2][3].Overlay; o != 2 {
		t.Fatalf("overlay %d, want the tile's slot 1 (+1)", o)
	}
}

// TestEndedFightBeastStays: a 2x2 beast turning (EvMove Turn, To unset)
// then struck stays on its square until its EvDown, then its body does
// (a turn moved it to 0,0: BR2's beast vanished before its fall).
func TestEndedFightBeastStays(t *testing.T) {
	r := &fightRules{}
	g := NewGame(r, game.Data{})
	g.lastCombat = &game.Combat{Size: image.Pt(8, 8), Tiles: []string{"grass", "venuscom_tile_003"},
		Units: []game.Unit{{ID: 5, At: image.Pt(1, 1), HP: 9}, {ID: 7, At: image.Pt(3, 2), Size: 2, HP: 3, Foe: true}}}
	g.events = []game.Event{{Kind: game.EvMove, Unit: 7, N: 2, Turn: true},
		{Kind: game.EvAttack, Unit: 5, Target: 7, Pose: 1, Ms: 400}, {Kind: game.EvDamage, Unit: 7, Target: 7, N: 3},
		{Kind: game.EvDown, Unit: 7, Target: 7, Picture: game.Picture{Name: "venuscom_tile_003"}, Ms: 500}, {Kind: game.EvMessage, Text: "won", Ms: 100}}
	v := g.Settle()
	beast := func(i int) bool {
		for _, u := range v.Events[i].Battle.Units {
			if u.ID == 7 {
				return u.X == 3 && u.Y == 2
			}
		}
		return false
	}
	if len(v.Events) != 5 || v.Events[1].Battle == nil || !beast(1) || !beast(3) {
		t.Fatal("the beast is not on its square for the blow and its fall")
	}
	if b := v.Events[4].Battle; beast(4) || b.Ground[2][3].Overlay != 2 {
		t.Fatalf("after the fall: beast %v, overlay %d", beast(4), b.Ground[2][3].Overlay)
	}
}
