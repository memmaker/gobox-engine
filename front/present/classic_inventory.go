package present

import (
	"slices"
	"strings"

	"gobox/core"
	"gobox/front"
	"gobox/game"
)

// DollSlots are the paper doll's slots by Inventory.Doll index, by
// game.Item.Slot name: core.DollNames (arrows and bolts share the Missiles
// slot), then the Launcher slot a launcher takes when the weapon slot is taken.
var DollSlots = append(core.DollNames[:], "launcher")

// dollSlot is the Doll index for it, -1 for none.
func dollSlot(it game.Item, weaponTaken bool) int {
	switch it.Slot {
	case "bolts":
		return slices.Index(DollSlots, "arrows")
	case "weapon":
		if weaponTaken {
			return slices.Index(DollSlots, "launcher")
		}
	}
	return slices.Index(DollSlots, it.Slot)
}

// powerActs are the inventory's member powers by kind.
var powerActs = map[game.PowerKind]front.ActKind{game.PowerCast: front.ActSpells, game.PowerDisplay: front.ActMagic, game.PowerHeal: front.ActHeal, game.PowerCure: front.ActCure}

// itemKinds are the party inventory's tabs by game.Item.Kind (the rest: misc).
var itemKinds = map[string]front.ItemKind{"weapon": front.KindWeapon, "armor": front.KindArmor}

// lines are an item's Look text as lines (nil: none).
func lines(t string) []string {
	if t == "" {
		return nil
	}
	return strings.Split(t, "\n")
}

// invItem is it as the inventory shows it: its icon's type and words, its
// tab.
func invItem(it game.Item) front.InvItem {
	return front.InvItem{Item: it.Handle, Name: it.Name, Type: byte(it.Type), Words: it.Words, Count: it.Count,
		Kind: itemKinds[it.Kind]}
}

// slotted: it has a doll slot (it can be readied).
func slotted(it game.Item) bool { return it.Kind != "" && it.Kind != "misc" }
