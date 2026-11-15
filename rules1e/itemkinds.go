package rules1e

import (
	"gobox/rules"
)

// ItemKinds is the dnd id of each of a game's base item types (rules.json
// itemKinds; nil: POR's, ItemIDs). The rules test an item's kind through it,
// never a POR item-type number: SotSB and PoD number their types otherwise
// (SotSB's 0x24 is armour).
type ItemKinds map[ItemType]rules.ItemID

// SotsbItemKinds are SotSB's base item types (sotsb-re: money.pas
// GenerateRandomItem, charhlp.pas's elf bonus 18, 19, 22..26, cbtacts.pas's
// missiles 2, 3, 5, 6, 7, 12, 28, 30, backstab 3, 4, 17..19); the weapons
// follow itemNames 1..28. 26 is the Fine Long Bow (words "Fine" "Long Bow", base stats as 24). PoD shares the numbering (pod-re charhlp.pas).
var SotsbItemKinds = ItemKinds{
	1: ItemIDBattleAxe, 2: ItemIDHandAxe, 3: ItemIDClub, 4: ItemIDDagger, 5: ItemIDDart,
	6: ItemIDHammer, 7: ItemIDJavelin, 8: ItemIDMace, 9: ItemIDMorningStar, 10: ItemIDMilitaryPick,
	11: ItemIDAwlPike, 12: ItemIDQuarrel, 13: ItemIDScimitar, 14: ItemIDSpear, 15: ItemIDQuarterStaff,
	16: ItemIDBastardSword, 17: ItemIDBroadSword, 18: ItemIDLongSword, 19: ItemIDShortSword,
	20: ItemIDTwoHandedSword, 21: ItemIDTrident, 22: ItemIDCompositeLongBow, 23: ItemIDCompositeShortBow,
	24: ItemIDLongBow, 25: ItemIDShortBow, 26: ItemIDLongBow, 27: ItemIDLightCrossbow,
	28: ItemIDSling, 30: ItemIDArrow, 31: ItemIDLeatherArmor, 32: ItemIDRingMail, 33: ItemIDScaleMail,
	34: ItemIDChainMail, 35: ItemIDBandedMail, 36: ItemIDPlateMail, 37: ItemIDShield,
	39: ItemIDMageScroll, 40: ItemIDClericScroll, 50: ItemIDBracers, 59: ItemIDProtectionRing,
	70: ItemIDFlail, 71: ItemIDHalberd,
	// named after the items of each type in SotSB's and PoD's ITEM blocks, checked against the base
	// table's slot (ITEMS: 41..46 slots 4..9, 72 slot 3) and pod-re (74 blessed bolt: effhndlr.pas effect 82;
	// 29 staff sling: cbtacts.pas with 28; 66 hornet's nest dart: cbtacts.pas missiles). 73 is SotSB's bundle
	// (Ord('I')), PoD's vorpal long sword.
	29: ItemIDStaffSling, 38: ItemIDProtectionScroll, 41: ItemIDHelm, 42: ItemIDBelt, 43: ItemIDRobe,
	44: ItemIDCloak, 45: ItemIDBoots, 46: ItemIDRing, 47: ItemIDMiscMagic, 48: ItemIDMirror,
	51: ItemIDWand, 52: ItemIDWand, 53: ItemIDStrengthPotion, 54: ItemIDYellowCanary, 58: ItemIDCloakDisplacement,
	61: ItemIDElvenChainMail, 62: ItemIDDrowChainMail, 63: ItemIDDrowLongSword, 65: ItemIDWizardryRing,
	66: ItemIDHornetsNestDart, 72: ItemIDGauntlets, 73: ItemIDVorpalLongSword, 74: ItemIDBlessedBolt,
	75: ItemIDCloakProtection,
}

// Kinds is the game's ItemKinds.
func (e *RuleState) Kinds() ItemKinds {
	if e.T().Kinds != nil {
		return e.T().Kinds
	}
	return ItemIDs
}

// kind is the dnd id of item type t (0: the game names none).
func (e *RuleState) kind(t ItemType) rules.ItemID { return e.Kinds()[t] }

// porNamed is POR's named items, as the rules look them up.
var porNamed = namedTable(PorNamedItems)

// iconKind is the id that picks it's icon: its own when it is one of the
// game's named items and identified (an unidentified one never shows what it
// is), else its base kind.
func (e *RuleState) iconKind(it *Item) rules.ItemID {
	named := e.T().Named
	if named == nil && e.T().Words == nil { // POR's own words: POR's items
		named = porNamed
	}
	if id, ok := named[it.Words]; ok && !it.Unknown() {
		return id
	}
	return e.kind(it.Type)
}

// isKind: item type t is one of ids.
func (e *RuleState) isKind(t ItemType, ids ...rules.ItemID) bool {
	k := e.kind(t)
	for _, id := range ids {
		if k != 0 && k == id {
			return true
		}
	}
	return false
}

// typeOf is the game's lowest item type of kind id (0: none).
func (e *RuleState) typeOf(id rules.ItemID) ItemType {
	for t := range ItemType(0xFF) {
		if t > 0 && e.kind(t) == id {
			return t
		}
	}
	return 0
}

// isBow: t is one of the four bows (the elves' bonus, Bow).
func (e *RuleState) isBow(t ItemType) bool {
	return e.isKind(t, ItemIDCompositeLongBow, ItemIDCompositeShortBow, ItemIDLongBow, ItemIDShortBow)
}
