package present

import (
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// The presenter's Appraise and Keep reach the game as play.Appraise, by the
// offer the game made.
func TestAppraiseThroughFront(t *testing.T) {
	f := &fake{st: play.Trading{Offers: []game.Offer{{Member: 2, Offer: 1, Kind: game.Jewelry}}}}
	g := NewGame(f, game.Data{})
	if err := g.Appraise(true); err != nil {
		t.Fatal(err)
	}
	if err := g.Appraise(false); err == nil {
		t.Error("gems appraised with none offered")
	}
	f.st = play.Trading{Offers: []game.Offer{{Member: 2, Offer: 1, Kind: game.Jewelry, Value: 50}}}
	if err := g.KeepAppraised(true); err != nil {
		t.Fatal(err)
	}
	want := []play.Action{play.Appraise{Member: 2, Offer: 1}, play.Appraise{Member: 2, Offer: 1, Keep: true}}
	if len(f.sent) != 2 || f.sent[0] != want[0] || f.sent[1] != want[1] {
		t.Fatalf("sent %v, want %v", f.sent, want)
	}
}
