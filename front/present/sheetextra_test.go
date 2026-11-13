package present

import (
	"strings"
	"testing"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
	"gobox/gpu"
)

// TestSheetDrawsExtraFields: the two-panel sheet (shop, temple) draws the
// weapon, armour and load the game fills, as the page layout does.
func TestSheetDrawsExtraFields(t *testing.T) {
	dst := gpu.NewImage(1280, 800)
	defer dst.Deallocate()
	var drawn strings.Builder
	g := func(c byte) [8]byte { drawn.WriteByte(c); return [8]byte{} }
	note := Note{Glyph: g, Font: func(c byte, _ bool) [8]byte { return g(c) }}
	sheetOn(&game.SheetPage{Weapon: "LONGSWORD", Armour: "CHAIN MAIL", Encumbrance: "NORMAL"}).drawSheet(dst, note, front.View{})
	for _, w := range []string{"LONGSWORD", "CHAIN MAIL", "NORMAL", "Load"} {
		if !strings.Contains(drawn.String(), w) {
			t.Errorf("%q not drawn", w)
		}
	}
}

// sheetOn is a presenter whose game has one member, A, with page as the sheet up.
func sheetOn(page *game.SheetPage) *Presenter {
	p := partyOn(game.Member{Name: "A", HP: 5, MaxHP: 5, Okay: true, Page: page})
	p.eng.r.sheet = 0
	p.bars = gpu.NewImageFromImage(barArt())
	return p
}

// partyOn is a presenter whose game explores with the members ms, the first selected.
func partyOn(ms ...game.Member) *Presenter {
	g, _ := playGame(play.Exploring{Common: play.Common{Party: ms}})
	p := &Presenter{}
	p.SetEngine(g)
	return p
}
