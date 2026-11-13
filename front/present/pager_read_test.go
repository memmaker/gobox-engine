package present

import (
	"testing"

	"gobox/game/play"
)

// The pager stays on its page while the game's Read is up (the same pointer
// every frame), and closes once it is gone.
func TestSyncReadKeepsPage(t *testing.T) {
	g, f := playGame(play.Scene{Kind: play.SceneHandout, Handout: "J", Lines: []string{"a", "b"}})
	p := &Presenter{}
	p.SetEngine(g)
	p.syncRead()
	p.pager.P = 1
	p.syncRead()
	if p.pager == nil || p.pager.P != 1 {
		t.Fatalf("pager reset: %+v", p.pager)
	}
	f.st = play.Exploring{}
	p.syncRead()
	if p.pager != nil {
		t.Errorf("pager kept after the read ended")
	}
}
