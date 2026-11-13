// Package classic is gobox's default front: the Gold Box UI built from a
// game's play.State alone (no game imports, no engine access).
package present

import (
	"cmp"
	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// partyFunctions are the party menu's functions by their labels.
var partyFunctions = map[string]front.ActKind{
	"Create Character": front.ActCreate, "Add Character": front.ActAdd, "Remove Character": front.ActRemove,
	"Modify Character": front.ActModify, "Train Character": front.ActTraining, "View Character": front.ActMember,
	"Load Saved Game": front.ActLoad, "Begin Adventuring": front.ActBegin, "Exit to DOS": front.ActQuit,
	"Drop Character": front.ActDropMember, "Human Change Class": front.ActNewClass, "Quick Start": front.ActQuickStart,
	"Forget Character": front.ActForget, "Purge Character": front.ActRemove,
	"Icon Select": front.ActIconEdit, "Save Current Game": ActSaveGame,
}

// PartyMenu is the party menu's open functions (nil: not up) and whether it
// is the training hall's.
func PartyMenu(mode game.Mode, training *game.Training, pending *play.Ask, hallHere bool) (open []front.ActKind, hall bool) {
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
