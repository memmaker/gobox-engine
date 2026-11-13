package ui

import (
	"slices"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestRenamedPartyRows: a series' names for the party menu's functions
// (BR2: Remove is "Purge Character", Drop is "Remove Character") open the
// functions they name, both rows kept.
func TestRenamedPartyRows(t *testing.T) {
	s := shown{Mode: game.ModeParty, Pending: &play.Ask{Kind: game.Choice, Options: []play.Option{
		{Label: "Purge Character", Func: "Remove Character"}, {Label: "Remove Character", Func: "Drop Character"}}}}
	open, _ := PartyMenu(s.Mode, s.Training, s.Pending, s.Where.Hall)
	if !slices.Contains(open, ActRemove) || !slices.Contains(open, ActDropMember) {
		t.Fatalf("open %v, want Remove and Drop", open)
	}
}
