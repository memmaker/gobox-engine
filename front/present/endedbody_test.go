package present

import (
	"image"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestEndedFightBody: the last blow ends the fight (no Combat after it):
// an EvDown's Picture is the body on the unit's square of the boards the
// events after it play on (POD's bodies vanished).
func TestEndedFightBody(t *testing.T) {
	r := &fightRules{}
	g := NewGame(r, game.Data{})
	g.lastCombat = &play.Combat{Size: image.Pt(8, 8), Units: []game.Unit{{ID: 4, At: image.Pt(3, 2), HP: 1}}}
	body := game.Picture{W: 1, H: 1, Pixels: []byte{5}, Clear: -1}
	g.events = []play.Report{{Kind: game.EvDown, Unit: 4, Picture: body, Ms: 1}, {Kind: game.EvMessage, Text: "won", Ms: 100}}
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
	g.lastCombat = &play.Combat{Size: image.Pt(8, 8), Tiles: []string{"grass", "venuscom_tile_003"},
		Squares: []game.Ground{{At: image.Pt(3, 2), Tile: 0}},
		Units:   []game.Unit{{ID: 1, At: image.Pt(3, 2), Size: 2, HP: 1}}}
	g.events = []play.Report{{Kind: game.EvDown, Unit: 1, Picture: game.Picture{Name: "venuscom_tile_003"}, Ms: 1}, {Kind: game.EvMessage, Text: "won", Ms: 100}}
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
// (a turn moved it to 0,0: BR2's beast vanished before its fall); the
// fall's moment shows the body, not the figure standing on it.
func TestEndedFightBeastStays(t *testing.T) {
	r := &fightRules{}
	g := NewGame(r, game.Data{})
	g.lastCombat = &play.Combat{Size: image.Pt(8, 8), Tiles: []string{"grass", "venuscom_tile_003"},
		Units: []game.Unit{{ID: 5, At: image.Pt(1, 1), HP: 9}, {ID: 7, At: image.Pt(3, 2), Size: 2, HP: 3, Foe: true}}}
	g.events = []play.Report{{Kind: game.EvMove, Unit: 7, N: 2, Turn: true},
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
	if len(v.Events) != 5 || v.Events[1].Battle == nil || !beast(1) {
		t.Fatal("the beast is not on its square for the blow")
	}
	for _, i := range []int{3, 4} { // the fall's own moment shows the body in its place (a figure on it hid it), and after
		if b := v.Events[i].Battle; beast(i) || b.Ground[2][3].Overlay != 2 {
			t.Fatalf("event %d: beast %v, overlay %d", i, beast(i), b.Ground[2][3].Overlay)
		}
	}
}

// TestLiveFallBody: a fall with a body (Ms 0) in a fight still up is held
// with the body on the unit's square and the unit off the board (POR's
// last blow: the page came straight after, Mirelle standing to the end).
func TestLiveFallBody(t *testing.T) {
	r := &fightRules{}
	r.st = play.Fighting{Field: play.Battlefield{Size: image.Pt(8, 8), Units: []game.Unit{{ID: 9, Member: -1, At: image.Pt(5, 5), HP: 9, Foe: true}}}}
	g := NewGame(r, game.Data{})
	g.lastCombat = &play.Combat{Size: image.Pt(8, 8), Units: []game.Unit{{ID: 2, At: image.Pt(3, 2), HP: 1}, {ID: 9, At: image.Pt(5, 5), HP: 9, Foe: true}}}
	body := game.Picture{W: 1, H: 1, Pixels: []byte{5}, Clear: -1}
	g.events = []play.Report{{Kind: game.EvDown, Unit: 2, Picture: body}}
	v := g.Settle()
	if len(v.Events) != 1 || v.Events[0].Ms == 0 || v.Events[0].Battle == nil {
		t.Fatalf("the fall has no moment: %+v", v.Events)
	}
	b := v.Events[0].Battle
	for _, u := range b.Units {
		if u.ID == 2 {
			t.Fatal("the fallen unit still stands")
		}
	}
	if b.Ground[2][3].Overlay == 0 {
		t.Fatal("no body on the square")
	}
}
