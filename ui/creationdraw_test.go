package ui

import (
	"testing"

	"gobox/game"
)

// The creation sheet's portrait and icon are the presenter's (the pool's,
// placed on its own layout); no picture of the engine's page reaches it.
func TestCreationArtPresenterPlaced(t *testing.T) {
	m := game.Member{PortraitHeadID: "head_000", PortraitBodyID: "body_000", IconHeadID: "combat_head_000", IconBodyID: "combat_body_000"}
	if poolPicture("portraits", m.PortraitHeadID) == nil {
		t.Skip("no pool art")
	}
	p, _ := draftOn(&game.Draft{Member: m})
	p.creating = true
	if pics := p.contentPictures(); len(pics) != 0 {
		t.Fatalf("engine pictures on the sheet: %+v", pics)
	}
	cr, port, icons := p.creation(), 0, 0
	for _, o := range p.screenPage().overs {
		switch {
		case o.At.In(cr.Head.Pixels().Union(cr.Body.Pixels())):
			port++
		case o.At.In(cr.Icon.Pixels()):
			icons++
		}
	}
	if port != 1 || icons != 2 {
		t.Fatalf("%d portraits and %d icon poses placed, want 1 and 2", port, icons)
	}
}
