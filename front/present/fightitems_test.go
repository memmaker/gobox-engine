package present

import (
	"strings"
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
	"gobox/gpu"
)

// TestFightItemsTitle: a fight's items are the member's own: the title is
// never "Party Items".
func TestFightItemsTitle(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	var drawn strings.Builder
	g := func(c byte) [8]byte { drawn.WriteByte(c); return [8]byte{} }
	note := Note{Glyph: g, Font: func(c byte, _ bool) [8]byte { return g(c) }}
	v := front.View{}
	p := partyOn(game.Member{Name: "Kade"})
	p.eng.r.items = game.ItemsOwn
	p.board = &play.Combat{}
	p.Inventory(dst, note, v)
	if s := strings.ToUpper(drawn.String()); !strings.Contains(s, "ITEMS") || strings.Contains(s, "PARTY ITEMS") {
		t.Fatalf("drawn %q", drawn.String())
	}
}
