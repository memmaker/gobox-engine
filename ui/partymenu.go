// partymenu.go is gobox's default front: the Gold Box UI built from a
// game's play.State alone (no game imports, no engine access).

package ui

import (
	"cmp"
	"gobox/game"
	"gobox/game/play"
)

// partyFunctions are the party menu's functions by their labels.
var partyFunctions = map[string]ActKind{
	"Create Character": FrontActCreate, "Add Character": ActAdd, "Remove Character": ActRemove,
	"Modify Character": ActModify, "Train Character": ActTraining, "View Character": ActMember,
	"Load Saved Game": ActLoad, "Begin Adventuring": ActBegin, "Exit to DOS": ActQuit,
	"Drop Character": ActDropMember, "Human Change Class": ActNewClass, "Quick Start": ActQuickStart,
	"Forget Character": ActForget, "Purge Character": ActRemove,
	"Icon Select": ActIconEdit, "Save Current Game": ActSaveGame,
}

// PartyMenu is the party menu's open functions (nil: not up) and whether it
// is the training hall's.
func PartyMenu(mode game.Mode, training *game.Training, pending *play.Ask, hallHere bool) (open []ActKind, hall bool) {
	inHall := mode == game.ModeTraining && training == nil // the hall's party menu before OpenTraining (SotSB publishes the same Question)
	if mode != game.ModeParty && !inHall || pending == nil {
		return nil, false
	}
	for _, o := range pending.Options {
		if k, ok := partyFunctions[cmp.Or(o.Func, o.Label)]; ok { // by its function, never its series name (BR2's Drop is labelled Remove)
			open = append(open, k)
		}
	}
	return open, hallHere || inHall
}
