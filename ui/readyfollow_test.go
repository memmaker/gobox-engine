package ui

import (
	"testing"

	"gobox/game"
	"gobox/gpu"
)

// TestReadyKeepsItemHighlighted: a readied item leaves the list for the
// doll; the highlight stays on its row, now the next item.
func TestReadyKeepsItemHighlighted(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	note := Note{Glyph: func(byte) [8]byte { return [8]byte{} }, Font: func(byte, bool) [8]byte { return [8]byte{} }}
	mace := game.Item{Handle: 1, Name: "Mace", Type: 1, Kind: "weapon"}
	sling := game.Item{Handle: 2, Name: "Sling", Type: 2, Kind: "weapon"}
	party := func(items ...game.Item) game.State {
		return game.Exploring{Common: game.Common{Party: []game.Member{{Name: "A", Items: items}}}}
	}
	g, f := playGame(party(mace, sling))
	g.r.items = game.ItemsOwn
	p := &Presenter{}
	p.SetEngine(g)
	p.Inventory(dst, note)
	p.inv.choose(false, 2)
	p.inv.removeX(ActReady, 2, 0)
	sling.Readied = true
	f.st = party(mace, sling)
	p.Inventory(dst, note)
	if it := p.inv.itemAt(p.inv.sel); it == nil || it.Name != "Mace" {
		t.Fatalf("highlight on %+v, want the Mace", it)
	}
}
