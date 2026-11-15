package ui

import (
	"strings"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// TestSheetShowsPowers: the sheet's menu offers the member's powers even
// when the game sent no sheet of its own (gobox's sheet, opened from a card).
func TestSheetShowsPowers(t *testing.T) {
	p := partyOn(game.Member{Name: "A", Powers: []game.PowerOffer{{Label: "Cast", Kind: game.PowerCast}, {Label: "Cure", Kind: game.PowerCure}}})
	var got []string
	for _, r := range p.sheetItems() {
		got = append(got, r.Text)
	}
	if len(got) != 2 || got[0] != "Cast" || got[1] != "Cure" {
		t.Errorf("sheet menu %v, want [Cast Cure]", got)
	}
}

// TestFightSheetShowsPowers: a fight's look-only sheet offers the powers a
// fight takes: the spell effects, and Cast to the member whose turn it is
// (Heal and Cure are not a fight's).
func TestFightSheetShowsPowers(t *testing.T) {
	powers := []game.PowerOffer{{Label: "Cast", Kind: game.PowerCast}, {Label: "Magic", Kind: game.PowerDisplay}, {Label: "Heal", Kind: game.PowerHeal}}
	st := game.Fighting{Common: game.Common{
		Party:   []game.Member{{Name: "A", Powers: powers, Spells: []game.SpellRef{{ID: 3, Name: "Bless"}}}, {Name: "B", Powers: powers}},
		Allowed: game.Allowed{Plain: []game.Action{game.Cast{Member: 0, Spell: 3}}}},
		Field: game.Battlefield{Size: rules.Size{W: 6, H: 6}, Turn: 1,
			Units: []game.Unit{{ID: 1, Member: 0, At: rules.BoardPos{X: 1, Y: 1}}, {ID: 2, Member: 1, At: rules.BoardPos{X: 2, Y: 1}}}}}
	p, _ := presenterOn(st)
	for look, want := range map[int]string{1: "Cast Magic", 2: "Magic"} {
		p.eng.r.lookSheet = look
		var got []string
		for _, r := range p.sheetItems() {
			got = append(got, r.Text)
		}
		if s := strings.Join(got, " "); s != want {
			t.Errorf("member %d's fight sheet menu %q, want %q", look-1, s, want)
		}
	}
}
