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
