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
	return open, s.Where.Hall
}

// View is v with the screens classic builds from s in place of the game's:
// the party menu (MainMenu, Hall), the question (Ask) and the treasure
// screen (Loot, under the game's inventory and sheet while they are up).
// The title's Logo and Backdrop stay the game's while the menu is up.
func View(v front.View, s game.State) front.View {
	v.MainMenu, v.Hall = PartyMenu(s)
	if v.Inventory == nil && v.Sheet == nil {
		v.Loot = Loot(s)
	}
	if q, ok := Ask(s); ok {
		v.Ask = q
	}
	if v.MainMenu == nil {
		v.Logo, v.Backdrop = nil, nil
	}
	return v
}
