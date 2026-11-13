package ui

import (
	"testing"

	"gobox/game"
)

// The pool's portrait and icon on the creation sheet (a Name, no File) are
// content the screen draws, as the game's DAX pictures there are.
func TestContentPoolPictures(t *testing.T) {
	m := game.Member{PortraitHeadID: "head_000", PortraitBodyID: "body_000", IconHeadID: "combat_head_000", IconBodyID: "combat_body_000"}
	if poolPicture("portraits", m.PortraitHeadID) == nil {
		t.Skip("no pool art")
	}
	p, _ := draftOn(&game.Draft{Member: m}) // the page's default layout: the portrait at 28,1, the icon at 29,13
	p.creating = true
	names := map[string]bool{}
	for _, pc := range p.contentPictures() {
		names[pc.Name] = true
	}
	if !names[m.PortraitHeadID] || !names[m.PortraitBodyID] || !names[m.IconHeadID] || names[""] {
		t.Fatalf("content pictures %v, want the pool's portrait and icon only", names)
	}
}
