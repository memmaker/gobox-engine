package present

import (
	"image"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// campPresenter is a presenter over a camp of one member, Ann.
func campPresenter() *Presenter {
	g, _ := playGame(play.Camping{Common: play.Common{Party: []game.Member{{Name: "Ann"}},
		Allowed: play.Allowed{Plain: []play.Action{play.EditIcon{Member: 0}, play.DropMember{Member: 0}}}}}) // the card: View, Icon (not drawn), Drop
	p := &Presenter{}
	p.SetEngine(g)
	return p
}

// TestCampCardKey: camp's card verbs answer their hotkey without Tab.
func TestCampCardKey(t *testing.T) {
	p := campPresenter()
	p.campKey('d')
	if p.camp.screen != campDrop {
		t.Fatalf("d: screen %d, want the drop's question", p.camp.screen)
	}
}

// TestCampCardClickRows: a click on a card's line acts the line drawn
// there (Icon is not drawn on the card, so row 1 is Drop).
func TestCampCardClickRows(t *testing.T) {
	p := campPresenter()
	p.active = true
	p.campCtx = [][]image.Rectangle{{image.Rect(0, 0, 50, 8), image.Rect(0, 8, 50, 16)}}
	p.Click(10, 12)
	if p.camp.screen != campDrop {
		t.Fatalf("click on Drop: screen %d, sent %v", p.camp.screen, p.inv.acts)
	}
}
