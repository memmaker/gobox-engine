// Package classic is gobox's default front: the Gold Box UI built from a
// game's game.State alone (no game imports, no engine access).
package classic

import (
	"gobox/front"
	"gobox/game"
)

// partyFunctions are the party menu's functions by their labels.
var partyFunctions = map[string]front.ActKind{
	"Create Character": front.ActCreate, "Add Character": front.ActAdd, "Remove Character": front.ActRemove,
	"Modify Character": front.ActModify, "Train Character": front.ActTraining, "View Character": front.ActMember,
	"Load Saved Game": front.ActLoad, "Begin Adventuring": front.ActBegin, "Exit to DOS": front.ActQuit,
	"Drop Character": front.ActDropMember, "Human Change Class": front.ActNewClass, "Quick Start": front.ActQuickStart,
	"Forget Character": front.ActForget,
}

// HallDetail is the Where.Detail key saying the party menu is the training hall's.
const HallDetail = "hall"

// PartyMenu is the party menu's open functions (nil: not up) and whether it
// is the training hall's.
func PartyMenu(s game.State) (open []front.ActKind, hall bool) {
	if s.Mode != game.ModeParty || s.Pending == nil {
		return nil, false
	}
	for _, o := range s.Pending.Options {
		if k, ok := partyFunctions[o.Label]; ok {
			open = append(open, k)
		}
	}
	return open, s.Where.Detail[HallDetail] != 0
}
