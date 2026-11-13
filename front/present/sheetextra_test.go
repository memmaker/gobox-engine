package present

import (
	"strings"
	"testing"

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
	sheetOn(&game.SheetPage{Weapon: "LONGSWORD", Armour: "CHAIN MAIL", Encumbrance: "NORMAL"}).drawSheet(dst, note)
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

// upOf is the inventory screen the state st puts up when the game has items up.
func upOf(st play.State, items game.ItemsUp) *invUp {
	g, _ := playGame(st)
	g.r.items = items
	return g.r.inventoryOf(g.r.g.State())
}

// menuOn is a presenter whose game waits on the party menu offering the functions named (by their gobox names).
func menuOn(funcs ...string) *Presenter {
	g, _ := playGame(play.Assembling{Common: play.Common{Ask: &play.Question{Replies: funcs}}})
	p := &Presenter{}
	p.SetEngine(g)
	return p
}

// draftAsked is draftOn with q asked over the creation.
func draftAsked(d *play.Draft, q *play.Question) *Presenter {
	p, _ := presenterOn(play.Assembling{Common: play.Common{Ask: q}, Draft: d})
	return p
}

// draftOn is a presenter whose game is making a character: its draft d.
func draftOn(d *play.Draft) (*Presenter, *fake) {
	g, f := playGame(play.Assembling{Draft: d})
	p := &Presenter{}
	p.SetEngine(g)
	return p, f
}
