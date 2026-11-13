package ui

import (
	"slices"
	"testing"

	"gobox/game"
)

func TestPartyMenu(t *testing.T) {
	s := shown{Mode: game.ModeParty, Where: game.Where{Hall: true},
		Pending: &game.Ask{Kind: game.Choice, Options: []game.Option{{Label: "Begin Adventuring"}, {Label: "Forget Character"}, {Label: "Exit to DOS"}}}}
	open, hall := PartyMenu(s.Mode, s.Training, s.Pending, s.Where.Hall)
	if !hall || !slices.Equal(open, []ActKind{ActBegin, ActForget, ActQuit}) {
		t.Fatal(open, hall)
	}
	if open, _ := PartyMenu(game.ModeExplore, nil, nil, false); open != nil {
		t.Fatal(open)
	}
}
