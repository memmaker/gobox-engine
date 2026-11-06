package classic

import (
	"slices"
	"testing"

	"gobox/front"
	"gobox/game"
)

// TestRenamedPartyRows: a series' names for the party menu's functions
// (BR2: Remove is "Purge Character", Drop is "Remove Character") open the
// functions they name, both rows kept.
func TestRenamedPartyRows(t *testing.T) {
	s := game.State{Mode: game.ModeParty, Pending: &game.Ask{Kind: game.Choice, Options: []game.Option{
		{Label: "Purge Character", Func: "Remove Character"}, {Label: "Remove Character", Func: "Drop Character"}}}}
	open, _ := PartyMenu(s)
	if !slices.Contains(open, front.ActRemove) || !slices.Contains(open, front.ActDropMember) {
		t.Fatalf("open %v, want Remove and Drop", open)
	}
}
