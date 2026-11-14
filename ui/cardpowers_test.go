package ui

import (
	"slices"
	"testing"

	"gobox/game"
)

// TestCardShowsPowers: a member's card menu offers its powers (a paladin's
// Heal and Cure), not only what the Allowed list names.
func TestCardShowsPowers(t *testing.T) {
	m := game.Member{Powers: []game.PowerOffer{{Label: "Lay on Hands", Kind: game.PowerHeal}, {Label: "Cure Disease", Kind: game.PowerCure}}}
	var names []string
	for _, e := range card(m, game.Allowed{}, 0) {
		names = append(names, e.Name)
	}
	if want := []string{"View", "Heal", "Cure"}; !slices.Equal(names, want) {
		t.Errorf("card menu %v, want %v", names, want)
	}
}
