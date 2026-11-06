package present

import (
	"image"
	"testing"

	"gobox/front"
)

func campCardView() front.View {
	return front.View{Camp: &front.Camp{}, Party: []front.Member{{Name: "Ann", Selected: true, Powers: []front.Power{
		{Name: "Icon", Kind: front.ActIconEdit}, {Name: "Drop", Kind: front.ActDropMember}}}}}
}

// TestCampCardKey: camp's card verbs answer their hotkey without Tab.
func TestCampCardKey(t *testing.T) {
	p := &Presenter{}
	p.campKey('d', campCardView())
	if p.camp.screen != campDrop {
		t.Fatalf("d: screen %d, want the drop's question", p.camp.screen)
	}
}

// TestCampCardClickRows: a click on a card's line acts the line drawn
// there (Icon is not drawn on the card, so row 0 is Drop).
func TestCampCardClickRows(t *testing.T) {
	v := campCardView()
	p := &Presenter{active: true, shownView: v}
	p.campCtx = [][]image.Rectangle{{image.Rect(0, 0, 50, 8)}}
	p.Click(10, 4)
	if p.camp.screen != campDrop {
		t.Fatalf("click on Drop: screen %d, sent %v", p.camp.screen, p.inv.acts)
	}
}
