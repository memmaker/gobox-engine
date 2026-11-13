package present

import (
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
	"gobox/gpu"
)

// TestReadyKeepsItemHighlighted: readying moves the item (readied gear
// first); the highlight follows the item, not its old row.
func TestReadyKeepsItemHighlighted(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	note := Note{Glyph: func(byte) [8]byte { return [8]byte{} }, Font: func(byte, bool) [8]byte { return [8]byte{} }}
	mace := game.Item{Handle: 1, Name: "Mace", Type: 1, Kind: "weapon"}
	sling := game.Item{Handle: 2, Name: "Sling", Type: 2, Kind: "weapon"}
	party := func(items ...game.Item) play.State {
		return play.Exploring{Common: play.Common{Party: []game.Member{{Name: "A", Items: items}}}}
	}
	g, f := playGame(party(mace, sling))
	g.r.items = game.ItemsOwn
	p := &Presenter{}
	p.SetEngine(g)
	v := front.View{}
	p.Inventory(dst, note, v)
	p.inv.choose(false, 2)
	p.inv.removeX(front.ActReady, 2, 0)
	sling.Readied = true // the rules list readied gear first (each keeps its handle)
	f.st = party(sling, mace)
	p.Inventory(dst, note, v)
	if it := p.inv.itemAt(p.inv.sel); it == nil || it.Name != "Sling" {
		t.Fatalf("highlight on %+v, want the Sling", it)
	}
}
