package rules1e

import (
	"gobox/core"
	"gobox/core/dnd"
)

// ItemKinds is the dnd id of each of a game's base item types (rules.json
// itemKinds; nil: POR's, ItemIDs). The rules test an item's kind through it,
// never a POR item-type number: SotSB and PoD number their types otherwise
// (SotSB's 0x24 is armour).
type ItemKinds map[ItemType]core.ItemID

// SotsbItemKinds are SotSB's base item types (sotsb-re: money.pas
// GenerateRandomItem, charhlp.pas's elf bonus 18, 19, 22..26, cbtacts.pas's
// missiles 2, 3, 5, 6, 7, 12, 28, 30, backstab 3, 4, 17..19); the weapons
// follow itemNames 1..28. 26 is the Fine Composite Long Bow; 29 the staff
// sling has no dnd id. PoD shares the numbering (pod-re charhlp.pas).
var SotsbItemKinds = ItemKinds{
	1: dnd.ItemBattleAxe, 2: dnd.ItemHandAxe, 3: dnd.ItemClub, 4: dnd.ItemDagger, 5: dnd.ItemDart,
	6: dnd.ItemHammer, 7: dnd.ItemJavelin, 8: dnd.ItemMace, 9: dnd.ItemMorningStar, 10: dnd.ItemMilitaryPick,
	11: dnd.ItemAwlPike, 12: dnd.ItemQuarrel, 13: dnd.ItemScimitar, 14: dnd.ItemSpear, 15: dnd.ItemQuarterStaff,
	16: dnd.ItemBastardSword, 17: dnd.ItemBroadSword, 18: dnd.ItemLongSword, 19: dnd.ItemShortSword,
	20: dnd.ItemTwoHandedSword, 21: dnd.ItemTrident, 22: dnd.ItemCompositeLongBow, 23: dnd.ItemCompositeShortBow,
	24: dnd.ItemLongBow, 25: dnd.ItemShortBow, 26: dnd.ItemCompositeLongBow, 27: dnd.ItemLightCrossbow,
	28: dnd.ItemSling, 30: dnd.ItemArrow, 31: dnd.ItemLeatherArmor, 32: dnd.ItemRingMail, 33: dnd.ItemScaleMail,
	34: dnd.ItemChainMail, 35: dnd.ItemBandedMail, 36: dnd.ItemPlateMail, 37: dnd.ItemShield,
	39: dnd.ItemMageScroll, 40: dnd.ItemClericScroll, 50: dnd.ItemBracers, 59: dnd.ItemProtectionRing,
	70: dnd.ItemFlail, 71: dnd.ItemHalberd,
}

// kinds is the game's ItemKinds.
func (e *RuleState) kinds() ItemKinds {
	if e.T().Kinds != nil {
		return e.T().Kinds
	}
	return ItemIDs
}

// kind is the dnd id of item type t (0: the game names none).
func (e *RuleState) kind(t ItemType) core.ItemID { return e.kinds()[t] }

// isKind: item type t is one of ids.
func (e *RuleState) isKind(t ItemType, ids ...core.ItemID) bool {
	k := e.kind(t)
	for _, id := range ids {
		if k != 0 && k == id {
			return true
		}
	}
	return false
}

// typeOf is the game's lowest item type of kind id (0: none).
func (e *RuleState) typeOf(id core.ItemID) ItemType {
	for t := range ItemType(0xFF) {
		if t > 0 && e.kind(t) == id {
			return t
		}
	}
	return 0
}

// isBow: t is one of the four bows (the elves' bonus, Bow).
func (e *RuleState) isBow(t ItemType) bool {
	return e.isKind(t, dnd.ItemCompositeLongBow, dnd.ItemCompositeShortBow, dnd.ItemLongBow, dnd.ItemShortBow)
}
