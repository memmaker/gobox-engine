// Package classic is gobox's default front: the Gold Box UI built from a
// game's game.State alone (no game imports, no engine access).
package classic

import (
	"cmp"
	"gobox/front"
	"gobox/game"
)

// partyFunctions are the party menu's functions by their labels.
var partyFunctions = map[string]front.ActKind{
	"Create Character": front.ActCreate, "Add Character": front.ActAdd, "Remove Character": front.ActRemove,
	"Modify Character": front.ActModify, "Train Character": front.ActTraining, "View Character": front.ActMember,
	"Load Saved Game": front.ActLoad, "Begin Adventuring": front.ActBegin, "Exit to DOS": front.ActQuit,
	"Drop Character": front.ActDropMember, "Human Change Class": front.ActNewClass, "Quick Start": front.ActQuickStart,
	"Forget Character": front.ActForget, "Purge Character": front.ActRemove,
	"Icon Select": front.ActIconEdit, "Save Current Game": front.ActSaveGame,
}

// PartyMenu is the party menu's open functions (nil: not up) and whether it
// is the training hall's.
func PartyMenu(mode game.Mode, training *game.Training, pending *game.Ask, hallHere bool) (open []front.ActKind, hall bool) {
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

// View is v with the screens classic builds from s in place of the game's:
// the party menu (MainMenu, Hall), the question (Ask) and the treasure
// screen (Loot, under the game's inventory and sheet while they are up);
// the map (Map), camp (Camp), temple (Temple), shop (Shop) and
// inventory (Inventory) over the game's.
// The title's Logo and Backdrop stay the game's while the menu is up.
func View(v front.View, s game.State) front.View {
	v.MainMenu, v.Hall = PartyMenu(s.Mode, s.Training, s.Pending, s.Where.Hall)
	v.Map, v.Camp = Map(v.Map, s.Map, s.Where), Camp(v.Camp, s.Party, s.Selected, s.RestHeals, s.RestStudies, s.Allowed["NoRest"], s.Bestiary)
	v.Temple, v.Shop, v.Inventory = Temple(v.Temple, s.Temple, s.TempleOffers, s.Purse, s.Allowed, s.Party, s.Selected), Shop(v.Shop, s.Shop, s.Purse, s.Allowed, s.Party), Inventory(v.Inventory, s.Mode, s.Party, s.Selected, s.PartyCap, s.PoolItems, s.Purse, s.Allowed, s.Shop, s.Loot)
	if q, ok := Ask(s.Mode, s.Pending); ok && q != nil { // none pending keeps v's own (the Game's "You found some loot!"): View lays over a Game's View
		v.Ask = q
	}
	if v.MainMenu == nil && v.Ask == nil && v.Camp == nil && v.Temple == nil && v.Shop == nil && v.Inventory == nil && s.Loot == nil && s.Training == nil &&
		s.Pending != nil && s.Mode >= game.ModeCamp && s.Mode <= game.ModeLoot { // a place's mode with no screen of its own yet: its question, never a blank frame
		if q, ok := Ask(game.ModeAsk, s.Pending); ok {
			v.Ask = q
		}
	}
	if v.MainMenu == nil {
		v.Logo, v.Backdrop = nil, nil
	}
	return v
}
