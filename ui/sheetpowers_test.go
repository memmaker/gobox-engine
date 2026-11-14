package ui

import (
	"testing"

	"gobox/game"
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
